# syntax=docker/dockerfile:1
FROM golang:1.23 AS build
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod     --mount=type=cache,target=/root/.cache/go-build     go build -o /out/quran-api ./cmd/quran-api

FROM gcr.io/distroless/base-debian12
WORKDIR /srv/app
COPY --from=build /out/quran-api /usr/local/bin/quran-api
# Data directory (mount a volume here)
ENV QURAN_DB_PATH=/data/quran.db
ENV QURAN_BIND=:8080
VOLUME ["/data"]
EXPOSE 8080
USER 65532:65532
# Healthcheck hits the internal /healthz via the same binary
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD ["/usr/local/bin/quran-api","-selfcheck"]
ENTRYPOINT ["/usr/local/bin/quran-api"]
