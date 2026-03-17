# Stage 1 — Build frontend
FROM node:22-slim AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
ARG VITE_SPOTIFY_CLIENT_ID
RUN npm run build

# Stage 2 — Build Go binary
FROM golang:1.25-bookworm AS backend
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o metaloreian ./cmd/server

# Stage 3 — Runtime (no Chromium needed)
FROM alpine:3.21
RUN apk add --no-cache wget ca-certificates && \
    adduser -D -s /sbin/nologin appuser
WORKDIR /app
COPY --from=backend /app/metaloreian .
COPY --from=frontend /app/frontend/dist ./dist
COPY backend/migrations ./migrations
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
CMD ["./metaloreian"]
