FROM golang:1.25.1 AS builder

WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/personal-page-be .

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/personal-page-be /app/personal-page-be

EXPOSE 9101

CMD ["./personal-page-be"]
