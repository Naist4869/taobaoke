package model

import (
	"fmt"
	"math"
	"strconv"
	"taobaoke/tools"
)

const (
	DBOrderVersion = 1
	DBOrderKey     = "orders"
)
const (
	IDField               = "_id"
	UserIDField           = "user_id"
	StatusField           = "status"
	DeletedField          = "deleted"
	TradeParentIDField    = "trade_parent_id"
	UpdateTimeField       = "update_time"
	AlipayTotalPriceField = "alipay_total_price"
	PaidTimeField         = "paid_time"
	CreateTimeField       = "create_time"
	SalaryField           = "salary"
	SalaryScaleField      = "salary_scale"
	CommissionField       = "commission"
	EarningTimeField      = "earning_time"
	WithDrawStatusField   = "withdraw_status"
	PayPriceField         = "pay_price"
)

type OrderStatus int

const (
	OrderBalance = 3
	OrderPaid    = 12
	OrderFailed  = 13
	OrderFinish  = 14
)

var OrderStatusMap = map[OrderStatus]string{
	OrderBalance: "订单结算",
	OrderPaid:    "订单付款",
	OrderFailed:  "订单失效",
	OrderFinish:  "订单成功",
}

func (s OrderStatus) String() string {
	if _, exist := OrderStatusMap[s]; exist {
		return OrderStatusMap[s]
	} else {
		return strconv.Itoa(int(s))
	}
}
func (s OrderStatus) Finish() bool {
	return s == OrderBalance || s == OrderFailed
}
func (s OrderStatus) Balance() bool {
	return s == OrderBalance
}

type Order struct {
	ID               string      `bson:"_id" json:"id"`                                // 对外显示的订单编号
	UserID           string      `bson:"user_id" json:"user_id"`                       // 用户ID
	ClickTime        tools.Time  `bson:"click_time" json:"click_time"`                 // 通过推广链接达到商品、店铺详情页的点击时间
	CreateTime       tools.Time  `bson:"create_time" json:"create_time"`               // 创建订单时间 (以远程订单为准)
	PaidTime         tools.Time  `bson:"paid_time" json:"paid_time"`                   // 付款时间
	EarningTime      tools.Time  `bson:"earning_time" json:"earning_time"`             // 订单确认收货后且商家完成佣金支付的时间
	UpdateTime       tools.Time  `bson:"update_time" json:"update_time"`               // 状态更新时间
	Title            string      `bson:"title" json:"title"`                           // 商品标题
	PicURL           string      `bson:"pic_url" json:"pic_url"`                       // 主图地址
	Count            int         `bson:"count" json:"count"`                           // 商品数量
	Price            int64       `bson:"price" json:"price"`                           // 商品单价  X100
	OriginalPrice    int64       `bson:"original_price" json:"original_price"`         // 商品原价 X100
	Coupon           int64       `bson:"coupon" json:"coupon"`                         // 优惠券金额 X100
	Commission       int64       `bson:"commission" json:"commission"`                 // 实际得到的佣金 X100
	Rebate           int64       `bson:"rebate" json:"rebate"`                         // 预估得到的佣金 X100
	Salary           int64       `bson:"salary" json:"salary"`                         // 真正返给用户的金额
	SalaryScale      int64       `bson:"salary_scale" json:"salary_scale"`             // 返还比例  %90表示为90
	WithDrawStatus   bool        `bson:"withdraw_status" json:"withdraw_status"`       // 是否已经提现
	ItemID           int64       `bson:"item_id" json:"item_id"`                       // 590141576510	商品id
	Status           OrderStatus `bson:"status" json:"status"`                         //已付款：指订单已付款，但还未确认收货 已收货：指订单已确认收货，但商家佣金未支付 已结算：指订单已确认收货，且商家佣金已支付成功 已失效：指订单关闭/订单佣金小于0.01元，订单关闭主要有：1）买家超时未付款； 2）买家付款前，买家/卖家取消了订单；3）订单付款后发起售中退款成功；3：订单结算，12：订单付款， 13：订单失效，14：订单成功
	AdzoneID         int64       `bson:"adzone_id" json:"adzone_id"`                   // 11	推广位管理下的推广位名称对应的ID，同时也是pid=mm_1_2_3中的“3”这段数字
	AlipayTotalPrice int64       `bson:"alipay_total_price" json:"alipay_total_price"` // X100 22.50存储2250	买家拍下付款的金额（不包含运费金额）
	PayPrice         int64       `bson:"pay_price" json:"pay_price"`                   // X100  9.11存储911	买家确认收货的付款金额（不包含运费金额）
	TradeID          string      `bson:"trade_id" json:"trade_id"`                     // 294159887445064307	买家通过购物车购买的每个商品对应的订单编号，此订单编号并未在淘宝买家后台透出
	TradeParentID    string      `bson:"trade_parent_id" json:"trade_parent_id"`       // 294159887445064307	买家在淘宝后台显示的订单编号
	TrendInfo        TrendInfo   `bson:"-" json:"trend_info"`                          // 价格趋势信息
	ShopName         string      `bson:"shop_name" json:"shop_name"`                   // 店铺名称
	ShopType         int         `bson:"shop_type" json:"shop_type"`                   // 店铺类型，0表示集市，1表示商城
	Deleted          bool        `bson:"deleted" json:"-"`                             // 是否被删除  false没删除 true已删除
	Meta             DbMeta      `bson:"meta" json:"meta"`                             // 版本
}

