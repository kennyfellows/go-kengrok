FROM golang:1.22.1-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN make proxyserver

FROM alpine:latest

# Install sqlite and its dependencies
RUN apk add --no-cache \
    sqlite \
    sqlite-dev \
    && mkdir -p /data/db \
    && chown -R nobody:nobody /data/db

RUN adduser -D kengrokuser

WORKDIR /app

COPY --from=builder /app/bin/* /app/

# Create directory for SQLite database and set permissions
RUN mkdir -p /app/data && chown -R kengrokuser:kengrokuser /app/data

USER kengrokuser

# Update the environment to point to the SQLite database location
ENV DB_PATH=/app/data/kengrok.db

# Initialize the SQLite database with the required table
RUN sqlite3 /app/data/kengrok.db "\
    CREATE TABLE IF NOT EXISTS port_mappings ( \
        subdomain TEXT PRIMARY KEY, \
        proxy_port INTEGER, \
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP \
    );" && \
    chmod 644 /app/data/kengrok.db

ENTRYPOINT ["/app/proxyserver", "3000"]
