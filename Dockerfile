FROM golang:1.23

WORKDIR /app

COPY . .

RUN go build -o main ./cmd

EXPOSE 8080

CMD ["./main"]