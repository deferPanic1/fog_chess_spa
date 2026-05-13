## Запуск

```bash
docker compose up --build
```

Frontend: [http://localhost:5173](http://localhost:5173)

После старта контейнеров нужно отдельно применить миграции:

```bash
make migrate-up DB_HOST=localhost
```


## Архитектура проекта

Проект разделен на три основные части:

- `frontend` - SPA на `Vue 3` + `Vite` + `Pinia`
- `backend` - API и real-time логика на `Go`
- `db` - `PostgreSQL`, поднимается через `docker compose`

### Frontend

Во фронтенде приложение инициализируется в `frontend/src/main.js`: создаются `Pinia`, роутер и восстанавливается пользовательская сессия через `auth` store.

Основные зоны ответственности:

- `frontend/src/router` - маршрутизация и guard-логика для приватных и админских страниц
- `frontend/src/stores` - клиентское состояние:
- `frontend/src/api` - HTTP-клиент поверх `axios` для REST-запросов к backend
- `frontend/src/views` и `frontend/src/components` - UI-компоненты

Frontend общается с сервером через:

- REST API для авторизации, профиля, списка лобби, истории
- WebSocket для событий лобби и для игровой логики

### Backend

Backend собран по слоистой структуре:

- `backend/cmd/app` - точка входа приложения
- `backend/internal/app` - wiring зависимостей и запуск HTTP-сервера
- `backend/internal/transport/http` - REST-обработчики, DTO, middleware, cookies и вспомогательные HTTP-утилиты
- `backend/internal/transport/websocket` - WebSocket-обработчики для лобби и матча
- `backend/internal/service` - бизнес-логика приложения
- `backend/internal/repository` - интерфейсы доступа к данным
- `backend/internal/repository/postgres` - реализация репозиториев поверх PostgreSQL/GORM
- `backend/internal/domain/models` - доменные модели
- `backend/internal/chessEngine` - собственный шахматный движок и логика fog-of-war режима

### Контейнеры

`compose.yaml` поднимает:

- `db` - PostgreSQL 17
- `backend` - Go-приложение
- `frontend` - production-сборку SPA, отдаваемую из контейнера
