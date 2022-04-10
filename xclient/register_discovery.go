package xclient

import (
	"strings"
	"fmt"
	"net/http"
	"time"
)

type RegisterDiscovery struct{
	*MultiServerDiscovery
	registryAddr string
	duration time.Duration
	lastUpdate time.Time
}

const DefaultUpdateTime = 10 * time.Second

func NewRegisterDiscovery(registerAddr string, duration time.Duration) *RegisterDiscovery{
	if duration == 0{
		duration = DefaultUpdateTime
	}
	return &RegisterDiscovery{
		registryAddr:registerAddr,
		duration:duration,
		MultiServerDiscovery:NewMultiServerDiscovery(make([]string, 0)),
	}
}

func (d *RegisterDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}

//客户端更新servers列表使用
func (d *RegisterDiscovery) Refresh() error{
	if d.lastUpdate.Add(DefaultUpdateTime).After(time.Now()){
		return nil
	}
	resp,err := http.Get(d.registryAddr)
	if err != nil{
		fmt.Println("client registry Get falied:",err)
	}
	respStr := resp.Header.Get("X-Myrpc-Servers")
	respSlice := strings.Split(respStr,",")
	d.MultiServerDiscovery.servers = make([]string, 0, len(respSlice))
	for _,serverAddr := range respSlice{
		if strings.TrimSpace(serverAddr) != ""{
			d.servers = append(d.servers, strings.TrimSpace(serverAddr))
		}
	}
	d.lastUpdate = time.Now()
	return nil
}

func (d *RegisterDiscovery) Get(mode SelectMode) (string,error){
	err := d.Refresh()
	if err != nil{
		return "",err
	}
	return d.MultiServerDiscovery.Get(mode)
}

func (d *RegisterDiscovery) GetAll() ([]string, error){
	err := d.Refresh()
	if err != nil{
		return nil,err
	}
	return d.MultiServerDiscovery.GetAll()
}



