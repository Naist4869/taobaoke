package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"taobaoke/internal/model"
	"time"

	"go.uber.org/zap"

	xtime "github.com/go-kratos/kratos/pkg/time"

	"github.com/go-redis/redis/v8"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/log"
)

type Fieldfun func() string

const (
	// id -> order
	_orderMap = "order_map"
	// tradeParentID -> order
	_matchMap = "match_map"
	_matchSet = "match_set"
	_unmatch  = "unmatch:%d.%d"
	_tkl      = "tkl:%s"
	redisOK   = "OK"
)

type RedisConfig struct {
	Name string // redis name, for trace
	// The network type, either tcp or unix.
	// Default is tcp.
	Network string
	// host:port address.
	Addr string
	// Optional password. Must match the password specified in the
	// requirepass server configuration option (if connecting to a Redis 5.0 instance, or lower),
	// or the User Password when connecting to a Redis 6.0 instance, or greater,
	// that is using the Redis ACL system.
	Password string
	// Database to be selected after connecting to the server.
	DB int
	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout xtime.Duration
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout xtime.Duration
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout xtime.Duration
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout xtime.Duration
	// Frequency of idle checks made by idle connections reaper.
	// Default is 1 minute. -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency xtime.Duration

	// Cluster
	// A seed list of host:port addresses of cluster nodes.
	Addrs []string
}

func NewRedis() (r *redis.ClusterClient, cf func(), err error) {
	var (
		cfg RedisConfig
		ct  paladin.Map
	)
	if err = paladin.Get("redis.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return
	}

	//r = redis.NewClient(&redis.Options{
	//	Network:            cfg.Network,
	//	Addr:               cfg.Addr,
	//	Password:           cfg.Password,
	//	DB:                 cfg.DB,
	//	DialTimeout:        time.Duration(cfg.DialTimeout),
	//	ReadTimeout:        time.Duration(cfg.ReadTimeout),
	//	WriteTimeout:       time.Duration(cfg.WriteTimeout),
	//	IdleTimeout:        time.Duration(cfg.IdleTimeout),
	//	IdleCheckFrequency: time.Duration(cfg.IdleCheckFrequency),
	//})
	//cf = func() { r.Close() }

	r = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:              cfg.Addrs,
		Password:           cfg.Password,
		DialTimeout:        time.Duration(cfg.DialTimeout),
		ReadTimeout:        time.Duration(cfg.ReadTimeout),
		WriteTimeout:       time.Duration(cfg.WriteTimeout),
		IdleTimeout:        time.Duration(cfg.IdleTimeout),
		IdleCheckFrequency: time.Duration(cfg.IdleCheckFrequency),
	})
	cf = func() { r.Close() }

	return
}

func (d *dao) PingRedis(ctx context.Context) (err error) {
	return d.redis.Ping(ctx).Err()
}

func unmatchKey(itemID int64, adZoneID int64) string {
	return fmt.Sprintf(_unmatch, itemID, adZoneID)
}
func (d *dao) SetNXToUnmatch(ctx context.Context, itemID, adZoneID int64, nonce string) (ok bool, err error) {
	//// https://github.com/redis/redis/issues/167 在field上设置过期时间在2020年还未实现...
	//if err = conn.Send("EXPIRE", _unmatch, 24*60*60); err != nil {
	//	log.Error("conn.Send(EXPIRE, %s, %d) error(%+v)", _unmatch, 24*60*60, err)
	//	return
	//}
	key := unmatchKey(itemID, adZoneID)
	// 保存5秒
	ok, err = d.redis.SetNX(ctx, key, nonce, time.Second*5).Result()
	if err != nil {
		log.Error("conn.Do(SetNX, %s, %s) error(%v)", key, nonce, err)
	}
	return
}
func (d *dao) SetToUnmatch(ctx context.Context, itemID, adZoneID int64, order *model.Order, nonce string) (ok bool, err error) {
	key := unmatchKey(itemID, adZoneID)
	defer func() {
		if err != nil {
			d.redis.Del(ctx, key)
		}
	}()
	marshal, err := json.Marshal(order)
	if err != nil {
		return
	}
	old, err := d.redis.GetSet(ctx, key, marshal).Result()
	if err != nil {
		err = fmt.Errorf("conn.Do(GetSet, %s, %s) error(%v)", key, marshal, err)
		return
	}

	if strings.Compare(old, nonce) != 0 {
		err = fmt.Errorf("nonce不匹配,old: %s,should: %s", old, nonce)
		return
	}

	result, err := d.redis.Expire(ctx, key, time.Hour*24*10).Result()
	if err != nil {
		err = fmt.Errorf("conn.Do(Expire, %s, %s) error(%v)", key, marshal, err)
		return
	}
	return result, nil

}
func (d *dao) UpdateFromUnmatch(ctx context.Context, itemID, adZoneID int64, order *model.Order) (ok bool, err error) {
	key := unmatchKey(itemID, adZoneID)
	marshal, err := json.Marshal(order)
	if err != nil {
		return
	}
	result, err := d.redis.Set(ctx, key, marshal, 0).Result()
	if err != nil {
		err = fmt.Errorf("conn.Do(Set, %s, %s) error(%v)", key, marshal, err)
		return
	}
	ok = result == redisOK
	return
}

