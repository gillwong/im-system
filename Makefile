pre:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install github.com/cloudwego/kitex/tool/cmd/kitex@latest
	go install github.com/cloudwego/thriftgo@latest

generate:
	mkdir -p ./http-server/proto_gen
	protoc -I=. --go_out=./http-server/proto_gen ./idl_http.proto
	cd http-server && kitex -module github.com/gillwong/im-system/http-server ../idl_rpc.thrift
	cd rpc-server && kitex -module github.com/gillwong/im-system/rpc-server ../idl_rpc.thrift
