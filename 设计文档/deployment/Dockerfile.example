# syntax=docker/dockerfile:1

FROM node:22-alpine AS frontend-build
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi
COPY frontend/ ./
RUN npm run build

FROM golang:1.23-alpine AS backend-build
WORKDIR /src/backend
RUN apk add --no-cache ca-certificates git
COPY backend/go.mod backend/go.sum* ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/nsh-guild-analytics ./cmd/server

FROM alpine:3.22 AS runtime
RUN apk add --no-cache ca-certificates tzdata wget && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=backend-build /out/nsh-guild-analytics /app/nsh-guild-analytics
COPY --from=frontend-build /src/frontend/dist /app/public
COPY deployment/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh && mkdir -p /app/data/uploads /app/backups /app/public && chown -R app:app /app /entrypoint.sh
USER app
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/nsh-guild-analytics"]
