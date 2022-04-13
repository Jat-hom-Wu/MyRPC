package MyRpc

import(
	"net"
	"testing"
	"time"
	"context"
)

//client代码耦合，测不了，直接main测...

type Bar int

func (b Bar) Timeout(argv int, reply *int) error {
	time.Sleep(time.Second * 3)
	return nil
}

func startServer(addr chan string) {
	var b Bar
	GlobalServerHandleTimeOut = 5*time.Second
	_ = Register(&b)
	// pick a free port
	l, _ := net.Listen("tcp", ":0")
	addr <- l.Addr().String()
	Accept(l)
}

func TestClient_Call(t *testing.T) {
	t.Parallel()
	addrCh := make(chan string)
	go startServer(addrCh)
	addr := <-addrCh
	// time.Sleep(time.Second)
	// t.Run("client timeout", func(t *testing.T) {
	// 	client := ClientDial(addr, 3*time.Second)
	// 	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	// 	var reply int
	// 	err := client.Call(ctx, "Bar.Timeout", 1, &reply)
	// 	if err != nil {
	// 		t.Errorf("client call error:%v",err)
	// 	}
	// })
	t.Run("server handle timeout", func(t *testing.T) {
		client := ClientDial(addr, 1*time.Second)
		var reply int
		err := client.Call(context.Background(), "Bar.Timeout", 1, &reply)
		// _assert(err != nil && strings.Contains(err.Error(), "handle timeout"), "expect a timeout error")

		if err != nil {
			t.Errorf("client call error:%v",err)
		}
	})
}