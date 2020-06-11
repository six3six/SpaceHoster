GRPC_FLAGS := \
    -I. \
    -I/usr/local/include \
    -I$(GOPATH)/src \
    -I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis

protoc -I $GOPATH/src/ -I . *.proto --go_opt=paths=source_relative --go_out=plugins=grpc:.
protoc -I $GOPATH/src/ -I . *.proto --grpc-gateway_out=logtostderr=true,paths=source_relative:.
protoc -I $GOPATH/src/ -I . *.proto --swagger_out=logtostderr=true:../swagger