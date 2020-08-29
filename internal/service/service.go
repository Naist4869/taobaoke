package service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"taobaoke/common"
	"taobaoke/internal/model"
	"taobaoke/tools"
	"time"

	"github.com/go-kratos/kratos/pkg/sync/errgroup"

	"github.com/teris-io/shortid"

	"go.uber.org/zap"

	"github.com/Naist4869/log"

	"github.com/extrame/xls"

	bm "github.com/go-kratos/kratos/pkg/net/http/blademaster"

	pb "taobaoke/api"
	"taobaoke/internal/dao"

	"github.com/go-kratos/kratos/pkg/conf/paladin"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
)

//go:generate kratos tool wire
var Provider = wire.NewSet(New, wire.Bind(new(pb.TBKServer), new(*Service)), NewLogger, NewBmClient, NewOrders)

// Service service.
type Service struct {
	ac          *paladin.Map
	dao         dao.Dao
	client      *bm.Client
	logger      *log.Logger
	orders      *orders // 订单缓存
	idGenerator common.IDGenerator
}

// New new a service and return.
func New(d dao.Dao, l *log.Logger, client *bm.Client, orders *orders) (s *Service, cf func(), err error) {
	var sid *shortid.Shortid
	sid, err = shortid.New(common.Taobaoke, shortid.DefaultABC, 2342)
	if err != nil {
		return
	}

	s = &Service{
		ac:          &paladin.TOML{},
		dao:         d,
		client:      client,
		logger:      l,
		orders:      orders,
		idGenerator: sid,
	}
	cf = s.Close
	err = paladin.Watch("application.toml", s.ac)
	// 查询DB中所有未完成的单添加到匹配队列
	now := tools.Now()
	if unfinishOrders, err := s.dao.QueryOrderByStatus(context.Background(), now.Add(-time.Hour*24*45), now, model.OrderPaid, model.OrderFinish); err != nil {
		s.logger.Error("Service初始化", zap.Error(err))
	} else {
		for _, order := range unfinishOrders {
			if _, err = s.dao.HSetNXToMatch(context.Background(), order); err != nil {
				s.logger.Error("Service初始化", zap.Error(err))
			}
		}
	}
	go s.Monitor()
	go s.MonitorMarch()
	return
}
func (s *Service) GetAppSecret() string {
	return paladin.String(s.ac.Get("appSecret"), "")
}
func (s *Service) GetURI() string {
	return paladin.String(s.ac.Get("uri"), "http://gw.api.taobao.com/router/rest")
}
func (s *Service) GetAppKey() string {
	return paladin.String(s.ac.Get("appKey"), "")
}
func (s *Service) GetSession() string {
	return paladin.String(s.ac.Get("session"), "")
}
func (s *Service) GenGetAdZoneID() func() int64 {
	var sliceStr []int64
	_ = s.ac.Get("adzoneID").Slice(&sliceStr)
	i := len(sliceStr)
	return func() int64 {
		if i > 0 {
			i--
			return sliceStr[i]
		}
		return 0
	}
}
func (s *Service) GetDefaultAdZoneID() int64 {
	var sliceStr []int64
	_ = s.ac.Get("adzoneID").Slice(&sliceStr)
	if len(sliceStr) > 0 {
		return sliceStr[0]
	}
	return 0
}
func (s *Service) GetTklAppKey() string {
	return paladin.String(s.ac.Get("taokoulingAppKey"), "")
}
func (s *Service) GetSalaryScale() int64 {
	return paladin.Int64(s.ac.Get("salaryScale"), 90)
}
func (s *Service) GetServerAddr() string {
	return paladin.String(s.ac.Get("serverAddr"), "")
}
func (s *Service) KeyConvert(ctx context.Context, req *pb.KeyConvertReq) (resp *pb.KeyConvertResp, err error) {
	deadline, _ := ctx.Deadline()
	s.logger.Info("KeyConvert", zap.Duration("过期时间", time.Until(deadline)))
	getadZoneID := s.GenGetAdZoneID()
	adZoneID := getadZoneID()
	id, err := s.idGenerator.Generate()
	if err != nil {
		return
	}
	keyInfo, err := s.analyzingKey(ctx, req.FromKey)
	if err != nil {
		return
	}
	r, err := s.analyzingItem(ctx, keyInfo.Content, keyInfo.PicURL, adZoneID)
	if err != nil {
		return
	}
	nonce := tools.MakeNonce()
	for ok, err := s.dao.SetNXToUnmatch(ctx, r.ItemID, adZoneID, nonce); err != nil || ok != true; adZoneID = getadZoneID() {
		s.logger.Info("KeyConvert", zap.Error(err), zap.Bool("OK", ok), zap.Int64("新AdZoneID", adZoneID))
		if adZoneID == 0 {
			err = fmt.Errorf("请稍后再试:(%w)", err)
			return nil, err
		}
		ok, err = s.dao.SetNXToUnmatch(ctx, r.ItemID, adZoneID, nonce)
	}

	order := model.NewOrder(id, req.UserID, adZoneID, r.Title, r.ItemID, r.PicURL, r.ShopName, r.ShopType, r.Price, r.ReservePrice, r.Coupon)
	if err = s.orders.Add(ctx, order, nonce); err != nil {
		return
	}
	query := url.Values{}
	query.Add("id", order.ID)
	query.Add("itemID", strconv.FormatInt(order.ItemID, 10))
	query.Add("adZoneID", strconv.FormatInt(order.AdzoneID, 10))
	rebate := strconv.FormatFloat(float64(r.Rebate)/100, 'f', -1, 64)
	coupon := strconv.FormatFloat(float64(order.Coupon)/100, 'f', -1, 64)
	price := strconv.FormatFloat(float64(order.Price-order.Coupon)/100, 'f', -1, 64)
	resp = &pb.KeyConvertResp{
		Price:   price,
		Rebate:  rebate,
		Coupon:  coupon,
		Title:   order.Title,
		PicURL:  order.PicURL,
		ItemURL: s.GetServerAddr() + "/item?" + query.Encode(),
	}
	s.logger.Info("保存订单信息", zap.String("标题", order.Title), zap.Int64("商品ID", order.ItemID), zap.Int64("广告位ID", order.AdzoneID), zap.String("预计返利", rebate), zap.String("优惠券", coupon), zap.String("到手价", price))
	return
}

