package main

import (
	"github.com/julianlee107/logagent/conf"
	"github.com/julianlee107/logagent/etcd"
	"github.com/julianlee107/logagent/kafka"
	"github.com/julianlee107/logagent/logger"
	"path"
	"time"
)

var (
	cfg = &conf.AppConfig
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
}
