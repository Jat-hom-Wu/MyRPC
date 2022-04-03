package MyRpc

import(
	"fmt"
	"net"
	"encoding/json"
	"reflect"
	// "strings"
	"sync"
	"io"
	"Global/CodeC"
)

type Server struct{}

type Option struct{
	MagicNumber int
	CodecType string
}

type request struct{
	header *CodeC.Header
	argv,reply reflect.Value
}

const MagicNumber = 0x3bef5c

func(s *Server)Accept(listener net.Listener) error{
	for{
		conn,err := listener.Accept()
		if err != nil{
			fmt.Println("server listen failed:", err)
		}
		go s.ServeOption(conn)
	}
}


func (s *Server)ServeOption(conn net.Conn){
	var opt Option
	err := json.NewDecoder(conn).Decode(&opt)	//是否取地址？
	if err != nil{
		fmt.Println("server json-encode option fialed,", err)
		return
	}
	if opt.MagicNumber != MagicNumber{
		fmt.Println("server receive invailed MagicNumber,",opt.MagicNumber)
		return
	}
	f,ok := CodeC.NewCodecModeMap[opt.CodecType]	
	if !ok{
		fmt.Println("server receive invailed CodecType,",opt.CodecType)
		return
	}
	s.ServeMesage(f(conn))
}

func (s *Server)readHeader(cc CodeC.Codec)(*CodeC.Header, error){
	var h CodeC.Header
	err := cc.ReadHeader(&h)
	if err != nil && err != io.EOF{
		fmt.Println("server read header failed,",err)
		return nil,err
	}
	return &h,err
}

func (s *Server)sendResponse(cc CodeC.Codec, h *CodeC.Header, body interface{}, sending *sync.Mutex){
	sending.Lock()
	defer sending.Unlock()
	err := cc.Write(h, body)
	if err != nil{
		fmt.Println("server write failed:",err)
	}
}

func (s *Server)handleRequest(cc CodeC.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup){
	defer wg.Done()
	fmt.Println("handleReqeust:",req.header, req.argv.Elem())
	req.reply = reflect.ValueOf(fmt.Sprintf("server geerpc resp %d", req.header.Seq))
	s.sendResponse(cc, req.header, req.reply.Interface(), sending)
}

func (s *Server)ServeMesage(codec CodeC.Codec){
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for{
		//decode header
		h,err := s.readHeader(codec)
		//error handle
		if err != nil{
			//should send error to client here
			break
		}
		//decode body
		req := &request{header:h}
		//suppose it is a string now
		req.argv = reflect.New(reflect.TypeOf(""))
		err = codec.ReadBody(req.argv.Interface())
		if err != nil{
			fmt.Println("server decode body failed:",err)
			var invalidRequest = struct{}{}
			s.sendResponse(codec, req.header, invalidRequest,sending)
			continue	//diefferent from readheaer here
		}
		//handle request
		//response
		wg.Add(1)
		go s.handleRequest(codec, req, sending, wg)
	}
	wg.Wait()
	_ = codec.Close()
}



var defaultServer =&Server{}


func Accept(listener net.Listener) error{
	return defaultServer.Accept(listener)
}



