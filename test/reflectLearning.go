package main

import (
	"fmt"
	"reflect"
)

type rePeople struct {
	Name string
	Age  int
	Sex  string
}

type printor interface {
	print()
}

type reTool struct {
	cap string
	key string
}

func (t *reTool) print() {
	fmt.Println("print t:", t.cap, t.key)
}

func test1() {
	var p1 printor
	t1 := &reTool{"a", "b"} //指针接收者限制
	p1 = t1
	p1.print()
}

func (p rePeople) Say(msg string) {
	fmt.Println("hello", msg)
}

func (p rePeople) PrintInfo(t *reTool) {
	fmt.Printf("姓名:%s, 年龄:%d, 性别:%s, 参数tool内容:%s %s\n", p.Name, p.Age, p.Sex, t.key, t.cap)
}

type service struct {
	servers map[string]reflect.Method
	val     reflect.Value
	typ     reflect.Type
}

func makeservice(rep interface{}) *service {
	ser := service{}
	ser.typ = reflect.TypeOf(rep)
	ser.val = reflect.ValueOf(rep)
	/*---------*/
	//get name by struct type or value
	name := reflect.Indirect(ser.val).Type().Name()
	fmt.Println("name:", name)
	fmt.Println("get name by type:", ser.typ.Name())
	/*-----------*/
	ser.servers = map[string]reflect.Method{}
	for i := 0; i < ser.typ.NumMethod(); i++ {
		method := ser.typ.Method(i)
		mname := method.Name
		mtype := method.Type //?
		fmt.Println("method.Type:", mtype)
		ser.servers[mname] = method
	}
	return &ser
}

func Run() {
	//map查找中可设置标志位ok判断是否存在所查找的key
	// var testMap map[string]int
	// testMap = map[string]int{}
	// testMap["test"] = 1
	// testReceiver,ok := testMap["test1"]
	// if ok{
	// 	fmt.Println("receive ok,testReceiver:", testReceiver)
	// }else{
	// 	fmt.Println("no this member")
	// }

	rec := reflect.TypeOf((*error)(nil)).Elem() //包装成type
	fmt.Printf("tt: %T\n", rec)
	fmt.Println(rec.Kind())

	p1 := rePeople{"xiaoming", 18, "man"}
	r := reflect.TypeOf(&p1)
	fmt.Println("test1:", r)
	r1 := reflect.New(r)
	fmt.Printf("r1:%T\n", r1)
	fmt.Println("r1 kind:", r.Kind())
	r1.Elem() //Elem的接收者需要是一个指针或其他指定类型，不能是struct。 elem的接受者的kind需要是ptr
	fmt.Printf("test here r:%T\n", r)
	fmt.Printf("test2:%v\n", r)
	ser := makeservice(p1)
	methodName := "PrintInfo"
	method, ok := ser.servers[methodName]
	if ok {
		replyType := method.Type.In(1) //i think this style is not good. In方法拿到的是指向该type的指针
		fmt.Println("method's type:", method.Type)
		fmt.Println("replyType:", replyType)
		replyType = replyType.Elem() // Elem会返回对	//为什么要经过这一步呢？
		fmt.Println("replyType2:", replyType)
		// New returns a Value representing a pointer to a new zero value for the specified type.
		replyv := reflect.New(replyType) //底层是指针	new：Type convert to value.The value is zero   //new可以改变kind的类型，struct转成ptr
		fmt.Println("replyv:", replyv)
		function := method.Func
		function.Call([]reflect.Value{ser.val, replyv})
	} else {
		fmt.Println("no this method")
	}
}

func TestConvertInterface() {
	var a interface{}
	a = 99
	c := 9
	//go中不支持隐式类型转换
	fmt.Printf("%T\n", c)
	b := float32(c)
	fmt.Printf("%T\n", b)
	fmt.Println(b)
	fmt.Println(a)
	p1 := rePeople{}
	fmt.Printf("%T\n", a)
	a = p1 	//是可以转的，空接口可以接收任意类型
	fmt.Printf("%T\n", a)
	d,ok := a.(int) //类型断言的类型检查？
	if ok{
		fmt.Println(d)
	}else{
		fmt.Println("a not int")
	}
	fmt.Println(t1())
}

func t1() int {
	var a interface{}
	a = 999
	return a.(int)
}
