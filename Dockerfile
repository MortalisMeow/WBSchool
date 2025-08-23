FROM golang:1.23-bookworm

WORKDIR /app


RUN apt-get update && apt-get install -y \
    gcc \
    musl-dev \
    librdkafka-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o wbschool ./cmd/

CMD ["./wbschool"]