// 匹配监控
func (s *Service) Monitor() {
	for range time.Tick(time.Minute) {
		now := tools.Now()
		result, err := s.execTbkOrderDetailsGet(context.Background(), TbkOrderDetailsGetReq{
			StartTime: now.Add(-time.Minute * 20),
			EndTime:   now,
		})
		if err != nil {
			if !errors.Is(err, QueryListEmpty{}) {
				s.logger.Error("Monitor", zap.Error(err))
			}
			continue
		}
		s.orders.Match(result)
	}
}

// 状态变更监控
func (s *Service) MonitorMarch() {
	for range time.Tick(time.Hour) {
		ctx := context.Background()
		orders, err := s.dao.MatchGetAll(ctx)
		if err != nil {
			s.logger.Error("MonitorMarch", zap.Error(err))
			continue
		}
		remoteOrders, err := s.QueryRemoteOrderByTradeParentID(ctx, orders)
		if err != nil {
			s.logger.Error("MonitorMarch", zap.Error(err))
			continue
		}
		ordersMap := make(map[string]*model.Order, len(orders))
		for _, order := range orders {
			if _, ok := remoteOrders.Load(order.ID); ok {
				ordersMap[order.ID] = order
			}
		}
		remoteOrders.Range(func(key, value interface{}) bool {
			id := key.(string)
			remoteOrder := value.(TbkOrderDetailsGetResult)
			localOrder := ordersMap[id]
			status := model.OrderStatus(remoteOrder.TkStatus)
			if localOrder.Status != status {
				s.logger.Info("MonitorMarch", zap.String("动作", "检测到远程订单状态变更"), zap.Any("本地订单", localOrder), zap.Any("远程订单", remoteOrder))
				var err error
				switch status {
				case model.OrderPaid:
					err = s.dao.UpdateOrderPaidStatus(ctx, id, remoteOrder.TkPaidTime, remoteOrder.AlipayTotalPrice, remoteOrder.IncomeRate, remoteOrder.PubSharePreFee, remoteOrder.ItemNum)
				case model.OrderFailed:
					err = s.dao.UpdateOrderFailedStatus(ctx, id, remoteOrder.TradeParentID)
				case model.OrderBalance:
					_, err = s.dao.FindOneAndUpdateOrderBalanceStatus(ctx, id, remoteOrder.TradeParentID, remoteOrder.TkEarningTime, remoteOrder.TotalCommissionFee, remoteOrder.PayPrice, s.GetSalaryScale())
				default:
					s.logger.Error("MonitorMarch", zap.Int("未知的状态:", remoteOrder.TkStatus), zap.Any("远程订单", remoteOrder), zap.Any("本地订单", localOrder))
				}
				if err != nil {
					s.logger.Error("MonitorMarch", zap.Error(err), zap.Any("远程订单", remoteOrder), zap.Any("本地订单", localOrder))
				}
			}
			return true
		})
	}
}

