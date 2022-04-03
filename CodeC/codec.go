package CodeC

import (
	"io"
	"encoding/gob"
	"bufio"
	"fmt"
)



type Header struct {
	ServiceMethod string
	Seq uint64
	Error string
}

type Codec interface{
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewCodecMode func(io.ReadWriteCloser) Codec //template

const(
	GobType string = "gob"
	JsonType string = "json"
)

var NewCodecModeMap map[string]NewCodecMode

func init(){
	NewCodecModeMap = make(map[string]NewCodecMode)
	NewCodecModeMap[GobType] = NewGobCodec
}

//gob implement
type GobCodec struct{
	dec *gob.Decoder
	enc *gob.Encoder
	buf *bufio.Writer
	conn io.ReadWriteCloser
}

var _ Codec = (*GobCodec)(nil) //attension

func NewGobCodec(conn io.ReadWriteCloser) Codec{
	buffer := bufio.NewWriter(conn)	//!
	return &GobCodec{
		conn : conn,
		dec : gob.NewDecoder(conn),
		enc : gob.NewEncoder(buffer),
		buf : buffer,
	}
}

func (c *GobCodec) ReadHeader(header *Header) error{
	return c.dec.Decode(header)
}
func (c *GobCodec) ReadBody(body interface{}) error{
	return c.dec.Decode(body)
}
//encode 
func (c *GobCodec) Write(header *Header, body interface{}) (err error){
	defer func(){
		_ = c.buf.Flush()
		if err != nil{
			_ = c.Close()
		}
	}()
	if err = c.enc.Encode(header); err != nil{
		fmt.Println("gob error encoding header:", err)
		return
	}
	if err = c.enc.Encode(body); err != nil{
		fmt.Println("gob error encoding body:", body)
		return
	}
	return
}
func (c *GobCodec) Close() error{
	return c.conn.Close()
}
