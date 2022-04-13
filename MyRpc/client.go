package MyRpc

import (
	"time"
	"Global/CodeC"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"errors"
	"context"
)

type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Error         error
	Done          chan *Call
}

func (call *Call) done() {
	call.Done <- call
}

type Client struct {
	cc         CodeC.Codec
	opt        *Option
	sending    sync.Mutex
	mu         sync.Mutex
	serviceMap map[uint64]*Call
	header     CodeC.Header
	seq        uint64
	closing    bool
	shutdowm   bool
}

func NewClient(cc CodeC.Codec, opt *Option) *Client {
	client := &Client{
		cc:         cc,
		opt:        opt,
		serviceMap: make(map[uint64]*Call),
	}
	return client
}

var ErrShutdown = errors.New("connection is shut down")

func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing {
		return ErrShutdown
	}
	client.closing = true
	return client.cc.Close()
}

func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.shutdowm && !client.closing
}

//for timeout
type clientDialResult struct{err error}

func ClientDial(addr string, t time.Duration) *Client {
	conn, errDial := net.DialTimeout("tcp", addr, t)
	if errDial != nil {
		fmt.Println("rpc client dial failed:", errDial)
		return nil
	}
	//send option to server first
	//直接设置
	opt := &Option{
		MagicNumber: MagicNumber,
		CodecType:   CodeC.GobType,
	}
	ch := make(chan clientDialResult)
	f := CodeC.NewCodecModeMap[opt.CodecType]
	go func(opt *Option, connection net.Conn){
		err := json.NewEncoder(connection).Encode(opt)
		ch <- clientDialResult{err:err}
	}(opt, conn)
	select{
	case <-time.After(t):
		fmt.Println("client parse option timeout")
		conn.Close()
		return nil
	case result := <-ch:
		if result.err != nil {
			fmt.Println("client encode opt failed:", result.err)
			conn.Close()
			return nil
		}
	}
	cc := f(conn)
	client := NewClient(cc, opt)
	go client.receive()
	return client
}

func (c *Client) receive() {
	var err error
	for err == nil {
		h := CodeC.Header{}
		err = c.cc.ReadHeader(&h)
		if err != nil {
			fmt.Println("client read header failed and close:", err)
			break
		}
		call := c.removeCall(h.Seq)
		if call == nil {
			//the call is not in servieceMap
			//ignore the call
			err = c.cc.ReadBody(nil) //输入为nil则返回值也为nil
		} else if h.Error != "" {
			call.Error = fmt.Errorf(h.Error)
			err = c.cc.ReadBody(nil)
			call.done()
		} else {
			err = c.cc.ReadBody(call.Reply)
			if err != nil {
				fmt.Println("client read body failed:", err)
			}
			call.done()
		}
	}
	c.TerminateCall(err)
}

func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing || c.shutdowm {
		return 0, ErrShutdown
	}
	call.Seq = c.seq
	c.serviceMap[call.Seq] = call
	// fmt.Println("register,",call.Seq)
	c.seq++
	return call.Seq, nil

}

func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	call := c.serviceMap[seq]
	delete(c.serviceMap, seq)
	return call
}

func (client *Client) TerminateCall(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()
	client.shutdowm = true
	for _, call := range client.serviceMap {
		call.Error = err
		call.done()
	}
}

func (c *Client) Call(ctx context.Context, serviceMethod string, argv interface{}, reply interface{}) error {
	call := c.Go(serviceMethod, argv, reply, make(chan *Call, 1))
	select{
	case <-ctx.Done():
		c.removeCall(c.seq)
		fmt.Println("client call timeout")
		return errors.New("client: call failed:"+ctx.Err().Error())
	case result := <-call.Done:
		return result.Error
	}
}

func (c *Client) Go(serviceMethod string, argv interface{}, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          argv,
		Reply:         reply,
		Done:          done,
	}
	c.send(call)
	return call
}

func (client *Client) send(call *Call) {
	client.sending.Lock()
	defer client.sending.Unlock()

	seq, err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}
	client.header.Seq = seq
	client.header.ServiceMethod = call.ServiceMethod
	client.header.Error = ""
	err = client.cc.Write(&client.header, call.Args)
	if err != nil {
		fmt.Println("client write failed:", err)
		call := client.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}

}