// 订单状态变更监控 实现有两种  一种是提现的时候才查单   一种是每个小时都查一边数据库里的单 目前先实现第一种吧 因为是实时的
func (s *Service) WithDraw(ctx context.Context, req *pb.WithDrawReq) (*pb.WithDrawResp, error) {
	orders, err := s.dao.QueryNotWithDrawOrderByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	remoteOrders, err := s.QueryRemoteOrderByTradeParentID(ctx, orders)
	if err != nil {
		return nil, err
	}
	ordersMap := make(map[string]*model.Order, len(orders))
	for _, order := range orders {
		if _, ok := remoteOrders.Load(order.ID); ok {
			ordersMap[order.ID] = order
		}
	}
	var totalSalary float64
	var withdrawSlice []string
	remoteOrders.Range(func(key, value interface{}) bool {
		id := key.(string)
		remoteOrder := value.(TbkOrderDetailsGetResult)
		localOrder := ordersMap[id]
		if !localOrder.Status.Balance() {
			if model.OrderStatus(remoteOrder.TkStatus).Balance() {
				afterOrder, err := s.dao.FindOneAndUpdateOrderBalanceStatus(ctx, id, localOrder.TradeParentID, remoteOrder.TkEarningTime, remoteOrder.TotalCommissionFee, remoteOrder.PayPrice, s.GetSalaryScale())
				if err != nil {
					s.logger.Error("WithDraw", zap.Error(err), zap.Any("localOrder", localOrder), zap.Any("RemoteOrder", remoteOrder))
					return true
				}
				if afterOrder.Salary == 0 {
					s.logger.Error("WithDraw", zap.String("原因", "返给用户的金额为0"), zap.Any("localOrder", localOrder), zap.Any("RemoteOrder", remoteOrder))
					return true
				}
				totalSalary += float64(afterOrder.Salary)
				withdrawSlice = append(withdrawSlice, afterOrder.ID)
				return true
			}
			return true
		} else {
			if localOrder.Salary == 0 {
				s.logger.Error("WithDraw", zap.String("原因", "返给用户的金额为0"), zap.Any("localOrder", localOrder), zap.Any("RemoteOrder", remoteOrder))
				return true
			}
			totalSalary += float64(localOrder.Salary)
			withdrawSlice = append(withdrawSlice, localOrder.ID)
			return true
		}
	})
	err = s.dao.UpdateManyWithDrawStatus(ctx, withdrawSlice)
	if err != nil {
		s.logger.Error("WithDraw", zap.Error(err), zap.Strings("待更新的withdrawSlice", withdrawSlice))
		return nil, err
	}
	return &pb.WithDrawResp{
		Rebate:   strconv.FormatFloat(totalSalary/100, 'f', -1, 64),
		OrderIDs: withdrawSlice,
	}, nil
}

