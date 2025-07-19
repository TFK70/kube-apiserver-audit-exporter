FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o build/kube-apiserver-audit-exporter cmd/kube-apiserver-audit-exporter/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/build/kube-apiserver-audit-exporter .

ENTRYPOINT ["/app/kube-apiserver-audit-exporter"]
