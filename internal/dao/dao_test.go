package dao

import (
	"context"
	"flag"
	"os"
	"taobaoke/internal/model"
	"taobaoke/tools"
	"testing"

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
	ok, err := d.SetNXToUnmatch(ctx, 123, 123, "444")
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = d.SetToUnmatch(ctx, 123, 123, &model.Order{}, "444")
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

func TestDao_SetToMatch(t *testing.T) {
	Convey("设置键值至匹配map", t, func() {
		_, err := d.DelFromUnmatchAndSetToMatch(ctx, &model.Order{})
		So(err, ShouldBeNil)
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
