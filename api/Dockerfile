FROM golang

RUN go get -u github.com/uber/prototool/cmd/prototool

WORKDIR /api
COPY go.mod go.mod
COPY version.go .

RUN go get -d -v .

RUN go install github.com/golang/protobuf/protoc-gen-go

ENV PATH "$PATH:$(go env GOPATH)/bin"

COPY protocol /api/protocol
COPY swagger /api/swagger
RUN cd /api/protocol && make clean && make all

RUN cd /api/
COPY . .
RUN go install -v .

CMD ["api"]