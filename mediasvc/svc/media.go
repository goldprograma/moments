package svc

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"mime/multipart"
	"gitlab.moments.im/pkg"
	"strings"

	"github.com/gin-gonic/gin"
)

type MediaService struct {
	pkg.BaseComponent
}

//DownloadAliOSS 下载服务
// @Summary 下载
// @Tags 媒体
// @Param name path string false "文件名称"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /mediasvc/download/{name} [get]
func (ms *MediaService) DownloadAliOSS(c *gin.Context) {
	// var err error
	// var fileDataSlice string
	// var fileData string
	fileName := c.Param("name")

	// var hasCache int64
	if fileName != "" {
		//检查redis是否存在，如果存在则从redis取返回给客户端
		// if hasCache, err = p.Redis.Exists(FilePerfixKey + fileName).Result(); err != nil {
		// 	Response(c, false, "500", "检测文件"+fileName+"缓存是否存在失败", err.Error())
		// 	return
		// }
		// if hasCache > 0 {
		// 	t := time.Now()
		// 	p.Log.Debug("redis存在文件直接返回")
		// 	// fileData := new(bytes.Buffer)

		// 	fileInfo := strings.Split(strings.Split(fileName, ".")[0], "_")
		// 	var fileSlice int64
		// 	if fileSlice, err = strconv.ParseInt(fileInfo[2], 0, 64); err != nil {
		// 		p.Log.Error("解析文件大小失败", zap.Error(err))
		// 		return
		// 	}
		// 	fmt.Println("fileInfo=", fileInfo, "fileSlice", fileSlice)

		// 	for index := 1; index < int(fileSlice)+1; index++ {
		// 		p.Log.Debug("获取文件分片", zap.String("文件名", FilePerfixKey+fileName), zap.Int("索引", index))
		// 		if fileDataSlice, err = p.Redis.HGet(FilePerfixKey+fileName, strconv.Itoa(index)).Result(); err != nil {
		// 			Response(c, false, "500", "获取文件"+fileName+"第"+strconv.Itoa(index)+"缓存失败", err.Error())
		// 			return
		// 		}
		// 		c.Writer.WriteString(fileDataSlice)
		// 	}

		// 	p.Log.Debug("返回文件", zap.Float64("耗时", time.Now().Sub(t).Seconds()))

		// 	c.Writer.Flush()
		// 	return
		// }
		// var bucket *oss.Bucket
		// if bucket, err = p.AliOss.Bucket(basic_lib.GetConfig().Ali.OSS.BucketName); err != nil {
		// 	Response(c, false, "500", "获取Ali OSS Bucket失败", err.Error())
		// 	return
		// }
		// if body, err := bucket.GetObject(fileName); err != nil {
		// 	Response(c, false, "500", "获取Ali OSS Bucket fileName 失败", err.Error())
		// 	return
		// } else {
		// 	defer body.Close()
		// 	data, _ := ioutil.ReadAll(body)
		// 	c.Writer.Write(data)
		// }
		//走阿里CDN
		aliOSS := ms.Cfg.Ali.OSS
		var urlStr = aliOSS.Scheme + "://" + aliOSS.BucketName + "." + aliOSS.CDNRegion + "." + aliOSS.Host + "/" + fileName
		// p.Log.Debug("转发文件请求", zap.String("url", urlStr))
		// var removte, err = url.Parse(urlStr)
		// director := func(req *http.Request) {
		// 	req.URL = removte
		// 	req.Proto = "HTTP/1.1"
		// 	req.ProtoMajor = 1
		// 	req.ProtoMinor = 1
		// 	req.Host = removte.Host
		// }

		// errorHandler := func(write http.ResponseWriter, request *http.Request, err error) {
		// 	fmt.Println(err)
		// }
		// proxy := &httputil.ReverseProxy{Director: director, ErrorHandler: errorHandler}
		// proxy.ServeHTTP(c.Writer, c.Request)
		// Response(c, true, "302", "请前往新地址下载文件", fileName)
		c.Redirect(301, urlStr)
	}
}

//Upload 上传服务
// @Summary 文件上传
// @Tags 媒体
// @Accept multipart/form-data
// @Param file formData file true "file"
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /mediasvc/upload [post]
func (ms *MediaService) Upload(c *gin.Context) {
	var err error
	var fileHeader *multipart.FileHeader
	// var thumbnailHeader *multipart.FileHeader

	if fileHeader, err = c.FormFile("file"); err != nil {
		ms.Response(c, "APPVERSION_UPLOAD", err, "获取文件失败", nil)
		return
	}
	names := strings.Split(fileHeader.Filename, ".")
	if len(names) < 2 {
		ms.Response(c, "APPVERSION_UPLOAD", err, "获取文件格式错误", nil)
		return
	}

	if fileHeader.Size < 1 {
		ms.Response(c, "APPVERSION_UPLOAD", err, "文件数据为空", nil)
		return
	}

	var f multipart.File
	buffer := make([]byte, fileHeader.Size)
	if f, err = fileHeader.Open(); err != nil {
		ms.Response(c, "APPVERSION_UPLOAD", err, "文件打开", nil)
		return
	}
	defer f.Close()
	f.Read(buffer)

	h := sha1.New()
	h.Write(buffer)
	hash := hex.EncodeToString(h.Sum(nil))
	name := hash + "." + names[len(names)-1]
	err = ms.PutAliOSS(name, bytes.NewBuffer(buffer))
	//文件上传
	ms.Response(c, "APPVERSION_UPLOAD", err, "文件上传", gin.H{"Name": name, "FileName": fileHeader.Filename, "Hash": hash})

}
