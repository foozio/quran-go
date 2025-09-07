# syntax=docker/dockerfile:1
FROM golang:1.22 AS build
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod     --mount=type=cache,target=/root/.cache/go-build     go build -o /out/quran-api ./cmd/quran-api

FROM gcr.io/distroless/base-debian12
WORKDIR /srv/app
COPY --from=build /out/quran-api /usr/local/bin/quran-api
COPY quran.db .
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/quran-api"]
