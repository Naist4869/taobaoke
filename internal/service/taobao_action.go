package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"taobaoke/internal/model"
	"taobaoke/tools"
)

type RawMessage = json.RawMessage
type Time = tools.Time

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
type convertMyKeyResult struct {
	Title    string
	ItemID   int64
	PicURL   string
	ShopName string
	ShopType int
	Price    int64
	Rebate   int64
	Coupon   int64
}
type HighCommissionResp struct {
	Result struct {
		Data HighCommissionResult `json:"data"`
	} `json:"result"`
	RequestID string `json:"request_id"`
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
}
type HighCommissionResult struct {
	CategoryID        int    `json:"category_id"`
	CouponClickURL    string `json:"coupon_click_url"`
	CouponEndTime     string `json:"coupon_end_time"`
	CouponInfo        string `json:"coupon_info"`
	CouponRemainCount int    `json:"coupon_remain_count"`
	CouponStartTime   string `json:"coupon_start_time"`
	CouponTotalCount  int    `json:"coupon_total_count"`
	ItemID            int64  `json:"item_id"`
	ItemURL           string `json:"item_url"`
	MaxCommissionRate string `json:"max_commission_rate"`
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

func (t *tbkitemInfoGetResp) Error() error {
	if e := t.RespCommon.Error(); e != nil {
		return e
	}
	if len(t.ItemInfoGetResponse.Results.TbkItemInfoGetResults) == 0 {
		return errors.New("查询列表为空")
	}
	return nil
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
	queryMap["adzone_id"] = fmt.Sprintf("%d", t.AdzoneId)
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

func (t *tbkDgMaterialOptionalResp) Error() error {
	if e := t.RespCommon.Error(); e != nil {
		return e
	}
	if t.TbkDgMaterialOptionalResponse.TotalResults == 0 {
		return NewQueryListEmptyError()
	}
	return nil
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
	queryMap["start_time"] = t.StartTime.String()
	queryMap["end_time"] = t.EndTime.String()
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
		return NewQueryListEmptyError()
	}
	return nil
}
func (t *TbkOrderDetailsGetResult) FillContext() *model.UpdateArgument {
	return &model.UpdateArgument{
		PaidTime:           t.TkPaidTime,
		EarningTime:        t.TkEarningTime,
		Status:             model.OrderStatus(t.TkStatus),
		AlipayTotalPrice:   t.AlipayTotalPrice,
		IncomeRate:         t.IncomeRate,
		PubSharePreFee:     t.PubSharePreFee,
		ItemNum:            t.ItemNum,
		TotalCommissionFee: t.TotalCommissionFee,
		PayPrice:           t.PayPrice,
	}
}

func (t TbkScInvitecodeReq) Code() int {
	return tbkScInvitecodeGetCode
}

func (t TbkScInvitecodeReq) Response() Response {
	return &tbkScInvitecodeGetResp{}
}

func (t TbkScInvitecodeReq) Name() string {
	return "淘宝客邀请码生成-社交"
}

func (t TbkScInvitecodeReq) Query(queryMap map[string]string) {
	queryMap["relation_app"] = t.RelationApp
	queryMap["code_type"] = strconv.Itoa(t.CodeType)
	if t.RelationID != 0 {
		queryMap["relation_id"] = strconv.FormatInt(t.RelationID, 10)
	}
}

func (t TbkScPublisherInfoSaveReq) Code() int {
	return tbkScPublisherInfoSaveCode
}

func (t TbkScPublisherInfoSaveReq) Response() Response {
	return &tbkScPublisherInfoSaveResp{}
}

func (t TbkScPublisherInfoSaveReq) Name() string {
	return " 淘宝客信息备案"
}

func (t TbkScPublisherInfoSaveReq) Query(queryMap map[string]string) {
	queryMap["info_type"] = strconv.Itoa(t.InfoType)
	queryMap["inviter_code"] = t.InviterCode
	if t.Note != "" {
		queryMap["note"] = t.Note
	}
	if t.OfflineScene != "" {
		queryMap["offline_scene"] = t.OfflineScene
	}
	if t.OnlineScene != "" {
		queryMap["online_scene"] = t.OnlineScene
	}
	if len(t.RegisterInfo) != 0 {
		queryMap["register_info"] = string(t.RegisterInfo)
	}
	if t.RelationFrom != "" {
		queryMap["relation_from"] = t.RelationFrom
	}
	return
}

func (t TbkScPublisherInfoGetReq) Code() int {
	return tbkScPublisherInfoGetCode
}

func (t TbkScPublisherInfoGetReq) Response() Response {
	return &tbkScPublisherInfoGetResp{}
}

func (t TbkScPublisherInfoGetReq) Name() string {
	return "淘宝客信息查询"
}

func (t TbkScPublisherInfoGetReq) Query(queryMap map[string]string) {
	queryMap["info_type"] = strconv.Itoa(t.InfoType)
	queryMap["relation_app"] = t.RelationApp
	if t.ExternalId != "" {
		queryMap["external_id"] = t.ExternalId
	}
	if t.PageNo != 0 {
		queryMap["page_no"] = strconv.Itoa(t.PageNo)
	}
	if t.PageSize != 0 {
		queryMap["page_size"] = strconv.Itoa(t.PageSize)
	}
	if t.RelationId != 0 {
		queryMap["relation_id"] = strconv.FormatInt(t.RelationId, 10)
	}
	if t.SpecialId != "" {
		queryMap["special_id"] = t.SpecialId
	}
	return
}

func (t *tbkScPublisherInfoGetResp) Error() error {
	if e := t.RespCommon.Error(); e != nil {
		return e
	}
	if t.TbkScPublisherInfoGetResponse.Data.TotalCount == 0 {
		return errors.New("查询列表为空")
	}
	return nil
}
