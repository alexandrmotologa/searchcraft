# Stage 1: Build Frontend & Widget
FROM node:22-alpine AS web-builder
WORKDIR /app
COPY widget/package*.json ./widget/
RUN cd widget && npm install
COPY widget/ ./widget/
RUN cd widget && npm run build

COPY web/package*.json ./web/
RUN cd web && npm install
COPY web/ ./web/
RUN cd web && npm run build

# Stage 2: Build Go Binary
FROM golang:1.23-alpine AS go-builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=web-builder /app/widget/dist ./server/widget_dist
COPY --from=web-builder /app/web/dist ./server/web_dist

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /searchcraft ./cmd/searchcraft

# Stage 3: Minimal Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /searchcraft /app/searchcraft

VOLUME ["/data"]
ENV SEARCHCRAFT_DB="/data/searchcraft.db"
ENV SEARCHCRAFT_PORT="7700"

EXPOSE 7700
ENTRYPOINT ["/app/searchcraft", "serve", "--port", "7700", "--db", "/data/searchcraft.db"]
