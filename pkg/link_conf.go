package pkg

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/gob"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
)

type (
	ConfigService struct {
		MysqlRead  Mysql
		MysqlWrite Mysql
		Common     Mysql
		Jaeger     Jaeger
		Redis      Redis
		Kafka      Kafka
		S3CA       S3CA
		Temp       Temp
		Mongo      Mongo
		Etcd       Etcd
		Server     Server
	}
	Mysql struct {
		User   string
		Pwd    string
		Ip     string
		Port   string
		Schema string
	}
	Jaeger struct {
		Address string
	}
	Redis struct {
		Nodes    []string
		Password string
	}

	Mongo struct {
		Url []string
	}
	Kafka struct {
		Address []string
		Flag    bool
	}
	S3CA struct {
		Credentials string
	}
	Temp struct {
		Path string
	}
	Etcd struct {
		Ips  []string
		TTL  int64
		Keep int64
	}
	Server struct {
		Comment   Comment
		Contact   Contact
		Operating Operating
		Upload    Upload
		Download  Download
	}
)

// func LoadConfig(filePath string) *ConfigService {
// 	var config ConfigService
// 	if filePath != "" {
// 		if _, err := toml.DecodeFile(filePath, &config); err != nil {
// 			panic(err)
// 		}
// 		return &config
// 	}
// 	return &config
// }

// ProdLoadConfig : 加载线上环境配置
// 		1. 读取 link_infomation 库
// 		2. 查询所需配置
// 		3. 解密
// 		4. 将解密结果作为配置内容返回
func ProdLoadConfig(mysql Mysql, link dblink) *ConfigService {

	var config ConfigService

	dataSourceName := mysql.User + ":" + mysql.Pwd +
		"@tcp(" + mysql.Ip + ":" + mysql.Port + ")/" + mysql.Schema +
		"?charset=utf8mb4&parseTime=true&loc=Local"
	log.Println("linkinfo 配置", dataSourceName)
	conn, err := sqlx.Open("mysql", dataSourceName)
	if nil != err {
		log.Fatalf("DB Instance initialize error,errMsg='%s'", err.Error())
	}
	err = conn.Ping()
	if err != nil {
		log.Fatalf("conn.Ping() error: %#v\n", err)
	}
	defer conn.Close()

	var (
		// watchSync  MysqlConfig
		rmysqlConf MysqlConfig
		wmysqlConf MysqlConfig
		redisConf  RedisConfig
		// kafkaConf  KafkaConfig
		etcdConf EtcdConfig
		// mongoConf  MongoConfig
		// ptCommon   MysqlConfig
	) //pt_wbasic pt_basictomoments
	// err = GetMysqlInfo(conn, "pt_basictomoments", &watchSync)
	// if nil != err {
	// 	log.Fatal(err)
	// }
	// err = GetMysqlInfo(conn, "pt_common", &ptCommon)
	// if nil != err {
	// 	log.Fatal(err)
	// }
	err = GetMysqlInfo(conn, link.Mysqlr, &rmysqlConf)
	if nil != err {
		log.Fatalf("GetMysqlInfo error: %#v\n", err)
	}

	err = GetMysqlInfo(conn, link.Mysqlw, &wmysqlConf)
	if nil != err {
		log.Fatalf("GetMysqlInfo error: %#v\n", err)
	}

	err = GetRedisInfo(conn, link.Redis, &redisConf)
	if nil != err {
		log.Fatalf("GetRedisInfo error: %#v\n", err)
	}
	// err = GetKafkaInfo(conn, "pt_moments_kafka", &kafkaConf)
	// if nil != err {
	// 	log.Fatalf("GetKafkaInfo error: %#v\n", err)
	// }
	err = GetEtcdInfo(conn, link.Etcd, &etcdConf)
	if nil != err {
		log.Fatalf("GetEtcdInfo error: %#v\n", err)
	}
	// err = GetMongoInfo(conn, "pt_moments_mongodb", &mongoConf)
	// if nil != err {
	// 	log.Fatalf("GetMongoInfo error: %#v\n", err)
	// }

	return &ConfigService{
		MysqlRead: Mysql{
			User:   rmysqlConf.UserName,
			Pwd:    rmysqlConf.UserPasswd,
			Ip:     strings.Split(rmysqlConf.HostInfo, ":")[0],
			Port:   strings.Split(rmysqlConf.HostInfo, ":")[1],
			Schema: link.Schema,
		},
		MysqlWrite: Mysql{
			User:   wmysqlConf.UserName,
			Pwd:    wmysqlConf.UserPasswd,
			Ip:     strings.Split(wmysqlConf.HostInfo, ":")[0],
			Port:   strings.Split(wmysqlConf.HostInfo, ":")[1],
			Schema: link.Schema,
		},
		// Common: Mysql{
		// 	User:   ptCommon.UserName,
		// 	Pwd:    ptCommon.UserPasswd,
		// 	Ip:     strings.Split(ptCommon.HostInfo, ":")[0],
		// 	Port:   strings.Split(ptCommon.HostInfo, ":")[1],
		// 	Schema: "potato_common",
		// },
		Redis: Redis{
			Nodes:    strings.Split(redisConf.HostInfo, ","),
			Password: redisConf.Passwd,
		},
		// Kafka: Kafka{
		// 	Address: strings.Split(kafkaConf.KafkaInfo, ","),
		// 	Flag:    true,
		// },
		// Jaeger: Jaeger{config.Jaeger.Address},
		// S3CA:   config.S3CA,
		// Temp:   config.Temp,
		// Mongo: Mongo{
		// 	Url: strings.Split(mongoConf.HostInfo, ","),
		// },
		Etcd: Etcd{
			Ips:  strings.Split(etcdConf.HostInfo, ","),
			TTL:  20,
			Keep: 5,
		},
		Server: config.Server,
	}

}

