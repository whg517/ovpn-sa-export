# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /ovpn-sa-export ./cmd/ovpn-sa-export

# Runtime stage
FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /ovpn-sa-export /usr/local/bin/ovpn-sa-export

EXPOSE 9176

ENTRYPOINT ["ovpn-sa-export"]
