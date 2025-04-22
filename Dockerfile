FROM golang:1.24

WORKDIR /app
COPY . .

RUN go build -o url-shortener ./cmd/server/

CMD ["./url-shortener"]