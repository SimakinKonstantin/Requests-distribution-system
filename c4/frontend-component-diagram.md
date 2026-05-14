# Диаграмма компонентов frontend

## Зависимости файлов

```plantuml
@startuml
left to right direction
skinparam packageStyle rectangle

rectangle "frontend/index.html" as index
rectangle "frontend/src/main.tsx" as main
rectangle "frontend/src/App.tsx" as app

package "Страницы (frontend/src/pages)" as pages {
  rectangle "EmployeesPage.tsx" as employees
  rectangle "ClientsPage.tsx" as clients
  rectangle "ThemesPage.tsx" as themes
  rectangle "SubthemesPage.tsx" as subthemes
  rectangle "SlotsPage.tsx" as slots
  rectangle "AppealsPage.tsx" as appeals
  rectangle "AppealDetailPage.tsx" as appealDetail
  rectangle "TeamsPage.tsx" as teams
  rectangle "WorkflowsPage.tsx" as workflows
}

package "Общие модули (frontend/src)" as core {
  rectangle "api.ts" as api
  rectangle "types.ts" as types
  rectangle "validation.ts" as validation
  rectangle "hooks/useCrud.ts" as useCrud
  rectangle "hooks/usePolling.ts" as usePolling
  rectangle "components/Modal.tsx" as modal
  rectangle "vite-env.d.ts" as viteEnv
}

rectangle "frontend/vite.config.ts" as vitecfg

index --> main
main --> app

app --> employees
app --> clients
app --> themes
app --> subthemes
app --> slots
app --> appeals
app --> appealDetail
app --> teams
app --> workflows

api --> types

employees --> api
employees --> useCrud
employees --> modal
employees --> types
employees --> validation

clients --> api
clients --> useCrud
clients --> modal
clients --> types
clients --> validation

themes --> api
themes --> useCrud
themes --> modal
themes --> types
themes --> validation

subthemes --> api
subthemes --> useCrud
subthemes --> modal
subthemes --> types
subthemes --> validation

slots --> api
slots --> useCrud
slots --> types

appeals --> api
appeals --> useCrud
appeals --> usePolling
appeals --> modal
appeals --> types
appeals --> validation

appealDetail --> api
appealDetail --> usePolling
appealDetail --> types

teams --> api
teams --> useCrud
teams --> modal
teams --> types
teams --> validation

workflows --> api
workflows --> modal
workflows --> types
workflows --> validation

vitecfg ..> index : сборка/dev-server/proxy API
vitecfg ..> api : VITE_* env
viteEnv ..> api : типы Vite env
@enduml
```

## Описание файлов frontend

| Файл | Описание | Зависит от | Используется в |
|---|---|---|---|
| `frontend/index.html` | HTML-входная точка SPA, содержит контейнер `#root` и подключает `src/main.tsx`. | - | Браузер, Vite/serve |
| `frontend/package.json` | NPM-манифест проекта: скрипты (`dev`, `build`, `preview`) и зависимости React/Vite/TS. | - | npm, Vite, TypeScript |
| `frontend/package-lock.json` | Зафиксированное дерево npm-зависимостей для воспроизводимых установок. | `package.json` | npm |
| `frontend/tsconfig.json` | Конфигурация TypeScript (strict mode, JSX, include `src`). | - | TypeScript, Vite |
| `frontend/vite.config.ts` | Конфиг сборщика Vite: React-плагин, порт dev-сервера и proxy на backend. | `@vitejs/plugin-react` | Vite |
| `frontend/README.md` | Документация по запуску, структуре проекта и API-слою фронтенда. | - | Разработчики |
| `frontend/src/main.tsx` | React-entrypoint: создаёт root и оборачивает `App` в `BrowserRouter`. | `App.tsx` | `index.html` |
| `frontend/src/App.tsx` | Корневой layout с sidebar и маршрутизацией на все страницы приложения. | Все файлы из `src/pages` | `main.tsx` |
| `frontend/src/api.ts` | Единый HTTP-слой (`fetch` + CRUD API-объекты по сущностям). | `types.ts`, `vite-env` (`import.meta.env`) | Почти все страницы |
| `frontend/src/types.ts` | Набор TypeScript-интерфейсов доменных сущностей и workflow-структур. | - | `api.ts`, страницы |
| `frontend/src/validation.ts` | Общие правила валидации текстовых/числовых/email полей и helper для форм. | - | Страницы с формами |
| `frontend/src/vite-env.d.ts` | Типизация `ImportMetaEnv` (в частности `VITE_API_URL`) для TypeScript. | `vite/client` types | `api.ts`, TypeScript |
| `frontend/src/hooks/useCrud.ts` | Универсальный CRUD-хук: загрузка списка, create/update/remove, обработка ошибок. | React hooks | Большинство CRUD-страниц |
| `frontend/src/hooks/usePolling.ts` | Хук short-polling для периодического обновления данных по таймеру. | React hooks | `AppealsPage.tsx`, `AppealDetailPage.tsx` |
| `frontend/src/components/Modal.tsx` | Переиспользуемый модальный контейнер для create/edit форм. | React | Большинство страниц |
| `frontend/src/pages/EmployeesPage.tsx` | CRUD-страница сотрудников, включая привязку к командам и валидацию формы. | `api.ts`, `useCrud.ts`, `Modal.tsx`, `types.ts`, `validation.ts` | `App.tsx` |
| `frontend/src/pages/ClientsPage.tsx` | CRUD-страница клиентов с формой и поддержкой VIP-признака. | `api.ts`, `useCrud.ts`, `Modal.tsx`, `types.ts`, `validation.ts` | `App.tsx` |
| `frontend/src/pages/ThemesPage.tsx` | CRUD-страница тем обращений. | `api.ts`, `useCrud.ts`, `Modal.tsx`, `types.ts`, `validation.ts` | `App.tsx` |
| `frontend/src/pages/SubthemesPage.tsx` | CRUD-страница подтем обращений. | `api.ts`, `useCrud.ts`, `Modal.tsx`, `types.ts`, `validation.ts` | `App.tsx` |
| `frontend/src/pages/SlotsPage.tsx` | Страница просмотра слотов сотрудников с фильтром по сотруднику. | `api.ts`, `useCrud.ts`, `types.ts` | `App.tsx` |
| `frontend/src/pages/AppealsPage.tsx` | Страница списка обращений: фильтр, модалки create/edit, polling и переход в детали. | `api.ts`, `useCrud.ts`, `usePolling.ts`, `Modal.tsx`, `types.ts`, `validation.ts` | `App.tsx` |
| `frontend/src/pages/AppealDetailPage.tsx` | Детальная страница обращения с автообновлением и действием закрытия обращения. | `api.ts`, `usePolling.ts`, `types.ts` | `App.tsx`, роут `/appeals/:id` |
| `frontend/src/pages/TeamsPage.tsx` | CRUD-страница команд с управлением связями тема/подтема/VIP. | `api.ts`, `useCrud.ts`, `Modal.tsx`, `types.ts`, `validation.ts` | `App.tsx` |
| `frontend/src/pages/WorkflowsPage.tsx` | Конструктор и CRUD автоматизаций workflow (условия, действия, сериализация узлов). | `api.ts`, `Modal.tsx`, `types.ts`, `validation.ts` | `App.tsx` |

