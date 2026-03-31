# Stage 1: Frontend bauen
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Backend bauen
FROM golang:1.23-alpine AS backend
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
COPY --from=frontend /app/frontend/build ./static
RUN CGO_ENABLED=1 GOOS=linux go build -o /vereinstool ./main.go

# Stage 3: Minimales Runtime-Image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /vereinstool .
ENV DB_PATH=/data/vereinstool.db
EXPOSE 8080
CMD ["./vereinstool"]
