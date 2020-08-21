package dao

import (
	"context"
	"flag"
	"os"
	"taobaoke/internal/model"
	"testing"

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
