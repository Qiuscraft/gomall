# syntax=docker/dockerfile:1
FROM golang:1.23 AS builder
WORKDIR /workspace
COPY go.work go.work
COPY go.work.sum go.work.sum
COPY . .
WORKDIR /workspace/app/frontend
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /workspace/bin/frontend ./

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/* && mkdir -p /app/bin
COPY --from=builder /workspace/bin/frontend /app/bin/frontend
COPY --from=builder /workspace/app/frontend/conf /app/conf
COPY --from=builder /workspace/app/frontend/static /app/static
COPY --from=builder /workspace/app/frontend/template /app/template
COPY --from=builder /workspace/app/frontend/types /app/types
ENV GO_ENV=k8s
ENV OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://jaeger:4317
ENTRYPOINT ["/app/bin/frontend"]
