package pkg

import (
	"fmt"
	"log"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (b *BaseComponent) SetUpZapLogger() {
	var conf zap.Config
	var err error
	var zapLog *zap.Logger
	conf = zap.NewProductionConfig()
	if b.Cfg.Logs.FilePath != "" {
		conf.OutputPaths = append(conf.OutputPaths, b.Cfg.Logs.FilePath+b.Cfg.ServiceName+".log")
	}

	conf.EncoderConfig.EncodeTime = TimeEncoder
	conf.Level = zap.NewAtomicLevelAt(zapcore.Level(b.Cfg.Logs.Level))
	if zapLog, err = conf.Build(); err != nil {
		log.Fatalln("logger init fail", err)
	}
	logger := zapLog.Named(b.Cfg.ServiceName)

	logger.Debug("logger init success")
	b.Log = logger
}

func (b *BaseComponent) sendRotMsg(e zapcore.Entry) error {
	if b.Cfg.RunMode == "prd" && e.Level > zapcore.InfoLevel {
		fmt.Println(sendRotMsg(b.Cfg.Robot, b.Cfg.RunMode+":"+e.LoggerName+":"+e.Message+":"+strings.ReplaceAll(strings.ReplaceAll(e.Stack, "\n", ""), "\t", "")))
	}
	return nil
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}
