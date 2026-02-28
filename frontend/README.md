# Frontend — Система распределения обращений

Веб-интерфейс для администрирования всех сущностей системы.  
Написан на **React 18 + TypeScript + Vite**, без сторонних UI-библиотек.

---

## Стек

| Инструмент | Версия | Роль |
|---|---|---|
| React | 18 | UI-фреймворк |
| TypeScript | 5 | Статическая типизация |
| Vite | 5 | Сборщик и dev-сервер |
| react-router-dom | 6 | Клиентская маршрутизация |
| serve | — | Раздача статики в Docker-контейнере |

---

## Запуск

### В Docker Compose (основной способ)
```bash
# из корня проекта
docker compose up --build
```
Приложение будет доступно на `http://localhost:3000`.

### Локально (для разработки)
```bash
cd frontend
npm install
npm run dev          # http://localhost:3000
```

При локальном запуске `vite.config.ts` настраивает **прокси**: запросы к путям
`/employees`, `/slots`, `/appeals`, `/subthemes` автоматически проксируются на
`http://localhost:8080` (бэкенд), чтобы избежать CORS-ошибок в dev-режиме.

---

## Переменные окружения

| Переменная | По умолчанию | Описание |
|---|---|---|
| `VITE_API_URL` | `""` (пустая строка) | Базовый URL бэкенда. Пустая строка означает «тот же origin» — браузер шлёт запросы на `http://localhost:3000/employees` и т. д. Nginx / прокси перенаправляет их на бэкенд |

В `Dockerfile` аргумент задаётся при сборке:
```dockerfile
ARG VITE_API_URL=http://localhost:8080
```
Переменная вшивается в статические файлы на этапе `npm run build` — изменить её
после сборки **нельзя** без пересборки образа.

---

## Структура исходного кода

```
src/
├── main.tsx              # точка входа: React root + BrowserRouter
├── App.tsx               # корневой компонент: сайдбар + маршруты
├── types.ts              # TypeScript-интерфейсы доменных сущностей
├── api.ts                # HTTP-клиент: один объект на ресурс
├── hooks/
│   ├── useCrud.ts        # универсальный хук для CRUD-операций
│   └── usePolling.ts     # хук периодического опроса (short polling)
├── components/
│   └── Modal.tsx         # универсальная модальная форма
└── pages/
    ├── EmployeesPage.tsx
    ├── ClientsPage.tsx
    ├── ThemesPage.tsx
    ├── SubthemesPage.tsx
    ├── SlotsPage.tsx
    ├── AppealsPage.tsx        # список обращений с фильтром
    └── AppealDetailPage.tsx   # детальная страница обращения
```

---

## Слой API (`src/api.ts`)

Все запросы к бэкенду идут через единую функцию `request<T>`:

```typescript
async function request<T>(path: string, init?: RequestInit): Promise<T>
```

- Добавляет заголовок `Content-Type: application/json`.
- Статус `204 No Content` возвращает `undefined` (не пытается парсить тело).
- Любой не-2xx статус выбрасывает `Error("STATUS TEXT")`.

### Объекты API

| Экспорт | Ресурс | Методы |
|---|---|---|
| `employeeApi` | `GET/POST /employees`, `GET/PUT/DELETE /employees/:id` | getAll, getById, create, update, delete |
| `clientApi` | `/clients` | getAll, getById, create, update, delete |
| `themeApi` | `/themes` | getAll, getById, create, update, delete |
| `subthemeApi` | `/subthemes` | getAll, getById, create, update, delete |
| `slotApi` | `/slots` | getAll, getById, create, update, delete |
| `appealApi` | `/appeals` | getAll, getById, create, update, delete, **close** |

Метод `appealApi.close(id)` отправляет `POST /appeals/:id/close` и возвращает
обновлённый объект `Appeal` (статус изменён на `"closed"`).

---

## Типы сущностей (`src/types.ts`)

```typescript
interface Employee {
  id: number; name: string; surname: string;
  limit: number; teamId: number; email: string
}

interface Client  { id: number; email: string }
interface Theme   { id: number; name: string }
interface Subtheme{ id: number; name: string }
interface Slot    { id: number; employeeId: number; appealId: number }

interface Appeal {
  id: number
  clientId: number
  employeeId: number | null  // null = сотрудник ещё не назначен
  themeId: number
  subthemeId: number
  text: string
  status: 'active' | 'closed'
}
```

> `Appeal.employeeId` — **nullable**. Бэкенд отдаёт `null` пока сотрудник не
> назначен. Фронтенд отображает «не назначен» для `null` и email сотрудника для
> числового значения.

---

## Хук `useCrud` (`src/hooks/useCrud.ts`)

Инкапсулирует типичный CRUD-цикл: загрузка при монтировании, оптимистичное
обновление локального состояния после мутаций.

```typescript
const { items, loading, error, create, update, remove, reload } =
  useCrud<Appeal, EditForm>(appealApi)
```

| Возвращаемое значение | Тип | Описание |
|---|---|---|
| `items` | `T[]` | Текущий список сущностей |
| `loading` | `boolean` | `true` во время выполнения `getAll` |
| `error` | `string \| null` | Сообщение последней ошибки |
| `create(data)` | `async` | POST, добавляет в конец `items` |
| `update(id, data)` | `async` | PUT, заменяет элемент в `items` |
| `remove(id)` | `async` | DELETE, удаляет элемент из `items` |
| `reload` | `async` | Повторный вызов `getAll` (стабильная ссылка) |

