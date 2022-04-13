package registry

import(
	"fmt"
	"time"
	"net/http"
	"sync"
	"strings"
)

const (
	specificServerS = "X-Myrpc-Servers"
	specificServer = "X-Myrpc-Server"

	registerDefaultTimeout = 5 * time.Minute
	regiterDefaultPath = "/myrpc/register"
)

type Registy struct{
	Timeout time.Duration
	mu sync.Mutex
	ServersAddr map[string]*Serveritem
}

type Serveritem struct{
	Addr string
	Start time.Time
}

func (r *Registy)PutServer(addr string){
	r.mu.Lock()
	defer r.mu.Unlock()
	server := r.ServersAddr[addr]
	if server == nil{
		r.ServersAddr[addr] = &Serveritem{
			Addr:addr,
			Start:time.Now(),
		}
	}else{
		r.ServersAddr[addr].Start = time.Now()
	}
}

func (r *Registy)AliveServer() []string{
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr,serveritem := range r.ServersAddr{
		if r.Timeout == 0 || serveritem.Start.Add(r.Timeout).After(time.Now()){
			//没超时
			alive = append(alive, addr)
		}else{
			delete(r.ServersAddr, addr)
		}
	}
	return alive
}

func (r *Registy)ServeHTTP(w http.ResponseWriter, req *http.Request){
	method := req.Method
	switch method{
	case "GET":
		//返回所有alive的server
		var resp []string
		resp = r.AliveServer()
		respStr := strings.Join(resp, ",")
		w.Header().Set(specificServerS, respStr)
	case "POST":
		//注册or更新start
		addr := req.Header.Get(specificServer)
		if addr == ""{
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("register: get invailed address")
			return
		}
		r.PutServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Registy)HandleHTTP(registerPath string){
	http.Handle(registerPath, r)
	fmt.Println("rpc registry path:", registerPath)
}

func HandleHTTP() {
	DefaultRegister.HandleHTTP(regiterDefaultPath)
}

func HeartBeat(registerAddr, serverAddr string, duration time.Duration){
	if duration == 0{
		duration = registerDefaultTimeout - time.Duration(1)*time.Minute
	}
	err := sendHeartBeat(registerAddr, serverAddr)
	go func(){
		ticker := time.NewTicker(duration)
		for err != nil{
			<-ticker.C
			err = sendHeartBeat(registerAddr,serverAddr)
		}
	}()
}

func sendHeartBeat(registerAddr, serverAddr string) error{
	client := http.Client{}
	req,_ := http.NewRequest("POST", registerAddr, nil)
	req.Header.Set(specificServer, serverAddr)
	_,err := client.Do(req)
	if err != nil{
		fmt.Println("server send to register failed:",err)
		return err
	}
	fmt.Println(serverAddr, "send heart beat to registry", registerAddr)
	return nil
}

var DefaultRegister = NewRegister(registerDefaultTimeout)

func NewRegister(duration time.Duration) *Registy{
	return &Registy{
		Timeout:duration,
		ServersAddr:make(map[string]*Serveritem),
	}
}

