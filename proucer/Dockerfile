# STEP 1
FROM golang:alpine as builder

RUN adduser -D -g '' appuser
RUN apk add libgcc g++ git gcc
RUN git clone https://github.com/0x0BSoD/kafka.git  $GOPATH/src/github.com/0x0BSoD/kafka
#COPY . $GOPATH/src/github.com/0x0BSoD/kafka/producer
WORKDIR $GOPATH/src/github.com/0x0BSoD/kafka/producer
RUN ls -l && go mod download -x
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/producer

# STEP 2 build a small image
FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/producer /go/bin/producer

#COPY consumer.yml /go/bin/producer.yml

USER appuser
EXPOSE 2112
ENTRYPOINT ["/go/bin/consumer"]