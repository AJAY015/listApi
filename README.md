# Go List Manipulation API

Maintain a FIFO list of integers and manipulate it by posting numbers to an API.  
The rules are simple:

- If the input number has the **same sign** as the numbers already in the list → **append** it.
- If the input number has the **opposite sign** → **consume** values from the **front** of the list (FIFO) until the input is exhausted.
- If the list is emptied and there’s still remainder, append the remainder with the input’s sign.

---

## Requirements
- Go 1.21+  
- Gin v1.10.0 (managed via `go.mod`)

---

##  Run the Server
```bash
go mod tidy
go run .
````

Server starts on `:8080` by default.

---

## 📡 API Endpoints

### Health check

```bash
GET /healthz
```

### Read current list

```bash
GET /numbers
```

### Apply a number

```bash
POST /numbers
Content-Type: application/json

{
  "value": 5
}
```

### Reset state

```bash
POST /reset
```

### Run the demo example (5, 10, -6)

```bash
POST /example
```

---

## Example Walkthrough

Sequence:

1. Input: `5`
   → List: `[5]`

2. Input: `10`
   → List: `[5, 10]`

3. Input: `-6`

   * Consume `5` completely → left to consume = `1`
   * Consume `1` from `10` → `9` remains
     → List: `[9]`

### Curl demo

```bash
curl -sS -X POST http://localhost:8080/reset
curl -sS -X POST http://localhost:8080/numbers -H 'Content-Type: application/json' -d '{"value":5}'
curl -sS -X POST http://localhost:8080/numbers -H 'Content-Type: application/json' -d '{"value":10}'
curl -sS -X POST http://localhost:8080/numbers -H 'Content-Type: application/json' -d '{"value":-6}'
curl -sS http://localhost:8080/numbers
```

Expected final output:

```json
{"list":[9]}
```

Or run the built-in example:

```bash
curl -sS -X POST http://localhost:8080/example
```

---

## How It Works

* The **list** is stored in memory, protected by a `sync.Mutex` (safe for concurrent requests).
* **Apply(n)** rules:

  * `0` → ignored (no-op).
  * Empty list → append `n`.
  * Same sign as list → append to the end.
  * Opposite sign → consume from head (FIFO):

    * Subtract from the oldest element until `n` is exhausted.
    * If an element reaches 0, drop it.
    * If the list empties and `n` is not fully consumed, append the leftover with the sign of `n`.

---

## Edge Cases

1. **Zero input** → ignored.
2. **Empty list** → starts with given number (positive or negative).
3. **Opposite input larger than total** → list empties, leftover is appended with input’s sign.

   * Example: `[4, 3]` + `-10` → `[-3]`
4. **Opposite input exactly equals total** → list becomes empty.
5. **Exact element consumption** → element is removed completely, no zeros are stored.
6. **Start with negative** → valid (`[-5]`, then add `-3` → `[-5, -3]`).
7. **Concurrency** → protected with a mutex.
8. **Validation** → only integer values accepted, invalid JSON returns `400`.

---

## Logs

* Gin logs all HTTP requests.
* Custom logs describe list operations (append, consume, flip).

---

## Tests

Unit tests are included in `store_test.go`.

Run:

```bash
go test ./...
```

Covers:

* Same sign append
* Opposite sign partial consumption
* Opposite sign exact consumption
* Opposite input larger than total (sign flip)
* Zero no-op
* Start with negative input

---