func (s *Service) QueryRemoteOrderByTradeParentID(ctx context.Context, orders []*model.Order) (remoteOrders sync.Map, err error) {
	group := errgroup.WithCancel(ctx)
	group.GOMAXPROCS(30)
	for _, order := range orders {
		o := order
		group.Go(func(ctx context.Context) error {
			result, err := s.execTbkOrderDetailsGet(ctx, TbkOrderDetailsGetReq{
				QueryType: 2,
				StartTime: o.PaidTime,
				EndTime:   o.PaidTime,
			})
			if err != nil {
				if !errors.Is(err, QueryListEmpty{}) {
					s.logger.Error("查询失败", zap.Error(err))
				}
				return err
			}
			for _, remoteOrder := range result {
				if remoteOrder.TradeParentID == o.TradeParentID {
					// ID -> TbkOrderDetailsGetResult
					remoteOrders.Store(order.ID, remoteOrder)
					return nil
				}
			}
			s.logger.Error("QueryRemoteOrderByTradeParentID", zap.Error(err), zap.Any("远程订单", result), zap.Any("待匹配订单", o))
			return nil
		})
	}
	err = group.Wait()
	if err != nil {
		s.logger.Error("QueryRemoteOrderByTradeParentID", zap.Error(err), zap.Any("待匹配订单", orders))
		err = fmt.Errorf("QueryRemoteOrderByTradeParentID error :%w", err)
		return
	}
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return e, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
}
func (s *Service) methodPost(ctx context.Context, req Request, resp Response, method string) error {
	uri := s.GetURI()
	appKey := s.GetAppKey()
	session := s.GetSession()
	query := url.Values{}
	queryMap := map[string]string{
		"method":      method,
		"app_key":     appKey,
		"format":      "json",
		"sign_method": "md5",
		"v":           "2.0",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"session":     session,
		//"simplify":    "true",
	}
	req.Query(queryMap)
	param := make([]string, 0, 20)
	for key, value := range queryMap {
		query.Set(key, value)
		param = append(param, key+value)
	}

	sign := s.Sign(param...)
	query.Set("sign", sign)
	if err := s.client.Post(ctx, uri, "", query, resp); err != nil {
		return err
	}
	if resp.Error() != nil {
		return resp.Error()
	}
	return nil
}

// 获取历史价格趋势
func (s *Service) PriceTrend(ctx context.Context, itemID int64) (trendInfo model.TrendInfo, err error) {
	price_trend_uri := "https://m.gwdang.com/trend/data_new"
	query := url.Values{}
	query.Add("opt", "trend")
	query.Add("dp_id", strconv.FormatInt(itemID, 10)+"-83")
	query.Add("from", "m")
	query.Add("period", "180")
	query.Add("is_coupon", "0")
	//query.Add("price")
	//query.Add("org_price","0")
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s?%s", price_trend_uri, query.Encode()), nil)
	if err != nil {
		err = fmt.Errorf("priceTrend创建请求: %w", err)
		return
	}
	request.Header.Set("user-agent", "Mozilla/5.0 (Linux; Android 5.0; SM-N9100 Build/LRX21V) > AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 > Chrome/37.0.0.0 Mobile Safari/537.36 > MicroMessenger/6.0.2.56_r958800.520 NetType/WIFI")
	resp := &model.PriceTrendResp{}
	if err = s.client.JSON(ctx, request, resp); err != nil {
		err = fmt.Errorf("priceTrend返回响应: %w", err)
		return
	}
	if resp.Code != 0 || len(resp.Data.Series) == 0 {
		err = fmt.Errorf("priceTrend解析响应: %w", err)
		return
	}
	list := resp.Data.Series[0]
	for i, v := range list.Data {
		list.Data[i].Y = v.Y / 100
	}
	marshal, err := json.Marshal(list.Data)
	if err != nil {
		err = fmt.Errorf("priceTrend把价格趋势JsonArray转化为Raw失败: %w", err)
		return
	}
	trendInfo.RawJsonTrend = string(marshal)
	trendInfo.CurrentPrice = strconv.FormatFloat(list.Current/100, 'f', -1, 64)
	trendInfo.MaxPrice = strconv.FormatFloat(list.Max/100, 'f', -1, 64)
	trendInfo.MinPrice = strconv.FormatFloat(list.Min/100, 'f', -1, 64)
	trendInfo.OriginalPrice = strconv.FormatFloat(list.Original/100, 'f', -1, 64)
	trendInfo.Period = list.Period
	trendInfo.EffectiveDate = tools.Now().DayStart()
	switch list.Trend {
	case -1:
		trendInfo.TrendMsg = "价格下降"
	case 1:
		trendInfo.TrendMsg = "价格上涨"
	case 0:
		trendInfo.TrendMsg = "价格平稳"
	case -2:
		trendInfo.TrendMsg = "历史最低"
	}
	return
}

