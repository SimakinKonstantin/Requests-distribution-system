package balancer

import "crud-service/internal/config"

// Чтобы не было циклических зависимостей.
type Config = config.Config
