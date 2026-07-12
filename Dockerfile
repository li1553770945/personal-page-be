ARG GO_BUILDER_IMAGE=docker.io/library/golang:1.25.12-bookworm@sha256:a9c020ee3d1508c7be5435c262434e3d3fc1d0e76a11afeb9ddae7d60bc86aa4

FROM ${GO_BUILDER_IMAGE} AS builder

WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY main.go router.go router_gen.go ./
COPY biz ./biz
RUN install -d -m 1777 /runtime-root/tmp && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /app/personal-page-be .

FROM scratch

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /runtime-root /
COPY --from=builder /app/personal-page-be /app/personal-page-be

EXPOSE 9101

ENTRYPOINT ["/app/personal-page-be"]
