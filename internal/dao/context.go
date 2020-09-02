package dao

import (
	"context"
	"taobaoke/internal/model"

	"github.com/Naist4869/log"
)

type Context struct {
	ctx        context.Context
	engine     OrderMonitor
	index      int8
	localOrder *model.Order
	updateArg  *model.UpdateArgument
	handlers   []HandlerFunc
	logger     *log.Logger
}
type HandlerFunc func(*Context)

func (c *Context) Next() {
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
}
