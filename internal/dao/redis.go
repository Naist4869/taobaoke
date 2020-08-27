package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"taobaoke/internal/model"
	"time"

	xtime "github.com/go-kratos/kratos/pkg/time"

	"github.com/go-redis/redis/v8"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/log"
)

type Fieldfun func() string

const (
	_orderMap   = "order_map:%s"
	_unmatchMap = "unmatch_map:%d.%d"
	_tkl        = "tkl:%s"
	redisOK     = "OK"
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
}

func NewRedis() (r *redis.Client, cf func(), err error) {
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
	r = redis.NewClient(&redis.Options{
		Network:            cfg.Network,
		Addr:               cfg.Addr,
		Password:           cfg.Password,
		DB:                 cfg.DB,
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
	return fmt.Sprintf(_unmatchMap, itemID, adZoneID)
}

func (d *dao) SetNXToUnmatch(ctx context.Context, itemID, adZoneID int64, orderNo string) (ok bool, err error) {
	//// https://github.com/redis/redis/issues/167 在field上设置过期时间在2020年还未实现...
	//if err = conn.Send("EXPIRE", _unmatchMap, 24*60*60); err != nil {
	//	log.Error("conn.Send(EXPIRE, %s, %d) error(%+v)", _unmatchMap, 24*60*60, err)
	//	return
	//}
	key := unmatchKey(itemID, adZoneID)
	// 保存5秒
	ok, err = d.redis.SetNX(ctx, key, orderNo, time.Second*5).Result()
	if err != nil {
		log.Error("conn.Do(SetNX, %s, %s) error(%v)", key, orderNo, err)
	}
	return
}

func (d *dao) SetToUnmatch(ctx context.Context, itemID, adZoneID int64, order *model.Order) (ok bool, err error) {
	key := unmatchKey(itemID, adZoneID)
	marshal, err := json.Marshal(order)
	if err != nil {
		return
	}
	// 保存10天
	result, err := d.redis.Set(ctx, key, marshal, time.Hour*24*10).Result()
	if err != nil {
		err = fmt.Errorf("conn.Do(Set, %s, %s) error(%v)", key, marshal, err)
		return
	}
	if result == redisOK {
		ok = true
		return
	}
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
func (d *dao) UnmatchGet(ctx context.Context, itemID, adZoneID int64) (order *model.Order, err error) {
	key := unmatchKey(itemID, adZoneID)
	result, err := d.redis.Get(ctx, key).Result()
	if err != nil {
		err = fmt.Errorf("conn.Do(Get, %s) error(%v)", key, err)
		return
	}
	if result == "" {
		err = fmt.Errorf("UnmatchGet not found order")
		return
	}
	if err = json.Unmarshal([]byte(result), order); err != nil {
		err = fmt.Errorf("UnmatchGet Unmarshal error: (%w),data: %s", err, result)
		return
	}
	return
}
func (d *dao) UnmatchGetAll(ctx context.Context) ([]string, error) {
	index := strings.Index(_unmatchMap, ":")
	iter := d.redis.Scan(ctx, 0, _unmatchMap[:index+1]+"*", 0).Iterator()
	all := make([]string, 0)
	for iter.Next(ctx) {
		all = append(all, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}
	return all, nil
}
func (d *dao) DelFromUnmatchMap(ctx context.Context, itemID, adZoneID int64) (int64, error) {
	key := unmatchKey(itemID, adZoneID)
	result, err := d.redis.Del(ctx, key).Result()
	if err != nil {
		log.Error("conn.Do(DEL, %s) error(%v)", key, err)
	}
	return result, err
}
