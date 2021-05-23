FROM golang:1.13-alpine


ENV PROJECT_PATH=github.com/Frosin/shoplist-telegram-bot

RUN apk add --no-cache git
RUN mkdir -p ${GOPATH}/src/${PROJECT_PATH}
WORKDIR ${GOPATH}/src/${PROJECT_PATH}
RUN git clone https://github.com/Frosin/shoplist-telegram-bot.git .

COPY shoplist-bot.yaml cert.pem key.pem ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.version=develop" -o shoplist-bot .

ENTRYPOINT [ "./shoplist-bot"]
EXPOSE 443