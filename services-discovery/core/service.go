package core

import (
	"math/rand"
	"strings"
	"sync"
)

type InstanceInfo struct {
	AppKey string `json:"app_key"`
	ID     string `json:"id"`
	IP     string `json:"ip"`
	Port   int    `json:"port"`
}

type instancePool struct {
	sync.RWMutex
	// <name,<id,node>>
	instances map[string]map[string]*InstanceInfo
}

func newNodePool() (m *instancePool) {
	return &instancePool{
		instances: map[string]map[string]*InstanceInfo{},
	}
}

func (n *instancePool) addInstance(node *InstanceInfo) {
	if node == nil {
		return
	}

	n.Lock()
	defer n.Unlock()

	if _, exist := n.instances[node.AppKey]; !exist {
		n.instances[node.AppKey] = map[string]*InstanceInfo{}
	}

	n.instances[node.AppKey][node.ID] = node
}

func (n *instancePool) pick(name string) *InstanceInfo {
	//读锁
	n.RLock()
	defer n.RUnlock()

	if nodes, exist := n.instances[name]; !exist {
		return nil
	} else {
		// 纯随机取节点
		idx := rand.Intn(len(nodes))
		for _, v := range nodes {
			if idx == 0 {
				return v
			}
			idx--
		}
	}
	return nil
}

func (n *instancePool) delInstance(id string) {
	sli := strings.Split(id, "/")
	name := sli[len(sli)-2]

	n.Lock()
	defer n.Unlock()

	if _, exist := n.instances[name]; exist {
		delete(n.instances[name], id)
	}
}
