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
	mkdir -p google/api
	curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > google/api/annotations.proto
	curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > google/api/http.proto
	$(PROTOC) google/api/*.proto $(LIB_FLAGS) $(COMPILE) $(CLIENT_FLAGS) $(GRPC_FLAGS)
	$(PROTOC) $(LIB_FLAGS) $(GRPC_GW_FLAGS) $(COMPILE)
	$(PROTOC) $(LIB_FLAGS) $(GRPC_SWAGGER_FLAGS) $(COMPILE)


clean:
	rm google -rf
	rm *.go
	rm ../swagger/*.json