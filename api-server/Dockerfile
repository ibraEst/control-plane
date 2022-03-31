FROM golang:1.18-alpine as builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /root/go/src/control-plane/api
COPY . .
RUN go get -d -v ./...
RUN go build -o api-server .


FROM scratch  
WORKDIR /root/
COPY --from=builder /root/go/src/control-plane/api .
CMD ["./api-server"]  
