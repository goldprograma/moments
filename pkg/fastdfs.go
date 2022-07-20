package pkg

import (
	"errors"

	"github.com/giantpoplar/fdfs/cluster"
	"go.uber.org/zap"
)

//SetUpFastDFSClinet 初始化FastDFS
func (bs *BaseComponent) SetUpFastDFSClinet() {
	if err := cluster.Init(bs.Cfg.FastDFS.Tracker.Address, cluster.TrackerConfig{}, cluster.StorageConfig{}); err != nil {
		bs.Log.Error("初始化FastDFS出错", zap.Error(err))
	}
	bs.Log.Debug("初始化FastDFS客户端成功")
}

//UploadFile 上传文件
func UploadFile(fileData []byte, groupName string, ext string) (string, error) {
	// 上传文件到storage组g1,并指定返回文件后缀为jpg

	fid, err := cluster.Upload(fileData, groupName, ext)
	if detailedErr, ok := err.(*cluster.Error); ok && detailedErr != nil {
		err = errors.New(detailedErr.Error())
		return "", err
	}
	return fid, nil
}

//DownloadFile 下载文件
func DownloadFile(fid string) ([]byte, error) {
	// 上传文件到storage组g1,并指定返回文件后缀为jpg
	file, err := cluster.Download(fid)
	if detailedErr, ok := err.(*cluster.Error); ok && detailedErr != nil {
		err = errors.New(detailedErr.Error())
		return nil, err
	}
	return file, nil
}

//DeleteFile 删除文件
func DeleteFile(fid string) error {
	// 上传文件到storage组g1,并指定返回文件后缀为jpg
	err := cluster.Delete(fid)
	if detailedErr, ok := err.(*cluster.Error); ok && detailedErr != nil {
		err = errors.New(detailedErr.Error())
		return err
	}
	return nil
}
