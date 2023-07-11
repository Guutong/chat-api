FROM golang:1.18-buster as builder

WORKDIR /app

COPY . ./

RUN go mod download
RUN go build -v -o server

FROM debian:buster-slim

WORKDIR /app

COPY --from=builder /app/server /app/server

EXPOSE 8080

CMD ["/app/server"]