func (s *Service) analyzingItem(ctx context.Context, title, picUrl string, adZoneID int64) (result convertMyKeyResult, err error) {

	materialResult, err := s.execTbkDgMaterialOptional(ctx, TbkDgMaterialOptionalReq{
		AdzoneId: adZoneID,
		Q:        title,
		Sort:     "total_sales_des",
	})
	if err != nil {
		return
	}
	result.AdzoneID = adZoneID
	for _, item := range materialResult {
		index := strings.Index(item.PictURL, "uploaded")
		if !strings.Contains(picUrl, item.PictURL[index+12:]) {
			continue
		}
		var (
			commissionRate              int64 = 1
			price, reservePrice, coupon float64
		)
		result.Title = item.Title
		result.ItemID = item.ItemID
		s.logger.Info("商品信息", zap.String("标题", item.Title), zap.Int64("商品ID", item.ItemID), zap.String("一口价", item.ReservePrice), zap.String("折扣价", item.ZkFinalPrice), zap.String("佣金比率", item.CommissionRate))
		if price, err = strconv.ParseFloat(item.ZkFinalPrice, 64); err != nil {
			s.logger.Error("analyzingItem", zap.Error(err), zap.String("折扣价", item.ZkFinalPrice))
		}
		if reservePrice, err = strconv.ParseFloat(item.ReservePrice, 64); err != nil {
			s.logger.Error("analyzingItem", zap.Error(err), zap.String("一口价", item.ReservePrice))
		}
		if item.CouponAmount != "" {
			if coupon, err = strconv.ParseFloat(item.CouponAmount, 64); err != nil {
				s.logger.Error("analyzingItem", zap.Error(err), zap.String("优惠券金额", item.CouponAmount))
			}
		}
		commissionRate, err = strconv.ParseInt(item.CommissionRate, 10, 64)
		if err != nil {
			s.logger.Error("analyzingItem", zap.Error(err), zap.String("佣金比率", item.CommissionRate))
		}
		result.ShopName = item.Nick
		result.PicURL = item.PictURL
		result.ShopType = item.UserType
		result.ReservePrice = int64(reservePrice * 100)
		result.Price = int64(price * 100)
		// 商品价格-优惠券价格*佣金比率=预计收入佣金
		result.Rebate = int64((price-coupon)*float64(commissionRate)) / 100
		result.Coupon = int64(coupon * 100)
		return
	}
	err = errors.New("目标商品未找到")
	return
}

func (s *Service) GetTklByItemID(ctx context.Context, itemID int64, adZoneID int64, title string) (tkl string, URL, CouponShareURL string, err error) {
	materialResult, err := s.execTbkDgMaterialOptional(ctx, TbkDgMaterialOptionalReq{
		AdzoneId: adZoneID,
		Q:        title,
		Sort:     "total_sales_des",
	})
	if err != nil {
		err = fmt.Errorf("GetTklByItemID fail: %w)", err)
		return
	}
	for _, item := range materialResult {
		var (
			parseUrl         *url.URL
			tpwdCreateResult TbkTpwdCreateResult
		)

		if item.ItemID != itemID {
			continue
		}

		if item.CouponShareURL != "" {
			URL = item.CouponShareURL
		} else {
			URL = item.URL
		}
		parseUrl, err = url.Parse(URL)
		if err != nil {
			return
		}
		parseUrl.Scheme = "https"
		URL = parseUrl.String()
		tpwdCreateResult, err = s.execTbkTpwdCreate(ctx, TbkTpwdCreateReq{
			Text: "啊实打实的撒大苏打萨达萨达萨达是的观点",
			URL:  URL,
		})
		if err != nil {
			return
		}
		tkl = tpwdCreateResult.Model
		return
	}
	err = errors.New("GetTklByItemID: 目标商品未找到")
	return
}

