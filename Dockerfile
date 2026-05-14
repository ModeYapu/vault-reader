# Stage 1: Build
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /vault-reader ./cmd/vault-reader

# Stage 2: Runtime
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /vault-reader /usr/local/bin/vault-reader

EXPOSE 3000

ENTRYPOINT ["vault-reader"]
