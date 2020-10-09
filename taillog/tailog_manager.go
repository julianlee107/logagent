package taillog

import (
	"fmt"
	"github.com/julianlee107/logagent/etcd"
	"github.com/julianlee107/logagent/logger"
	"time"
)

var taskManager *tailLogManager

type tailLogManager struct {
	taskMap     map[string]*TailTask
	logEntry    []*etcd.LogEntry
	newConfChan chan []*etcd.LogEntry
}

func Init(logEntryConf []*etcd.LogEntry) {
	taskManager = &tailLogManager{
		taskMap:     make(map[string]*TailTask, 10),
		logEntry:    logEntryConf,
		newConfChan: make(chan []*etcd.LogEntry),
	}
	for _, log := range logEntryConf {
		tailObj := NewTailTask(log.Path, log.Topic)
		mark := fmt.Sprintf("%s_%s", log.Path, log.Topic)
		taskManager.taskMap[mark] = tailObj
	}
	go taskManager.run()
}

func (t *tailLogManager) run() {
	for {
		select {
		case newConf := <-t.newConfChan:
			for _, conf := range newConf {
				mark := fmt.Sprintf("%s_%s", conf.Path, conf.Topic)
				_, ok := t.taskMap[mark]
				if ok {
					// 原有配置不操作
					continue
				} else {
					tailTask := NewTailTask(conf.Path, conf.Topic)
					t.taskMap[mark] = tailTask
				}
			}
			for deleteConf := range etcd.DeleteSig {
				mark := fmt.Sprintf("%s_%s", string(deleteConf.Kv.Key), string(deleteConf.Kv.Value))
				t.taskMap[mark].cancelFun()
			}
			logger.Log.Debug("new conf arrived")
		default:
			time.Sleep(time.Second)
		}

	}
}

func NewConfChan() chan<- []*etcd.LogEntry {
	return taskManager.newConfChan
}
