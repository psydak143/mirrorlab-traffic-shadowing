# ===== build stage =====
FROM golang:1.25-alpine AS build
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o proxy .

# ===== runtime stage =====
FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/proxy /app/proxy

ENV LISTEN_ADDR=:8080
EXPOSE 8080

ENTRYPOINT ["/app/proxy"]
