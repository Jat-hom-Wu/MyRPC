package main

import(
	"fmt"
	"net"
	"encoding/json"
	"time"
)

type Option struct {
	MagicNumber int        // MagicNumber marks this's a geerpc request
	CodecType   string // client may choose different Codec to encode body
}

func main(){
	client()	
}

func client(){
	fmt.Println("client")
	conn,err := net.Dial("tcp", "127.0.0.1:9527")
	conn.Write([]byte("asdfasdf"))
	if err != nil{
		fmt.Println("dial failed",err)
	}
	opt := Option{
		MagicNumber:99,
		CodecType:"xiaoming",
	}
	opt2 := &Option{
		MagicNumber:199,
		CodecType:"xiaoming",
	}
	time.Sleep(3*time.Second)
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		fmt.Println("rpc client: options error: ", err)
		_ = conn.Close()
	}
	if err := json.NewEncoder(conn).Encode(opt2); err != nil {
		fmt.Println("rpc client: options error: ", err)
		_ = conn.Close()
	}
	if err := json.NewEncoder(conn).Encode("asdfasdf"); err != nil {
		fmt.Println("rpc client: options error: ", err)
		_ = conn.Close()
	}
	for{}
}