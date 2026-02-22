FROM golang:1.25-alpine

RUN apk add --no-cache git

WORKDIR /app

ENV GOFLAGS=-mod=mod

RUN go install github.com/air-verse/air@latest

COPY go.mod ./
RUN go mod download || true

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
