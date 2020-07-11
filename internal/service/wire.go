// +build wireinject

// The build tag makes sure the stub is not built in the final build.
package service

import (
	"taobaoke/internal/dao"

	"github.com/google/wire"
)

//go:generate kratos tool wire
func newTestService() (*Service, func(), error) {
	panic(wire.Build(dao.Provider, Provider))
}
