FROM golang:1.19.1-alpine3.16 AS builder
ARG TARGETARCH
WORKDIR /app
COPY . ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -a -ldflags="-s -w" -installsuffix cgo -v -o /service-intesis ./

FROM builder AS test
RUN go test ./...

FROM scratch
COPY --from=builder /service-intesis /service-intesis
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/service-intesis"]