type Comment struct {
	BaseConfig
}
type Contact struct {
	BaseConfig
}
type Operating struct {
	BaseConfig
}
type Upload struct {
	BaseConfig
}
type Download struct {
	BaseConfig
}
type BaseConfig struct {
	GrpcPort  string
	DebugPort string
	Address   string
	Name      string
	Dep       map[string]string
	Tarcer    bool
	Kafka     bool
	Auth      bool
	LogOut    string
	LogEnv    string
}
type MongoConfig struct {
	UserName   string `db:"user_name" json:"userName"`
	UserPasswd string `db:"user_passwd" json:"userPasswd"`
	HostInfo   string `db:"host_info" json:"hostInfo"`
}

type EtcdConfig struct {
	HostInfo string `db:"host_info" json:"hostInfo"`
}

type KafkaConfig struct {
	KafkaInfo     string `db:"kafka_info" json:"kafkaInfo"`
	ZookeeperInfo string `db:"zookeeper_info" json:"zookeeperInfo"`
}

type RedisConfig struct {
	HostInfo string `db:"host_info" json:"hostInfo"`
	Passwd   string `db:"passwd" json:"passwd"`
}

type NatsConfig struct {
	UserName string `db:"user_name" json:"userName"`
	Passwd   string `db:"passwd" json:"passwd"`
	HostInfo string `db:"host_info" json:"hostInfo"`
}

type MysqlConfig struct {
	UserName   string `db:"user_name" json:"userName"`
	UserPasswd string `db:"user_passwd" json:"userPasswd"`
	HostInfo   string `db:"host_info" json:"hostInfo"`
}

func (cs *ConfigService) Serialize() (string, error) {
	//构造缓冲区
	buf := bytes.NewBuffer(nil)
	//生成gob编码器
	g := gob.NewEncoder(buf)
	//编码
	err := g.Encode(cs) /*参数值类型或指针类型都可以*/
	return hex.EncodeToString(buf.Bytes()), err
}

func (cs *ConfigService) Deserialize(data string) error {
	byts, _ := hex.DecodeString(data)
	//构造阅读器
	r := bytes.NewReader(byts)
	//构造gob解码器
	dg := gob.NewDecoder(r)
	//解码
	return dg.Decode(cs)
}