`reload` оборачивается в `useCallback` и **не меняет ссылку** при ре-рендерах,
поэтому его безопасно передавать в `usePolling`.

---

## Хук `usePolling` (`src/hooks/usePolling.ts`)

Реализует **short polling** — периодический опрос бэкенда.

```typescript
usePolling(fn, intervalMs?, enabled?)
// intervalMs по умолчанию 3000 мс
```

Сохраняет функцию в `ref`, чтобы обновление ссылки на `fn` не пересоздавало
таймер. Первый вызов **не** происходит немедленно при монтировании (начальную
загрузку делает `useCrud`/компонент самостоятельно); последующие — каждые
`intervalMs` мс. При размонтировании компонента таймер очищается.

---

## Маршрутизация (`src/App.tsx`)

Клиентская маршрутизация через `react-router-dom` v6 (`BrowserRouter`).

| Путь | Компонент |
|---|---|
| `/` | `EmployeesPage` (редирект по умолчанию) |
| `/employees` | `EmployeesPage` |
| `/clients` | `ClientsPage` |
| `/themes` | `ThemesPage` |
| `/subthemes` | `SubthemesPage` |
| `/slots` | `SlotsPage` |
| `/appeals` | `AppealsPage` |
| `/appeals/:id` | `AppealDetailPage` |

Все маршруты — **SPA-маршруты** (без перезагрузки страницы). При прямом
переходе по URL (`http://localhost:3000/appeals/5`) сервер (контейнер `serve`)
должен отдавать `index.html` — флаг `-s` в `CMD` это обеспечивает.

---

## Страницы

### `AppealsPage` — список обращений

- Загружает список обращений через `useCrud`.  
- Включает **short polling** (`usePolling`) каждые 3 с — список автоматически
  обновляется, когда сторонний сервис назначит сотрудника.
- **Фильтр по клиенту**: выпадающий список email-адресов; фильтрация
  выполняется на стороне браузера по уже загруженным данным — дополнительных
  запросов не создаёт.
- Дропдауны «Клиент», «Тема», «Подтема» в формах создания/редактирования
  подгружают справочники через отдельные запросы к `/clients`, `/themes`,
  `/subthemes` при монтировании страницы.
- При создании обращения `employeeId` всегда отправляется как `null` — назначение
  сотрудника выполняется внешним сервисом.
- Кнопка **«Детали»** выполняет `navigate('/appeals/:id')` без перезагрузки.

### `AppealDetailPage` — детальная страница обращения

- Параметр `:id` извлекается через `useParams`.
- Постоянный **short polling** каждые 3 с, пока страница открыта: как только
  внешний сервис присвоит `employeeId`, обновление отразится на экране без
  действий пользователя.
- Кнопка **«Закрыть обращение»** вызывает `POST /appeals/:id/close` и
  немедленно обновляет локальное состояние из ответа сервера. Кнопка скрыта,
  если `status === "closed"`.
- Запрашивает справочники (`/clients`, `/employees`, `/themes`, `/subthemes`)
  один раз при монтировании для отображения email/имён вместо числовых ID.

### Остальные страницы (`EmployeesPage`, `ClientsPage`, `ThemesPage`, `SubthemesPage`, `SlotsPage`)

Единообразная структура: таблица + модальная форма создания + модальная форма
редактирования. Используют `useCrud` и компонент `Modal`.

---

## Взаимодействие с бэкендом

```
Браузер
  └─ fetch("/appeals", { method: "GET" })
       │
       │  (в Docker: браузер обращается к http://localhost:8080 напрямую)
       │  (в dev: Vite proxy перенаправляет /appeals → localhost:8080)
       │
       └─► Go HTTP-сервер (порт 8080)
             └─► PostgreSQL
```

Бэкенд добавляет заголовки CORS (`Access-Control-Allow-Origin: *`), поэтому
запросы из браузера проходят без ошибок в production-конфигурации Docker.

### Формат запросов

- Тело: `application/json`, camelCase-поля (например, `clientId`, `employeeId`).
- Go-теги `json:"clientId"` на структурах модели обеспечивают корректный маппинг.
- `PUT /resource/:id` — полная замена объекта (не PATCH).
- `DELETE /resource/:id` → `204 No Content`, тело отсутствует.
- `POST /appeals/:id/close` → `200 OK` с телом обновлённого объекта `Appeal`.

---

## Сборка в Docker

```dockerfile
# Stage 1: сборка статики
FROM node:20-alpine AS builder
ARG VITE_API_URL=http://localhost:8080   # вшивается в JS-бандл
RUN npm run build                         # → /app/dist

# Stage 2: раздача статики
FROM node:20-alpine
RUN npm install -g serve
CMD ["serve", "-s", "/app/dist", "--listen", "tcp://0.0.0.0:3000"]
#   -s  — SPA-режим: все 404 → index.html (нужно для react-router)
#   0.0.0.0 — слушать все интерфейсы (иначе порт не проброситься из контейнера)
```
