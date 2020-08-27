package service

type QueryListEmpty struct{}

func NewQueryListEmptyError() error {
	return QueryListEmpty{}
}
func (q QueryListEmpty) Error() string {
	return "查询列表为空"
}
func (q QueryListEmpty) Is(target error) bool {
	switch target.(type) {
	case *QueryListEmpty, QueryListEmpty:
		return true
	default:
		return false
	}
}
