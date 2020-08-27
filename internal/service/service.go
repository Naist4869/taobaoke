package service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
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

	"github.com/360EntSecGroup-Skylar/excelize/v2"

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
var Ngrok = "http://123.56.29.61"

// Service service.
type Service struct {
	ac          *paladin.Map
	dao         dao.Dao
	client      *bm.Client
	logger      *log.Logger
	orders      *orders // 订单缓存
	idGenerator common.IDGenerator
}

func (s *Service) GetItem(itemID string) {
	panic("implement me")
}
func (s *Service) GetAppSecret() string {
	return paladin.String(s.ac.Get("appSecret"), "")
}
func (s *Service) KeyConvertKey(ctx context.Context, req *pb.KeyConvertKeyReq) (resp *pb.KeyConvertKeyResp, err error) {
	deadline, _ := ctx.Deadline()
	s.logger.Info("KeyConvertKey", zap.Duration("过期时间", time.Until(deadline)))
	getadZoneID := s.GenGetadZoneID()
	adZoneID := getadZoneID()
	orderNo, err := s.idGenerator.Generate()
	if err != nil {
		return
	}
	keyInfo, err := s.analyzingKey(ctx, req.FromKey)
	if err != nil {
		return
	}
	title := keyInfo.Content
	picURL := keyInfo.PicURL
	r, err := s.convertMyKey(ctx, title, picURL, adZoneID)
	if err != nil {
		return
	}
	var ok bool
	for adZoneID != 0 {
		ok, err = s.dao.SetNXToUnmatch(ctx, r.ItemID, adZoneID, orderNo)
		if err != nil || !ok {
			s.logger.Warn("KeyConvertKey更换AdZoneID", zap.Error(err), zap.Int64("原AdZoneID", adZoneID))
			adZoneID = getadZoneID()
			s.logger.Warn("KeyConvertKey更换AdZoneID", zap.Error(err), zap.Int64("新AdZoneID", adZoneID))
			continue
		}
		if r.Key, r.URL, r.CouponShareURL, err = s.GetTklByItemID(ctx, r.ItemID, adZoneID, title); err != nil {
			s.logger.Warn("KeyConvertKey", zap.Error(err), zap.Int64("itemID", r.ItemID), zap.Int64("adZoneID", adZoneID), zap.String("title", title))
			//s.dao.DelFromUnmatchMap(context.Background(), r.ItemID, adZoneID)
			continue
		}
		break
	}
	if !ok {
		err = fmt.Errorf("添加商品至未匹配队列失败: %w", err)
		return
	}

	// todo 这里要放入map
	order := model.NewOrder(orderNo, req.UserID, adZoneID, r.Title, r.ItemID, r.PicURL, r.ShopName, r.ShopType, r.Price, r.ReservePrice, r.Coupon, r.URL, r.CouponShareURL)

	if err = s.orders.Add(ctx, order); err != nil {
		return
	}

	trendInfo, err := s.PriceTrend(ctx, order.ItemID)
	if err != nil {
		return
	}
	order.TrendInfo = trendInfo
	order.TrendInfo.TKL = r.Key
	resp = &pb.KeyConvertKeyResp{
		ToKey:   r.Key,
		Price:   strconv.FormatFloat(float64(order.Price-order.Coupon)/100, 'f', -1, 64),
		Rebate:  strconv.FormatFloat(float64(r.Rebate)/100, 'f', -1, 64),
		Coupon:  strconv.FormatFloat(float64(order.Coupon)/100, 'f', -1, 64),
		Title:   order.Title,
		PicURL:  order.PicURL,
		ItemURL: Ngrok + "/item/" + strconv.FormatInt(order.AdzoneID, 10) + "/" + strconv.FormatInt(order.ItemID, 10) + "/" + order.ID,
	}
	s.logger.Info("保存订单信息", zap.String("标题", order.Title), zap.Int64("商品ID", order.ItemID), zap.Int64("价格", order.Price), zap.String("淘口令", r.Key))
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
				s.logger.Error("查询失败", zap.Error(err))
			}
			continue
		}
		s.orders.Match(result)
	}
}

