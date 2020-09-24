FROM golang:latest

LABEL maintainer="Shikhar Kannoje <shikharkannoje09@gmail.com>"

WORKDIR /app

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY . .

ENV CACHE_DOMAIN localhost:9000
ENV API_GATE_SERVER localhost:9080
ENV DS_HOSTPORT 5432
ENV DS_HOSTNAME localhost
ENV DS_USERNAME postgres
ENV DS_PASSWORD root
ENV DS_DATABASE employee
ENV DS_SERVERRUN localhost:9000 

RUN go build

CMD ["cachingService.exe", "APIGateway.exe"]