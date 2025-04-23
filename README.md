# URL Shortener

Микросервис для сокращения ссылок с метриками на Go.

## Features
- [x] Сокращение URL
- [x] Редирект по короткой ссылке
- [x] Мониторинг через Prometheus/Grafana
- [ ] Tests

## Quick Start

```bash
git clone https://github.com/weddya/url-shortener.git
cd url-shortener
docker-compose up --build
```

API будет доступно на http://localhost:8080.

## API Endpoints

| Метод | Путь           | Описание                      |
|-------|----------------|-------------------------------|
| POST  | `/create`      | Создать короткую ссылку       |
| GET   | `/{short_code}` | Редирект на оригинальный URL  |

