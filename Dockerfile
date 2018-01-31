# --------
# Stage 1: Build
# -------
FROM golang:alpine as builder

RUN apk --no-cache add git

WORKDIR /go/src/app
COPY . .

ENV CGO_ENABLED=0

RUN go-wrapper download
RUN go-wrapper install

# --------
# Stage 2: Release
# --------
FROM gcr.io/distroless/base

COPY --from=builder /go/bin/app /

ENV CONFIG_FILE=/config.json \
    FIRMWARE_DIR=/firmware \
    CACHE_DIR=/cache \
    FLATTEN=true \
    MANIFEST= \
    INCLUDE_MODELS= \
    INCLUDE_FILES= \
    EXCLUDE_MODELS= \
    EXCLUDE_FILES= \
    LATEST=

WORKDIR /firmware
WORKDIR /cache

VOLUME /firmware
VOLUME /cache

CMD ["/app"]
