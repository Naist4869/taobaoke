package gdbc

import (
	"reflect"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

// Spec 对数据库表的抽象描述
type Spec struct {
	collection *mongo.Collection
	version    int
	generator  func() interface{}
	updater    func(data interface{}) error
}

/*NewSpec 创建一个新的数据库版本规则
参数:
*	version   	int							// 期望版本，必须>1
*	generator 	func() interface{}			// 生成一个对应的对象，必须为指针
*	updater   	func(interface{}) error		// 升级方法，参数是所有版本低于version的数据，并且是调用generator()生成
返回值:
*	*Spec	*Spec
*	error	error
*/
func NewSpec(version int, generator func() interface{}, updater func(interface{}) error) (*Spec, error) {

	if version < 1 {
		return nil, errors.New("version参数必须大于等于1")
	}
	if generator == nil {
		return nil, errors.New("generator 参数不能为空")
	}
	if data := generator(); reflect.TypeOf(data).Kind() != reflect.Ptr || data == nil {
		return nil, errors.New("generator()方法必须返回一个非空指针")
	}
	if updater == nil {
		return nil, errors.New("updater参数不能为空")
	}
	return &Spec{
		version:   version,
		generator: generator,
		updater:   updater,
	}, nil
}
func (s *Spec) SetCollection(collection *mongo.Collection) {
	s.collection = collection
}
