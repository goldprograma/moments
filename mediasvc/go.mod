module gitlab.moments.im/mediasvc

go 1.13

require (
	github.com/gin-gonic/gin v1.7.7
	github.com/goinggo/mapstructure v0.0.0-20140717182941-194205d9b4a9
	github.com/prometheus/client_golang v1.12.2
	gitlab.moments.im/pkg v0.0.0
)

replace gitlab.moments.im/pkg v0.0.0 => ../pkg
