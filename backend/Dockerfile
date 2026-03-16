# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o serverdock ./cmd/server/

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app
COPY --from=builder /app/serverdock .

RUN mkdir -p /app/data

EXPOSE 8000

CMD ["./serverdock"]
