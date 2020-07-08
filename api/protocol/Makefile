LIB_FLAGS := \
    -I. \
    -I./google/api \
    -I/usr/local/include \
    -I$(GOPATH)/src \

GRPC_FLAGS := --go_out=plugins=grpc:.
GRPC_GW_FLAGS := --grpc-gateway_out=logtostderr=true,paths=source_relative:.
GRPC_SWAGGER_FLAGS := --swagger_out=logtostderr=true:../swagger

CLIENT_FLAGS := \
    --go_opt=paths=source_relative

COMPILE := \
	*.proto \

PROTOC = protoc

all:
	$(PROTOC) $(LIB_FLAGS) $(COMPILE) $(CLIENT_FLAGS) $(GRPC_FLAGS)

clean:
	rm google -rf
	rm *.go
	rm ../swagger/*.json