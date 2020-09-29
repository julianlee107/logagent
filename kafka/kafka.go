package kafka

import (
	"github.com/Shopify/sarama"
	"time"
)

type logData struct {
	topic string
	data  string
}

var (
	client  sarama.SyncProducer // 声明一个全局生产者
	logChan chan *logData
)

func Init(addrs []string, maxSize int) (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner // 新选出⼀个 partition
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel返回
	client, err = sarama.NewSyncProducer(addrs, config)
	if err != nil {
		return
	}
	logChan = make(chan *logData, maxSize)
	// 开启后台的goroutine从通道中取取数据发往kafka
	go sendToKafka()
	return
}

// SendToChan 给外部暴露的一个函数，该函数只把日志数据发送到一个内部的channel中
func SendToChan(topic, data string) {
	msg := &logData{
		topic: topic,
		data:  data,
	}
	logChan <- msg
}

func sendToKafka() {
	for {
		select {
		case l := <-logChan:
			msg := &sarama.ProducerMessage{}
			msg.Topic = l.topic
			msg.Value = sarama.StringEncoder(l.data)

			//	发送到kafka
			_, _, err := client.SendMessage(msg)
			if err != nil {
				return
			}
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}
