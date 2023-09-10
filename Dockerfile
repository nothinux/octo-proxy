from golang:1.21-alpine as builder


WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/octo ./cmd/octo

FROM scratch

COPY --from=builder /usr/local/bin/octo /usr/local/bin/octo

CMD ["/usr/local/bin/octo"]