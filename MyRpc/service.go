package MyRpc

import (
	"fmt"
	"go/ast"
	"reflect"
	"sync/atomic"
)

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  uint64 //统计方法调用次数
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

//???
func (m *methodType) newArgv() reflect.Value {
	var value reflect.Value
	if m.ArgType.Kind() == reflect.Ptr {
		value = reflect.New(m.ArgType.Elem()) //elem ?
	} else {
		value = reflect.New(m.ArgType).Elem()
	}
	return value
}

//confuses me here
func (m *methodType) newReplyv() reflect.Value {
	// reply must be a pointer type			//why?
	replyv := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}

type service struct {
	seviceName string
	typ        reflect.Type
	self       reflect.Value //结构体的实例本身，在调用时需要 self 作为第 0 个参数
	methods    map[string]*methodType
}

func newService(self interface{}) *service {
	s := &service{
		// seviceName : reflect.ValueOf(self).Type().Name(), //what the fuck??? 为什么不行？？？
		seviceName: reflect.Indirect(reflect.ValueOf(self)).Type().Name(),
		typ:        reflect.TypeOf(self),
		self:       reflect.ValueOf(self),
	}
	s.methods = make(map[string]*methodType)
	if !ast.IsExported(s.seviceName) {
		fmt.Println("rpc server: ", s.seviceName, "is not a valid service name")
	}
	s.register()
	return s
}

func (s *service) register() {
	length := s.typ.NumMethod()
	for i := 0; i < length; i++ {
		method := s.typ.Method(i)
		mType := method.Type

		//TODO:输入异常处理

		argvType := mType.In(1)
		replyType := mType.In(2)
		s.methods[method.Name] = &methodType{
			method:    method,
			ArgType:   argvType,
			ReplyType: replyType,
		}
		fmt.Printf("server register serviceMethod: %s.%s\n",s.seviceName,method.Name)
	}
}

func (s *service) Call(m *methodType, argv reflect.Value, reply reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	result := f.Call([]reflect.Value{s.self, argv, reply})
	if err := result[0].Interface(); err != nil {
		return err.(error) //类型断言
	}
	return nil
}