func GetEtcdInfo(db *sqlx.DB, platform string, out *EtcdConfig) error {
	var err error
	findCmd := fmt.Sprintf(`SELECT host_info FROM link_information.etcd_info WHERE platform='%s'`, platform)
	err = db.Get(out, findCmd)
	if nil != err {
		return err
	}

	out.HostInfo, err = RsaDecrypt(out.HostInfo)
	if nil != err {
		return err
	}

	return err
}

func GetRedisInfo(db *sqlx.DB, platform string, out *RedisConfig) error {
	var err error

	findCmd := fmt.Sprintf(`SELECT host_info, passwd FROM link_information.redis_info WHERE platform = '%s'`, platform)
	err = db.Get(out, findCmd)
	if nil != err {
		return err
	}

	out.Passwd, err = RsaDecrypt(out.Passwd)
	if nil != err {
		return err
	}
	out.HostInfo, err = RsaDecrypt(out.HostInfo)
	if nil != err {
		return err
	}

	return err
}

func GetMysqlInfo(db *sqlx.DB, platform string, out *MysqlConfig) error {
	var err error

	findCmd := fmt.Sprintf(`SELECT user_name, user_passwd, host_info FROM link_information.mysql_info t WHERE t.platform = '%s'`, platform)
	err = db.Get(out, findCmd)
	if nil != err {
		return err
	}

	out.HostInfo, err = RsaDecrypt(out.HostInfo)
	if nil != err {
		return err
	}
	out.UserName, err = RsaDecrypt(out.UserName)
	if nil != err {
		return err
	}
	out.UserPasswd, err = RsaDecrypt(out.UserPasswd)
	if nil != err {
		return err
	}

	return err
}

func GetMongoInfo(db *sqlx.DB, platform string, out *MongoConfig) error {
	var err error

	findCmd := fmt.Sprintf(`SELECT user_name, user_passwd, host_info FROM link_information.mongo_info WHERE platform='%s'`, platform)
	err = db.Get(out, findCmd)
	if nil != err {
		return err
	}

	out.UserName, err = RsaDecrypt(out.UserName)
	if nil != err {
		return err
	}
	out.UserPasswd, err = RsaDecrypt(out.UserPasswd)
	if nil != err {
		return err
	}
	out.HostInfo, err = RsaDecrypt(out.HostInfo)
	if nil != err {
		return err
	}

	return err
}

func GetNatsInfo(db *sqlx.DB, platform string, out *NatsConfig) error {
	var err error

	findCmd := fmt.Sprintf(`SELECT user_name, passwd, host_info FROM link_information.nats_info WHERE platform = '%s'`, platform)
	err = db.Get(out, findCmd)
	if nil != err {
		return err
	}

	out.Passwd, err = RsaDecrypt(out.Passwd)
	if nil != err {
		return err
	}
	out.HostInfo, err = RsaDecrypt(out.HostInfo)
	if nil != err {
		return err
	}
	out.UserName, err = RsaDecrypt(out.UserName)
	if nil != err {
		return err
	}

	return err
}

func GetKafkaInfo(db *sqlx.DB, platform string, out *KafkaConfig) error {
	var err error

	findCmd := fmt.Sprintf(`SELECT kafka_info, zookeeper_info FROM link_information.kafka_info WHERE platform = '%s'`, platform)
	err = db.Get(out, findCmd)
	if nil != err {
		return err
	}

	out.KafkaInfo, err = RsaDecrypt(out.KafkaInfo)
	if nil != err {
		return err
	}
	out.ZookeeperInfo, err = RsaDecrypt(out.ZookeeperInfo)
	if nil != err {
		return err
	}

	return err
}

func RsaDecrypt(str string) (string, error) {
	ciphertext := []byte(str)
	block, _ := pem.Decode(privateKey) //将密钥解析成私钥实例
	if block == nil {
		return "", fmt.Errorf("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes) //解析pem.Decode（）返回的Block指针实例
	if err != nil {
		return "", err
	}
	rst, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext) //RSA算法解密
	if nil != err {
		return "", err
	}
	return string(rst), nil
}

var privateKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
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
B2/X8LVUbWjJ
-----END RSA PRIVATE KEY-----
`)
