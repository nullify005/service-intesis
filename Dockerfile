FROM golang:1.19.5-alpine3.16 AS builder
RUN apk --no-cache add build-base
ARG TARGETARCH
WORKDIR /src
RUN go install golang.org/x/vuln/cmd/govulncheck@latest
COPY src/ ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -a -ldflags="-s -w" -installsuffix cgo -v -o /service-intesis .

FROM builder AS test
RUN go test ./...
RUN govulncheck ./...

FROM scratch AS final
COPY --from=builder /service-intesis /service-intesis
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/service-intesis"]