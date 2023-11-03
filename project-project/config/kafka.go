package config

import (
	"context"
	"go.uber.org/zap"
	"project-common/kafkas"
	"project-project/internal/dao"
	"project-project/internal/repo"
	"time"
)

var kw *kafkas.KafkaWriter

func InitKafkaWriter() func() {
	kw = kafkas.GetWriter(AppConf.Kc.Addr[0])
	return kw.Close
}

func SendLog(data []byte) {
	kw.Send(kafkas.LogData{
		Topic: "msproject_log",
		Data:  data,
	})
}

func SendCache(data []byte) {
	kw.Send(kafkas.LogData{
		Topic: "msproject_cache",
		Data:  data,
	})
}

type KafkaCache struct {
	R     *kafkas.KafkaReader
	cache repo.Cache
}

func NewCacheReader() *KafkaCache {
	reader := kafkas.GetReader(AppConf.Kc.Addr, AppConf.Kc.Group, AppConf.Kc.Topic)
	return &KafkaCache{
		R:     reader,
		cache: dao.Rc,
	}
}

func (c *KafkaCache) DeleteCache() {
	for {
		message, err := c.R.R.ReadMessage(context.Background())
		if err != nil {
			zap.L().Error("DeleteCache ReadMessage err", zap.Error(err))
			continue
		}
		zap.L().Info("收到缓存", zap.String("value", string(message.Value)))
		if string(message.Value) != "" { //如果kafka获得的message value为 task，则说明是需要在redis 中拿取 string(message.Value)
			fields, err := c.cache.HKeys(context.Background(), string(message.Value))
			if err != nil {
				zap.L().Error("DeleteCache HKeys err", zap.Error(err))
				continue
			}
			time.Sleep(1 * time.Second) //延时删除
			c.cache.Delete(context.Background(), fields)
		}
	}
}
