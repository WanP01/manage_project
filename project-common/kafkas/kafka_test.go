package kafkas

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"testing"
	"time"
)

func TestKafkaProducer(t *testing.T) {
	// to produce messages
	topic := "my-topic"
	partition := 0

	//连接kafka的leader
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	//设置超时时间
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	//发送消息
	_, err = conn.WriteMessages(
		kafka.Message{Value: []byte("one!")},
		kafka.Message{Value: []byte("two!")},
		kafka.Message{Value: []byte("three!")},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	// MQ 关闭
	if err := conn.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}

func TestKafkaConsumer(t *testing.T) {
	// to consume messages
	topic := "my-topic"
	partition := 0

	//连接kafka的leader
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	//设置超时时间
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	//从kafka读取数据
	batch := conn.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max

	b := make([]byte, 10e3) // 10KB max per message
	for {
		n, err := batch.Read(b)
		if err != nil {
			break
		}
		fmt.Println(string(b[:n]))
	}

	if err := batch.Close(); err != nil {
		log.Fatal("failed to close batch:", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close connection:", err)
	}
}

func TestProducer(t *testing.T) {
	k := GetWriter("localhost:9092")
	m := make(map[string]string)
	m["projectCode"] = "1200"
	bytes, _ := json.Marshal(m)
	k.Send(LogData{
		Topic: "msproject_log",
		Data:  bytes,
	})
	time.Sleep(2 * time.Second)
}

func TestConsumer(t *testing.T) {
	GetReader([]string{"localhost:9092"}, "group1", "msproject_log")
	for {
	}
}
