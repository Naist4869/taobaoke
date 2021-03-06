package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"taobaoke/internal/model"
	"taobaoke/tools"

	"go.mongodb.org/mongo-driver/bson"

	"go.uber.org/zap"
)

// 空对象
const cacheNull = "{null}"

//FGetKey 根据id获取key
type FGetKey func(id string, opt ...interface{}) (key string, field []string)

// ICacheObj 对象缓存接口
type ICacheObj interface {
	// GetKey 获取对象缓存的key
	GetKey(id string, opt ...interface{}) (key string, field []string)
	// AppendString--v is from cache, to decode & append to dest, && check null
	ICacheDecoder
	// AppendRaw--get from db, to encode & append to dest & to convert to string
	FillFromDB(ctx context.Context, d CacheObjService, miss []string, kopt []interface{}, opt ...interface{}) (missCache map[string]interface{}, err error)
}

type ICacheDecoder interface {
	io.Writer
}

func (d *dao) QueryOrderByTradeParentID(ctx context.Context, tradeParentIDs []string, onlyUnfinished bool) (results []*model.Order, err error) {
	objDest := &tradeParentIDsToOrderArr{}
	if err = d.objCache(ctx, tradeParentIDs, objDest, nil); err != nil {
		return
	}
	if !onlyUnfinished {
		results = objDest.dest
		return
	}
	for _, item := range objDest.dest {
		if !item.Status.Finish() {
			results = append(results, item)
		}
	}
	return
}

func (d *dao) QueryNotWithDrawOrderByUserID(ctx context.Context, id string) (result []*model.Order, err error) {
	query := bson.M{
		model.UserIDField:         id,
		model.WithDrawStatusField: false,
	}
	d.logger.Info("QueryNotWithDrawOrderByUserID订单查询", ZapBsonM("条件", query))
	orders, _, err := d.orderClient.queryOrder(ctx, query, nil, 0, 0, nil, nil)
	if err != nil {
		d.logger.Error(fmt.Sprint("QueryNotWithDrawOrderByUserID订单查询失败"), zap.Error(err), ZapBsonM("条件", query))
		return nil, err
	}
	return orders, nil
}
func (d *dao) QueryOrderByStatus(ctx context.Context, start, end tools.Time, status ...model.OrderStatus) ([]*model.Order, error) {
	query := bson.M{model.CreateTimeField: bson.M{GTE: start, LTE: end}}
	if len(status) > 0 {
		query[model.StatusField] = bson.M{IN: status}
	}
	d.logger.Info("订单查询", ZapBsonM("条件", query))
	orders, _, err := d.orderClient.queryOrder(ctx, query, []string{model.CreateTimeField}, 0, 0, nil, nil)
	if err != nil {
		d.logger.Error(fmt.Sprint("订单查询失败"), zap.Error(err), ZapBsonM("条件", query))
		return nil, err
	}
	return orders, nil
}

/*UpdateSingleOrderGeneric 通用的更新单个订单方法
参数:
*	id              	string	id/订单号
*	additionalFilter	bson.M	额外的过滤条件,可以为nil
*	operation       	bson.M	操作条件
返回值:
*	error	error
*/
func (d *dao) UpdateSingleOrderGeneric(ctx context.Context, id string, additionalFilter, operation bson.M) (err error) {
	d.logger.Info("准备操作订单", zap.String("订单编号", id), ZapBsonM("额外的过滤", additionalFilter), ZapBsonM("操作", operation))
	if additionalFilter == nil {
		additionalFilter = bson.M{}

	}
	additionalFilter = CombineBsonM(additionalFilter, bson.M{model.IDField: id, model.DeletedField: false})
	d.logger.Info("准备操作订单-组合条件", zap.String("订单编号", id), ZapBsonM("完整的过滤", additionalFilter), ZapBsonM("操作", operation))
	if err = d.UpdateOrder(ctx, additionalFilter, operation, true); err != nil {
		return fmt.Errorf("UpdateSingleOrderGeneric: %w", err)
	}
	return nil
}

