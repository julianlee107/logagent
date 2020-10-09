package etcd

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"github.com/julianlee107/logagent/logger"
	"time"
)

var (
	client    *clientv3.Client
	DeleteSig = make(chan *clientv3.Event, 10)
)

type LogEntry struct {
	Path  string `json:"path"`  // 日志存放路径
	Topic string `json:"topic"` // kafka的topic
}

func Init(addr string, timeout time.Duration) (err error) {
	client, err = clientv3.New(
		clientv3.Config{
			Endpoints:   []string{addr},
			DialTimeout: timeout,
		})
	if err != nil {
		return
	}

	return
}

// 获取配置信息
func GetConf(key string) (LogEntryConf []*LogEntry, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := client.Get(ctx, key)
	cancel()
	if err != nil {
		return
	}
	for _, conf := range resp.Kvs {
		err = json.Unmarshal(conf.Value, &LogEntryConf) // 会自动append进去
		if err != nil {
			return
		}
	}
	return
}

func WatchConf(key string, newConfCh chan<- []*LogEntry) {
	ch := client.Watch(context.Background(), key)
	for resp := range ch {
		for _, event := range resp.Events {
			logger.Log.Infof("Type:%v Key:%v Value:%v \n", event.Type, string(event.Kv.Key), string(event.Kv.Value))
			var newConf []*LogEntry
			if event.Type != clientv3.EventTypeDelete {
				err := json.Unmarshal(event.Kv.Value, &newConf)
				if err != nil {
					logger.Log.Warnf("unmarshal failed,err:%v\n", err)
					continue
				}
			} else {
				DeleteSig <- event
			}
			logger.Log.Infof("Get new conf :%v \n", newConf)
			newConfCh <- newConf
		}
	}
}
