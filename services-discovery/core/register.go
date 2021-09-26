package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	_ttl = 10
)

type Register struct {
	cli     *clientv3.Client
	lease   clientv3.Lease
	leaseId clientv3.LeaseID

	id   string
	info *InstanceInfo

	closeChan chan error
}

func NewRegister(info *InstanceInfo, conf clientv3.Config) (reg *Register, err error) {
	r := &Register{}
	r.closeChan = make(chan error)
	r.id = "discovery/" + info.AppKey + "/" + info.ID
	r.info = info
	r.cli, err = clientv3.New(conf)
	return r, err
}

func (r *Register) Run() {
	dur := time.Second
	timer := time.NewTicker(dur)
	r.register()

	for {
		select {
		case <-timer.C:
			r.keepAlive()
		case <-r.closeChan:
			goto EXIT
		}
	}

EXIT:
	log.Printf("[Register] Run exit...")
}

func (r *Register) register() error {
	r.lease = clientv3.NewLease(r.cli)
	leaseResp, err := r.lease.Grant(context.TODO(), _ttl)
	if err != nil {
		err = errors.Wrapf(err, "[Register] register Grant err")
		return err
	}

	kv := clientv3.NewKV(r.cli)
	data, _ := json.Marshal(r.info)
	_, err = kv.Put(context.TODO(), r.id, string(data), clientv3.WithLease(leaseResp.ID))
	if err != nil {
		err = errors.Wrapf(err, "[Register] register kv.Put err %s-%+v", r.id, string(data))
		return err
	}

	r.leaseId = leaseResp.ID
	return nil
}

func (r *Register) keepAlive() error {
	_, err := r.lease.KeepAliveOnce(context.TODO(), r.leaseId)
	if err != nil {
		// 租约丢失，重新注册
		if err == rpctypes.ErrLeaseNotFound {
			r.register()
			err = nil
		} else {
			err = errors.Wrapf(err, "[Register] keepAlive err")
		}
	}

	log.Printf(fmt.Sprintf("[Register] keepalive... leaseId:%+v", r.leaseId))
	return err
}

func (r *Register) Stop() {
	r.revoke()
	close(r.closeChan)
}

func (r *Register) revoke() error {
	// 撤销
	_, err := r.cli.Revoke(context.TODO(), r.leaseId)
	if err != nil {
		err = errors.Wrapf(err, "[Register] revoke err")
		return err
	}

	log.Printf(fmt.Sprintf("[Register] revoke node:%+v", r.leaseId))
	return nil
}
