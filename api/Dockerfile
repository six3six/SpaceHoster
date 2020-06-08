FROM golang

WORKDIR /api
COPY go.mod go.mod
COPY version.go .

RUN go get -d -v .
COPY . .
RUN go install -v .
CMD ["api"]