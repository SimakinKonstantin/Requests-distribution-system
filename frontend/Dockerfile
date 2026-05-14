FROM node:20-alpine

WORKDIR /app
ENV NODE_ENV=development

COPY package.json package-lock.json* ./
COPY node_modules ./node_modules
COPY . .

# Адрес бэкенда, доступный из браузера (порт проброшен на хост)
ARG VITE_API_URL=http://localhost:8080
ENV VITE_API_URL=$VITE_API_URL

RUN test -x ./node_modules/.bin/tsc && test -x ./node_modules/.bin/vite
RUN npm run build

EXPOSE 3000
CMD ["./node_modules/.bin/vite", "preview", "--host", "0.0.0.0", "--port", "3000"]
