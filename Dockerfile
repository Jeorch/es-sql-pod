# builder 源镜像
FROM golang:1.12.4-alpine as builder

# 安装git
RUN apk add --no-cache git

ENV GOPROXY="https://goproxy.cn"

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build

# prod 源镜像
FROM alpine:latest as prod

RUN apk --no-cache add ca-certificates

# 程序配置的环境变量可在部署服务时设置(yaml文件中)
#ENV PROJECT_NAME="es-sql-pods"
#ENV BP_LOG_TIME_FORMAT="2006-01-02 15:04:05"
#ENV BP_LOG_OUTPUT="console"
#ENV BP_LOG_LEVEL="info"
#ENV ES_SERVER="http://59.110.31.215:9200"

WORKDIR /app

COPY --from=0 /app/es-sql-pods .

EXPOSE 3000
ENTRYPOINT ["/app/es-sql-pods"]