func (o *Order) MakeMatched(clickTime tools.Time, createTime tools.Time, status int, tradeID string, tradeParentID string, count int, pubSharePreFee string) error {
	rebate, err := strconv.ParseFloat(pubSharePreFee, 64)
	if err != nil {
		return err
	}
	o.Rebate = int64(rebate * 100)
	o.ClickTime = clickTime
	o.CreateTime = createTime
	o.Status = OrderStatus(status)
	o.TradeID = tradeID
	o.TradeParentID = tradeParentID
	o.Count = count
	return nil
}

func (o *Order) MakeCommission(earningTime tools.Time, totalCommissionFee string, PayPrice string, salaryScale int64, status int) error {
	commission, err := strconv.ParseFloat(totalCommissionFee, 64)
	if err != nil {
		return err
	}
	payPrice, err := strconv.ParseFloat(PayPrice, 64)
	if err != nil {
		return err
	}
	o.Commission = int64(commission * 100)
	o.SalaryScale = salaryScale
	o.Salary = int64(commission*100) * salaryScale / 100
	o.EarningTime = earningTime
	o.Status = OrderStatus(status)
	o.PayPrice = int64(payPrice * 100)
	return nil
}
func (o *Order) MakePaid(paidTime tools.Time, status int, AlipayTotalPrice string, IncomeRate string) error {
	alipayTotalPrice, err := strconv.ParseFloat(AlipayTotalPrice, 64)
	if err != nil {
		return err
	}

	incomeRate, err := strconv.ParseFloat(IncomeRate, 64)
	if err != nil {
		return err
	}
	o.AlipayTotalPrice = int64(alipayTotalPrice * 100)

	o.PaidTime = paidTime
	o.Status = OrderStatus(status)
	calculateCommission := alipayTotalPrice * float64(o.Count) * incomeRate / 10000
	// 保留两位小数四舍五入
	roundCommission := math.Round(calculateCommission*100) / 100
	if int64(roundCommission)*100 != o.Rebate {
		return NewCalculateCommissionInconsistentError(float64(o.Rebate/100), roundCommission, calculateCommission)
	}
	return nil
}

type CalculateCommissionInconsistent struct {
	pubSharePreFee      float64
	roundCommission     float64
	calculateCommission float64
}

func NewCalculateCommissionInconsistentError(pubSharePreFee float64, roundCommission float64, calculateCommission float64) *CalculateCommissionInconsistent {
	return &CalculateCommissionInconsistent{pubSharePreFee: pubSharePreFee, roundCommission: roundCommission, calculateCommission: calculateCommission}
}
func (e CalculateCommissionInconsistent) Error() string {
	return fmt.Sprintf("计算得出的佣金和预估佣金不同, 预估佣金: %f, 计算佣金: %f, 为保留2位小数的计算佣金: %f", e.pubSharePreFee, e.roundCommission, e.calculateCommission)
}
func (e CalculateCommissionInconsistent) Is(target error) bool {
	switch target.(type) {
	case *CalculateCommissionInconsistent, CalculateCommissionInconsistent:
		return true
	default:
		return false
	}
}

type TrendInfo struct {
	Period        int        `json:"period"`         // 天数 默认为180
	MaxPrice      string     `json:"max_price"`      // 近Period天最高价格 X100
	MinPrice      string     `json:"min_price"`      // 近Period天最低价格 X100
	RawJsonTrend  string     `json:"raw_json_trend"` // 趋势json字符串
	OriginalPrice string     `json:"original_price"` // 原价  X100
	CurrentPrice  string     `json:"current_price"`  // 现价  X100
	TrendMsg      string     `json:"trend_msg"`      // 价格平稳  描述趋势字符串
	TKL           string     `json:"tkl"`            // 淘口令
	EffectiveDate tools.Time `json:"add_time"`       // 添加时间
}

// DbMeta 数据库元数据信息
type DbMeta struct {
	Version int `bson:"version"` // 版本
}

func NewOrder(id string, userID string, adzoneID int64, title string, itemID int64, picURL string, shopName string, shopType int, price int64, reservePrice int64, coupon int64) *Order {
	return &Order{ID: id, UserID: userID, AdzoneID: adzoneID, Title: title, ItemID: itemID, PicURL: picURL, ShopName: shopName, ShopType: shopType, Price: price, Coupon: coupon, OriginalPrice: reservePrice}
}
