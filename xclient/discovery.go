package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const(
	RandomSelect SelectMode = iota
	RoundRobinSelect
)

type Discovery interface {
	Get(mode SelectMode) (string,error)
	GetAll() ([]string, error)
}

type MultiServerDiscovery struct {
	r       *rand.Rand
	mu      sync.Mutex
	servers []string
	index   int
}

func NewMultiServerDiscovery(servers []string) *MultiServerDiscovery {
	m := &MultiServerDiscovery{
		r:       rand.New(rand.NewSource(time.Now().UnixNano())), //一个随机数
		servers: servers,
	}
	m.index = m.r.Intn(math.MaxInt32 - 1) //返回一个在设定区间内的随机数int值
	return m
}

func (m *MultiServerDiscovery)Get(mode SelectMode) (string,error){
	m.mu.Lock()
	defer m.mu.Unlock()
	length := len(m.servers)
	if length == 0{
		return "", errors.New("rpc discovery failed: no available server")
	}
	switch mode{
	case RandomSelect:
		n := m.r.Intn(length)
		return m.servers[n],nil
	case RoundRobinSelect:
		//循环法需要用锁
		s := m.servers[m.index % length]
		m.index = (m.index + 1) % length
		return s, nil
	default:
		return "",errors.New("rpc discovery failed, input invailed select mode")
	}
}

func (m *MultiServerDiscovery)GetAll() ([]string, error){
	m.mu.Lock()
	defer m.mu.Unlock()
	length := len(m.servers)
	servers := make([]string, length)
	if length == 0{
		return servers,errors.New("rpc discovery failed: no avaliable server")
	}
	copy(servers,m.servers)
	return servers,nil
}
