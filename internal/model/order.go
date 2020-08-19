package model

import (
	"time"
)

const (
	DBOrderVersion = 1
	DBOrderKey     = "orders"
)
const (
	IDField      = "_id"
	UserIDField  = "user_id"
	StatusField  = "status"
	DeletedField = "deleted"
)

type Order struct {
	ID               string    `bson:"_id" json:"id"`                                // id,也就是编号
	UserID           string    `bson:"user_id" json:"user_id"`                       // 用户ID
	ClickTime        time.Time `bson:"click_time" json:"click_time"`                 // 通过推广链接达到商品、店铺详情页的点击时间
	CreateTime       time.Time `bson:"create_time" json:"create_time"`               // 创建订单时间 (以远程订单为准)
	PaidTime         time.Time `bson:"paid_time" json:"paid_time"`                   // 付款时间
	EarningTime      time.Time `bson:"earning_time" json:"earning_time"`             // 订单确认收货后且商家完成佣金支付的时间
	UpdateTime       time.Time `bson:"update_time" json:"update_time"`               // 状态更新时间
	Title            string    `bson:"title" json:"title"`                           // 商品标题
	PicURL           string    `bson:"pic_url" json:"pic_url"`                       // 主图地址
	Count            int       `bson:"count" json:"count"`                           // 商品数量
	Price            int64     `bson:"price" json:"price"`                           // 商品单价  X100
	ReservePrice     int64     `bson:"reserve_price" json:"reserve_price"`           // 商品原价 X100
	Commission       int64     `bson:"commission" json:"commission"`                 // 大约的佣金 X100
	Rebate           int64     `bson:"rebate" json:"rebate"`                         // 返利金额
	ItemID           int64     `bson:"item_id" json:"item_id"`                       // 590141576510	商品id
	Status           int       `bson:"status" json:"status"`                         //已付款：指订单已付款，但还未确认收货 已收货：指订单已确认收货，但商家佣金未支付 已结算：指订单已确认收货，且商家佣金已支付成功 已失效：指订单关闭/订单佣金小于0.01元，订单关闭主要有：1）买家超时未付款； 2）买家付款前，买家/卖家取消了订单；3）订单付款后发起售中退款成功；3：订单结算，12：订单付款， 13：订单失效，14：订单成功
	AdzoneID         int64     `bson:"adzone_id" json:"adzone_id"`                   // 11	推广位管理下的推广位名称对应的ID，同时也是pid=mm_1_2_3中的“3”这段数字
	AlipayTotalPrice string    `bson:"alipay_total_price" json:"alipay_total_price"` // 22.50	买家拍下付款的金额（不包含运费金额）
	PayPrice         string    `bson:"pay_price" json:"pay_price"`                   // 9.11	买家确认收货的付款金额（不包含运费金额）
	TradeID          string    `bson:"trade_id" json:"trade_id"`                     // 294159887445064307	买家通过购物车购买的每个商品对应的订单编号，此订单编号并未在淘宝买家后台透出
	TradeParentID    string    `bson:"trade_parent_id" json:"trade_parent_id"`       // 294159887445064307	买家在淘宝后台显示的订单编号
	CouponShareURL   string    `bson:"coupon_share_url" json:"coupon_share_url"`     // uland.xxx	链接-宝贝+券二合一页面链接
	TrendInfo        TrendInfo `bson:"trend_info" json:"trend_info"`                 // 价格趋势信息
	ShopName         string    `bson:"shop_name" json:"shop_name"`                   // 店铺名称
	ShopType         int       `bson:"shop_type" json:"shop_type"`                   // 店铺类型，0表示集市，1表示商城
	URL              string    `bson:"url" json:"url"`                               // s.click.xxx	链接-宝贝推广链接
	Key              string    `bson:"key" json:"key"`                               // 淘口令
	Deleted          bool      `bson:"deleted" json:"-"`                             // 是否被删除  false没删除 true已删除
	Meta             DbMeta    `bson:"meta" json:"meta"`                             // 版本
}

type TrendInfo struct {
	Period        int    // 天数 默认为180
	MaxPrice      string // 近Period天最高价格 X100
	MinPrice      string // 近Period天最低价格 X100
	RawJsonTrend  string // 趋势json字符串
	OriginalPrice string // 原价  X100
	CurrentPrice  string // 现价  X100
	TrendMsg      string // 价格平稳  描述趋势字符串
}

// DbMeta 数据库元数据信息
type DbMeta struct {
	Version int `bson:"version"` // 版本
}

func NewOrder(id string, userID string, adzoneID int64, title string, itemID int64, picURL string, shopName string, shopType int, price int64, reservePrice int64, rebate int64, URL string, couponShareURL string, key string) *Order {
	return &Order{ID: id, UserID: userID, AdzoneID: adzoneID, Title: title, ItemID: itemID, PicURL: picURL, ShopName: shopName, ShopType: shopType, Price: price, ReservePrice: reservePrice, Rebate: rebate, URL: URL, CouponShareURL: couponShareURL, Key: key}
}
