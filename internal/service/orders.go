package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"taobaoke/internal/model"
	"time"

	"go.uber.org/zap"

	"github.com/Naist4869/log"
)

const (
	max                 = 30   // 单个号码最大下单数
	workerOrderCapacity = 1000 // 每个工人历史订单的起始容量
)

var (
	errFill      = fmt.Errorf("已经达到下单上限%d", max)
	errDuplicate = errors.New("有相同用户且点击时间相同的订单存在") // 无法匹配
)

type orders struct {
	max       int                     // 最大下单成功订单数量
	current   int                     // 当前下单成功数量
	unmatched map[string]*model.Order // 未匹配队列，用户ID+点击时间戳->订单
	matched   map[string]*model.Order // 已经匹配的，具备联通订单号的订单,联通订单号->订单
	lock      *sync.RWMutex           // 并发锁
	logger    *log.Logger             // 日志器
}

func NewOrders(logger *log.Logger) *orders {
	return &orders{
		max:       max,
		unmatched: make(map[string]*model.Order, 1000),
		matched:   make(map[string]*model.Order, 100),
		lock:      &sync.RWMutex{},
		logger:    logger,
	}

}
func (o *orders) String() string {
	o.lock.RLock()
	defer o.lock.RUnlock()
	builder := &strings.Builder{}
	_, err := fmt.Fprintf(builder, "上限[%d],当前数量[%d],未匹配[%s],已匹配[%s]", o.max, o.current, o.unmatchedString(), o.matchedString())
	if err == nil {
		return builder.String()
	}
	return err.Error()
}
func (o *orders) unmatchedString() string {
	keys := make([]string, 0, len(o.unmatched))
	for key := range o.unmatched {
		keys = append(keys, key)
	}
	return strings.Join(keys, ",")
}
func (o *orders) matchedString() string {
	keys := make([]string, 0, len(o.matched))
	//for k, order := range o.matched {
	//	keys = append(keys, fmt.Sprintf("%s-%s-%s-%d", key(order.ChargePhone, order.ChargeMoney), order.Status.String(), k, order.PlaceTime.Unix()))
	//}
	return strings.Join(keys, ",")
}

func (o *orders) Match(orders ...[]model.Order) {
	return
}
func (o *orders) Finished() bool {
	o.lock.RLock()
	defer o.lock.RUnlock()

	return len(o.matched)+len(o.unmatched) == 0
}

func (o *orders) Add(order *model.Order) error {
	o.logger.Info("添加本地订单记录", zap.Int64("渠道ID", order.AdzoneID), zap.String("用户ID", order.UserID), zap.Int64("商品ID", order.ItemID))
	o.lock.Lock()
	defer o.lock.Unlock()
	if o.current == o.max {
		return errFill
	}
	if o.Exist(key(order.ItemID)) {
		return errDuplicate
	}
	o.addUnmatched(order)
	o.current++
	return nil

}

func (o *orders) addUnmatched(order *model.Order) {
	key := key(order.ItemID)
	o.unmatched[key] = order
}

func (o *orders) Exist(key string) bool {
	o.logger.Info("查询本地订单记录是否存在", zap.String("key", key))
	_, exist := o.unmatched[key]
	return exist
}

func key(itemID int64) string {
	return fmt.Sprintf("%d", itemID)
}

func (o *orders) MatchingUnmatched(unmatchedOrders []TbkOrderDetailsGetResult) {
	for _, remoteOrder := range unmatchedOrders {
		o.logger.Info("MatchingUnmatched 逐个遍历", zap.Any("远程订单", remoteOrder), zap.Any("本地订单", o.unmatched))
		key := key(remoteOrder.ItemID)
		if !o.Exist(key) {
			o.logger.Info("订单本地不存在，跳过")
			continue
		}
		localOrder := o.unmatched[key]
		clickTime, err := time.Parse(TimeFormat, remoteOrder.ClickTime)
		if err != nil {
			o.logger.Error("MatchingUnmatched 解析点击时间", zap.Error(err), zap.String("远程订单点击时间", remoteOrder.ClickTime))
			continue
		}
		if !isClickTimeMatch(clickTime, localOrder.ClickTime) {
			o.logger.Info("订单本地不匹配,跳过")
			continue
		}

		o.logger.Info("订单已经匹配", zap.Any("查单结果", remoteOrder), zap.Any("本地", localOrder))
		o.makeMatched(localOrder, remoteOrder)

	}
}
func isClickTimeMatch(clickTime, localClickTime time.Time) bool {
	distance := clickTime.Second() - localClickTime.Second()
	return -5 <= distance && distance <= 5
}
func (o *orders) makeMatched(localOrder *model.Order, remoteOrder TbkOrderDetailsGetResult) {
	return
}
