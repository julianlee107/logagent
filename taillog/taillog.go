package taillog

import (
	"context"
	"github.com/astaxie/beego/logs"
	"github.com/hpcloud/tail"
	"github.com/julianlee107/logagent/kafka"
)

type TailTask struct {
	path      string
	topic     string
	instance  *tail.Tail
	ctx       context.Context
	cancelFun context.CancelFunc
}

func NewTailTask(path, topic string) (task *TailTask) {
	ctx, cancel := context.WithCancel(context.Background())
	task = &TailTask{
		path:      path,
		topic:     topic,
		ctx:       ctx,
		cancelFun: cancel,
	}
	// 根据路径去打开对应的日志
	task.init()
	return
}

func (t *TailTask) init() {
	config := tail.Config{
		Location:  &tail.SeekInfo{Offset: 0, Whence: 1}, // Whence 0表示相对于文件的原点，1表示相对于当前偏移量，2表示相对于结束。
		ReOpen:    true,                                 // 重新打开
		Follow:    true,                                 // 是否跟随
		MustExist: false,                                // 文件不存在不报错
		Poll:      true,
	}

	var err error
	t.instance, err = tail.TailFile(t.path, config)
	if err != nil {
		logs.Warn("tail file failed,err:%v", err)
	}
	go t.run()
}

func (t *TailTask) run() {
	for {
		select {
		case <-t.ctx.Done():
			logs.Info("tail task %s_%s finished...", t.path, t.topic)
			return
		case line := <-t.instance.Lines:
			logs.Info("get log data from %s success, log:%v\n", t.path, line.Text)
			kafka.SendToChan(t.topic, line.Text)
		}
	}
}
