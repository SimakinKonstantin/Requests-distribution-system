FROM golang:1.26.1

WORKDIR /app
ENV GOFLAGS=-mod=vendor

COPY . .

RUN test -f vendor/modules.txt
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o main ./cmd/server/
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o goose github.com/pressly/goose/v3/cmd/goose

EXPOSE 8080

CMD ./goose -dir ./migrations postgres "$DATABASE_URL" up && ./main