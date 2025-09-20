FROM alpine:3.17
LABEL authors="xiangyt"
WORKDIR /go/bin
ENV TZ=Asia/Shanghai
RUN apk add tzdata && cp /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone
COPY ./cmd/worker/main /go/bin/worker
VOLUME /go/bin/configs
RUN chmod +x /go/bin/worker
CMD ./worker