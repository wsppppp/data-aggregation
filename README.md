## Поднимаем сервисы:
- Только БД: `docker compose up -d --build db`
- Применить миграции: `docker compose run --rm migrate`
- Приложение: `docker compose up -d app`
- ВСЕ СРАЗУ `docker compose up -d --build`
---
##  Swagger UI
### [http://localhost:8080/docs/index.html#/Subscriptions/get_api_v1_subscriptions](http://localhost:8080/docs/index.html#/Subscriptions/get_api_v1_subscriptions)

---


## Примеры запросов (curl)

- Создать подписку:
  ```
  curl -X POST http://localhost:8080/api/v1/subscriptions \
    -H "Content-Type: application/json" \
    -d '{"service_name":"Yandex Plus","price":400,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025"}'
  ```

- Получить по id:
  ```
  curl http://localhost:8080/api/v1/subscriptions/<id>
  ```

- Листинг по пользователю:
  ```
  curl "http://localhost:8080/api/v1/subscriptions?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus&limit=100&offset=0"
  ```

- Total за период:
  ```
  curl "http://localhost:8080/api/v1/subscriptions/total?from=07-2025&to=12-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus"
  ```

- Формат дат:
    - Во входных данных: `MM-YYYY`.
    - В ответах: строки `MM-YYYY`.
    - В БД хранится DATE как первое число месяца (например, `2025-07-01`), для того, чтобы были быстрые сравнения и индексация.