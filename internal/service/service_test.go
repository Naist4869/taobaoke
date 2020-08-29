package service

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	pb "taobaoke/api"
	"taobaoke/internal/model"
	"taobaoke/tools"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/stretchr/testify/require"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/testing/lich"
)

var (
	testService *Service
	ctx         = context.Background()
)
var test2 = []byte(`(function(){
        var AsyncUrlUtils = new Object();
AsyncUrlUtils.loadUrl = function(src, redirect) {
        function callCountDown() {
                AsyncUrlUtils.countDown(redirect);
        }
        var img = document.createElement("img");
        img.onload = callCountDown;
        img.onerror = callCountDown;
        img.onabort = callCountDown;
        img.src = src;
}
AsyncUrlUtils.initCounter = function(initValue) {
        this.imgCounter = initValue;
}
AsyncUrlUtils.countDown = function(redirect) {
        this.imgCounter--;
        if (0 == this.imgCounter) {
                redirect();
        }
}
        function successHandler(){
                callback({"code":200,"data":{"st":"1Thvji3eLabwQLMljvNP23Q"}});
        }

        var asyncUrls = [];

        if (asyncUrls.length == 0) {
                successHandler();
                return;
        }

        setTimeout(successHandler, 500);
        AsyncUrlUtils.initCounter(asyncUrls.length);

        for (var i in asyncUrls) {
                AsyncUrlUtils.loadUrl(asyncUrls[i], successHandler);
        }
})();
`)

func TestMain(m *testing.M) {
	flag.Set("conf", "../../test")
	//flag.Set("f", "../../test/docker-compose.yaml")
	flag.Parse()
	os.Setenv("DISABLE_LICH", "true")
	disableLich := os.Getenv("DISABLE_LICH") != ""
	if !disableLich {
		if err := lich.Setup(); err != nil {
			panic(err)
		}
	}
	var err error
	if err = paladin.Init(); err != nil {
		panic(err)
	}
	var cf func()
	if testService, cf, err = newTestService(); err != nil {
		panic(err)
	}

	ret := m.Run()
	if cf != nil {
		cf()
	}

	if !disableLich {
		_ = lich.Teardown()
	}
	os.Exit(ret)
}

func TestService_ItemInfoGet(t *testing.T) {
	result, err := testService.execTbkItemInfoGet(ctx, TbkItemInfoGetReq{
		NumIDs:   "603587505846",
		Platform: 2,
	})
	require.NoError(t, err)
	spew.Dump(result)
}

//  https://oauth.taobao.com/authorize?response_type=code&client_id=30055875&redirect_uri=http://127.0.0.1:12345/error&state=1212&view=web

// https://mos.m.taobao.com/inviter/register?inviterCode=7DXRK3&src=pub&app=common&rtag=成
func TestService_TbkTpwdCreate(t *testing.T) {
	result, err := testService.execTbkTpwdCreate(ctx, TbkTpwdCreateReq{
		Text: "哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈哈",
		URL:  "https://mos.m.taobao.com/inviter/register?inviterCode=7DXRK3&src=pub&app=common&rtag=cheng",
	})
	require.NoError(t, err)
	spew.Dump(result)
}
func TestService_TbkDgMaterialOptional(t *testing.T) {

	testService.execTbkDgMaterialOptional(ctx, TbkDgMaterialOptionalReq{
		AdzoneId: 110790300374,
		Q:        "华为5G CPE Pro 无线路由器千兆端口双宽带插卡5G全网通随身WiFi\n",
	})
}

func TestOpenXLS(t *testing.T) {
	OpenXLS()
}

func TestService_HighCommission(t *testing.T) {
	result, err := testService.HighCommission(ctx, 608813238220)
	require.NoError(t, err)
	spew.Dump(result)
}

func TestService_QueryOrder(t *testing.T) {
	parseTime, err := tools.ParseTimeInLength("2020-06-27 12:14:32")
	if err != nil {
		t.Fatal(err)
	}

	result, err := testService.execTbkOrderDetailsGet(ctx, TbkOrderDetailsGetReq{
		QueryType: 2,
		StartTime: parseTime,
		EndTime:   parseTime,
	})
	require.NoError(t, err)
	t.Logf("%#v", result)
}
func TestTbkScInvitecodeGet(t *testing.T) {
	result, err := testService.execTbkScInvitecodeGet(ctx, TbkScInvitecodeReq{
		RelationApp: "common",
		CodeType:    1,
	})
	require.NoError(t, err)
	// 7DXRK3
	t.Logf("%#v", result)
}

