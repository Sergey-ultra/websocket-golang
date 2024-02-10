FROM golang:1.20-alpine AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash gcc gettext musl-dev

COPY ["app/go.mod", "app/go.sum", "./"]
RUN go mod download

COPY app ./
RUN go build -o ./bin/app   app/cmd/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/app /
#COPY app/configs/config.yaml /config.yaml

CMD ["/app"]