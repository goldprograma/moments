module gitlab.moments.im/dbsvc

go 1.13

require (
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/jmoiron/sqlx v1.3.5
	gitlab.moments.im/pkg v0.0.0
	go.uber.org/zap v1.21.0
	google.golang.org/grpc v1.46.2
)

replace (
	gitlab.moments.im/pkg v0.0.0 => ../pkg
	gitlab.moments.im/pkg/handlehttp => ../pkg/handlehttp
	gitlab.moments.im/pkg/middleware => ../pkg/middleware
)
