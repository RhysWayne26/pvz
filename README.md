# **PVZ CLI**
Консольное приложение для управления логикой пункта выдачи заказов (ПВЗ)

### Установка и запуск
Собрать бинарь `go build -o pvz.exe ./cmd/main.go`

Запустить `./pvz.exe`

Или сразу в интерактивном режиме: `go run ./cmd/main.go`

### Доступные команды
#### 1) accept-order 
Принять заказ от курьера. 

`--order-id <id> --user-id <id> --expires <yyyy-mm-dd>`

#### 2) process-orders
Выдать заказы или принять возврат клиента.

`process-orders --user-id <id> --action <issue|return> --order-ids <id1,id2,...>`

#### 3) return-order

Вернуть заказ курьеру

`return-order --order-id <id>`

#### 4) list-orders
Получить список заказов.

**Флаги:**
- `--in-pvz` — только заказы в статусе `ACCEPTED` или `RETURNED`
- `--last-id <id>` — курсор: вернуть заказы **после** указанного `order_id`
- `--last <N>` — вернуть заказы **начиная с N-го**, аналогично `offset`
- `--page <N> --limit <M>` — классическая пагинация (номер страницы и размер страницы)

`list-orders --user-id <id> [--in-pvz] [--last-id <id>] [--last <N>] [--page <N> --limit <M>]`

#### 5) list-returns

Получить список возвратов (пагинация).

`list-returns [--page <N> --limit <M>]`

#### 6) order-history

Показать историю всех операций со всеми заказами.

`order-history`

#### 7) import-orders

Импортировать заказы из JSON-файла.

`import-orders --file <path/to/orders.json>`

##### Формат JSON:

```json
[
  {
    "order_id": "abc123",
    "user_id": "user42",
    "expires_at": "2025-06-01"
  },
  {
    "order_id": "def456",
    "user_id": "user42",
    "expires_at": "2025-06-02"
  }
]
```

#### 8) scroll-orders

Бесконечная прокрутка списка заказов (cursor-based).

`scroll-orders --user-id <id> [--limit <N>]`

#### 9) help
Показать список доступных команд.

`help`