FROM golang:1.24 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/api ./cmd/api && CGO_ENABLED=0 GOOS=linux go build -o bin/db ./cmd/db && CGO_ENABLED=0 GOOS=linux go build -o bin/worker ./cmd/worker

FROM scratch as api

WORKDIR /app

COPY --from=builder /app/bin/api .

EXPOSE 5000

CMD ["./api"]

FROM scratch as db

WORKDIR /app

COPY --from=builder /app/bin/db .

CMD ["./db"]

FROM scratch as worker

WORKDIR /app

COPY --from=builder /app/bin/worker .

CMD ["./worker"]