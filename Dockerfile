FROM golang:1.23.3-alpine3.20 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o ./app ./cmd/app

FROM scratch

# COPY ./assets /assets
COPY .env.prod /.env
COPY --from=builder /app/app /app

USER 1000

CMD ["/app"]
