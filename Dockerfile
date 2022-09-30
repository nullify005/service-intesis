FROM golang:1.19.1-alpine3.16 AS builder
RUN apk --no-cache add build-base
ARG TARGETARCH
WORKDIR /app
COPY src/cmd/ ./cmd
COPY src/pkg/ ./pkg
COPY src/go.mod ./
COPY src/go.sum ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -a -ldflags="-s -w" -installsuffix cgo -v -o /service-intesis ./cmd/service-intesis.go

FROM builder AS test
RUN go test ./...

FROM scratch
COPY --from=builder /service-intesis /service-intesis
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/service-intesis"]