package service

import (
	"errors"
	"fmt"
	"strconv"
)

const TimeFormat = "2006-01-02 15:04:05"

type analyzingKeyResp struct {
	Ret       string `json:"ret"`
	URL       string `json:"url"`
	Content   string `json:"content"`
	PicURL    string `json:"picUrl"`
	ValidDate string `json:"validDate"`
	Pjk       string `json:"pjk"`
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
}

type HighCommissionResp struct {
	Result struct {
		Data HighCommissionResult `json:"data"`
	} `json:"result"`
	RequestID string `json:"request_id"`
}
type HighCommissionResult struct {
	CategoryID          int    `json:"category_id"`
	CouponClickURL      string `json:"coupon_click_url"`
	CouponEndTime       string `json:"coupon_end_time"`
	CouponInfo          string `json:"coupon_info"`
	CouponStartTime     string `json:"coupon_start_time"`
	ItemID              int64  `json:"item_id"`
	MaxCommissionRate   string `json:"max_commission_rate"`
	CouponTotalCount    int    `json:"coupon_total_count"`
	CouponRemainCount   int    `json:"coupon_remain_count"`
	MmCouponRemainCount int    `json:"mm_coupon_remain_count"`
	MmCouponTotalCount  int    `json:"mm_coupon_total_count"`
	MmCouponClickURL    string `json:"mm_coupon_click_url"`
	MmCouponEndTime     string `json:"mm_coupon_end_time"`
	MmCouponStartTime   string `json:"mm_coupon_start_time"`
	MmCouponInfo        string `json:"mm_coupon_info"`
	CouponType          int    `json:"coupon_type"`
	ItemURL             string `json:"item_url"`
}

type HighCommissionErrorMsg struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
type RespCommon struct {
	ErrorResponse struct {
		SubMsg    string `json:"sub_msg"`
		Code      int    `json:"code"`
		SubCode   string `json:"sub_code"`
		Msg       string `json:"msg"`
		RequestID string `json:"request_id,omitempty"`
	} `json:"error_response,omitempty"`
}

func (e *RespCommon) Error() error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("错误代码: %d,错误信息: %s,子错误代码: ,%s,子错误信息: %s", e.ErrorResponse.Code, e.ErrorResponse.Msg, e.ErrorResponse.SubCode, e.ErrorResponse.SubMsg)
}

type Request interface {
	Code() int
	Response() Response
	Name() string
	Query(queryMap map[string]string)
}

type Response interface {
	Error() error
}

func (i TbkItemInfoGetReq) Code() int {
	return itemInfoGetCode
}

func (i TbkItemInfoGetReq) Response() Response {
	return &itemClickExtractResp{}
}
func (i TbkItemInfoGetReq) Name() string {
	return "淘宝客商品详情查询（简版）"
}

func (i TbkItemInfoGetReq) Query(queryMap map[string]string) {
	queryMap["num_iids"] = i.NumIDs
	if i.Ip != "" {
		queryMap["ip"] = i.Ip

	}
	if i.Platform != 0 {
		queryMap["platform"] = strconv.Itoa(i.Platform)
	}
}
func (t TbkTpwdCreateReq) Code() int {
	return tbkTpwdCreateCode
}

func (t TbkTpwdCreateReq) Response() Response {
	return &tbkTpwdCreateResp{}
}

func (t TbkTpwdCreateReq) Name() string {
	return "淘宝客-公用-淘口令生成"
}

func (t TbkTpwdCreateReq) Query(queryMap map[string]string) {
	queryMap["text"] = t.Text
	queryMap["url"] = t.URL
	if t.Ext != "" {
		queryMap["ext"] = t.Ext
	}
	if t.Logo != "" {
		queryMap["logo"] = t.Logo
	}
	if t.UserID != "" {
		queryMap["user_id"] = t.UserID
	}
}
func (t TbkDgMaterialOptionalReq) Code() int {
	return tbkDgMaterialOptionalCode
}

func (t TbkDgMaterialOptionalReq) Response() Response {
	return &tbkDgMaterialOptionalResp{}
}

func (t TbkDgMaterialOptionalReq) Name() string {
	return "淘宝客-推广者-物料搜索"
}

