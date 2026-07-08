# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /acsgo .

# ---- Runtime ----
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -s /bin/false acsgo

COPY --from=builder /acsgo /usr/local/bin/acsgo

RUN mkdir -p /data && chown acsgo:acsgo /data

USER acsgo
WORKDIR /data

ENV CWMP_PORT=7547
ENV PORT=7548
ENV DB_TYPE=sqlite
ENV DB_URL=/data/acsgo.db

EXPOSE 7547 7548

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
  CMD wget -qO- http://localhost:7548/api/health || exit 1

ENTRYPOINT ["/usr/local/bin/acsgo"]
