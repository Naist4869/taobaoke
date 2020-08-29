package model

import (
	"net/url"
	"strconv"
	"strings"
	"taobaoke/tools"
)

var XlsToJsonMap = map[string]string{
	"点击时间":      "click_time",
	"创建时间":      "tk_create_time",
	"付款时间":      "tk_paid_time",
	"结算时间":      "tk_earning_time",
	"商品ID":      "item_id",
	"商品图片":      "item_img",
	"商品标题":      "item_title",
	"店铺名称":      "seller_shop_title",
	"商品数量":      "item_num",
	"商品单价":      "item_price",
	"淘宝订单编号":    "trade_parent_id",
	"淘宝子订单号":    "trade_id",
	"订单状态":      "tk_status",
	"订单类型":      "order_type",
	"付款金额":      "alipay_total_price",
	"结算金额":      "pay_price",
	"佣金比率":      "total_commission_rate",
	"佣金金额":      "total_commission_fee",
	"补贴比率":      "subsidy_rate",
	"补贴金额":      "subsidy_fee",
	"收入比率":      "income_rate",
	"分成比率":      "pub_share_rate",
	"提成":        "tk_total_rate",
	"技术服务费率":    "alimama_rate",
	"技术服务费":     "alimama_share_fee",
	"付款预估收入":    "pub_share_pre_fee",
	"结算预估收入":    "pub_share_fee",
	"媒体ID":      "site_id",
	"媒体名称":      "site_name",
	"推广位ID":     "adzone_id",
	"推广位名称":     "adzone_name",
	"内容专项服务费率":  "tk_commission_rate_for_media_platform",
	"预估内容专项服务费": "tk_commission_pre_fee_for_media_platform",
	"结算内容专项服务费": "tk_commission_fee_for_media_platform",
}

type XLSOrder struct {
	TradeParentID                      string     `json:"trade_parent_id"`
	SubsidyFee                         float64    `json:"subsidy_fee"`
	TkCommissionFeeForMediaPlatform    float64    `json:"tk_commission_fee_for_media_platform"`
	ClickTime                          tools.Time `json:"click_time"`
	ItemImg                            string     `json:"item_img"`
	AlipayTotalPrice                   string     `json:"alipay_total_price"`
	TotalCommissionFee                 string     `json:"total_commission_fee"`
	PubShareFee                        float64    `json:"pub_share_fee"`
	TkCreateTime                       tools.Time `json:"tk_create_time"`
	TotalCommissionRate                string     `json:"total_commission_rate"`
	AlimamaShareFee                    float64    `json:"alimama_share_fee"`
	ItemTitle                          string     `json:"item_title"`
	SubsidyRate                        string     `json:"subsidy_rate"`
	IncomeRate                         string     `json:"income_rate"`
	PubSharePreFee                     string     `json:"pub_share_pre_fee"`
	ItemID                             string     `json:"item_id"`
	OrderType                          string     `json:"order_type"`
	SiteID                             string     `json:"site_id"`
	TkPaidTime                         tools.Time `json:"tk_paid_time"`
	SellerShopTitle                    string     `json:"seller_shop_title"`
	TkStatus                           string     `json:"tk_status"`
	TkTotalRate                        string     `json:"tk_total_rate"`
	TkEarningTime                      tools.Time `json:"tk_earning_time"`
	ItemNum                            int        `json:"item_num"`
	SiteName                           string     `json:"site_name"`
	TkCommissionRateForMediaPlatform   string     `json:"tk_commission_rate_for_media_platform"`
	AdzoneID                           string     `json:"adzone_id"`
	AdzoneName                         string     `json:"adzone_name"`
	TkCommissionPreFeeForMediaPlatform float64    `json:"tk_commission_pre_fee_for_media_platform"`
	ItemPrice                          float64    `json:"item_price"`
	TradeID                            string     `json:"trade_id"`
	PayPrice                           string     `json:"pay_price"`
	PubShareRate                       string     `json:"pub_share_rate"`
	AlimamaRate                        string     `json:"alimama_rate"`
}

var XlsConvertRuleMap = map[string]func(string) string{
	"订单类型":      shopTypeConvert,
	"订单状态":      tkStatusConvert,
	"提成":        removePercent,
	"分成比率":      removePercent,
	"技术服务费率":    removePercent,
	"佣金比率":      removePercent,
	"补贴比率":      removePercent,
	"内容专项服务费率":  removePercent,
	"收入比率":      removePercent,
	"商品数量":      isNumber,
	"商品单价":      isNumber,
	"补贴金额":      isNumber,
	"技术服务费":     isNumber,
	"结算预估收入":    isNumber,
	"预估内容专项服务费": isNumber,
	"结算内容专项服务费": isNumber,
	"商品图片":      CompleteURL,
}

func tkStatusConvert(s string) string {
	for k, v := range OrderStatusMap {
		if s == v {
			return strconv.Itoa(int(k))
		}
	}
	return "0"
}
func shopTypeConvert(s string) string {
	for k, v := range OrderTypeMap {
		if s == v {
			return strconv.Itoa(int(k))
		}
	}
	return "0"
}
func removePercent(s string) string {
	i := strings.Index(s, "%")
	return s[:i]
}

func isNumber(s string) string {
	return "number"
}

func CompleteURL(s string) string {
	parse, err := url.Parse(s)
	if err != nil {
		return ""
	}
	parse.Scheme = "https"
	return parse.String()
}
