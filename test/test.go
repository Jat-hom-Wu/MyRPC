package main

import (
	"time"
	// "os"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"reflect"
	"strings"
	"math/rand"
)

func main() {
	fmt.Println("test")
	// sample()
	// gobSample()
	// ReflectSample()
	// Run()
	// test1()
	// TestConvertInterface()
	// mapLearning()

	// var err error
	// if err == nil{
	// 	fmt.Println("error == nil")
	// }else{
	// 	fmt.Println("error != nil")
	// }

	// p1 := People{Name:"xiaoming",Age:18}
	// interfaceLearning1(p1)

	// reflectLearning()
	// t(1,2,"asdfa")

	// var a interface{}
	// var b int = 2
	// a = b
	// r1,r2 := a.(float32)
	// fmt.Println(r1,r2)
	// fmt.Printf("%T\n", r2)
	// fmt.Printf("%T\n",r1)
	// goroutineLearning()
	// randLearning()
	slieceLearning()
}

func slieceLearning(){
	s := make([]int, 0)
	s = append(s,2)
	s = append(s,3)
	s2 := make([]int, 0)
	s2 = make([]int,len(s))
	copy(s2,s)
	for key,value := range s2{
		fmt.Println(key,value)
	}

	s3 := make(map[int]string)
	s3[1] = "xiaoming"
	value,ok := s3[3]
	fmt.Println(value,ok)
}

func randLearning(){
	s1 := rand.Int()
	fmt.Println(s1)
	s2 := rand.New(rand.NewSource(time.Now().UnixNano()))
	fmt.Println(s2.Int())
}

func goroutineLearning(){
	go func(){
		time.Sleep(2*time.Second)
		fmt.Println("litttle goroutine")
	}()
	time.Sleep(3*time.Second)
	fmt.Println("big goroutine")
}

func t(arr ...interface{}) {
	if len(arr) == 0 {
		fmt.Println("no input")
		return
	} else {
		for _, v := range arr {
			fmt.Println(v)
		}
	}
}

func reflectLearning() {
	var a reflect.Value
	p1 := People{
		Name: "xiaoming",
		Age:  18,
	}

	a = reflect.New(reflect.TypeOf(p1))
	b := a.Elem() //elem就是取地址
	fmt.Printf("%v\n", a.Type())
	fmt.Printf("%v\n", b.Type())
	c := reflect.TypeOf(p1).Name()
	fmt.Println(c)
	fmt.Println(reflect.ValueOf(p1).Type().Name())
	if a == b {
		fmt.Println("equal")
	}
	// b := a.Interface()
	// b = "asdfasdf"
	// fmt.Println(b)
	// fmt.Println(a.Kind())
	// c := a.Elem().String()
	// fmt.Printf("%v\n",b)
	// fmt.Println(c)
	// fmt.Printf("%T\n", c)
	// aa := reflect.ValueOf(p1)
	// fmt.Println(aa.Interface())
	// fmt.Printf("%T\n",aa.Interface())
	// fmt.Println(b)
}

func interfaceLearning1(e interface{}) {
	var test interface{}
	test = 99
	test = "asdf"
	fmt.Println(test)
	p2 := e
	var tt interface{}
	// var p3 interface{}
	// p3 = e
	/* 这是错误的，不能直接声明一个类型去接收interface。如果一定要声明的话需要做类型断言（2种方法）。
		可以直接通过interface进行接收。
	var p People
	p = e
	*/
	s := Student{Name: "xiaoli", Age: 16}
	f := 5
	e = f
	tt = f
	fmt.Println("tt1:", tt)
	fmt.Printf("%T\n", tt)
	tt = s
	fmt.Println("tt2:", tt)
	fmt.Printf("%T\n", tt)
	fmt.Println(p2)
	fmt.Printf("%T\n", e)
	fmt.Println(tt)
}

func ReadFrom(reader io.Reader, num int) ([]byte, error) {
	p := make([]byte, num)
	n, err := reader.Read(p)
	if err != nil {
		fmt.Println("err not nil")
	} else {
		fmt.Println("err is nil")
	}
	fmt.Println("n:", n)
	if n > 0 {
		return p, nil
	} else {
		return p, err
	}
}

func sample() {
	data, _ := ReadFrom(strings.NewReader("from string"), strings.Count("from string", "")-10) //返回字符串长度加一
	fmt.Println("counts: ", strings.Count("f", ""))
	fmt.Println(data)
	fmt.Printf("%T\n", data)
	s_data := string(data)
	fmt.Println(s_data)
}

type People struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Student struct {
	Name string
	Age  int
}

func gobSample() {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	p := People{
		Name: "xiaoming",
		Age:  18,
	}
	err := encoder.Encode(p)
	if err != nil {
		fmt.Println("encode failed, err: ", err)
		return
	}
	fmt.Println(string(buf.Bytes()))

	//decode
	decoder := gob.NewDecoder(bytes.NewReader(buf.Bytes())) //bytes convert to reader(interface)
	var s1 Student
	decoder.Decode(&s1)
	fmt.Println("name: ", s1.Name, ", age: ", s1.Age)
}

func ReflectSample() {
	var a int = 10
	var b float32 = 6.66
	v := reflect.TypeOf(a) //v is not a type
	fmt.Printf("type:%T\n", v)
	fmt.Println(v)
	t := reflect.ValueOf(b)
	tt := t.Float() //类型要对应才行
	fmt.Printf("value's type:%T\n", t)
	fmt.Printf("tt's type:%T\n", tt)
	fmt.Println(t)

	p1 := People{
		Name: "xiaoming",
		Age:  18,
	}
	r := reflect.TypeOf(p1)
	fmt.Printf("r's type:%T\n", r)
	fmt.Println(r.Name(), r.Kind())
	for i := 0; i < r.NumField(); i++ {
		field := r.Field(i)
		fmt.Printf("name:%s index:%d type:%v json tag:%v\n", field.Name, field.Index, field.Type, field.Tag.Get("json"))
	}

	nameField, _ := r.FieldByName("Name")
	fmt.Println(nameField.Name, nameField.Index, nameField.Type)

	e := reflect.ValueOf(&p1)
	fmt.Println(e.Kind())
	fmt.Printf("e's type:%T\n", e)
	el := e.Elem() //这里e的kind不是interface或者ptr会出错
	fmt.Println("Elem:", el)

}

func mapLearning() {
	//若不存在，返回对应value类型的零值。其中string的零值是"",不是nil
	var m map[int]*string
	m = make(map[int]*string)
	t := "a"
	m[1] = &t
	re, ok := m[3]
	delete(m, 3)
	re, ok = m[3]
	if ok {
		fmt.Println("ok is true")
	} else {
		fmt.Println("ok is false")
	}
	if re == nil {
		fmt.Println("nil")
	}
	fmt.Println(re)
	fmt.Printf("%T\n", re)
}