/*UpdateOrder 更新订单
参数:
*	filter   	bson.M		过滤条件
*	operation	bson.M		操作
*	strict   	bool		是否严格模式，严格模式表示如果没有数据被更新，那么返回错误
返回值:
*	error	error
*/
func (d *dao) UpdateOrder(ctx context.Context, filter, operation bson.M, strict bool) error {
	result, err := d.orderClient.collections[model.DBOrderKey].UpdateMany(ctx, filter, operation)
	if err != nil {
		return fmt.Errorf("更新数据库错误: %w,条件[%s],操作[%s]", err, filter, operation)
	}
	if strict {
		if result.MatchedCount == 0 {
			return NewUnMatchedError(filter)
		}
	}
	return nil
}

func (d *dao) objCache(ctx context.Context, ids []string, dest ICacheObj, kopt []interface{}, opt ...interface{}) (err error) {
	//get from cache
	miss := d.getFromCache(ctx, ids, dest.GetKey, dest, kopt)
	if len(miss) == 0 {
		return
	}
	//get miss from db
	missCache, err := dest.FillFromDB(ctx, d.orderClient, miss, kopt, opt...)
	if err != nil {
		return
	}
	//set miss to cache
	if missCache == nil {
		missCache = map[string]interface{}{}
	}
	var key string
	for _, missid := range miss {
		var fields []string
		key, fields = dest.GetKey(missid, kopt...)
		for _, missk := range fields {
			if _, exist := missCache[missk]; exist {
				continue
			}
			missCache[missk] = cacheNull
		}
	}
	d.logger.Info("objCache", zap.String("missCache", fmt.Sprintf("%+v", missCache)))
	d.orderCacheCh <- map[string]interface{}{key: missCache}
	return
}

// Error represents an error returned in a command reply.
type Error string

func (err Error) Error() string { return string(err) }

// Strings is a helper that converts an array command reply to a []string. If
// err is not equal to nil, then Strings returns nil, err. Nil array items are
// converted to "" in the output slice. Strings returns an error if an array
// item is not a bulk string or nil.
func Strings(reply interface{}, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	switch reply := reply.(type) {
	case []interface{}:
		result := make([]string, len(reply))
		for i := range reply {
			if reply[i] == nil {
				continue
			}
			p, ok := reply[i].(string)
			if !ok {
				return nil, fmt.Errorf("redigo: unexpected element type for Strings, got type %T", reply[i])
			}
			result[i] = p
		}
		return result, nil
	case nil:
		return nil, errors.New("redigo: nil returned")
	case Error:
		return nil, reply
	}
	return nil, fmt.Errorf("redigo: unexpected type for Strings, got type %T", reply)
}
func (d *dao) getFromCache(ctx context.Context, ids []string, getKey FGetKey, dest ICacheDecoder, kopt []interface{}) (miss []string) {

	var (
		err    error
		caches []string
		key    string
	)

	if len(ids) == 0 {
		return
	}
	// 去重
	ids = tools.Unique(ids, false)

	keymap := make(map[string]string, len(ids)) // 关键词->id
	fields := make([]string, 0, len(ids))       // []关键字

	for _, id := range ids {
		var field []string
		key, field = getKey(id, kopt...)

		for _, k := range field {
			if _, exist := keymap[k]; exist {
				continue
			}
			keymap[k] = id
			fields = append(fields, k)
		}
	}

	if caches, err = Strings(d.redis.HMGet(ctx, key, fields...).Result()); err != nil {
		miss = ids
		d.logger.Warn("getFromCache", zap.Error(err), zap.String("key", key), zap.Strings("fields", fields))
		err = nil
		return
	}
	// 去重
	missMap := map[string]bool{}
	for i, item := range caches {
		id := keymap[fields[i]]
		if item != "" && item != cacheNull {
			if _, err = dest.Write([]byte(item)); err == nil {
				continue
			}
			d.logger.Warn("getFromCache", zap.Error(err), zap.String("id", id), zap.String("item", item))
		}
		if _, exist := missMap[id]; exist || item == cacheNull {
			continue
		}
		miss = append(miss, id)
		missMap[id] = true
	}
	d.logger.Info("getFromCache", zap.String("key", key), zap.Strings("fields", fields), zap.Strings("cache", caches), zap.Strings("miss", miss))
	return
}