func TestTbkScPublisherInfoSave(t *testing.T) {
	result, err := testService.execTbkScPublisherInfoSave(ctx, TbkScPublisherInfoSaveReq{
		InviterCode: "7DXRK3",
		InfoType:    1,
	})
	require.NoError(t, err)
	t.Logf("%#v", result)
}

func TestTbkScPublisherInfoGet(t *testing.T) {
	result, err := testService.execTbkScPublisherInfoGet(ctx, TbkScPublisherInfoGetReq{
		InfoType:    1,
		RelationApp: "common",
	})
	require.NoError(t, err)
	t.Logf("%#v", result)
}

func TestService_analyzingKey(t *testing.T) {
	resp, err := testService.analyzingKey(ctx, "$nniWccD1nru$")
	require.NoError(t, err)
	spew.Dump(resp)

}

func TestService_KeyConvert(t *testing.T) {
	convertKey, err := testService.KeyConvert(ctx, &pb.KeyConvertReq{
		FromKey: `₤QHUEcd1NPHR₤`,
		UserID:  "1",
	})
	require.NoError(t, err)
	t.Logf("%v", convertKey)
}

func TestRegexp2(t *testing.T) {
	re := regexp.MustCompile(`"data":{"st":"(.[^\"]*)`)
	matches := re.FindSubmatch(test2)
	fmt.Printf("%s", matches[1])
}
func TestOrders_TemplateMsgSend(t *testing.T) {
	testService.orders.TemplateMsgSend(ctx, &pb.TemplateMsgSendReq{
		UserID:           "oqeBd0fGbtYTmoVGhHzZ5Nf3-Egc",
		OrderID:          "043HR",
		Title:            "Julia编程基础",
		PaidTime:         tools.Now().String(),
		AlipayTotalPrice: "12.8",
		Rebate:           "0.12",
	})
}

// -race 测试通过
func TestService_WithDraw(t *testing.T) {
	text := `{
    "orders": [
        {
            "paid_time": "2020-06-30 14:44:15",
            "trade_parent_id": "1093949954060945186"
        },
        {
            "paid_time": "2020-06-29 13:45:39",
            "trade_parent_id": "1092126338847747471"
        },
        {
            "paid_time": "2020-06-29 13:40:11",
            "trade_parent_id": "1091523488582747471"
        },
        {
            "paid_time": "2020-06-27 12:14:32",
            "trade_parent_id": "1088049344508740781"
        },
        {
            "paid_time": "2020-06-16 11:02:58",
            "trade_parent_id": "1065911329613747471"
        },
        {
            "paid_time": "2020-06-11 09:19:47",
            "trade_parent_id": "1054723971619972844"
        },
        {
            "paid_time": "2020-06-11 07:34:42",
            "trade_parent_id": "1053946560561568653"
        },
        {
            "paid_time": "2020-06-10 04:21:44",
            "trade_parent_id": "1052810275721747471"
        }
    ]
}
`
	v := new(struct {
		Orders []struct {
			PaidTime      tools.Time `json:"paid_time"`
			TradeParentID string     `json:"trade_parent_id"`
		} `json:"orders"`
	})
	err := json.Unmarshal([]byte(text), v)
	if err != nil {
		t.Fatal(err)
	}
	var orders []*model.Order
	for _, item := range v.Orders {
		orders = append(orders, &model.Order{
			PaidTime:      item.PaidTime,
			TradeParentID: item.TradeParentID,
		})
	}

	results, err := testService.QueryRemoteOrderByTradeParentID(ctx, orders)
	if err != nil {
		t.Fatal(err)
	}
	results.Range(func(key, value interface{}) bool {
		t.Logf("%s: %v", key.(string), value.(TbkOrderDetailsGetResult))
		return true
	})

}
func TestOrders_String(t *testing.T) {
	s := testService.orders.String()
	t.Log(s)
}
