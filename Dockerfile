# build stage
FROM golang:1.26.4-alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build \
    -ldflags="-s -w" \
    -o doc-api-service .

# runtime stage
FROM alpine:3.22

RUN apk add --no-cache tzdata

ENV TZ=Asia/Jakarta

RUN addgroup -g 1001 binarygroup && \
    adduser -D -u 1001 -G binarygroup userapp

WORKDIR /app 

COPY --from=builder --chown=userapp:binarygroup /app/doc-api-service .

USER userapp

EXPOSE 4040

CMD ["./doc-api-service"]