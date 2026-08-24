FROM golang:1.23.12

ENV GOPROXY=off GOSUMDB=off
ENV ENVMONITOR_ADDR=0.0.0.0:7790
WORKDIR /app

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .

RUN go build -mod=vendor ./...

EXPOSE 7790
CMD ["go", "run", "-mod=vendor", "./cmd/envmonitor"]
