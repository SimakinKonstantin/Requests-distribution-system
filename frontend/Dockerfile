# ── Stage 1: build ────────────────────────────────────────────────────────
FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm install

COPY . .

# Адрес бэкенда, доступный из браузера (порт проброшен на хост)
ARG VITE_API_URL=http://localhost:8080
ENV VITE_API_URL=$VITE_API_URL

RUN npm run build

# ── Stage 2: serve static ──────────────────────────────────────────────────
FROM node:20-alpine

RUN npm install -g serve

COPY --from=builder /app/dist /app/dist

EXPOSE 3000
CMD ["serve", "-s", "/app/dist", "--listen", "tcp://0.0.0.0:3000"]
