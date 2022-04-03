package main

import(
	"net"
	"fmt"
	"encoding/json"
	// "io"
	// "time"
	// "bufio"
)

type Option struct {
	MagicNumber int        // MagicNumber marks this's a geerpc request
	CodecType   string // client may choose different Codec to encode body
}

func main(){
	fmt.Println("test")
	listener,err := net.Listen("tcp","127.0.0.1:9527")
	if err != nil{
		fmt.Println("listen error",err)
	}
	for{
		conn,err := listener.Accept()
		if err != nil{
			fmt.Println("accept failed",err)
		}
		fmt.Println("connected")
		var res []byte
		res = make([]byte,256)
		conn.Read(res)
		fmt.Println("res:",string(res))
		var opt Option
		if err := json.NewDecoder(conn).Decode(&opt); err != nil {
			fmt.Println("rpc server: options error: ", err)
			fmt.Println(opt)
			return
		}
		fmt.Println("option:",opt)
		// if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		// 	fmt.Println("rpc server: options error: ", err)
		// 	fmt.Println(opt)
		// 	return
		// }
		var res3 []byte
		res3 = make([]byte,256)
		conn.Read(res3)
		fmt.Println("res3:",string(res3))

		fmt.Println("run here")
		fmt.Println("option:",opt)
		if err := json.NewDecoder(conn).Decode(&opt); err != nil {
			fmt.Println("rpc server: options error: ", err)
			return
		}
		fmt.Println("finaly")
		var res2 []byte
		res2 = make([]byte,256)
		conn.Read(res2)
		fmt.Println("res2:",string(res2))
	}
	
}

// func client(){
// 	time.Sleep(2*time.Second)
// 	conn,err := net.Dial("tcp", "127.0.0.1:9527")
// 	if err != nil{
// 		fmt.Println("dial failed",err)
// 	}
// 	opt := &Option{
// 		MagicNumber:99,
// 		CodecType:"xiaoming"
// 	}
// 	if err := json.NewEncoder(conn).Encode(opt); err != nil {
// 		fmt.Println("rpc client: options error: ", err)
// 		_ = conn.Close()
// 	}

// }