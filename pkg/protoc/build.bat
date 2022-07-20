protoc-go-inject-tag -input=./moment/notice.pb.go
protoc-go-inject-tag -input=./moment/comment.pb.go
protoc-go-inject-tag -input=./moment/follow.pb.go
protoc-go-inject-tag -input=./moment/forum.pb.go
protoc-go-inject-tag -input=./moment/media.pb.go
protoc-go-inject-tag -input=./moment/notice.pb.go
protoc-go-inject-tag -input=./moment/thumb.pb.go
protoc-go-inject-tag -input=./moment/topic.pb.go
protoc-go-inject-tag -input=./moment/base.pb.go
protoc-go-inject-tag -input=./moment/token.pb.go

protoc --proto_path=./imapigateway/ -I=./moment --gogofast_out=plugins=grpc:./moment/   ./moment/xxxxx.proto