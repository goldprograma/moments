package middleware

import (
	"github.com/didip/tollbooth"
	"github.com/gin-gonic/gin"
	"gitlab.moments.im/pkg/handlehttp"
)

var limiter = tollbooth.NewLimiter(10, nil)

func TollboothLimiter(ctx *gin.Context) {
	limitErr := tollbooth.LimitByRequest(limiter, ctx.Writer, ctx.Request)
	if limitErr != nil {
		handlehttp.HandleResp429(ctx, "TOO_MUCH_REQUEST")
		ctx.Abort()
		return
	}
	ctx.Next()
}
