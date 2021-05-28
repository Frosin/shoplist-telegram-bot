FROM golang:1.16-alpine

ENV GO111MODULE=off
ENV PROJECT_PATH=github.com/Frosin/shoplist-telegram-bot

RUN apk add --no-cache git
RUN apk add --no-cache build-base
RUN apk add --no-cache sqlite

RUN mkdir -p ${GOPATH}/src/${PROJECT_PATH}
WORKDIR ${GOPATH}/src/${PROJECT_PATH}
COPY . .

#COPY shoplist-bot.yaml cert.pem key.pem ./
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o shoplist-bot .

ENTRYPOINT [ "./shoplist-bot"]
EXPOSE 443