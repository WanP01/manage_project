package config

import "project-common/kafkas"

var kw *kafkas.KafkaWriter

func InitKafkaWriter() func() {
	kw = kafkas.GetWriter("localhost:9092")
	return kw.Close
}

func SendLog(data []byte) {
	kw.Send(kafkas.LogData{
		Topic: "msproject_log",
		Data:  data,
	})
}
