package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taobaoke/internal/di"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/log"
)

var (
	Version, Build string
)

func main() {
	flag.Parse()
	fmt.Println("Version:", Version)
	fmt.Println("Build Time:", Build)
	fmt.Println("HOSTNAME: ", os.Getenv("HOSTNAME"))
	log.Init(nil) // debug flag: log.dir={path}
	defer log.Close()
	log.Info("taobaoke start")
	paladin.Init()
	_, closeFunc, err := di.InitApp()
	if err != nil {
		panic(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeFunc()
			log.Info("taobaoke exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
