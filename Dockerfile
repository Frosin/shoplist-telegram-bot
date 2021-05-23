FROM golang:1.16-alpine

ENV GO111MODULE=off
ENV PROJECT_PATH=github.com/Frosin/shoplist-telegram-bot

RUN apk add --no-cache git
RUN mkdir -p ${GOPATH}/src/${PROJECT_PATH}
WORKDIR ${GOPATH}/src/${PROJECT_PATH}
COPY . .

#COPY shoplist-bot.yaml cert.pem key.pem ./
RUN CGO_ENABLED=0 GOOS=linux go build -o shoplist-bot .

ENTRYPOINT [ "./shoplist-bot"]
EXPOSE 443