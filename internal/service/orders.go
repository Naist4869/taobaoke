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
	"taobaoke/tools"
	"time"

	"github.com/go-kratos/kratos/pkg/net/rpc/warden"
	xtime "github.com/go-kratos/kratos/pkg/time"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"go.mongodb.org/mongo-driver/bson"

	"go.uber.org/zap"

	"github.com/Naist4869/log"
)

const (
	max                 = 30   // 单个号码最大下单数
	workerOrderCapacity = 1000 // 每个工人历史订单的起始容量
)

var (
	errFill = fmt.Errorf("已经达到下单上限%d", max)
)

type orders struct {
	max     int // 最大下单成功订单数量
	current int // 当前下单成功数量
	dao     dao.OrderMatchService
	client  pb.WechatClient
	doStart func()
	logger  *log.Logger // 日志器
}

func NewOrders(dao dao.OrderMatchService, logger *log.Logger) *orders {
	o := &orders{
		max:    max,
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
	_, err := fmt.Fprintf(builder, "上限[%d],当前数量[%d],未匹配[%s],已匹配[%s]", o.max, o.current, o.unmatchedString(), o.matchedString())
	if err == nil {
		return builder.String()
	}
	return err.Error()
}

func (o *orders) unmatchedString() string {
	all, err := o.dao.UnmatchGetAll(context.Background())
	if err != nil {
		o.logger.Error("unmatchedString", zap.Error(err))
		return err.Error()
	}
	return strings.Join(all, ",")
}

func (o *orders) matchedString() string {
	matched, err := o.dao.QueryOrderByStatus(context.Background(), tools.Time{}, tools.Now(), model.OrderPaid, model.OrderFinish)
	if err != nil {
		o.logger.Error("matchedString", zap.Error(err))
		return err.Error()
	}
	strs := make([]string, 0, len(matched))
	for _, order := range matched {
		strs = append(strs, fmt.Sprintf("%s-%s-%s", order.ID, order.CreateTime, order.Status.String()))
	}
	return strings.Join(strs, ",")
}

func (o *orders) Match(orders []TbkOrderDetailsGetResult) {
	o.logger.Warn("开始查单", zap.Any("remoteOrder", orders))
	unmatched := make([]TbkOrderDetailsGetResult, 0, len(orders))
	localOrder := map[string]*model.Order{}
	tradeParentIDs := make([]string, 0, len(orders))
	for _, order := range orders {
		tradeParentIDs = append(tradeParentIDs, order.TradeParentID)
	}
	results, err := o.dao.QueryOrderByTradeParentID(context.Background(), tradeParentIDs, true)
	if err != nil {
		o.logger.Error("Match", zap.Error(err), zap.Any("remoteOrder", orders))
		return
	}
	for _, order := range results {
		localOrder[order.TradeParentID] = order
	}

	changedTradeParentIDs := make([]string, 0, len(orders)/2)
	for i, remoteOrder := range orders {
		if matchedOrder, matched := localOrder[remoteOrder.TradeParentID]; matched { // 已匹配的单，更新更新时间和状态
			if int(matchedOrder.Status) != (remoteOrder.TkStatus) {
				if err := o.dao.UpdateSingleOrderGeneric(context.Background(), matchedOrder.ID, nil, bson.M{
					dao.SET: bson.M{
						model.UpdateTimeField:       tools.Now(),
						model.StatusField:           model.OrderStatus(remoteOrder.TkStatus),
						model.AlipayTotalPriceField: remoteOrder.AlipayTotalPrice,
						model.PaidTimeField:         remoteOrder.TkPaidTime,
					},
				}); err != nil {
					o.logger.Error("Match", zap.Error(err), zap.Any("localOrder", matchedOrder), zap.Any("remoteOrder", remoteOrder))
					continue
				}
				changedTradeParentIDs = append(changedTradeParentIDs, matchedOrder.TradeParentID)
			}
			o.logger.Info("进入到已匹配的单，更新更新时间和状态", zap.Any("localOrder", matchedOrder), zap.Any("remoteOrder", remoteOrder))
		} else { // 未匹配的单，添加到未匹配队列
			unmatched = append(unmatched, orders[i])
		}
	}
	_, err = o.dao.DelFromOrderCache(context.Background(), changedTradeParentIDs)
	if err != nil {
		o.logger.Error("Match", zap.Error(err), zap.Strings("changedTradeParentIDs", changedTradeParentIDs))
	}
	o.MatchingUnmatched(unmatched)
	return
}

func (o *orders) Add(ctx context.Context, order *model.Order) (err error) {
	o.logger.Info("添加本地订单记录", zap.Int64("渠道ID", order.AdzoneID), zap.String("用户ID", order.UserID), zap.Int64("商品ID", order.ItemID))
	if o.current == o.max {
		return errFill
	}
	ok, err := o.dao.SetToUnmatch(ctx, order.ItemID, order.AdzoneID, order)
	if err != nil || !ok {
		err = fmt.Errorf("orders Add fail: %w", err)
		return err
	}
	o.current++
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
		localOrder, err := o.dao.UnmatchGet(context.Background(), remoteOrder.ItemID, remoteOrder.AdzoneID)
		if err != nil {
			o.logger.Error("MatchingUnmatched", zap.Error(err))
			continue
		}
		o.logger.Info("MatchingUnmatched 逐个遍历", zap.Any("远程订单", remoteOrder), zap.Any("本地订单", localOrder))
		if !isRecent(remoteOrder.ClickTime, localOrder.UpdateTime) {
			o.logger.Info("订单本地不匹配,跳过")
			continue
		}
		o.logger.Info("订单已经匹配", zap.Any("查单结果", remoteOrder), zap.Any("本地", localOrder))
		if err = o.makeMatched(localOrder, remoteOrder); err != nil {
			o.logger.Error("MatchingUnmatched", zap.Error(err))
			continue
		}
	}
}

