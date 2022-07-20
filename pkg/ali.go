package pkg

import (
	"bytes"
	"log"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
)

//SetUpAliOssClient 初始化AliOSS 上传
func (bc *BaseComponent) SetUpAliOssClient() {
	var err error
	endpoint := bc.Cfg.Ali.OSS.Scheme + "://" + bc.Cfg.Ali.OSS.Region + "." + bc.Cfg.Ali.OSS.Host
	log.Println("OSS URL:", endpoint)
	var ossClient *oss.Client
	if ossClient, err = oss.New(endpoint, bc.Cfg.Ali.OSS.AccessKeyID, bc.Cfg.Ali.OSS.AccessKeySecret); err != nil {
		log.Panic("获取阿里OSS上传客户端出错", err.Error())
	}
	bc.AliOSS = ossClient

}

//PutAliOSS 文件上传
func (bc *BaseComponent) PutAliOSS(fileName string, fileData *bytes.Buffer) error {
	t := time.Now()
	bc.Log.Debug("文件开始上传AliOSS", zap.String("文件名", fileName), zap.Int("大小", fileData.Len()>>10))
	var bucket *oss.Bucket
	var err error
	if bucket, err = bc.AliOSS.Bucket(bc.Cfg.Ali.OSS.BucketName); err != nil {
		bc.Log.Error("获取阿里OSS Bucket 出错", zap.Error(err))
	}

	if err = bucket.PutObject(fileName, bytes.NewReader(fileData.Bytes()), []oss.Option{oss.ObjectACL(oss.ACLPublicRead)}...); err != nil {
		bc.Log.Error("阿里OSS 文件上传出错", zap.Error(err))
	}
	//上传成功
	bc.Log.Debug("上传成功", zap.Int("大小KB", fileData.Len()>>10), zap.Float64("耗时", time.Now().Sub(t).Seconds()))
	return err
}