func (t TbkDgMaterialOptionalReq) Query(queryMap map[string]string) {
	queryMap["adzone_id"] = strconv.Itoa(t.AdzoneId)
	if t.StartDsr != 0 {
		queryMap["start_dsr"] = strconv.Itoa(t.StartDsr)
	}
	if t.PageSize != 0 {
		queryMap["page_size"] = strconv.Itoa(t.PageSize)
	}
	if t.PageNo != 0 {
		queryMap["page_no"] = strconv.Itoa(t.PageNo)
	}
	if t.Platform != 0 {
		queryMap["platform"] = strconv.Itoa(t.Platform)
	}
	if t.EndTkRate != 0 {
		queryMap["end_tk_rate"] = strconv.Itoa(t.EndTkRate)
	}
	if t.StartTkRate != 0 {
		queryMap["start_tk_rate"] = strconv.Itoa(t.StartTkRate)
	}
	if t.EndPrice != 0 {
		queryMap["end_price"] = strconv.Itoa(t.EndPrice)
	}
	if t.StartPrice != 0 {
		queryMap["start_price"] = strconv.Itoa(t.StartPrice)
	}
	if t.IsOverseas {
		queryMap["is_overseas"] = strconv.FormatBool(t.IsOverseas)
	}
	if t.IsTmall {
		queryMap["is_tmall"] = strconv.FormatBool(t.IsTmall)
	}
	if t.Sort != "" {
		queryMap["sort"] = t.Sort
	}
	if t.Itemloc != "" {
		queryMap["itemloc"] = t.Itemloc
	}
	if t.Cat != "" {
		queryMap["cat"] = t.Cat
	}
	if t.Q != "" {
		queryMap["q"] = t.Q
	}
	if t.MaterialId != 0 {
		queryMap["material_id"] = strconv.Itoa(t.MaterialId)
	}
	if t.HasCoupon {
		queryMap["has_coupon"] = strconv.FormatBool(t.HasCoupon)
	}
	if t.Ip != "" {
		queryMap["ip"] = t.Ip
	}
	if t.NeedFreeShipment {
		queryMap["need_free_shipment"] = strconv.FormatBool(t.NeedFreeShipment)
	}
	if t.NeedPrepay {
		queryMap["need_prepay"] = strconv.FormatBool(t.NeedPrepay)
	}
	if t.IncludePayRate30 {
		queryMap["include_pay_rate_30"] = strconv.FormatBool(t.IncludePayRate30)
	}
	if t.IncludeGoodRate {
		queryMap["include_good_rate"] = strconv.FormatBool(t.IncludeGoodRate)
	}
	if t.IncludeRfdRate {
		queryMap["include_rfd_rate"] = strconv.FormatBool(t.IncludeRfdRate)
	}
	if t.NpxLevel != 0 {
		queryMap["npx_level"] = strconv.Itoa(t.NpxLevel)
	}
	if t.EndKaTkRate != 0 {
		queryMap["end_ka_tk_rate"] = strconv.Itoa(t.EndKaTkRate)
	}
	if t.StartKaTkRate != 0 {
		queryMap["start_ka_tk_rate"] = strconv.Itoa(t.StartKaTkRate)
	}
	if t.DeviceEncrypt != "" {
		queryMap["device_encrypt"] = t.DeviceEncrypt
	}
	if t.DeviceValue != "" {
		queryMap["device_value"] = t.DeviceValue
	}
	if t.DeviceType != "" {
		queryMap["device_type"] = t.DeviceType
	}
	if t.LockRateEndTime != 0 {
		queryMap["lock_rate_end_time"] = strconv.FormatInt(t.LockRateEndTime, 10)
	}
	if t.LockRateStartTime != 0 {
		queryMap["lock_rate_start_time"] = strconv.FormatInt(t.LockRateStartTime, 10)
	}
	if t.Longitude != "" {
		queryMap["longitude"] = t.Longitude
	}
	if t.Latitude != "" {
		queryMap["latitude"] = t.Latitude
	}
	if t.CityCode != "" {
		queryMap["city_code"] = t.CityCode
	}
	if t.SellerIds != "" {
		queryMap["seller_ids"] = t.SellerIds
	}
}
func (t TbkOrderDetailsGetReq) Code() int {
	return tbkOrderDetailsGetCode
}

func (t TbkOrderDetailsGetReq) Response() Response {
	return &tbkOrderDetailsGetResp{}
}

func (t TbkOrderDetailsGetReq) Name() string {
	return "淘宝客-推广者-所有订单查询"
}

func (t TbkOrderDetailsGetReq) Query(queryMap map[string]string) {
	queryMap["start_time"] = t.StartTime.Format(TimeFormat)
	queryMap["end_time"] = t.EndTime.Format(TimeFormat)
	if t.QueryType != 0 {
		queryMap["query_type"] = strconv.Itoa(t.QueryType)
	}
	if t.PositionIndex != "" {
		queryMap["position_index"] = t.PositionIndex
	}
	if t.PageSize != 0 {
		queryMap["page_size"] = strconv.Itoa(t.PageSize)
	}
	if t.MemberType != 0 {
		queryMap["member_type"] = strconv.Itoa(t.MemberType)
	}
	if t.TkStatus != 0 {
		queryMap["tk_status"] = strconv.Itoa(t.TkStatus)
	}
	if t.JumpType != 0 {
		queryMap["jump_type"] = strconv.Itoa(t.JumpType)
	}
	if t.PageNo != 0 {
		queryMap["page_no"] = strconv.Itoa(t.PageNo)
	}
	if t.OrderScene != 0 {
		queryMap["order_scene"] = strconv.Itoa(t.OrderScene)
	}
}
func (t *tbkOrderDetailsGetResp) Error() error {
	if e := t.RespCommon.Error(); e != nil {
		return e
	}
	if len(t.TbkOrderDetailsGetResponse.Data.Results.TbkOrderDetailsGetResults) == 0 {
		return errors.New("查询列表为空")
	}
	return nil
}