func isRecent(clickTime, localTime tools.Time) bool {
	distance := clickTime.Sub(localTime).Seconds()
	return distance <= 864000
}

func (o *orders) makeMatched(localOrder *model.Order, remoteOrder TbkOrderDetailsGetResult) (err error) {
	if err = localOrder.MakeMatched(remoteOrder.ClickTime, remoteOrder.TkCreateTime, remoteOrder.TkStatus, remoteOrder.TradeID, remoteOrder.TradeParentID, remoteOrder.ItemNum, remoteOrder.PubSharePreFee); err != nil {
		o.logger.Error("makeMatched", zap.Error(err), zap.Any("RemoteOrder", remoteOrder), zap.Any("LocalOrder", localOrder))
	}
	err = o.dao.Insert(context.Background(), localOrder)
	if err != nil {
		return err
	}
	// 删除匹配队列里的键
	if err = tools.Retry(func() (err error, mayRetry bool) {
		_, err = o.dao.DelFromUnmatchMap(context.Background(), localOrder.ItemID, localOrder.AdzoneID)
		return
	}); err != nil {
		return fmt.Errorf("makeMatched Retry del key distance > 180s, errors: (%w)", err)
	}
	o.TemplateMsgSend(context.Background(), &pb.TemplateMsgSendReq{
		UserID:           localOrder.UserID,
		OrderID:          localOrder.TradeParentID,
		Title:            localOrder.Title,
		PaidTime:         localOrder.PaidTime.String(),
		AlipayTotalPrice: strconv.FormatInt(localOrder.AlipayTotalPrice, 10),
		Rebate:           strconv.FormatInt(localOrder.Rebate, 10),
	})
	return
}

func (o *orders) TemplateMsgSend(ctx context.Context, in *pb.TemplateMsgSendReq, opts ...grpc.CallOption) (*empty.Empty, error) {
	if o.doStart != nil {
		o.doStart()
	}
	return o.client.TemplateMsgSend(ctx, in, opts...)
}
