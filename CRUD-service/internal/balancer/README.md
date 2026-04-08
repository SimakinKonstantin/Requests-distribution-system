# golang-ver (demo)

Упрощённая Go-версия балансировщика для демо:

- **RabbitMQ**: входящие ивенты `APPEAL_NEEDS_DISTRIBUTION`, `APPEAL_CLOSED`
- **asynq**: внутренняя очередь джобов (вместо BullMQ)
- **Postgres**: хранение состояния (вместо Redis sorted set). Один pod → **распределённые локи не нужны**.

## Архитектура

- `event-handler-service.go`: консьюмер RabbitMQ → batch/dedupe → группировка → enqueue `asynq` jobs
- `balancer-update-service.go`: обработчик batch jobs → обновление state в Postgres (pending appeals)
- `matcher.go`: раз в 1 секунду запускается distribution job → читает pending appeals + available managers → считает назначения → enqueue assign jobs
- `assigner.go`: транзакционно назначает в Postgres и удаляет pending запись

## Запуск (docker-compose)

1) Поднять зависимости:

```bash
cd golang-ver
docker compose up -d
```

2) Запустить приложение (в той же pod/машине):

```bash
# Windows PowerShell пример
$env:ROLE="all"
$env:POSTGRES_DSN="postgres://postgres:postgres@localhost:5432/balancer_demo?sslmode=disable"
$env:REDIS_ADDR="localhost:6379"
$env:RABBIT_URL="amqp://guest:guest@localhost:5672/"
$env:RABBIT_QUEUE="balancer.demo.events"

go run .
```

3) Опубликовать тест-ивент (пример payload см. ниже) и посмотреть логи.

## Формат ивента

Сообщение в очереди RabbitMQ — JSON:

```json
{
  "name": "APPEAL_NEEDS_DISTRIBUTION",
  "payload": { "appealId": 123, "teamId": "team-a" }
}
```

или

```json
{
  "name": "APPEAL_CLOSED",
  "payload": { "appealId": 123, "teamId": "team-a" }
}
```

