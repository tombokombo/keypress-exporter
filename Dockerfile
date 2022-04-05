FROM golang:1.18-alpine as builder

WORKDIR /build

COPY go.mod ./
COPY go.sum ./
COPY main.go ./


RUN go mod download && go build -o ./keypress-exporter

FROM alpine:3.14

WORKDIR /

COPY --from=builder /build/keypress-exporter /bin/keypress-exporter

CMD ["/bin/keypress-exporter"]

