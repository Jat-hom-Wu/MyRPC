package xclient

import(
	"fmt"
	"context"
	. "Global/MyRpc"
	"sync"
	"time"
	"errors"
	"reflect"
	"strings"
	// "log"
)

type XClient struct{
	d Discovery	//get,getall接口
	mode SelectMode
	mu sync.Mutex
	opt *Option
	clients map[string]*Client
}

func NewXClient(d Discovery, mode SelectMode, opt *Option) *XClient{
	return &XClient{
		d:d,
		mode:mode,
		opt:opt,
		clients:make(map[string]*Client),
	}
}

//一关就全部client都关了
func (xc *XClient)Close(){
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for key,client := range xc.clients{
		client.Close()
		delete(xc.clients, key)
	}
}

	//switch avalible client
	//use client call

//嘿嘿嘿
type BoardCastFlag struct{
	Ok bool
	Addr string
}

func (xc *XClient)Call(ctx context.Context, serviceMethod string, args, reply interface{}, t time.Duration, boardcast BoardCastFlag) error{
	xc.mu.Lock()
	defer xc.mu.Unlock()
	addr,err := xc.d.Get(xc.mode)
	if err != nil{
		fmt.Println("xclient call failed:", err)
		return err
	}
	if boardcast.Ok{
		addr = boardcast.Addr
	}
	index := strings.Index(addr, "]")
	addr = addr[index + 2:]
	addr = "127.0.0.1:" + addr
	client,ok := xc.clients[addr]
	if ok && !client.IsAvailable(){
		fmt.Println("here")
		client.Close()
		delete(xc.clients, addr)
		client = nil
	}
	if client == nil{
		client = ClientDial(addr, t)
		if client == nil{
			fmt.Println("client dial failed")
			return errors.New("client dial failed")
		}
		xc.clients[addr] = client
	}
	err = client.Call(ctx, serviceMethod, args,reply)
	if err != nil{
		// log.Fatal("client call failed:", err)
		fmt.Println("client call failed:",err)
		return err
	}
	return nil
}

func (xc *XClient)BoardCast(ctx context.Context, serviceMethod string, args, reply interface{}, t time.Duration) error{
	//get the slice of server and range it 
	//invoke call

	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex // protect e and replyDone
	var e error
	replyDone := reply == nil // if reply is nil, don't need to set value
	ctx, cancel := context.WithCancel(ctx)
	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var clonedReply interface{}
			if reply != nil {
				clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			boardcast := BoardCastFlag{
				Ok:true,
				Addr:rpcAddr,
			}
			err := xc.Call(ctx, serviceMethod, args, clonedReply, t, boardcast)
			mu.Lock()
			if err != nil && e == nil {
				e = err
				cancel() // if any call failed, cancel unfinished calls
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())
				replyDone = true
			}
			mu.Unlock()
		}(rpcAddr)
	}
	wg.Wait()
	return e

}