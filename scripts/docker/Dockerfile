# Build stage
FROM golang:1.12.3 AS builder

ARG GOPROXY=http://artifactory.ep.chehejia.com

ENV GOPROXY=$GOPROXY

ENV APP_NAME=golang-starter

WORKDIR /$APP_NAME
COPY . .
RUN /bin/bash scripts/build.sh

# Production stage
FROM reg.chehejia.com/chj-k8s-change/public/scratch-alpine:3.8

COPY --from=builder /go/bin/$APP_NAME /chj/app/$APP_NAME

ENV WORKDIR=/chj/app/

CMD ["/chj/app/$APP_NAME start"]
