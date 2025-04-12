FROM golang:1.23.3 AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o main cmd/risken-mcp-server/main.go

FROM alpine:3.20
COPY --from=builder /app/main /usr/local/bin/
RUN apk --no-cache add ca-certificates
ENTRYPOINT ["/usr/local/bin/main"]
