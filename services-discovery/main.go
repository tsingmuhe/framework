package main

import (
	"fmt"
	"time"

	"github.com/tsingmuhe/framework/services-discovery/core"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	dis, _ := core.NewDiscovery(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})

	reg, _ := core.NewRegister(&core.InstanceInfo{
		AppKey: "com.reg1",
		ID:     "1",
		IP:     "ip1",
		Port:   123,
	}, clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})

	reg2, _ := core.NewRegister(&core.InstanceInfo{
		AppKey: "com.reg2",
		ID:     "2",
		IP:     "ip2",
		Port:   123,
	}, clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})

	reg3, _ := core.NewRegister(&core.InstanceInfo{
		AppKey: "com.reg2",
		ID:     "3",
		IP:     "ip2",
		Port:   123,
	}, clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})

	go reg.Run()
	time.Sleep(time.Second * 2)

	dis.Pull()
	go dis.Watch()

	dis.Dump()
	time.Sleep(time.Second * 1)

	go reg2.Run()
	time.Sleep(time.Second * 1)

	go reg3.Run()
	time.Sleep(time.Second * 1)

	fmt.Println(dis.Pick("com.reg2"))
	fmt.Println(dis.Pick("com.reg2"))
	fmt.Println(dis.Pick("com.reg2"))
	fmt.Println(dis.Pick("com.reg2"))
	fmt.Println(dis.Pick("com.reg2"))
	fmt.Println(dis.Pick("com.reg2"))
	time.Sleep(time.Second * 5)

}
