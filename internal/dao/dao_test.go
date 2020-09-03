package dao

import (
	"context"
	"flag"
	"fmt"
	"os"
	"reflect"
	pb "taobaoke/api"
	"taobaoke/internal/model"
	"taobaoke/tools"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/davecgh/go-spew/spew"

	"github.com/stretchr/testify/require"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/testing/lich"
)

var d *dao
var ctx = context.Background()

func TestMain(m *testing.M) {
	flag.Set("conf", "../../test")
	flag.Set("f", "../../test/docker-compose.yaml")
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
	if d, cf, err = newTestDao(); err != nil {
		panic(err)
	}
	ret := m.Run()
	cf()
	if !disableLich {
		_ = lich.Teardown()
	}
	os.Exit(ret)
}
func TestOrderClient_Insert(t *testing.T) {
	err := d.Insert(ctx, &model.Order{ID: "123", UserID: "123"})
	require.NoError(t, err)
}

func TestDao_SetNXToUnmatch(t *testing.T) {
	ok, err := d.SetNXToUnmatch(ctx, 123, 123, "123")
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = d.SetNXToUnmatch(ctx, 123, 123, "123")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestDao_QueryOrderByTradeParentID(t *testing.T) {
	err := d.Insert(ctx, &model.Order{ID: "123", UserID: "123", TradeParentID: "123"})
	require.NoError(t, err)
	orders, err := d.QueryOrderByTradeParentID(ctx, []string{"123", "123", "12", "1", ""}, true)
	require.NoError(t, err)
	spew.Dump(orders)
}

func TestDao_SetToUnmatch(t *testing.T) {
	//ok, err := d.SetNXToUnmatch(ctx, 123, 123, "444")
	//require.NoError(t, err)
	//require.True(t, ok)
	//time.Sleep(time.Second * 6)
	ok, err := d.SetToUnmatch(ctx, 123, 123, &model.Order{}, "444")
	require.NoError(t, err)
	require.True(t, ok)
}
func TestDao_UpdateToUnmatch(t *testing.T) {
	Convey("更新redis中未匹配map的值", t, func() {
		ok, err := d.UpdateFromUnmatch(ctx, 123, 123, &model.Order{ID: "1"})
		So(ok, ShouldBeTrue)
		So(err, ShouldBeNil)
	})
}
func TestDao_UnmatchGet(t *testing.T) {
	Convey("获取redis中未匹配map的值", t, func() {
		order, err := d.GetUnmatch(ctx, 123, 123)
		So(err, ShouldBeNil)
		So(order, ShouldNotBeNil)
		t.Logf("%#v", order)
	})
}
func TestDao_UnmatchGetAll(t *testing.T) {
	all, err := d.GetAllUnmatch(ctx)
	require.NoError(t, err)
	spew.Dump(all)
}

func TestDao_MatchGetAll(t *testing.T) {
	Convey("获取redis中匹配map全部数据", t, func() {
		all, err := d.MatchGetAll(ctx)
		So(err, ShouldBeNil)
		So(all, ShouldNotBeNil)
		for _, order := range all {
			t.Log(order)
		}
	})

}
func TestDao_FindOrderByID(t *testing.T) {
	Convey("从DB中获取Order", t, func() {
		order, err := d.FindOrderByID(ctx, "rd2NH6rYM")
		So(err, ShouldBeNil)
		So(order, ShouldNotBeNil)
	})
}

func TestDao_DelFromUnmatchAndSetToMatch(t *testing.T) {
	Convey("从未匹配队列中删除并更新到已匹配队列", t, func() {
		order, err := d.FindOrderByID(ctx, "rd2NH6rYM")
		So(err, ShouldBeNil)
		So(order, ShouldNotBeNil)
		ok, err := d.DelFromUnmatchAndSetToMatch(ctx, order)
		So(err, ShouldBeNil)
		So(ok, ShouldBeTrue)

	})
}
func TestDao_QueryOrderByStatus(t *testing.T) {
	Convey("获取45天内DB中所有未完成的订单", t, func() {
		now := tools.Now()
		unfinishOrders, err := d.QueryOrderByStatus(context.Background(), now.Add(-time.Hour*24*45), now, model.OrderPaid, model.OrderFinish)
		So(err, ShouldBeNil)
		for _, order := range unfinishOrders {
			_, err = d.HSetNXToMatch(context.Background(), order)
			So(err, ShouldBeNil)
		}
	})
}

func TestDao_Insert(t *testing.T) {
	parseTime, err := tools.ParseTimeInLength("2020-06-27 12:14:32")
	if err != nil {
		t.Fatal(err)
	}
	err = d.Insert(ctx, &model.Order{
		ID:       "1",
		UserID:   "oqeBd0fGbtYTmoVGhHzZ5Nf3-Egc",
		PaidTime: parseTime,
		Deleted:  false,
	})
	if err != nil {
		t.Fatal(err)
	}
}
func Test_OrderMap(t *testing.T) {
	var statues = []model.OrderStatus{model.OrderCreate, model.OrderPaid, model.OrderFinish, model.OrderBalance}
	var fun HandlerFunc
	var statusesMap = tools.NewOrderedMap(tools.NewKeys(func(i interface{}, j interface{}) int8 {
		if i.(model.OrderStatus) == j.(model.OrderStatus) {
			return 0
		}
		var Ifinded, Jfinded int8
		for _, status := range statues {
			// 先找到的肯定比后找到的小
			switch status {
			case i.(model.OrderStatus):
				Ifinded += 1
			case j.(model.OrderStatus):
				Jfinded += 1
			default:
				continue
			}
			return Jfinded - Ifinded
		}
		return 0
	}, reflect.TypeOf(model.OrderCreate)), reflect.TypeOf(fun))
	statusesMap.Put(model.OrderFinish, HandlerFunc(func(c *Context) {
		fmt.Println("1")
	}))
	statusesMap.Put(model.OrderPaid, HandlerFunc(func(c *Context) {
		fmt.Println("2")
	}))
	statusesMap.Put(model.OrderBalance, HandlerFunc(func(c *Context) {
		fmt.Println("4")
	}))
	statusesMap.Put(model.OrderCreate, HandlerFunc(func(c *Context) {
		fmt.Println("1")
	}))
	t.Logf("%s", statusesMap)
}
func Test_UpdateStatus(t *testing.T) {
	d.UpdateStatus(ctx, &model.Order{
		Status: model.OrderIllegal,
	}, &model.XLSOrder{
		TkStatus: "13",
	}, 90)
}
func TestDao_BalanceTemplateMsgSend(t *testing.T) {
	d.BalanceTemplateMsgSend(ctx, &pb.BalanceTemplateMsgSendReq{
		UserID:      "oqeBd0fGbtYTmoVGhHzZ5Nf3-Egc",
		OrderID:     "abcabc",
		Title:       "asdasdASDASDasd",
		EarningTime: tools.Now().String(),
		Salary:      "12.34",
		Balance:     "16.32",
	})
}
func TestDao_QueryNotWithDrawOrderByUserID(t *testing.T) {
	Convey("查询当前用户所有未提现的单", t, func() {
		orders, err := d.QueryNotWithDrawOrderByUserID(ctx, "oqeBd0fGbtYTmoVGhHzZ5Nf3-Egc")
		So(err, ShouldBeNil)
		for _, order := range orders {
			t.Logf("%v", order)
		}
	})

}
