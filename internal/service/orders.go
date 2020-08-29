package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	pb "taobaoke/api"
	"taobaoke/internal/dao"
	"taobaoke/internal/model"
	"time"

	"github.com/go-kratos/kratos/pkg/net/rpc/warden"
	xtime "github.com/go-kratos/kratos/pkg/time"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"go.uber.org/zap"

	"github.com/Naist4869/log"
)

type orders struct {
	dao     dao.OrderMatchService
	client  pb.WechatClient
	doStart func()
	logger  *log.Logger // 日志器
}

func NewOrders(dao dao.OrderMatchService, logger *log.Logger) *orders {
	o := &orders{
		logger: logger.With(zap.String("模块名", "订单管理")),
		dao:    dao,
	}
	var once sync.Once
	o.doStart = func() {
		once.Do(func() {
			if o.client == nil {
				var err error
				o.client, err = pb.NewClient(&warden.ClientConfig{Timeout: xtime.Duration(time.Second)})
				if err != nil {
					// todo 也许应该重试
					panic(err)
				}
			}
		})
	}
	return o

}

func (o *orders) String() string {
	builder := &strings.Builder{}
	unmatchedCount, unmatched := o.unmatchedString()
	matchedCount, matched := o.matchedString()
	_, err := fmt.Fprintf(builder, "未匹配数量[%d],已匹配数量[%d],未匹配[%s],已匹配[%s]", unmatchedCount, matchedCount, unmatched, matched)
	if err == nil {
		return builder.String()
	}
	return err.Error()
}

func (o *orders) unmatchedString() (int, string) {
	unmatch, err := o.dao.GetAllUnmatch(context.Background())
	if err != nil {
		o.logger.Error("unmatchedString", zap.Error(err))
		return 0, err.Error()
	}
	strs := make([]string, 0, len(unmatch))
	for _, order := range unmatch {
		strs = append(strs, fmt.Sprintf("%s-%d-%d-%s-%s-%s-%s", order.ID, order.ItemID, order.AdzoneID, order.Title, order.UpdateTime.String(), order.TrendInfo.TKL, order.Status.String()))
	}
	return len(unmatch), strings.Join(strs, ",")
}

func (o *orders) matchedString() (int, string) {
	matched, err := o.dao.MatchGetAll(context.Background())
	if err != nil {
		o.logger.Error("matchedString", zap.Error(err))
		return 0, err.Error()
	}
	strs := make([]string, 0, len(matched))
	for _, order := range matched {
		strs = append(strs, fmt.Sprintf("%s-%s-%s-%d-%s-%s", order.ID, order.Title, order.TradeParentID, order.AlipayTotalPrice, order.PaidTime.String(), order.Status.String()))
	}
	return len(matched), strings.Join(strs, ",")
}

func (o *orders) Match(orders []TbkOrderDetailsGetResult) {
	o.logger.Info("开始查单", zap.Any("remoteOrder", orders))
	o.MatchingUnmatched(orders)
	return
}

func (o *orders) Add(ctx context.Context, order *model.Order, nonce string) (err error) {
	o.logger.Info("添加本地订单记录", zap.Int64("渠道ID", order.AdzoneID), zap.String("用户ID", order.UserID), zap.Int64("商品ID", order.ItemID))
	ok, err := o.dao.SetToUnmatch(ctx, order.ItemID, order.AdzoneID, order, nonce)
	if err != nil || !ok {
		err = fmt.Errorf("orders Add fail: %w", err)
		return err
	}
	return
}

func (o *orders) Exist(itemID, adZoneID int64) bool {
	o.logger.Info("查询本地订单记录是否存在", zap.Int64("itemID", itemID), zap.Int64("adZoneID", adZoneID))
	exist, err := o.dao.ExistInUnmatch(context.Background(), itemID, adZoneID)
	if err != nil {
		o.logger.Error("查询本地订单记录是否存在", zap.Error(err), zap.Int64("itemID", itemID), zap.Int64("adZoneID", adZoneID))
		return false
	}
	return exist
}

func (o *orders) MatchingUnmatched(unmatchedOrders []TbkOrderDetailsGetResult) {
	for _, remoteOrder := range unmatchedOrders {
		if !o.Exist(remoteOrder.ItemID, remoteOrder.AdzoneID) {
			o.logger.Info("订单本地不存在，跳过")
			continue
		}
		if remoteOrder.AlipayTotalPrice == "" {
			o.logger.Info("订单还没有付款金额，跳过")
			continue
		}
		localOrder, err := o.dao.GetUnmatch(context.Background(), remoteOrder.ItemID, remoteOrder.AdzoneID)
		if err != nil {
			o.logger.Error("MatchingUnmatched", zap.Error(err))
			continue
		}
		o.logger.Info("订单已经匹配", zap.Any("远程订单", remoteOrder), zap.Any("本地订单", localOrder))
		if err = o.makeMatched(localOrder, remoteOrder); err != nil {
			o.logger.Error("MatchingUnmatched", zap.Error(err))
			continue
		}
	}
}

func (o *orders) makeMatched(localOrder *model.Order, remoteOrder TbkOrderDetailsGetResult) (err error) {
	if err = localOrder.MakeMatched(remoteOrder.ClickTime, remoteOrder.TkCreateTime, remoteOrder.TkStatus, remoteOrder.TradeID, remoteOrder.TradeParentID, remoteOrder.ItemNum, remoteOrder.PubSharePreFee); err != nil {
		o.logger.Error("makeMatched", zap.Error(err), zap.Any("RemoteOrder", remoteOrder), zap.Any("LocalOrder", localOrder))
		return err
	}
	err = o.dao.Insert(context.Background(), localOrder)
	if err != nil {
		return err
	}

	if _, err = o.dao.DelFromUnmatchAndSetToMatch(context.Background(), localOrder); err != nil {
		o.logger.Error("makeMatched", zap.Error(err))
	}

	_, _ = o.TemplateMsgSend(context.Background(), &pb.TemplateMsgSendReq{
		UserID:           localOrder.UserID,
		OrderID:          localOrder.TradeParentID,
		Title:            localOrder.Title,
		PaidTime:         localOrder.PaidTime.String(),
		AlipayTotalPrice: strconv.FormatFloat(float64(localOrder.AlipayTotalPrice)/100, 'f', -1, 64),
		Rebate:           strconv.FormatFloat(float64(localOrder.Rebate)/100, 'f', -1, 64),
	})
	o.logger.Info("查单队列", zap.String("状态", o.String()))
	return
}

func (o *orders) TemplateMsgSend(ctx context.Context, in *pb.TemplateMsgSendReq, opts ...grpc.CallOption) (*empty.Empty, error) {
	if o.doStart != nil {
		o.doStart()
	}
	return o.client.TemplateMsgSend(ctx, in, opts...)
}
