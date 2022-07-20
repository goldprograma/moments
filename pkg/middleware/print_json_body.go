package middleware

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PrintPostBody(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.Next()
		return
	}

	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Println("PrintPostBody Error ->", err)
		return
	}
	log.Println(string(body))
	//写回body到Request
	ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	ctx.Next()
}
