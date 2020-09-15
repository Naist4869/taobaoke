package service

import (
	"strconv"
	"taobaoke/tools"
	"time"

	"github.com/Naist4869/base/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// 下单次数
	placeCounters = "tbk_place_counters"
	// 下单耗时
	placeHistogram = "tbk_place_since"
	// 跟单超时失败次数
	followFailCounters = "tbk_follow_fail_counters"
	// 跟单次数
	followHistogram = "tbk_follow"
	// 提现金额
	withdrawCounters = "tbk_withdraw_counters"
	// 平台净利润
	profitCounters = "tbk_profit_counters"
	// 未结算订单数量
	unbalanceOrderGauge = "tbk_unbalance_order_num"
)

type tbkMetrics struct {
	// 下单次数
	placeCounters *prometheus.CounterVec
	// 下单成功耗时
	placeSuccessHistogram *prometheus.HistogramVec
	// 跟单超时失败次数
	followFailCounters *prometheus.CounterVec
	// 跟单次数
	followHistogram *prometheus.HistogramVec
	// 提现金额
	withdrawCounters prometheus.Counter
	// 平台净利润  暂时不用
	profitCounters prometheus.Counter
	// 未结算订单数量
	unbalanceOrderNum *prometheus.GaugeVec
}

func (t *tbkMetrics) init(metrics metrics.Metrics) error {
	t.placeCounters = metrics.CounterVec[placeCounters]
	t.placeSuccessHistogram = metrics.HistogramVec[placeHistogram]
	t.followFailCounters = metrics.CounterVec[followFailCounters]
	t.withdrawCounters = metrics.Counters[withdrawCounters]
	t.profitCounters = metrics.Counters[profitCounters]
	t.unbalanceOrderNum = metrics.GaugeVec[unbalanceOrderGauge]
	t.followHistogram = metrics.HistogramVec[followHistogram]
	return tools.NotNil(t, nil, nil)
}

func (t *tbkMetrics) addPlaceCount(terminalID string, success bool, takes time.Duration) {
	t.placeCounters.With(map[string]string{
		"terminalID": terminalID,
		"success":    strconv.FormatBool(success),
	}).Inc()
	t.placeSuccessHistogram.With(map[string]string{
		"terminalID": terminalID,
		"success":    strconv.FormatBool(success),
	}).Observe(float64(takes.Milliseconds()))
}
func (t *tbkMetrics) addFollowSince(terminalID string, success bool, takes time.Duration) {
	t.followHistogram.With(map[string]string{
		"terminalID": terminalID,
		"success":    strconv.FormatBool(success),
	}).Observe(float64(takes.Milliseconds()))
}

func (t *tbkMetrics) addFollowFailCounters(terminalID string) {
	t.followFailCounters.With(map[string]string{
		"terminalID": terminalID,
	}).Inc()
}

func (t *tbkMetrics) setUnbalanceOrderNum(terminalID string, num float64) {
	t.unbalanceOrderNum.With(map[string]string{
		"terminalID": terminalID,
	}).Set(num)
}

func NewMetrics() (*tbkMetrics, error) {

	m, err := metrics.MakeMetrics(&metrics.MetricArguments{
		Counters: []prometheus.CounterOpts{
			{
				Name: withdrawCounters,
				Help: "提现金额",
			},
			{
				Name: profitCounters,
				Help: "平台净利润",
			},
		},
		CounterVec: []prometheus.CounterOpts{
			{
				Name: placeCounters,
				Help: "下单次数",
			},
			{
				Name: followFailCounters,
				Help: "跟单超时失败次数",
			},
		},
		CounterVecLabels: [][]string{
			// 每次启动都有自己的ID,从k8s环境变量HOSTNAME中获取 https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#envvarsource-v1-core
			{"terminalID", "success"},
			{"terminalID"},
		},
		HistogramVec: []prometheus.HistogramOpts{
			{
				Name: placeHistogram,
				Help: "下单耗时",
				// 微信回复消息单次最长时间为5s
				Buckets: []float64{1000, 2000, 3000, 4000, 5000},
			},
			{
				Name: followHistogram,
				Help: "跟单耗时",
				// 3分钟内  30分钟内 1小时内 一天内 3天内 7天内  最长在unmatched队列中是10天暂时先不统计了
				Buckets: []float64{180000, 1800000, 3600000, 86400000, 25920000, 604800000},
			},
		},
		HistogramVecLabels: [][]string{
			{"terminalID", "success"},
			{"terminalID", "success"},
		},
		GaugeVec: []prometheus.GaugeOpts{
			{
				Name: unbalanceOrderGauge,
				Help: "未结算订单数量",
			},
		},
		GaugeVecLabels: [][]string{
			{"terminalID"},
		},
	})
	if err != nil {
		return nil, err
	}
	tbkMetrics := &tbkMetrics{}
	if err := tbkMetrics.init(m); err != nil {
		return nil, err
	}
	return tbkMetrics, nil
}
