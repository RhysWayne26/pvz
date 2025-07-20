# Минимальная документация по кеш-пакету

## Механизмы инвалидации

Пакет поддерживает три простых способа сброса устаревших или изменённых данных:

1. `Invalidate(key K)`  
   Удаляет запись с точным ключом.

2. `InvalidatePattern(pattern string)`  
   Удаляет все записи, чьи ключи соответствуют заданному регулярному выражению.

3. `InvalidateFunc(fn func(key K) bool)`  
   Удаляет все записи, для которых функция-предикат возвращает `true`.

**Примеры использования**
```go
// 1) Удалить какой-то элемент
cacheObj.Invalidate("Order:456")

// 2) Дропнуть все кеши списков заказов
cacheObj.InvalidatePattern("^ListOrders:")

// 3) Дропнуть все «старые» записи
cacheObj.InvalidateFunc(func(key string) bool {
    return strings.HasSuffix(key, ":old")
})
```
## Ограничение ресурсов
### 1. Политики вытеснения
TTL (Time To Live) - базовая политика

```go
// Только TTL без ограничений по размеру
cacheObj := cache.NewInMemoryShardedCache[string, Order](
    16,
    policies.NewTTLPolicy[string, Order](),
)

// Элементы удаляются только по истечении времени
cacheObj.Set("Order:123", order, 5*time.Minute)
```

LRU (Least Recently Used) - ограничение по количеству
```go
//LRU с лимитом в 100 элементов
cacheObj := cache.NewInMemoryShardedCache[string, Order](
16,
policies.NewLRUPolicy[string, Order](100),
)

// При превышении лимита удаляются наименее используемые элементы
cacheObj.Set("Order:456", order, 10*time.Minute)
```
Примечание: можно выбрать LRU стратегию, но базово "фоновый очиститель" устаревших записей по TTL всё равно будет работать

Пример инита кеша:
```go
shards := 16
policy := policies.NewLRUPolicy[string, Order](100)
cacheObj := cache.NewInMemoryShardedCache[string, Order](shards, policy)
cacheObj.Set("Order:456", order, 5*time.Minute)
```