// 淘宝客签名 https://open.taobao.com/doc.htm?docId=101617&docType=1
func (s *Service) Sign(strs ...string) (signature string) {
	sort.Strings(strs)
	tmpstr := strings.Join(strs, "")
	secret := s.GetAppSecret()
	str := secret + tmpstr + secret
	signature = fmt.Sprintf("%X", md5.Sum([]byte(str)))
	return
}
func EazyCopyStruct(from interface{}, to interface{}) {
	formValue := reflect.ValueOf(from)
	toValue := reflect.Indirect(reflect.ValueOf(to))
	toType := reflect.TypeOf(to)
	if toType.Kind() == reflect.Ptr {
		toType = toType.Elem()
	}
	if toValue.CanSet() {
		formValue = DFSFindStruct(formValue, toType)
		if formValue.IsValid() {
			toValue.Set(formValue)
		}

	}
}
func DFSFindStruct(fromValue reflect.Value, toType reflect.Type) reflect.Value {
	if fromValue.Kind() == reflect.Struct {
		for i := 0; i < fromValue.NumField(); i++ {
			field := fromValue.Field(i)
			if field.Type().AssignableTo(toType) {
				return field
			}
			dest := DFSFindStruct(field, toType)
			if dest.IsValid() {
				return dest
			}
		}
	}
	return reflect.Value{}
}
func (s *Service) UnmatchGet(ctx context.Context, itemID, adZoneID int64) (*model.Order, error) {
	return s.dao.GetUnmatch(ctx, itemID, adZoneID)
}

func (s *Service) UpdateToUnmatch(ctx context.Context, itemID, adZoneID int64, order *model.Order) (ok bool, err error) {
	return s.dao.UpdateFromUnmatch(ctx, itemID, adZoneID, order)
}

func OpenXLS() {
	if xlFile, err := xls.Open("OrderDetail-2020-06-13.xls", "utf-8"); err == nil {
		if sheet1 := xlFile.GetSheet(0); sheet1 != nil {
			fmt.Print("Total Lines ", sheet1.MaxRow, sheet1.Name)
			for row := 0; row <= int(sheet1.MaxRow); row++ {
				currentRow := sheet1.Row(row)
				for col := 0; col < currentRow.LastCol(); col++ {
					fmt.Printf("%s ", currentRow.Col(col))
				}
				fmt.Println()
			}
		}
	}
}

func (s *Service) HighCommission(ctx context.Context, numIid int64) (result HighCommissionResult, err error) {
	adzoneID := s.GetDefaultAdZoneID()
	appKey := s.GetTklAppKey()
	itemID := strconv.FormatInt(numIid, 10)
	param := url.Values{}
	param.Set("apikey", appKey)
	param.Set("itemid", itemID)
	param.Set("siteid", "43474861")
	param.Set("adzoneid", fmt.Sprintf("%d", adzoneID))
	param.Set("uid", "2329747174")
	var request *http.Request
	request, err = s.client.NewRequest(http.MethodGet, "https://api.taokouling.com/tkl/TbkPrivilegeGet", "", param)
	if err != nil {
		return
	}

	var resp HighCommissionResp
	if err = s.client.JSON(ctx, request, &resp); err != nil {
		return
	}
	if resp.Code != 0 {
		err = fmt.Errorf("HighCommission: 错误代码: %d, 错误信息: %s", resp.Code, resp.Msg)
		return
	}
	EazyCopyStruct(resp, &result)
	return
}
func (s *Service) analyzingKey(ctx context.Context, fromKey string) (resp analyzingKeyResp, err error) {
	param := url.Values{}
	appKey := s.GetTklAppKey()
	param.Set("apikey", appKey)
	param.Set("tkl", fromKey)
	var request *http.Request
	request, err = s.client.NewRequest(http.MethodGet, "https://api.taokouling.com/tkl/tkljm", "", param)
	if err != nil {
		return
	}
	if err = s.client.JSON(ctx, request, &resp); err != nil {
		return
	}
	if resp.Code == 0 {
		err = fmt.Errorf("analyzingKey: 错误信息:%s", resp.Msg)
		return
	}

	return
}
