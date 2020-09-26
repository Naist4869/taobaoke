package service

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	pb "taobaoke/api"
	"taobaoke/internal/dao"
	"taobaoke/internal/model"
	"taobaoke/tools"
	"time"

	"go.uber.org/zap"

	"github.com/Naist4869/log"
)

type orders struct {
	terminalID string
	dao        dao.OrderMatchService
	client     pb.WechatClient
	doStart    func()
	logger     *log.Logger // 日志器
	metrics    *tbkMetrics // 度量指标
}

func NewOrders(dao dao.OrderMatchService, logger *log.Logger, metrics *tbkMetrics) *orders {
	o := &orders{
		logger:     logger.With(zap.String("模块名", "订单管理")),
		dao:        dao,
		terminalID: os.Getenv("HOSTNAME"),
		metrics:    metrics,
	}
	go o.Monitor()
	return o

}
func (o *orders) Monitor() {
	for message := range o.dao.PSubscribeKeyspace() {
		o.logger.Info("监听到消息", zap.String("message", message.String()))
		o.metrics.addFollowFailCounters(o.terminalID)
	}
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
		strs = append(strs, fmt.Sprintf("%s-%d-%d-%s-%s-%s", order.ID, order.ItemID, order.AdzoneID, order.Title, order.TrendInfo.TKL, order.Status.String()))
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

func (o *orders) Add(ctx context.Context, order *model.Order, nonce string, deadline time.Time) (err error) {
	defer func() {
		o.metrics.addPlaceCount(o.terminalID, err == nil, time.Since(deadline))
	}()
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
		localOrder, err := o.dao.GetUnmatch(context.Background(), remoteOrder.ItemID, remoteOrder.AdzoneID)
		if err != nil {
			o.logger.Error("MatchingUnmatched", zap.Error(err))
			continue
		}
		if err = o.makeMatched(localOrder, remoteOrder); err != nil {
			o.logger.Error("MatchingUnmatched", zap.Error(err))
			continue
		}
		o.logger.Info("订单已经匹配", zap.Any("远程订单", remoteOrder), zap.Any("本地订单", localOrder))
	}
}

func (o *orders) makeMatched(localOrder *model.Order, remoteOrder TbkOrderDetailsGetResult) (err error) {

	defer func() {
		checkedTime := tools.Time{}
		for _, action := range localOrder.Timelines {
			if action.Action == model.OrderChecked.String() {
				checkedTime = action.Time
				continue
			}
			if action.Action == model.OrderCreate.String() && !checkedTime.IsZero() {
				o.metrics.addFollowSince(o.terminalID, err == nil, action.Time.Sub(checkedTime))
				return
			}
		}
		o.logger.Error("makeMatched", zap.String("原因", "本地订单没有下单时间,创建下单时间指标失败"), zap.Any("远程订单", remoteOrder), zap.Any("本地订单", localOrder))
	}()
	if err = localOrder.MakeMatched(remoteOrder.ClickTime, remoteOrder.TkCreateTime, remoteOrder.TradeID, remoteOrder.TradeParentID, remoteOrder.PubSharePreFee, remoteOrder.ItemPrice); err != nil {
		o.logger.Error("makeMatched", zap.Error(err), zap.Any("RemoteOrder", remoteOrder), zap.Any("LocalOrder", localOrder))
		return err
	}
	err = o.dao.Insert(context.Background(), localOrder)
	if err != nil {
		return err
	}

	if ok, err := o.dao.DelFromUnmatchAndSetToMatch(context.Background(), localOrder); err != nil || !ok {
		// 这里用局部变量err 出错了打个日志然后继续流程
		o.logger.Error("makeMatched", zap.Error(err), zap.Bool("是否成功", ok))
	}

	go tools.Retry(func() (err error, mayRetry bool) {
		_, err = o.dao.MatchedTemplateMsgSend(context.Background(), &pb.MatchedTemplateMsgSendReq{
			UserID:           localOrder.UserID,
			OrderID:          localOrder.TradeParentID,
			Title:            localOrder.Title,
			PaidTime:         localOrder.PaidTime.String(),
			AlipayTotalPrice: strconv.FormatFloat(float64(localOrder.AlipayTotalPrice)/100, 'f', -1, 64),
			Rebate:           strconv.FormatFloat(float64(localOrder.Rebate)/100, 'f', -1, 64),
		})
		if err != nil {
			o.logger.Error("makeMatched", zap.Error(err))
		}
		return err, true
	})

	o.logger.Info("查单队列", zap.String("状态", o.String()))
	return
}
