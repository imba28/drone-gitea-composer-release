FROM golang:1.18 as builder

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN make gitea-composer-release


FROM alpine:3
RUN apk add --no-cache ca-certificates

COPY --from=builder /build/gitea-composer-release /usr/local/bin/gitea-composer-release

ENTRYPOINT ["/usr/local/bin/gitea-composer-release"]