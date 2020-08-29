package dao

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

// ErrOrderIDNotFound 指定ID的用户不存在错误
type ErrOrderIDNotFound struct {
	id []string
}

func NewErrIDNotFound(id []string) ErrOrderIDNotFound {
	return ErrOrderIDNotFound{id: id}
}
func (e ErrOrderIDNotFound) Error() string {
	return fmt.Sprintf("未找到OrderID为%s的订单", e.id)
}
func (e ErrOrderIDNotFound) Is(target error) bool {
	switch target.(type) {
	case *ErrOrderIDNotFound, ErrOrderIDNotFound:
		return true
	default:
		return false
	}
}

// ErrTradeParentIDNotFound 指定的淘宝订单号不存在错误
type ErrTradeParentIDNotFound struct {
	id []string
}

func NewErrTradeParentIDNotFound(id []string) ErrTradeParentIDNotFound {
	return ErrTradeParentIDNotFound{id: id}
}
func (e ErrTradeParentIDNotFound) Error() string {
	return fmt.Sprintf("未找到淘宝订单号为%s的订单", e.id)
}
func (e ErrTradeParentIDNotFound) Is(target error) bool {
	switch target.(type) {
	case *ErrTradeParentIDNotFound, ErrTradeParentIDNotFound:
		return true
	default:
		return false
	}
}

type UmMatchedError struct {
	filter bson.M
}

func NewUnMatchedError(filter bson.M) error {
	return UmMatchedError{filter: filter}
}
func (u UmMatchedError) Error() string {
	return fmt.Sprintf("没有符合条件的数据[%v]", u.filter)
}
func (u UmMatchedError) Is(target error) bool {
	switch target.(type) {
	case *UmMatchedError, UmMatchedError:
		return true
	default:
		return false
	}
}
