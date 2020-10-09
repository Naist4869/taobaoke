package dao

import (
	"context"
	"math"
	"taobaoke/internal/model"

	"go.uber.org/zap"

	"github.com/Naist4869/log"
)

const abortIndex int8 = math.MaxInt8 / 2

type Context struct {
	ctx         context.Context
	engine      OrderMonitor
	index       int8
	localOrder  *model.Order
	updateArg   *model.UpdateArgument
	SalaryScale int64 // 返还比例  %90表示为90
	handlers    []HandlerFunc
	logger      *log.Logger
}
type HandlerFunc func(*Context)

func (c *Context) Next() {
	defer func() {
		if _, err := c.engine.DelFromMatchCache(c.ctx, c.localOrder.TradeParentID); err != nil {
			c.logger.Error("UpdateStatus", zap.Error(err), zap.String("tradeParentID", c.localOrder.TradeParentID))
		}
	}()
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context) reset() {
	c.handlers = nil
	c.index = -1
	c.updateArg = nil
	c.localOrder = nil
	c.ctx = nil
	c.SalaryScale = 0
}

func (c *Context) Abort() {
	c.index = abortIndex
}
