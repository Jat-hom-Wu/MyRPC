package main

import (
	"net"
	"net/http"
	// "Global/CodeC"
	"Global/MyRpc"
	"Global/registry"
	"Global/xclient"
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

func startServerDay06(addrCh chan string) {
	var foo Foo
	l, _ := net.Listen("tcp", ":0")
	server := MyRpc.NewServer()
	_ = server.Register(&foo)
	addrCh <- l.Addr().String()
	server.Accept(l)
}

func startServerDay07(addrCh chan string,registryAddr string) {
	var foo Foo
	l, _ := net.Listen("tcp", ":0")
	server := MyRpc.NewServer()
	addrCh <- l.Addr().String()
	_ = server.Register(&foo)
	fmt.Println("start server addr:","tcp@"+l.Addr().String())
	registry.HeartBeat(registryAddr, "tcp@"+l.Addr().String(), 0)
	server.Accept(l)
}

func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	wg.Done()
	_ = http.Serve(l, nil)
}

func main() {
	// Day03()
	// day06()
	day07()
}

func Day03() {
	fmt.Println("rpc test day03")
	go StartServerDay03()
	MyRpc.GlobalServerHandleTimeOut = 0 //测试reflect调用要3秒....默认timeout为0,即没有超时时间
	client := MyRpc.ClientDial("127.0.0.1:9527", 2*time.Second)
	if client == nil {
		fmt.Println("client dial error and shutdown")
		return
	}
	defer client.Close()
	time.Sleep(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
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

func day06() {
	fmt.Println("rpc test day06")
	ch1 := make(chan string)
	ch2 := make(chan string)
	go startServerDay06(ch1)
	go startServerDay06(ch2)
	addr1 := <-ch1
	addr2 := <-ch2
	time.Sleep(time.Second)
	fmt.Println("wake up")
	call(addr1, addr2)
	call(addr1, addr2)
	call(addr1, addr2)
	boardcast(addr1, addr2)
}

func day07() {
	fmt.Println("rpc test day07")
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()
	registryAddr := "http://127.0.0.1:9999/myrpc/register"
	ch1 := make(chan string)
	ch2 := make(chan string)
	go startServerDay07(ch1, registryAddr)
	go startServerDay07(ch2, registryAddr)
	<-ch1
	<-ch2
	time.Sleep(time.Second)
	call07()
	time.Sleep(time.Second)
	boardcast07()
}

func call(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{addr1, addr2}) //默认均为tcp //day06
	xc := xclient.NewXClient(d, xclient.RoundRobinSelect, nil)   //option直接写定了
	defer xc.Close()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var reply int
			var err error
			b := xclient.BoardCastFlag{
				Ok: false,
			}
			ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
			err = xc.Call(ctx, "Foo.Sum", &Args{A: i, B: i * i}, &reply, 2*time.Second, b)
			if err != nil {
				// log.Fatal("client call failed:", err)
				fmt.Println("client call error", err)
			} else {
				fmt.Printf("%s %s success: %d + %d = %d\n", "call", "Foo.Sum", i, i*i, reply)
			}
		}(i)
	}
	wg.Wait()
}

func call07() {
	d := xclient.NewRegisterDiscovery("http://127.0.0.1:9999/myrpc/register",0)
	xc := xclient.NewXClient(d, xclient.RoundRobinSelect, nil)   //option直接写定了
	defer xc.Close()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var reply int
			var err error
			b := xclient.BoardCastFlag{
				Ok: false,
			}
			ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
			err = xc.Call(ctx, "Foo.Sum", &Args{A: i, B: i * i}, &reply, 2*time.Second, b)
			if err != nil {
				// log.Fatal("client call failed:", err)
				fmt.Println("client call error", err)
			} else {
				fmt.Printf("%s %s success: %d + %d = %d\n", "call", "Foo.Sum", i, i*i, reply)
			}
		}(i)
	}
	wg.Wait()
}

func boardcast(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{addr1, addr2}) //默认均为tcp
	xc := xclient.NewXClient(d, xclient.RoundRobinSelect, nil)   //option直接写定了
	defer xc.Close()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var reply int
			var err error
			ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
			err = xc.BoardCast(ctx, "Foo.Sum", &Args{A: i, B: i * i}, &reply, 2*time.Second)
			if err != nil {
				fmt.Printf(" %s error: %v\n", "Foo.Sum", err)
			} else {
				fmt.Printf("boardcast,%s %s success: %d + %d = %d\n", "call", "Foo.Sum", i, i*i, reply)
			}
		}(i)
	}
	wg.Wait()
}

func boardcast07() {
	d := xclient.NewRegisterDiscovery("http://127.0.0.1:9999/myrpc/register",0)
	xc := xclient.NewXClient(d, xclient.RoundRobinSelect, nil)   //option直接写定了
	defer xc.Close()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var reply int
			var err error
			ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
			err = xc.BoardCast(ctx, "Foo.Sum", &Args{A: i, B: i * i}, &reply, 2*time.Second)
			if err != nil {
				fmt.Printf(" %s error: %v\n", "Foo.Sum", err)
			} else {
				fmt.Printf("boardcast,%s %s success: %d + %d = %d\n", "call", "Foo.Sum", i, i*i, reply)
			}
		}(i)
	}
	wg.Wait()
}

