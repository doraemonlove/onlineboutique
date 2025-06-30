FROM golang:1.14.7 AS builder

WORKDIR /build/src
ENV GOPROXY=https://proxy.golang.com.cn,direct
COPY . .

RUN go build -tags netgo -o locust_exporter

FROM alpine as release

COPY --from=builder /build/src/locust_exporter /locust_exporter

EXPOSE 9646

ENTRYPOINT ["/locust_exporter"]
