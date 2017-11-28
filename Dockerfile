FROM alpine:3.6

LABEL maintainer "henrik@loodse.com"

RUN apk add -U ca-certificates && rm -rf /var/cache/apk/*

ADD _output/nodeport-exposer /nodeport-exposer

ENTRYPOINT ["/nodeport-exposer"]