func (d *dao) ExistInUnmatch(ctx context.Context, itemID, adZoneID int64) (exist bool, err error) {
	key := unmatchKey(itemID, adZoneID)
	result, err := d.redis.Exists(ctx, key).Result()
	if err != nil {
		err = fmt.Errorf("conn.Exists(Set, %s) error(%v)", key, err)
		return
	}
	if result == 1 {
		exist = true
		return
	}
	return
}
func (d *dao) GetUnmatch(ctx context.Context, itemID, adZoneID int64) (*model.Order, error) {
	key := unmatchKey(itemID, adZoneID)
	result, err := d.redis.Get(ctx, key).Result()
	if err != nil {
		err = fmt.Errorf("conn.Do(Get, %s) error(%v)", key, err)
		return nil, err
	}
	if result == "" {
		err = fmt.Errorf("GetUnmatch not found order")
		return nil, err
	}
	v := &model.Order{}
	if err = json.Unmarshal([]byte(result), v); err != nil {
		err = fmt.Errorf("GetUnmatch Unmarshal error: (%w),data: %s", err, result)
		return nil, err
	}
	return v, nil
}
func (d *dao) GetAllUnmatch(ctx context.Context) (map[string]*model.Order, error) {
	index := strings.Index(_unmatch, ":")
	iter := d.redis.Scan(ctx, 0, _unmatch[:index+1]+"*", 0).Iterator()
	all := map[string]*model.Order{}
	kSlice := make([]string, 0)
	for iter.Next(ctx) {
		kSlice = append(kSlice, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	pipe := d.redis.Pipeline()
	stringCmdSlice := make([]*redis.StringCmd, len(kSlice))
	for i, key := range kSlice {
		stringCmdSlice[i] = pipe.Get(ctx, key)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		d.logger.Error("GetAllUnmatch", zap.Error(err))
		return nil, err
	}
	for i, stringCmd := range stringCmdSlice {
		if stringCmd.Err() != redis.Nil {
			if jsonByte, err := stringCmd.Bytes(); err != nil {
				continue
			} else {
				v := &model.Order{}
				if err := json.Unmarshal(jsonByte, v); err != nil {
					d.logger.Error("GetAllUnmatch", zap.Error(err), zap.ByteString("值", jsonByte))
					continue
				}
				if _, exist := all[kSlice[i]]; !exist {
					all[kSlice[i]] = v
				}
			}

		}
	}

	// https://github.com/go-redis/redis/issues/291  用cluster 的mget会出现CROSSSLOT Keys in request don't hash to the same slot错误
	//vSlice, err := d.redis.MGet(ctx, kSlice...).Result()
	//if err != nil {
	//	return nil, err
	//}
	//for i, value := range vSlice {
	//	if value == nil {
	//		continue
	//	}
	//	v := &model.Order{}
	//	if err := json.Unmarshal([]byte(value.(string)), v); err != nil {
	//		d.logger.Error("GetAllUnmatch", zap.Error(err), zap.String("值", iter.Val()))
	//		continue
	//	}
	//	if _, exist := all[kSlice[i]]; !exist {
	//		all[kSlice[i]] = v
	//	}
	//
	//}
	return all, nil
}

//func (d *dao) MatchGetAll(ctx context.Context) ([]*model.Order, error) {
//	iter := d.redis.HScan(ctx, _matchMap, 0, "", 0).Iterator()
//	all := make([]*model.Order, 0)
//	kvSlice := make([]string, 0)
//	for iter.Next(ctx) {
//		kvSlice = append(kvSlice, iter.Val())
//	}
//	if err := iter.Err(); err != nil {
//		return nil, err
//	}
//	for i := 1; i < len(kvSlice); i += 2 {
//		if kvSlice[i] == cacheNull {
//			continue
//		}
//		v := &model.Order{}
//		if err := json.Unmarshal([]byte(kvSlice[i]), v); err != nil {
//			d.logger.Error("MatchGetAll", zap.Error(err), zap.String("值", iter.Val()))
//			continue
//		}
//		all = append(all, v)
//	}
//	return all, nil
//}

func (d *dao) MatchGetAll(ctx context.Context) ([]*model.Order, error) {
	var tradeParentIDs []string
	iter := d.redis.SScan(ctx, _matchSet, 0, "", 0).Iterator()
	for iter.Next(ctx) {
		tradeParentIDs = append(tradeParentIDs, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return d.QueryOrderByTradeParentID(ctx, tradeParentIDs, true)
}

func (d *dao) DelFromMatchCache(ctx context.Context, tradeParentIDs ...string) (n int64, err error) {
	if len(tradeParentIDs) == 0 {
		return
	}
	n, err = d.redis.HDel(ctx, _matchMap, tradeParentIDs...).Result()
	if err != nil {
		err = fmt.Errorf("conn.Do(Del, %d, %s) error(%w)", n, tradeParentIDs, err)
		return
	}
	return
}
func (d *dao) PSubscribeKeyspace() <-chan *redis.Message {
	// https://redis.io/topics/notifications
	index := strings.Index(_unmatch, ":")
	// psubscribe '__keyspace@0__:*'
	// __keyspace@0__:unmatch:*
	channels := "__keyspace@0__" + _unmatch[:index+1] + "*"
	subscribe := d.redis.PSubscribe(context.Background(), channels)
	return subscribe.Channel()
}

//func (d *dao) DelFromUnmatchAndSetToMatch(ctx context.Context, order *model.Order) (ok bool, err error) {
//	itemID := order.ItemID
//	adZoneID := order.AdzoneID
//	tradeParentID := order.TradeParentID
//	marshal, err := json.Marshal(order)
//	if err != nil {
//		err = fmt.Errorf("DelFromUnmatchAndSetToMatch failed error: (%w),order: %v", err, order)
//		return
//	}
//	key := unmatchKey(itemID, adZoneID)
//	pipeline := d.redis.TxPipeline()
//	del := pipeline.Del(ctx, key)
//	set := pipeline.HSet(ctx, _matchMap, tradeParentID, marshal)
//	_, err = pipeline.Exec(ctx)
//	if err != nil {
//		err = fmt.Errorf("DelFromUnmatchAndSetToMatch failed error: (%w),order: %v", err, order)
//		return
//	}
//	ok = del.Val() == 1 && set.Val() == 1
//	return
//}

func (d *dao) DelFromUnmatchAndSetToMatch(ctx context.Context, order *model.Order) (ok bool, err error) {
	key := unmatchKey(order.ItemID, order.AdzoneID)
	pipeline := d.redis.TxPipeline()
	del := pipeline.Del(ctx, key)
	set := pipeline.SAdd(ctx, _matchSet, order.TradeParentID)
	_, err = pipeline.Exec(ctx)
	if err != nil {
		err = fmt.Errorf("DelFromUnmatchAndSetToMatch failed error: (%w),order: %v", err, order)
		return
	}
	ok = del.Val() == 1 && set.Val() == 1
	return
}
func (d *dao) REMFromMatchSet(ctx context.Context, tradeParentID ...string) (n int64, err error) {
	rem := d.redis.SRem(ctx, _matchSet, tradeParentID)
	return rem.Val(), rem.Err()
}
func (d *dao) HSetNXToMatch(ctx context.Context, order *model.Order) (ok bool, err error) {
	marshal, err := json.Marshal(order)
	if err != nil {
		err = fmt.Errorf("HSetNXToMatch failed error: (%w),order: %v", err, order)
		return
	}
	ok, err = d.redis.HSetNX(ctx, _matchMap, order.TradeParentID, marshal).Result()
	if err != nil {
		err = fmt.Errorf("HSetNXToMatch failed error: (%w),order: %v", err, order)
	}
	return
}
