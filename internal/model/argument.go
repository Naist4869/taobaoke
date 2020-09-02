package model

import "taobaoke/tools"

type UpdateArgument struct {
	PaidTime    tools.Time  `bson:"paid_time" json:"paid_time"`       // 付款时间
	EarningTime tools.Time  `bson:"earning_time" json:"earning_time"` // 订单确认收货后且商家完成佣金支付的时间
	Status      OrderStatus `bson:"status" json:"status"`             //已付款：指订单已付款，但还未确认收货 已收货：指订单已确认收货，但商家佣金未支付 已结算：指订单已确认收货，且商家佣金已支付成功 已失效：指订单关闭/订单佣金小于0.01元，订单关闭主要有：1）买家超时未付款； 2）买家付款前，买家/卖家取消了订单；3）订单付款后发起售中退款成功；3：订单结算，12：订单付款， 13：订单失效，14：订单成功
	// AlipayTotalPrice 11.22	买家拍下付款的金额（不包含运费金额）
	AlipayTotalPrice string `json:"alipay_total_price"`
	// IncomeRate 9.99	订单结算的佣金比率+平台的补贴比率
	IncomeRate string `json:"income_rate"`
	// PubSharePreFee 0	付款预估收入=付款金额*提成。指买家付款金额为基数，预估您可能获得的收入。因买家退款等原因，可能与结算预估收入不一致
	PubSharePreFee string `json:"pub_share_pre_fee"`
	// ItemNum 2	商品数量
	ItemNum int `json:"item_num"`
	// TotalCommissionFee 0	佣金金额=结算金额*佣金比率
	TotalCommissionFee string `json:"total_commission_fee"`
	// PayPrice 9.11	买家确认收货的付款金额（不包含运费金额）
	PayPrice string `json:"pay_price"`
}
type Fill interface {
	FillContext() *UpdateArgument
}
