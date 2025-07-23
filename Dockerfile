# Dockerfile
FROM golang:1.24.5

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o server ./cmd/server

CMD ["/app/server"]
