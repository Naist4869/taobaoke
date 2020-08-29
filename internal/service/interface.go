package service

import (
	"context"
	pb "taobaoke/api"
	"taobaoke/internal/model"
)

type Server interface {
	pb.TBKServer
	TaobaokeService
}

// 淘宝客服务
type TaobaokeService interface {
	GetServerAddr() string
	PriceTrend(ctx context.Context, itemID int64) (trendInfo model.TrendInfo, err error)
	UnmatchGet(ctx context.Context, itemID, adZoneID int64) (*model.Order, error)
	UpdateToUnmatch(ctx context.Context, itemID, adZoneID int64, order *model.Order) (ok bool, err error)
	GetTklByItemID(ctx context.Context, itemID int64, adZoneID int64, title string) (tkl string, URL, CouponShareURL string, err error)
}
