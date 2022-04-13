package MyRpc

import (
	"fmt"
	"reflect"
	"testing"
)

type Args struct {
	a, b int
}

type Foo struct{}

func (f *Foo) Sum(args Args, reply *int) error {
	*reply = args.a + args.b
	return nil
}

func TestCall(t *testing.T) {
	var f Foo
	a := 1
	b := 2
	argvs := Args{
		a: a,
		b: b,
	}
	s := newService(&f)
	mType := s.methods["Sum"]
	argv := mType.newArgv()
	replyv := mType.newReplyv()
	argv.Set(reflect.ValueOf(Args{a: 1, b: 2}))
	fmt.Println(replyv)
	err := s.Call(mType, argv, replyv)
	if err != nil {
		t.Errorf("call failed: %#v", err)
	}
	want := new(int)
	f.Sum(argvs, want)
	fmt.Println(*want)
	fmt.Println(*replyv.Interface().(*int))
	if *replyv.Interface().(*int) != *want { //???
		t.Errorf("not equal")
	}
}


