package service

import (
	"github.com/Naist4869/log"
	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	MaxSize int
	MaxAge  int
	LogDir  string
	Name    string
	Console bool
	Debug   bool
	Level   map[string]zapcore.Level
}

func NewLogger() (l *log.Logger, cf func(), err error) {

	var (
		ct  paladin.TOML
		cfg Config
	)
	if err = paladin.Get("logger.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Logger").UnmarshalTOML(&cfg); err != nil {
		return
	}
	l = log.NewLogger(cfg.MaxSize, cfg.MaxAge, cfg.LogDir, cfg.Name, cfg.Console, cfg.Debug, cfg.Level["main"])
	cf = func() { l.Close() }
	return
}
