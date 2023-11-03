package kafkas

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

type LogData struct {
	Topic string
	Data  []byte //可以存入json数据
}

type KafkaWriter struct {
	w        *kafka.Writer
	dataChan chan LogData //日志消息通道
}

func GetWriter(addr string) *KafkaWriter {
	w := &kafka.Writer{
		Addr:     kafka.TCP(addr),
		Balancer: &kafka.LeastBytes{},
	}
	k := &KafkaWriter{
		w:        w,
		dataChan: make(chan LogData, 100),
	}
	go k.sendKafka() // 持续的go协程监控是否有日志发出
	return k
}

func (w *KafkaWriter) Send(data LogData) {
	w.dataChan <- data
}

func (w *KafkaWriter) Close() {
	if w.w != nil {
		w.w.Close()
	}
}

func (w *KafkaWriter) sendKafka() {
	for {
		select {
		case data := <-w.dataChan:
			messages := []kafka.Message{
				{
					Topic: data.Topic,
					Key:   []byte("logMsg"),
					Value: data.Data,
				},
			}
			var err error
			const retries = 3

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			for i := 0; i < retries; i++ {
				// attempt to create topic prior to publishing the message
				err = w.w.WriteMessages(ctx, messages...)
				if errors.Is(err, kafka.LeaderNotAvailable) || errors.Is(err, context.DeadlineExceeded) {
					log.Printf("kafka send writemessage err %s \n", err.Error())
					//time.Sleep(time.Millisecond * 250)
					continue
				}

				if err != nil {
					log.Printf("kafka send writemessage err %s \n", err.Error())
					continue
				}

				break
			}
		}
	}
}
