package service

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	pb "taobaoke/api"
	"taobaoke/tools"
	"testing"
	"time"

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
		NumIDs: "603587505846",
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
	adzoneID := testService.GetadzoneID()
	testService.execTbkDgMaterialOptional(ctx, TbkDgMaterialOptionalReq{
		AdzoneId: int(adzoneID),
		Q:        "华为5G CPE Pro 无线路由器千兆端口双宽带插卡5G全网通随身WiFi\n",
	})
}

func TestService_Convert(t *testing.T) {
	type args struct {
		ctx   context.Context
		title string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "标题转成淘口令",
			args: args{
				ctx:   ctx,
				title: "吉列威锋双层剃须刀手动刮胡刀老式刮脸刀吉利刀头刀架套装",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testService.Convert(tt.args.ctx, tt.args.title)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Convert() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpenExcel(t *testing.T) {
	err := OpenExcel()
	require.NoError(t, err)
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
	now := tools.Now()
	result, err := testService.execTbkOrderDetailsGet(ctx, TbkOrderDetailsGetReq{
		StartTime: now.Add(-time.Hour),
		EndTime:   now,
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

func TestService_ConvertMyKey(t *testing.T) {
	key, err := testService.keyConvertKey(ctx, "₴amEKcX3FfiF₴")
	require.NoError(t, err)
	spew.Dump(key)
}
func TestService_KeyConvertKey(t *testing.T) {
	convertKey, err := testService.KeyConvertKey(ctx, &pb.KeyConvertKeyReq{
		FromKey: `$nniWccD1nru$`,
		UserID:  "1",
	})
	require.NoError(t, err)
	t.Logf("%s", convertKey.ToKey)
	time.Sleep(5 * time.Second)
}

func TestRegexp2(t *testing.T) {
	re := regexp.MustCompile(`"data":{"st":"(.[^\"]*)`)
	matches := re.FindSubmatch(test2)
	fmt.Printf("%s", matches[1])
}
