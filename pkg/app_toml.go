package pkg

import (
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/giantpoplar/pool"
)

type AppConfig struct {
	RunMode       string `toml:"runmode"`
	NodeID        int64
	ServiceName   string   `toml:"serviceName"`
	DB            database `toml:"database"`
	Redis         redisCfg `toml:"redis"`
	Aws           awsCfg   `toml:"aws"`
	Ali           aliCfg   `toml:"ali"`
	HTTPPort      int
	GRPCPort      int
	Logs          logs              `toml:"logs"`
	Etcd          etcd              `toml:"etcd"`
	Elasticsearch elasticSearch     `toml:"elasticsearch"`
	IP            string            `toml:"IP"`
	GrpcService   map[string]string `toml:"grpcservice"`
	Uploader      uploader          `toml:"uploader"`
	Depends       map[string]string `toml:"depends"`
	GRPCDepends   map[string]string `toml:"grpcdepends"`
	PprofAddr     int               `toml:"pprofPort"`
	FastDFS       fastDFS           `toml:"fastDFS"`
	DBLink        *ConfigService
	DBLinkCfg     dblink `toml:"dblink"`
	Robot         robot
	Identity      string
}

type etcd struct {
	Endpoints []string
}
type dblink struct {
	Mysqlr string
	Mysqlw string
	Schema string
	Redis  string
	Etcd   string
}
type fastDFS struct {
	Tracker tracker
	Storage storage
}
type storage struct {
	DownloadSizeLimit int64
	Pool              pool.Config
}
type tracker struct {
	Address []string
	Pool    pool.Config
}
type elasticSearch struct {
	Host  []string
	Index string
	Type  string
}

type redisCfg struct {
	Endpoints []string
}

type uploader struct {
	Mainland []string
	Abroad   []string
}

type cors struct {
	Origin string
}
type database struct {
	Type         string
	DriverName   string `toml:"driverName"`
	MaxOpenConns int    `toml:"maxOpenConns"`
	MaxIdleConns int    `toml:"maxIdleConns"`
	MaxLifetime  int64  `toml:"maxLifetime"`
	Location     string
	Nodes        []node `toml:"node"`
}

type aliCfg struct {
	OSS alioss `toml:"oss"`
}

type alioss struct {
	Host            string
	Scheme          string
	Region          string
	CDNRegion       string `toml:"cdnRegion"`
	AccessKeyID     string `toml:"accessKeyID"`
	AccessKeySecret string `toml:"accessKeySecret"`
	BucketName      string `toml:"bucketName"`
}

type awsCfg struct {
	S3 aws3 `toml:"s3"`
}
type aws3 struct {
	Host            string
	Scheme          string
	AccessKeyID     string `toml:"accessKeyID"`
	SecretAccessKey string `toml:"secretAccessKey"`
	Region          string
	Bucket          string
}
type node struct {
	Server   string
	Port     string
	UserName string `toml:"userName"`
	SID      string
	PassWord string
}

type logs struct {
	FilePath  string
	Level     int
	Formatter string
}

type robot struct {
	Endpoint string
	Token    string
	RoomType int
	RoomID   int
}

//SetUpSetting 初始化配置文件
func (bs *BaseComponent) SetUpSetting(filePath string) {
	var config = &AppConfig{}
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		panic("初始化 app.toml 配置文件失败\n" + err.Error())
	}
	log.Println("初始化 app.toml 配置文件成功", config)
	bs.Cfg = config
	bs.GenNodeID()
}

// GenNodeID  加载环境变量
func (bs *BaseComponent) GenNodeID() {
	var err error

	// nodeIDEnv := os.Getenv("NODE_ID")
	// if nodeIDEnv == "" {
	// 	log.Println("没设置NODE_ID环境变量,系统默认为0,如启动多个程序请设置NODE_ID")
	// 	bs.initNode(0)
	// 	return
	// }
	ip := getLocalIPV4()
	if ip == "" {
		ip = "0"
	}
	log.Println("本机IP", ip)
	ip = strings.ReplaceAll(ip, ".", "")
	var ipInt int64
	if ipInt, err = strconv.ParseInt(ip, 0, 64); err != nil {
		log.Println("解析IP出错,系统默认设置为0", err)
	}

	bs.initNode(ipInt % 1024)
}

//获取本机IP地址
func getLocalIPV4() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalln("获取本机IP出错")
	}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ip.IP.To4() == nil {
				continue
			}
			return ip.IP.To4().String()
		}
	}
	return ""
}
