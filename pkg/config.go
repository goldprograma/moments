package pkg

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

// 初始化一个http client
var hc = http.Client{
	Transport: &http.Transport{
		Dial: func(n, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(n, addr, time.Second*10)
			if err != nil {
				return nil, err
			}
			c.SetDeadline(time.Now().Add(time.Second * 20))
			return c, nil
		},
	},
}

// 发起http请求
func PostHttp(req *http.Request) ([]byte, error) {
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// 获取系统环境变量
func GetEnv(key string) string {
	return os.Getenv(key)
}

// 配置文件获取结构
type GetConfig struct {
	Server   string // 访问的ip  // 不需要传
	Tenant   string // 租客信息  // 不需要传
	DataId   string // 配置文件id // 不需要传
	Group    string // 组id // 不需要传
	SavePath string // 文件保存路径(不包含文件名)
	Suffix   string // 文件后缀
}

// 保存文件
func (config *GetConfig) SaveFile(data []byte) error {
	// 如果没有后缀，默认yml
	if config.Suffix == "" {
		config.Suffix = ".yml"
	}
	// 如果文件夹不为空
	if config.SavePath != "" {
		if _, err := os.Stat(config.SavePath); os.IsNotExist(err) {
			// 先创建文件夹
			err := os.MkdirAll(config.SavePath, 0777)
			if err != nil {
				return err
			}
			// 再修改权限
			err = os.Chmod(config.SavePath, 0777)
			if err != nil {
				return err
			}
		}
		if config.SavePath[len(config.SavePath)-1:] != "/" {
			config.SavePath = config.SavePath + "/"
		}
	}
	return ioutil.WriteFile(config.SavePath+config.DataId+config.Suffix, data, 0666) //写入文件(字节数组)
}

// 获取配置文件信息 saveDir=文件夹不包含文件名 suffix=文件后缀
func GetConfigFile(saveDir string, suffix string) error {
	conf := &GetConfig{
		Server: GetEnv("SERVER"),
		Tenant: GetEnv("TENANT"),
		DataId: GetEnv("DATA_ID"),
		Group:  GetEnv("GROUP")}
	if saveDir != "" {
		conf.SavePath = saveDir
	}
	if suffix != "" {
		conf.Suffix = suffix
	}
	log.Printf("%+v\n", conf)
	u := "http://" + conf.Server + "/nacos/v1/cs/configs"
	urlr, err := url.Parse(u)
	if err != nil {
		log.Println(err)
		return err
	}
	// 封装http Req
	req := &http.Request{
		URL: urlr,
	}
	q := urlr.Query()
	q.Add("dataId", conf.DataId)
	q.Add("group", conf.Group)
	q.Add("tenant", conf.Tenant)
	req.URL.RawQuery = q.Encode()
	data, err := PostHttp(req)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("%v\n", string(data))
	return conf.SaveFile(data)
}
