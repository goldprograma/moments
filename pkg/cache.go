package pkg

import (
	"log"

	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

//SetUpRedisClient 初始化Redis
func (b *BaseComponent) SetUpRedisClient() {

	b.Log.Info("redis配置", zap.Any("Nodes", b.Cfg.DBLink.Redis.Nodes))

	redisCli := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    b.Cfg.DBLink.Redis.Nodes,
		Password: b.Cfg.DBLink.Redis.Password,
	})

	if _, err := redisCli.Ping().Result(); nil != err {
		log.Panic("Redis连接失败,请检查linkCfg配置文件\n", b.Cfg.Redis, err.Error())
	}
	b.Redis = redisCli
}
