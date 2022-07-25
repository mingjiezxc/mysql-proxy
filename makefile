.PHONY: all build run clean alpine docker help

BINARY="mysql-proxy"

all: build

k8s: docker apply

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BINARY}

alpine:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -tags  netgo -o ${BINARY} 

docker:
	docker build . -t ${BINARY}:${Version}

apply:
	kubectl apply -f deploy/k8s/

run:
	@go run ./

clean:
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

help:
	@echo "make - 格式化 Go 代码, 并编译生成二进制文件"
	@echo "make build - 编译 Go 代码, 生成二进制文件"
	@echo "make alpine - 编译 Go 代码, 生成 alpine 二进制文件"
	@echo "make run - 直接运行 Go 代码"
	@echo "make clean - 移除二进制文件和 vim swap files"
	@echo "make docker - 构建 docker image 'make docker -e Version=v1' "

