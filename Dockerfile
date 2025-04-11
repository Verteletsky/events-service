FROM golang:1.23.4-alpine3.19 AS builder

WORKDIR /app

RUN apk add --no-cache upx

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -extldflags=-static" \
    -gcflags="-trimpath" \
    -asmflags="-trimpath" \
    -o /app/events-service \
    github.com/godev/events-service/cmd

RUN upx --best --lzma /app/events-service

FROM scratch

COPY --from=builder /app/events-service /events-service

WORKDIR /

EXPOSE 8080

CMD ["/events-service"]