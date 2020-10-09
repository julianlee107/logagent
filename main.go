package main

import (
	"github.com/julianlee107/logagent/conf"
	"github.com/julianlee107/logagent/etcd"
	"github.com/julianlee107/logagent/kafka"
	"github.com/julianlee107/logagent/logger"
	"github.com/julianlee107/logagent/taillog"
	"path"
	"sync"
	"time"
)

var (
	cfg = &conf.AppConfig
	wg  sync.WaitGroup
)

func main() {
	// 初始化配置文件列表
	conf.InitViperConf("dev")
	// 初始化配置
	conf.Init()
	// 加载日志文件
	err := logger.Init(path.Join((*cfg).Log.FilePath, (*cfg).Log.FileName), (*cfg).Log.LogLevel, time.Duration((*cfg).Log.MaxAge)*time.Hour*24)
	if err != nil {
		logger.Log.Warnf("初始化日志文件失败, err:%v\n", err)
		return
	}

	err = kafka.Init([]string{cfg.Kafka.KafkaAddr}, cfg.Kafka.ChanMaxSize)
	if err != nil {
		logger.Log.Errorf("init kafka failed,err:%v\n", err)
		return
	}
	logger.Log.Info("init kafka success")

	err = etcd.Init(cfg.Etcd.Address, time.Duration(cfg.Etcd.Timeout)*time.Second)
	if err != nil {
		logger.Log.Errorf("init etcd failed,err:%v\n", err)
		return
	}
	logger.Log.Info("init etcd success")

	// 2.1 从etcd中获取日志收集项的配置信息
	logEntryConf, err := etcd.GetConf(cfg.Etcd.Key)
	if err != nil {
		logger.Log.Errorf("get conf from etcd failed,err:%v\n", err)
		return
	}
	logger.Log.Infof("get conf from etcd success, %v\n", logEntryConf)
	for index, value := range logEntryConf {
		logger.Log.Debugf("index:%v value:%v\n", index, value)
	}

	// 3. 收集日志发往Kafka
	taillog.Init(logEntryConf)
	// 因为NewConfChan访问了tskMgr的newConfChan, 这个channel是在taillog.Init(logEntryConf) 执行的初始化
	newConfChan := taillog.NewConfChan() // 从taillog包中获取对外暴露的通道
	wg.Add(1)
	go etcd.WatchConf(cfg.Etcd.Key, newConfChan) // 哨兵发现最新的配置信息会通知上面的那个通道
	wg.Wait()
}
