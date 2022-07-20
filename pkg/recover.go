package pkg

import (
	"fmt"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.uber.org/zap"
)

//异常恢复
func RecoverInit(zlog *zap.Logger) []grpc_recovery.Option {
	return []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(recoverImplement(zlog)),
	}
}

func recoverImplement(zlog *zap.Logger) grpc_recovery.RecoveryHandlerFunc {

	return func(p interface{}) (err error) {

		//开发测试中抛出，上线后在开启恢复机制
		/**
		.....恢复逻辑
		*/
		zlog.Error("program exception:", zap.String("error", fmt.Sprintf("%v", p)))
		return
	}
}
