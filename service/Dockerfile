# Open-Skill API service
FROM golang:1.25-alpine AS build
WORKDIR /src
# Cache deps
COPY go.mod go.sum ./
RUN go mod download
# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/open-skill-service ./cmd/open-skill-service

FROM alpine:3.20
RUN adduser -D -H appuser \
    && mkdir -p /data /app \
    && chown appuser:appuser /data /app
WORKDIR /app
COPY --from=build /out/open-skill-service /app/open-skill-service
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/open-skill-service"]
CMD ["-db", "/data/open-skill.db"]