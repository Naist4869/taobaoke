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
	PriceTrend(ctx context.Context, itemID string) (trendInfo model.TrendInfo, err error)
	QueryTitleByItemID(ctx context.Context, itemID string) (title, picURL, shopName string, err error)
	GetTKL(ctx context.Context, title, picURL, itemID string) (tkl string, err error)
}
