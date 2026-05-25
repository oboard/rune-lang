# core/map

Map construction and receiver APIs.

Use `@map.new(keyType, valueType)` when a map literal cannot infer the desired
key or value type. Map literals such as `{ "a": 1 }` infer `Map[String, Int]`.
