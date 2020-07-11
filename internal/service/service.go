package service

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"taobaoke/common"
	"taobaoke/internal/model"
	"time"

	"github.com/teris-io/shortid"

	"go.uber.org/zap"

	"github.com/Naist4869/log"

	"github.com/extrame/xls"

	"github.com/360EntSecGroup-Skylar/excelize/v2"

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

func (s *Service) KeyConvertKey(ctx context.Context, req *pb.KeyConvertKeyReq) (resp *pb.KeyConvertKeyResp, err error) {
	var id, key string
	id, err = s.idGenerator.Generate()
	if err != nil {
		return
	}
	c := context.WithValue(ctx, "order", model.NewOrder(id, req.UserID))
	key, err = s.keyConvertKey(c, req.FromKey)
	if err != nil {
		return
	}
	order := c.Value("order").(*model.Order)
	resp = &pb.KeyConvertKeyResp{
		ToKey:  key,
		Price:  strconv.FormatInt(order.Price, 10),
		Rebate: strconv.FormatInt(order.Rebate, 10),
		Title:  order.Title,
		PicURL: order.PicURL,
	}
	s.logger.Info("保存订单信息", zap.String("标题", order.Title), zap.Int64("商品ID", order.ItemID), zap.Int64("价格", order.Price), zap.String("淘口令", order.Key))
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

func (s *Service) methodPost(ctx context.Context, req Request, resp Response, method string) (err error) {
	var uri, appKey string
	uri, err = s.ac.Get("uri").String()
	if err != nil {
		return
	}
	query := url.Values{}
	appKey, err = s.ac.Get("appKey").String()
	if err != nil {
		return
	}
	queryMap := map[string]string{
		"method":      method,
		"app_key":     appKey,
		"format":      "json",
		"sign_method": "md5",
		"v":           "2.0",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		//"simplify":    "true",
	}
	req.Query(queryMap)
	param := make([]string, 0, 20)
	for key, value := range queryMap {
		query.Set(key, value)
		param = append(param, key+value)
	}

	sign := Sign(param...)
	query.Set("sign", sign)
	if err = s.client.Post(ctx, uri, "", query, resp); err != nil {
		return
	}
	if resp.Error() != nil {
		err = resp.Error()
	}
	return
}

// 淘口令转高佣淘口令
func (s *Service) keyConvertKey(ctx context.Context, fromKey string) (string, error) {
	var Title, PicURL string
	if keyInfo, err := s.analyzingKey(ctx, fromKey); err != nil {
		return "", err
	} else {
		Title = keyInfo.Content
		PicURL = keyInfo.PicURL
	}
	if key, err := s.convertMyKey(ctx, Title, PicURL); err != nil {
		return "", err
	} else {
		return key, err
	}

}

func (s *Service) convertMyKey(ctx context.Context, title, picUrl string) (string, error) {
	order := ctx.Value("order").(*model.Order)
	var adzoneID int64
	if id, err := s.ac.Get("adzoneID").Int64(); err != nil {
		return "", err
	} else {
		adzoneID = id
	}

	result, err := s.execTbkDgMaterialOptional(ctx, TbkDgMaterialOptionalReq{
		AdzoneId: int(adzoneID),
		Q:        title,
		Sort:     "total_sales_des",
	})
	if err != nil {
		return "", err
	}
	order.AdzoneID = adzoneID
	if len(result) == 0 {
		return "", errors.New("搜索列表为空")
	}
	for _, item := range result {
		index := strings.Index(item.PictURL, "uploaded")
		if !strings.Contains(picUrl, item.PictURL[index+12:]) {
			continue
		}
		var (
			URL            string
			price          float64
			commissionRate int64 = 1
		)
		order.Title = item.Title
		order.ItemID = item.NumIid // 文档里说要废弃

		s.logger.Info("商品信息", zap.String("标题", item.Title), zap.Int64("商品ID", item.NumIid), zap.String("一口价", item.ReservePrice), zap.String("折扣价", item.ZkFinalPrice), zap.String("佣金比率", item.CommissionRate))
		if p, err := strconv.ParseFloat(item.ZkFinalPrice, 64); err != nil {
			s.logger.Error("convertMyKey", zap.Error(err), zap.String("折扣价", item.ZkFinalPrice))
			if p, err := strconv.ParseFloat(item.ReservePrice, 64); err != nil {
				s.logger.Error("convertMyKey", zap.Error(err), zap.String("一口价", item.ReservePrice))
			} else {
				price = p
			}
		} else {
			price = p
		}
		if p, err := strconv.ParseInt(item.CommissionRate, 10, 64); err != nil {
			s.logger.Error("convertMyKey", zap.Error(err), zap.String("佣金比率", item.CommissionRate))
		} else {
			commissionRate = p
		}
		order.PicURL = item.PictURL
		order.Price = int64(price * 100)
		order.Rebate = int64(price*float64(commissionRate)) / 100
		if highCommissionInfo, err := s.HighCommission(ctx, item.NumIid); err != nil {
			s.logger.Error("convertMyKey 转换高佣失败", zap.Error(err), zap.Int64("itemID", item.NumIid))
			order.URL = item.URL
			order.CouponShareURL = item.CouponShareURL
			if item.CouponShareURL != "" {
				URL = item.CouponShareURL
			} else {
				URL = item.URL
			}
		} else {
			order.URL = highCommissionInfo.ItemURL
			order.CouponShareURL = highCommissionInfo.CouponClickURL
			if item.CouponShareURL != "" {
				URL = highCommissionInfo.CouponClickURL
			} else {
				URL = highCommissionInfo.ItemURL
			}
		}

		if parseUrl, err := url.Parse(URL); err != nil {
			return "", err
		} else {
			parseUrl.Scheme = "https"
			URL = parseUrl.String()
		}

		if result, err := s.execTbkTpwdCreate(ctx, TbkTpwdCreateReq{
			Text: "啊实打实的撒大苏打萨达萨达萨达是的观点",
			URL:  URL,
		}); err != nil {
			return "", err
		} else {
			order.Key = result.Model
			return result.Model, nil
		}
	}
	return "", errors.New("目标商品未找到")
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
	clickTime := time.Now()
	order.ClickTime = clickTime
	s.logger.Info("ClickKey成功", zap.String("userID", order.UserID), zap.Int64("itemID", order.ItemID), zap.Time("本地点击时间", clickTime), zap.Int64("adzoneID", order.AdzoneID), zap.String("URL", order.URL))
	if err := s.orders.Add(order); err != nil {
		s.logger.Error("ClickKey 添加至本地订单失败", zap.Error(err), zap.String("userID", order.UserID), zap.Int64("itemID", order.ItemID), zap.Time("本地点击时间", clickTime), zap.Int64("adzoneID", order.AdzoneID), zap.String("URL", order.URL))
	}
	return
}
func (s *Service) Convert(ctx context.Context, title string) (string, error) {
	var adzoneID int64
	if id, err := s.ac.Get("adzoneID").Int64(); err != nil {
		return "", err
	} else {
		adzoneID = id
	}
	result, err := s.execTbkDgMaterialOptional(ctx, TbkDgMaterialOptionalReq{
		AdzoneId: int(adzoneID),
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
func Sign(strs ...string) (signature string) {
	sort.Strings(strs)
	tmpstr := strings.Join(strs, "")
	str := "a55fe6c77a947a439079640ecbcdc817" + tmpstr + "a55fe6c77a947a439079640ecbcdc817"
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

func (s *Service) HighCommission(ctx context.Context, numIid int64) (result HighCommissionResult, err error) {
	var (
		appKey   string
		adzoneID int64
	)

	if adzoneID, err = s.ac.Get("adzoneID").Int64(); err != nil {
		return
	}
	if appKey, err = s.ac.Get("taokoulingAppKey").String(); err != nil {
		return
	}
	itemID := strconv.FormatInt(numIid, 10)
	param := url.Values{}
	param.Set("apikey", appKey)
	param.Set("itemid", itemID)
	param.Set("siteid", "43474861")
	param.Set("adzoneid", strconv.FormatInt(adzoneID, 10))
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
	EazyCopyStruct(resp, &result)
	return
}
func (s *Service) analyzingKey(ctx context.Context, fromKey string) (resp analyzingKeyResp, err error) {
	param := url.Values{}
	var appKey string
	if appKey, err = s.ac.Get("taokoulingAppKey").String(); err != nil {
		return
	}
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
	return
}

//func (s *Service) QueryOrder(ctx context.Context) (result []PublisherOrderDto, err error) {
//	param := url.Values{}
//	param.Set("uid", "2329747174")
//	param.Set("query_type", "1") //  查询时间类型，1：按照订单淘客创建时间查询，2:按照订单淘客付款时间查询，3:按照订单淘客结算时间查询 不传为1
//	//param.Set("position_index", "") // 位点，除第一页之外，都需要传递；前端原样返回。不传为2222_334666
//	//param.Set("page_size", "1")     // 页大小，默认20，1~100	不传为20
//	//param.Set("member_type", "")    //推广者角色类型,2:二方，3:三方，不传，表示所有角色
//	//param.Set("tk_status", "")      // 淘客订单状态，12-付款，13-关闭，14-确认收货，15-结算成功;不传，表示所有状态
//	param.Set("end_time", "2020-06-11 07:39:59")   // 2019-04-23 12:28:22	订单查询结束时间	必填
//	param.Set("start_time", "2020-06-11 07:20:00") //2019-04-05 12:18:22	订单查询开始时间	必填
//	//param.Set("jump_typ", "")       //跳转类型，当向前或者向后翻页必须提供,-1: 向前翻页,1：向后翻页
//	//param.Set("page_no", "")        //几页，默认1，1~100，不传为第一页
//	//param.Set("order_scen", "")     //场景订单场景类型，1:常规订单，2:渠道订单，3:会员运营订单，默认为1
//	var request *http.Request
//	request, err = s.client.NewRequest(http.MethodGet, "https://api.taokouling.com/tbk/TbkScOrderDetailsGet", "", param)
//	if err != nil {
//		return
//	}
//	var resp tbkScOrderDetailsResp
//	if err = s.client.JSON(ctx, request, &resp); err != nil {
//		return
//	}
//	EazyCopyStruct(resp, &result)
//	return
//}
