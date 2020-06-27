FROM ysicing/gobuild as builder

ENV GO111MODULE=on

ENV GOPROXY=https://goproxy.cn

ENV TZ=Asia/Shanghai

COPY . /go/src

WORKDIR /go/src

RUN go build -o ginws

FROM ysicing/golang

COPY --from=builder /go/src/ginws /usr/local/bin/

RUN chmod +x /usr/local/bin/ginws

EXPOSE 9001

ENTRYPOINT /usr/local/bin/ginws

