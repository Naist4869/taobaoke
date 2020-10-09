package api

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/pkg/net/rpc/warden"

	"google.golang.org/grpc"
)

// AppID .
const AppID = "TODO: ADD APP ID"

// NewClient new grpc client
//  27.155.87.89
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (WechatClient, error) {
	client := warden.NewClient(cfg, opts...)
	cc, err := client.Dial(context.Background(), fmt.Sprintf("direct://default/wechat-svc:8001"))
	if err != nil {
		return nil, err
	}
	return NewWechatClient(cc), nil
}

// 生成 gRPC 代码
//go:generate kratos tool protoc --grpc --bm api.proto
