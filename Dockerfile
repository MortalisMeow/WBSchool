FROM golang:1.20.4

WORKDIR /app

COPY main.go .
RUN go build -o wbschool main.go

CMD ["./wbschool"]
