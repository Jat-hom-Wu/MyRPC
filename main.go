package main

import (
	"net"
	// "Global/CodeC"
	"Global/MyRpc"
	"fmt"
	"time"

	// "encoding/json"
	"context"
	"log"
	"sync"
)

type Foo struct{}
type Args struct{ A, B int }

func (f *Foo) Sum(args Args, reply *int) error {
	*reply = args.A + args.B
	return nil
}

func StartServer() {
	listener, err := net.Listen("tcp", "127.0.0.1:9527")
	if err != nil {
		fmt.Println("server listen failed:", err)
	}
	MyRpc.Accept(listener)
}

func StartServerDay03() {
	listener, err := net.Listen("tcp", "127.0.0.1:9527")
	if err != nil {
		fmt.Println("server listen failed:", err)
	}
	var foo Foo
	MyRpc.Register(&foo)
	MyRpc.Accept(listener)
}

func main() {
	fmt.Println("rpc test day03")
	go StartServerDay03()
	Day03()
}

func Day03() {
	MyRpc.GlobalServerHandleTimeOut = 0 //测试reflect调用要3秒....默认timeout为0,即没有超时时间
	client := MyRpc.ClientDial("127.0.0.1:9527", 2*time.Second)
	if client == nil {
		fmt.Println("client dial error and shutdown")
		return
	}
	defer client.Close()
	time.Sleep(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			args := &Args{A: j, B: j * j}
			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("client call failed:", err) //打印并退出！ 搞了1个小时，吐了！
			}
			fmt.Printf("%d + %d = %d\n", args.A, args.B, reply)
		}(i)
	}
	wg.Wait()
	fmt.Println("close")
}

//day2
// func main(){
// 	fmt.Println("rpc test")
// 	go StartServer()
// 	client := MyRpc.ClientDial("127.0.0.1:9527")
// 	defer client.Close()
// 	time.Sleep(time.Second)
// 	var wg sync.WaitGroup
// 	for i := 0; i < 5; i++{
// 		wg.Add(1)
// 		go func(j int){
// 			defer wg.Done()
// 			args := fmt.Sprintf("geerpc req %d", j)
// 			var reply string
// 			if err := client.Call("sum.foo", args, &reply);err != nil{
// 				fmt.Println("client call failed:",err)
// 			}
// 			fmt.Println("reply:",reply)
// 		}(i)
// 	}
// 	wg.Wait()
// 	fmt.Println("close")
// }

// func main(){
// 	fmt.Println("server test")
// 	go StartServer()
// 	//client implemnet
// 	client, err := net.Dial("tcp","127.0.0.1:9527")
// 	if err != nil{
// 		fmt.Println("client dial failed:",err)
// 	}
// 	defer func(){
// 		_ = client.Close()
// 	}()

// 	time.Sleep(time.Second)
// 	clientOption := MyRpc.Option{
// 		MagicNumber:MyRpc.MagicNumber,
// 		CodecType:CodeC.GobType,
// 	}
// 	err = json.NewEncoder(client).Encode(clientOption)
// 	if err != nil{
// 		fmt.Println("client encode option failed:",err)
// 	}
// 	cc := CodeC.NewGobCodec(client)
// 	for i := 0; i < 5; i++{
// 		h := &CodeC.Header{
// 			ServiceMethod:"Foo.Sum",
// 			Seq : uint64(i),
// 		}
// 		cc.Write(h,fmt.Sprintf("geerpc req %d",h.Seq))
// 		cc.ReadHeader(h)
// 		var reply string
// 		cc.ReadBody(&reply)
// 		fmt.Println("reply:", reply)
// 	}
// }
