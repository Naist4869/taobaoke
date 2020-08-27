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
	GetItem(itemID string)
	PriceTrend(ctx context.Context, itemID int64) (trendInfo model.TrendInfo, err error)
	QueryTitleByItemID(ctx context.Context, itemID string) (title, picURL, shopName string, err error)
	UnmatchGet(ctx context.Context, itemID, adZoneID int64) (*model.Order, error)
	SetToUnmatch(ctx context.Context, itemID, adZoneID int64, order *model.Order) (ok bool, err error)
	GetTKL(ctx context.Context, title, picURL string, adZoneID int64) (tkl string, err error)
}
