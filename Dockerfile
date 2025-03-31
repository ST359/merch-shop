FROM golang:1.23 AS builder

WORKDIR /app
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /merch-shop ./cmd/merch-shop/ \
    && go clean -cache -modcache

FROM alpine:latest
WORKDIR /
COPY --from=builder /merch-shop ./merch-shop
RUN ls -l

EXPOSE 8080

CMD ["/merch-shop"]