package conf

import (
	"bytes"
	"github.com/astaxie/beego/logs"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var ViperConfMap map[string]*viper.Viper
var AppConfig AppConf

type AppConf struct {
	Kafka KafkaConf
	Log   LogConf
}

type KafkaConf struct {
	KafkaAddr   string `mapstructure:"kafka_addr"`
	ChanMaxSize int    `mapstrucrure:"chan_max_size"`
}

type LogConf struct {
	FilePath string `mapstructure:"log_path"`
	FileName string `mapstructure:"log_files"`
	LogLevel string `mapstructure:"log_level"`
	MaxAge   int64  `mapstructure:"max_age"`
}

func Init() {
	if err := ViperConfMap["app"].Unmarshal(&AppConfig); err != nil {
		logs.Error("parse base config failed,config:%s,err:%v", err)
	}
}

func InitViperConf(mode string) error {
	currentWD, err := os.Getwd()
	if err != nil {
		logs.Error("get current config dir error: %v", err)
		return err
	}
	path := filepath.Join(currentWD, "conf", mode)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		logs.Error("read config dir error: %v", err)
	}
	for _, f := range files {
		if !f.IsDir() {
			conf, err := ioutil.ReadFile(filepath.Join(path, f.Name()))
			if err != nil {
				logs.Error("read config file error: %v", err)
				return err
			}
			v := viper.New()
			confType := strings.Split(f.Name(), ".")
			v.SetConfigType(confType[len(confType)-1])
			v.ReadConfig(bytes.NewBuffer(conf))
			if ViperConfMap == nil {
				ViperConfMap = make(map[string]*viper.Viper)
			}
			ViperConfMap[confType[0]] = v
		}
	}

	return nil
}
