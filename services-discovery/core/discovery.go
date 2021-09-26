package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Discovery struct {
	cli  *clientv3.Client
	pool *instancePool
}

func NewDiscovery(conf clientv3.Config) (dis *Discovery, err error) {
	d := &Discovery{}
	d.pool = newNodePool()
	d.cli, err = clientv3.New(conf)
	return d, err
}

func (d *Discovery) Pull() {
	kv := clientv3.NewKV(d.cli)

	resp, err := kv.Get(context.TODO(), "discovery/", clientv3.WithPrefix())
	if err != nil {
		log.Fatalf("[Discovery] kv.Get err:%+v", err)
		return
	}

	for _, v := range resp.Kvs {
		node := &InstanceInfo{}
		err = json.Unmarshal(v.Value, node)
		if err != nil {
			log.Fatalf("[Discovery] json.Unmarshal err:%+v", err)
			continue
		}

		d.pool.addInstance(node)
		log.Printf("[Discovery] pull node:%+v:%+v", string(v.Key), node)
	}
}

func (d *Discovery) Watch() {
	watcher := clientv3.NewWatcher(d.cli)
	watchChan := watcher.Watch(context.TODO(), "discovery", clientv3.WithPrefix())
	for {
		select {
		case resp := <-watchChan:
			d.watchEvent(resp.Events)
		}
	}
}

func (d *Discovery) watchEvent(evs []*clientv3.Event) {
	for _, ev := range evs {
		switch ev.Type {
		case clientv3.EventTypePut:
			node := &InstanceInfo{}
			err := json.Unmarshal(ev.Kv.Value, node)
			if err != nil {
				log.Fatalf("[Discovery] json.Unmarshal err:%+v", err)
				continue
			}

			d.pool.addInstance(node)
			log.Printf("[Discovery] watch node:%+v:%+v", string(ev.Kv.Key), node)
		case clientv3.EventTypeDelete:
			d.pool.delInstance(string(ev.Kv.Key))
			log.Printf(fmt.Sprintf("[Discovery] del node:%s data:%s", string(ev.Kv.Key), string(ev.Kv.Value)))
		}
	}
}
func (d *Discovery) Pick(appKey string) *InstanceInfo {
	return d.pool.pick(appKey)
}

func (d *Discovery) Dump() {
	for k, v := range d.pool.instances {
		for kk, vv := range v {
			log.Printf("[Discovery] Dump,Name:%s Id:%s Node:%+v", k, kk, vv)
		}
	}
}
