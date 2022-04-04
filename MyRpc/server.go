package MyRpc

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"

	// "strings"
	"Global/CodeC"
	"errors"
	"io"
	"strings"
	"sync"
)

type Server struct {
	//没有用锁，用了一个sync.Map
	serviceMap sync.Map
}

type Option struct {
	MagicNumber int
	CodecType   string
}

type request struct {
	header      *CodeC.Header
	argv, reply reflect.Value
	mType       *methodType
	serive      *service
}

const MagicNumber = 0x3bef5c

func (s *Server) Accept(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("server listen failed:", err)
		}
		go s.ServeOption(conn)
	}
}

func (s *Server) ServeOption(conn net.Conn) {
	var opt Option
	//so important! tell the client that something wrong happen
	defer func() {
		conn.Close()
	}()
	err := json.NewDecoder(conn).Decode(&opt) //是否取地址？
	if err != nil {
		fmt.Println("server json-encode option fialed,", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		fmt.Println("server receive invailed MagicNumber,", opt.MagicNumber)
		return
	}
	f, ok := CodeC.NewCodecModeMap[opt.CodecType]
	if !ok {
		fmt.Println("server receive invailed CodecType,", opt.CodecType)
		return
	}
	s.ServeMesage(f(conn))
}

func (s *Server) readHeader(cc CodeC.Codec) (*CodeC.Header, error) {
	var h CodeC.Header
	err := cc.ReadHeader(&h)
	if err != nil && err != io.EOF {
		fmt.Println("server read header failed,", err)
		return nil, err
	}
	return &h, err
}

func (s *Server) sendResponse(cc CodeC.Codec, h *CodeC.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	err := cc.Write(h, body)
	if err != nil {
		fmt.Println("server write failed:", err)
	}
}

func (s *Server) handleRequest(cc CodeC.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	// fmt.Println("handleReqeust:", req.header, req.argv.Elem())	//used to panic.The req.argv is not pointer,so Elem() panic
	//Invoke method
	req.serive.Call(req.mType, req.argv, req.reply)
	s.sendResponse(cc, req.header, req.reply.Interface(), sending)
}

var invalidRequest = struct{}{}

func (s *Server) ServeMesage(codec CodeC.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		//decode header
		h, err := s.readHeader(codec)
		//error handle
		if err != nil{
			//should send error to client here
			break
		}
		//decode body
		req := &request{header: h}

		svc, mType,err := s.findserviceMethod(req.header.ServiceMethod)
		if err != nil{
			fmt.Println("find serviceMethod failed:",err)
			//send error to client
			req.header.Error = "server can't find the service or method"
			s.sendResponse(codec, req.header,invalidRequest,sending)
			continue
		}
		req.mType = mType
		req.serive = svc
		req.argv = req.mType.newArgv()
		req.reply = req.mType.newReplyv()
		//??? confuse...   why and how
		argvi := req.argv.Interface()
		if req.argv.Type().Kind() != reflect.Ptr {
			argvi = req.argv.Addr().Interface()
		}

		err = codec.ReadBody(argvi)
		if err != nil {
			fmt.Println("server decode body failed:", err)
			var invalidRequest = struct{}{}
			s.sendResponse(codec, req.header, invalidRequest, sending)
			continue //diefferent from readheaer here
		}
		//handle request
		//response
		wg.Add(1)
		go s.handleRequest(codec, req, sending, wg)
	}
	wg.Wait()
	_ = codec.Close()
}

func (server *Server) Register(rcvc interface{}) error {
	s := newService(rcvc)
	if _, dup := server.serviceMap.LoadOrStore(s.seviceName, s); dup {
		return errors.New("rpc: service already defined:" + s.seviceName)
	}
	return nil
}

func (server *Server) findserviceMethod(serviceMethod string) (*service, *methodType, error) {
	//这里不用锁，因为只读不修改
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		fmt.Println("invalid serviceMethod name :", serviceMethod)
		return nil, nil, errors.New("invalid serviceMethod Name")
	}
	seviceName := serviceMethod[:dot]
	methodName := serviceMethod[dot+1:]
	svc, ok := server.serviceMap.Load(seviceName)
	if !ok {
		fmt.Println("can not found the serviceName:", seviceName)
		return nil, nil, errors.New("not found this serviceName")
	}
	svic, ok := svc.(*service) //interface 类型断言
	if !ok {
		fmt.Println("type assert falied")
		return nil, nil, errors.New("type assert falied")
	}
	mType := svic.methods[methodName]
	if mType == nil {
		fmt.Println("can't find the method name:", methodName)
		return nil, nil, errors.New("can't found the methodName")
	}
	return svic, mType, nil
}

var defaultServer = &Server{}

func Accept(listener net.Listener) error {
	return defaultServer.Accept(listener)
}

func Register(rcvr interface{}) error { return defaultServer.Register(rcvr) }