// 开箱即用
type tradeParentIDsToOrderArr struct {
	dest []*model.Order
}

func (o *tradeParentIDsToOrderArr) GetKey(tradeParentID string, opt ...interface{}) (key string, field []string) {
	return _matchMap, []string{tradeParentID}
}

func (o *tradeParentIDsToOrderArr) FillFromDB(ctx context.Context, d CacheObjService, miss []string, kopt []interface{}, opt ...interface{}) (missCache map[string]interface{}, err error) {

	orders, err := d.FindOrderByTradeParentID(ctx, miss)
	if err != nil && !errors.Is(err, ErrTradeParentIDNotFound{}) {
		err = fmt.Errorf("tradeParentIDsToOrderArr FillFromDB FindOrderByIDs error(%w) miss(%s)", err, miss)
		return
	}
	o.dest = append(o.dest, orders...)
	missCache = map[string]interface{}{}
	for _, item := range orders {
		bs, err := json.Marshal(item)
		if err != nil {
			//err = fmt.Errorf("FillFromDB json.Marshal error(%w) item(%+v)",err,item)
			continue
		}
		_, fields := o.GetKey(item.TradeParentID, kopt...)
		for _, k := range fields {
			missCache[k] = string(bs)
		}
	}
	return missCache, nil
}

func (o *tradeParentIDsToOrderArr) Write(p []byte) (n int, err error) {
	one := &model.Order{}
	if err = json.Unmarshal(p, one); err != nil {
		return
	}
	o.dest = append(o.dest, one)
	return
}

// 开箱即用
type orderIDsToOrderArr struct {
	dest []*model.Order
}

func (o *orderIDsToOrderArr) GetKey(id string, opt ...interface{}) (key string, field []string) {
	return _orderMap, []string{id}
}

func (o *orderIDsToOrderArr) Write(p []byte) (n int, err error) {
	one := &model.Order{}
	if err = json.Unmarshal(p, one); err != nil {
		return
	}
	o.dest = append(o.dest, one)
	return
}

func (o *orderIDsToOrderArr) FillFromDB(ctx context.Context, d CacheObjService, miss []string, kopt []interface{}, opt ...interface{}) (missCache map[string]interface{}, err error) {
	orders, err := d.FindOrderByIDs(ctx, miss)
	if err != nil && !errors.Is(err, ErrOrderIDNotFound{}) {
		err = fmt.Errorf("orderIDsToOrderArr FillFromDB FindOrderByIDs error(%w) miss(%s)", err, miss)
		return
	}
	o.dest = append(o.dest, orders...)
	missCache = map[string]interface{}{}
	for _, item := range orders {
		bs, err := json.Marshal(item)
		if err != nil {
			//err = fmt.Errorf("FillFromDB json.Marshal error(%w) item(%+v)",err,item)
			continue
		}
		_, fields := o.GetKey(item.ID, kopt...)
		for _, k := range fields {
			missCache[k] = string(bs)
		}
	}
	return missCache, nil
}

func (d *dao) DelFromOrderCache(ctx context.Context, IDs []string) (n int64, err error) {
	if len(IDs) == 0 {
		return
	}
	n, err = d.redis.HDel(ctx, _orderMap, IDs...).Result()
	if err != nil {
		err = fmt.Errorf("conn.Do(Del, %d, %s) error(%w)", n, IDs, err)
		return
	}
	return
}
