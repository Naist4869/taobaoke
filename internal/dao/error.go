package dao

import "fmt"

// ErrIDNotFound 指定ID的用户不存在错误
type ErrIDNotFound struct {
	id string
}

func NewErrIDNotFound(id string) ErrIDNotFound {
	return ErrIDNotFound{id: id}
}
func (e ErrIDNotFound) Error() string {
	return fmt.Sprintf("未找到ID为%s的订单", e.id)
}
