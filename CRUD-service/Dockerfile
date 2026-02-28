FROM golang:1.25-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/

EXPOSE 8080

CMD goose -dir ./migrations postgres "$DATABASE_URL" up && ./main