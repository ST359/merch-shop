FROM golang:1.22

WORKDIR ${GOPATH}/merch-shop/
COPY . ${GOPATH}/merch-shop/
RUN go build -o /build ./cmd/merch-shop/ \
    && go clean -cache -modcache

EXPOSE 8080

CMD ["/build"]