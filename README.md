# **PVZ CLI**
Консольное приложение для управления логикой пункта выдачи заказов (ПВЗ)

### Установка и запуск
Собрать бинарь `go build -o pvz.exe ./cmd/main.go`

Запустить `./pvz.exe`

Или сразу в интерактивном режиме: `go run ./cmd/main.go`

### Доступные команды
#### 1) accept-order 
Принять заказ от курьера. 

`accept-order --order-id <id> --user-id <id> --expires <yyyy-mm-dd> --weight <float> --price <float> [--package <bag|box|film|bag+film|box+film>] `

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
  { "order_id": "1", "user_id": "u1", "expires_at": "2025-06-20", "weight": "5", "price": "100", "package": "bag" },
  { "order_id": "2", "user_id": "u2", "expires_at": "2025-06-20", "weight": "25", "price": "200", "package": "box+film" },
  { "order_id": "3", "user_id": "u3", "expires_at": "2025-06-20", "weight": "50", "price": "50", "package": "film" },
  { "order_id": "4", "user_id": "u4", "expires_at": "2025-06-20", "weight": "15", "price": "100", "package": "bag" },
  { "order_id": "5", "user_id": "u5", "expires_at": "2025-06-20", "weight": "40", "price": "150", "package": "box" },
  { "order_id": "6", "user_id": "u6", "expires_at": "2025-06-20", "weight": "5", "price": "70",  "package": "trash" },
  { "order_id": "7", "user_id": "u7", "expires_at": "2025-06-20", "weight": "0", "price": "100", "package": "film" },
  { "order_id": "8", "user_id": "u8", "expires_at": "2025-06-20", "weight": "2", "price": "0",   "package": "film" },
  { "order_id": "9", "user_id": "u9", "expires_at": "2025-06-20", "weight": "3", "price": "100" }
]
```

#### 8) scroll-orders

Бесконечная прокрутка списка заказов (cursor-based).

`scroll-orders --user-id <id> [--limit <N>]`

#### 9) help
Показать список доступных команд.

`help`