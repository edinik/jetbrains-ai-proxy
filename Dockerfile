FROM golang:alpine as builder
WORKDIR /app

COPY ./ /app
RUN go build  -o jetbrains-ai-proxy

FROM alpine
LABEL maintainer="zouyq <zyqcn@live.com>"

COPY --from=builder /app/jetbrains-ai-proxy /usr/local/bin/

ENTRYPOINT ["jetbrains-ai-proxy"]


