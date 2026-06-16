# core/json

Declared JSON helpers.

`@json.stringify(value)` serializes object-like values to a JSON string.
Function-valued object fields are omitted.

`@json.parse(text)` deserializes JSON into the type declared on its binding:

```rune
user := @json.parse(text) : User
```

The target type must implement `&FromJson`.

Use `#json.object` on a struct to enable field configuration and generate its
static `fromJson(text: String) -> Self` implementation:

```rune
#json.object
User: {
  #json.name("user_name")
  name: String
  #json.ignore
  password: String
}

user := User::fromJson(
  "{\"user_name\":\"Ada\",\"password\":\"ignored\"}"
)
```

`#json.name` changes the field name in both directions. `#json.ignore` omits
the field when stringifying and leaves its zero value when parsing.