// 订单状态变更监控 实现有两种  一种是提现的时候才查单   一种是每个小时都查一边数据库里的单 目前实现第一种吧 因为是实时的
func (s *Service) WithDraw(ctx context.Context, req *pb.WithDrawReq) (*pb.WithDrawResp, error) {
	orders, err := s.dao.QueryNotGiveSalaryOrderByUserID(ctx, req.UserID)
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
	var changedOrderIDs []string
	remoteOrders.Range(func(key, value interface{}) bool {
		id := key.(string)
		remoteOrder := value.(TbkOrderDetailsGetResult)
		status := model.OrderStatus(remoteOrder.TkStatus)
		if status.Balance() {
			localOrder := ordersMap[id]
			err := localOrder.MakeCommission(remoteOrder.TkEarningTime, remoteOrder.TotalCommissionFee, 90, remoteOrder.TkStatus)
			if err != nil {
				s.logger.Error("WithDraw", zap.Error(err), zap.Any("RemoteOrder", remoteOrder))
			}
		}
		return true
	})
	for id, localOrder := range ordersMap {
		if err = s.dao.FindOneAndUpdateCommission(ctx, id, localOrder); err != nil {
			s.logger.Error("WithDraw", zap.Error(err), zap.Any("localOrder", localOrder))
			continue
		}
		totalSalary += float64(localOrder.Salary)
		changedOrderIDs = append(changedOrderIDs, id)
		if _, err := s.dao.DelFromOrderCache(ctx, []string{localOrder.TradeParentID}); err != nil {
			s.logger.Error("WithDraw", zap.Error(err), zap.String("changedTradeParentID", id))
		}
	}
	return &pb.WithDrawResp{
		Rebate:   strconv.FormatFloat(totalSalary/100, 'f', -1, 64),
		OrderIDs: changedOrderIDs,
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

func (s *Service) TitleConvertTBKey(ctx context.Context, req *pb.TitleConvertTBKeyReq) (resp *pb.TitleConvertTBKeyResp, err error) {
	var key string
	key, err = s.Convert(ctx, req.Title)
	if err != nil {
		return
	}
	resp = &pb.TitleConvertTBKeyResp{
		TBKey: key,
	}
	return
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
	go s.Monitor()
	return
}

//func (s *Service) historicalPrice(ctx context.Context) {
//	header := map[string]string{
//		"authority":       "www.gwdang.com",
//		"accept":          "application/json, text/javascript, */*; q=0.01",
//		"user-agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.106 Safari/537.36",
//		"referer":         "https://www.gwdang.com/trend?url=https%3A%2F%2Fdetail.tmall.com%2Fitem.htm%3Fid%3D603587505846&days=180",
//		"accept-language": "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
//	}
//	query := url.Values{}
//	query.Set("dp_id", "603587505846-83")
//	query.Set("show_prom", "true")
//	query.Set("v", "2")
//	query.Set("get_coupon", "0")
//	//query.Set("price","24.9")
//	s.client.NewRequest(http.MethodGet, "https://www.gwdang.com/trend/data_www")
//
//}

// SayHello grpc demo func.
func (s *Service) SayHello(ctx context.Context, req *pb.HelloReq) (reply *empty.Empty, err error) {
	reply = new(empty.Empty)
	fmt.Printf("hello %s", req.Name)
	return
}

// SayHelloURL bm demo func.
func (s *Service) SayHelloURL(ctx context.Context, req *pb.HelloReq) (reply *pb.HelloResp, err error) {
	reply = &pb.HelloResp{
		Content: "hello " + req.Name,
	}
	fmt.Printf("hello url %s", req.Name)
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
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
func Separate(number string) string {
	integerPart, decimalPart := separate(number)
	return integerPart + "." + decimalPart
}
func separate(number string) (integerPart string, decimalPart string) {
	switch len(number) {
	case 0:
		decimalPart = "00"
		integerPart = "0"
	case 1:
		decimalPart = "0" + number
		integerPart = "0"
	case 2:
		decimalPart = number
		integerPart = "0"
	default:
		integerPart = number[:len(number)-2]
		decimalPart = number[len(number)-2:]
	}
	return
}

type keyConvertKeyResult struct {
	convertMyKeyResult
	ValidDate string
}

// 淘口令转高佣淘口令
func (s *Service) keyConvertKey(ctx context.Context, fromKey string, adZoneID int64) (result keyConvertKeyResult, err error) {
	keyInfo, err := s.analyzingKey(ctx, fromKey)
	if err != nil {
		return
	}
	title := keyInfo.Content
	picURL := keyInfo.PicURL
	convertMyKeyInfo, err := s.convertMyKey(ctx, title, picURL, adZoneID)
	if err != nil {
		return
	}
	result.convertMyKeyResult = convertMyKeyInfo
	result.ValidDate = keyInfo.ValidDate
	return
}

type convertMyKeyResult struct {
	AdzoneID       int64
	Title          string
	ItemID         int64
	PicURL         string
	ShopName       string
	ShopType       int
	Price          int64
	Rebate         int64
	Coupon         int64
	URL            string
	CouponShareURL string
	Key            string
	ReservePrice   int64
}

func (s *Service) convertMyKey(ctx context.Context, title, picUrl string, adZoneID int64) (result convertMyKeyResult, err error) {

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
			URL                         string
			commissionRate              int64 = 1
			price, reservePrice, coupon float64
			highCommissionInfo          HighCommissionResult
			parseUrl                    *url.URL
			tpwdCreateResult            TbkTpwdCreateResult
		)
		result.Title = item.Title
		result.ItemID = item.ItemID

		s.logger.Info("商品信息", zap.String("标题", item.Title), zap.Int64("商品ID", item.ItemID), zap.String("一口价", item.ReservePrice), zap.String("折扣价", item.ZkFinalPrice), zap.String("佣金比率", item.CommissionRate))
		if price, err = strconv.ParseFloat(item.ZkFinalPrice, 64); err != nil {
			s.logger.Error("convertMyKey", zap.Error(err), zap.String("折扣价", item.ZkFinalPrice))
		}
		if reservePrice, err = strconv.ParseFloat(item.ReservePrice, 64); err != nil {
			s.logger.Error("convertMyKey", zap.Error(err), zap.String("一口价", item.ReservePrice))
		}
		if coupon, err = strconv.ParseFloat(item.CouponAmount, 64); err != nil {
			s.logger.Error("convertMyKey", zap.Error(err), zap.String("优惠券金额", item.CouponAmount))
		}
		commissionRate, err = strconv.ParseInt(item.CommissionRate, 10, 64)
		if err != nil {
			s.logger.Error("convertMyKey", zap.Error(err), zap.String("佣金比率", item.CommissionRate))
		}
		result.ShopName = item.Nick
		result.PicURL = item.PictURL
		result.ShopType = item.UserType
		result.ReservePrice = int64(reservePrice * 100)
		result.Price = int64(price * 100)
		// 商品价格-优惠券价格*佣金比率=预计收入佣金
		result.Rebate = int64((price-coupon)*float64(commissionRate)) / 100

		result.Coupon = int64(coupon * 100)
		result.URL = item.URL
		result.CouponShareURL = item.CouponShareURL
		if item.CouponShareURL != "" {
			URL = item.CouponShareURL
		} else {
			URL = item.URL
		}

		highCommissionInfo, err = s.HighCommission(ctx, item.ItemID)
		if err == nil {
			result.URL = highCommissionInfo.ItemURL
			result.CouponShareURL = highCommissionInfo.CouponClickURL
			//通知！通知！接到官方小二通知，非淘客链接不在支持生成淘口令，如果有需要请自己使用千牛App中的淘外推广进行创建口令
			if item.CouponShareURL != "" {
				URL = highCommissionInfo.CouponClickURL
			} else {
				URL = highCommissionInfo.ItemURL
			}
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
		result.Key = tpwdCreateResult.Model
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

func (s *Service) ClickKey(order *model.Order) {
	request, err := http.NewRequest(http.MethodGet, order.URL, nil)
	if err != nil {
		s.logger.Error("ClickKey 建立请求失败", zap.Error(err), zap.String("userID", order.UserID), zap.Int64("itemID", order.ItemID), zap.Int64("adzoneID", order.AdzoneID), zap.String("URL", order.URL))
	}
	if err := s.client.Do(context.Background(), request, nil); err != nil {
		dumpRequest, _ := httputil.DumpRequest(request, false)
		s.logger.Error("ClickKey DO请求失败", zap.Error(err), zap.Any("request", dumpRequest), zap.String("userID", order.UserID), zap.Int64("itemID", order.ItemID), zap.Int64("adzoneID", order.AdzoneID), zap.String("URL", order.URL))
	}
	clickTime := tools.Now()
	order.ClickTime = clickTime
	s.logger.Info("ClickKey成功", zap.String("userID", order.UserID), zap.Int64("itemID", order.ItemID), zap.String("本地点击时间", clickTime.String()), zap.Int64("adzoneID", order.AdzoneID), zap.String("URL", order.URL))
	if err := s.orders.Add(context.Background(), order); err != nil {
		s.logger.Error("ClickKey 添加至本地订单失败", zap.Error(err), zap.String("userID", order.UserID), zap.Int64("itemID", order.ItemID), zap.String("本地点击时间", clickTime.String()), zap.Int64("adzoneID", order.AdzoneID), zap.String("URL", order.URL))
	}
	return
}

func (s *Service) Convert(ctx context.Context, title string) (string, error) {
	adzoneID := s.GetdefultAdzoneID()
	result, err := s.execTbkDgMaterialOptional(ctx, TbkDgMaterialOptionalReq{
		AdzoneId: adzoneID,
		Q:        title,
		Sort:     "total_sales_des",
	})
	if err != nil {
		return "", err
	}
	for _, item := range result {

		URL := item.URL
		if item.CouponShareURL != "" {
			URL = item.CouponShareURL
		}
		if parseUrl, err := url.Parse(URL); err != nil {
			return "", err
		} else {
			parseUrl.Scheme = "https"
			URL = parseUrl.String()
		}

		if result, err := s.execTbkTpwdCreate(ctx, TbkTpwdCreateReq{
			Text: "啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊",
			URL:  URL,
		}); err != nil {
			return "", err
		} else {
			return result.Model, nil
		}
	}
	return "", nil
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
func (s *Service) GetTKL(ctx context.Context, title, picURL string, adZoneID int64) (tkl string, err error) {
	result, err := s.convertMyKey(ctx, title, picURL, adZoneID)
	if err != nil {
		err = fmt.Errorf("GetTKL failed: %w", err)
		return
	}
	tkl = result.Key
	return
}

func (s *Service) QueryTitleByItemID(ctx context.Context, itemID string) (title, picURL, shopName string, err error) {
	itemInfoGet, err := s.execTbkItemInfoGet(ctx, TbkItemInfoGetReq{
		NumIDs: itemID,
	})
	if err != nil {
		err = fmt.Errorf("QueryTitleByItemID failed: %w", err)
		return
	}
	title = itemInfoGet[0].Title
	picURL = itemInfoGet[0].PictURL
	shopName = itemInfoGet[0].Nick
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
func (s *Service) UnmatchGet(ctx context.Context, itemID, adZoneID int64) (*model.Order, error) {
	return s.dao.UnmatchGet(ctx, itemID, adZoneID)
}
func (s *Service) SetToUnmatch(ctx context.Context, itemID, adZoneID int64, order *model.Order) (ok bool, err error) {
	return s.dao.SetToUnmatch(ctx, itemID, adZoneID, order)
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

func OpenExcel() error {
	f, err := excelize.OpenFile("x")
	if err != nil {
		return err
	}
	// 获取工作表中指定单元格的值
	cell, err := f.GetCellValue("Page1", "B2")
	if err != nil {
		return err
	}
	fmt.Println(cell)
	// 获取 Sheet1 上所有单元格
	rows, err := f.GetRows("Page1")
	for _, row := range rows {
		for _, colCell := range row {
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}
	return nil
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

func (s *Service) GenGetadZoneID() func() int64 {
	var sliceStr []int64
	s.ac.Get("adzoneID").Slice(&sliceStr)
	i := len(sliceStr)
	return func() int64 {
		if i > 0 {
			i--
			return sliceStr[i]
		}
		return 0
	}
}

func (s *Service) GetdefultAdzoneID() int64 {
	var sliceStr []int64
	s.ac.Get("adzoneID").Slice(&sliceStr)
	if len(sliceStr) > 0 {
		return sliceStr[0]
	}
	return 0
}

func (s *Service) GettaokoulingAppKey() string {
	return paladin.String(s.ac.Get("taokoulingAppKey"), "")
}

func (s *Service) HighCommission(ctx context.Context, numIid int64) (result HighCommissionResult, err error) {
	adzoneID := s.GetdefultAdzoneID()
	appKey := s.GettaokoulingAppKey()
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
	appKey := s.GettaokoulingAppKey()
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
