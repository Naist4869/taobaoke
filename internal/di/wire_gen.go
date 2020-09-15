// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package di

import (
	"taobaoke/internal/dao"
	"taobaoke/internal/server/grpc"
	"taobaoke/internal/server/http"
	"taobaoke/internal/service"
)

// Injectors from wire.go:

func InitApp() (*App, func(), error) {
	logger, cleanup, err := service.NewLogger()
	if err != nil {
		return nil, nil, err
	}
	clusterClient, cleanup2, err := dao.NewRedis()
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	memcache, cleanup3, err := dao.NewMC()
	if err != nil {
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	db, cleanup4, err := dao.NewDB()
	if err != nil {
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	client, cleanup5, err := dao.NewMongo()
	if err != nil {
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	orderClient, err := dao.NewOrderClient(client, logger)
	if err != nil {
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	daoDao, cleanup6, err := dao.New(logger, clusterClient, memcache, db, client, orderClient)
	if err != nil {
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	blademasterClient, cleanup7, err := service.NewBmClient()
	if err != nil {
		cleanup6()
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	tbkMetrics, err := service.NewMetrics()
	if err != nil {
		cleanup7()
		cleanup6()
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	orders := service.NewOrders(daoDao, logger, tbkMetrics)
	serviceService, cleanup8, err := service.New(daoDao, logger, blademasterClient, orders)
	if err != nil {
		cleanup7()
		cleanup6()
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	engine, err := http.New(serviceService)
	if err != nil {
		cleanup8()
		cleanup7()
		cleanup6()
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	server, err := grpc.New(serviceService)
	if err != nil {
		cleanup8()
		cleanup7()
		cleanup6()
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	app, cleanup9, err := NewApp(serviceService, engine, server)
	if err != nil {
		cleanup8()
		cleanup7()
		cleanup6()
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	return app, func() {
		cleanup9()
		cleanup8()
		cleanup7()
		cleanup6()
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
	}, nil
}
