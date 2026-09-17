# Builds one of the three binaries in cmd/ selected by the CMD build arg.
FROM golang:1-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG CMD=api
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${CMD}

FROM gcr.io/distroless/static-debian12 AS run
COPY --from=build /out/app /app
USER 65532:65532
ENTRYPOINT ["/app"]
