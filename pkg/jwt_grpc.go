package pkg

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials"
)

type jwt_gprc struct {
	token string
}

// NewFromTokenFile return a jwt credential
func NewJwt() credentials.PerRPCCredentials {

	return jwt_gprc{token: "bearer some_good_token zhangjie789798"}
}

func (j jwt_gprc) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": j.token,
	}, nil
}

func (j jwt_gprc) RequireTransportSecurity() bool {
	return false
}
