package pkg

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/bwmarrin/snowflake"

	//"github.com/go-redis/redis/v7"
	"github.com/go-redis/redis"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type BaseComponent struct {
	Cfg    *AppConfig
	DB     *sqlx.DB
	Log    *zap.Logger
	Redis  *redis.ClusterClient
	AliOSS *oss.Client
	Node   *snowflake.Node
}
type ComponentsType int

const (
	Component_S3 ComponentsType = iota
	Component_REDIS
	Component_DB
	Component_FASHDFS
	Component_ALIOSS
	Component_ES
)

//RegisterComponent 初始化基础服务
func (bs *BaseComponent) RegisterComponent(confPath string, registerComponents ...ComponentsType) {
	//初始化配置文件
	// InitCfg("./app.toml")
	bs.SetUpSetting(confPath)
	bs.SetUpZapLogger() //日志必须
	for _, registerComponent := range registerComponents {
		switch registerComponent {
		case Component_REDIS:
			bs.SetUpRedisClient()
		case Component_DB:
			bs.SetUpDatabase()
		case Component_FASHDFS:
			bs.SetUpFastDFSClinet()
		case Component_ALIOSS:
			bs.SetUpAliOssClient()
		default:
			log.Fatalln("不支持的组件注册")
		}
	}

}

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

//GetGRPCConn 获取GRPC连接
func (bs *BaseComponent) GetGRPCConn(target, serviceName string) *grpc.ClientConn {

	bs.Log.Info("Dependence im name", zap.String("dep", serviceName), zap.String("url", bs.Cfg.GRPCDepends[serviceName]))
	// b := RegisterResolver(addr, target, serviceName)
	// conn, err := grpc.Dial(bs.Cfg.GRPCDepends[serviceName],
	// 	grpc.WithMaxMsgSize(20*1024*1024),
	// 	grpc.WithTimeout(time.Second*50),
	// 	grpc.WithKeepaliveParams(kacp),
	// 	grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name), grpc.WithPerRPCCredentials(NewJwt()), grpc.WithDisableRetry(), grpc.WithDisableHealthCheck(),
	// 	grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
	// 	grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
	// )

	//
	grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name))
	conn, err := grpc.Dial(bs.Cfg.GRPCDepends[serviceName],
		grpc.WithMaxMsgSize(20*1024*1024),
		grpc.WithTimeout(time.Second*50),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithInsecure(), grpc.WithPerRPCCredentials(NewJwt()), grpc.WithDisableRetry(), grpc.WithDisableHealthCheck(),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
	)

	if err != nil {
		log.Panic("initExternalDependence", zap.String("error", err.Error()))
	}

	return conn
}

//StartGRPCServer 开启GRPC 服务
func (bs *BaseComponent) StartGRPCServer(port int, fun func(*grpc.Server)) error {

	// basic_lib.Register(p.ConfigService.Etcd, basic_lib.GetConfig().ServiceName, basic_lib.GetConfig().GRPCPort)
	//grpc中间件初始化
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_recovery.UnaryServerInterceptor(RecoverInit(bs.Log)...),
		)),
		grpc.MaxConcurrentStreams(10000),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	grpc_prometheus.Register(s)

	defer s.GracefulStop()
	fun(s)
	// moment.RegisterMomentDBServiceServer(s, db)
	// Register  Service
	reflection.Register(s)
	//GRpc 启动
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("--------------2333--------------@> err : ", err)
		bs.Log.Error("failed to listen:", zap.String("error", err.Error()))
		return err
	}

	bs.Log.Info(bs.Cfg.ServiceName+" grpc server start", zap.Int("lis_addr", bs.Cfg.GRPCPort))

	if err := s.Serve(lis); err != nil {
		bs.Log.Error("failed to serve start:", zap.String("err", err.Error()))
	}
	fmt.Println("-------------------------------@> err : ", err)
	return err
}

func StructToMap(obj interface{}) map[string]interface{} {
	obj1 := reflect.TypeOf(obj)
	obj2 := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < obj1.Elem().NumField(); i++ {
		data[obj1.Elem().Field(i).Name] = obj2.Elem().Field(i).Interface()
	}
	return data
}

func (bs *BaseComponent) initNode(nodeID int64) {
	var err error
	bs.Node, err = snowflake.NewNode(nodeID)
	if err != nil {
		log.Fatalln("初始化IDNode出错", err)
	}
	log.Printf("初始化NodeID成功,NodeID = %d \n", nodeID)
}

var once sync.Once

// GetID 生成ID
func (bs *BaseComponent) GetID() int64 {
	// if bs.Node == nil {
	// 	once.Do(func() {
	// 		bs.initNode(bs.Cfg.NodeID)
	// 	})
	// }
	return bs.Node.Generate().Int64()
}
