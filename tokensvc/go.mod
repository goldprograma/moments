module gitlab.moments.im/tokensvc

go 1.13

require (
	github.com/gin-gonic/gin v1.7.7
	github.com/prometheus/client_golang v1.12.2
	gitlab.moments.im/pkg v0.0.0
	go.uber.org/zap v1.21.0
	google.golang.org/grpc v1.46.2
)

replace gitlab.moments.im/pkg v0.0.0 => ../pkg
