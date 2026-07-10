
````markdown
# Go Memory Analysis CLI Tools

This document lists command-line tools and techniques to analyze memory allocations in Go code. It is intended for LLMs or automation tools to follow and extract relevant commands and concepts.

---

## 🔍 1. `go build -gcflags="-m"`

**Purpose:** Detect variables that escape to the heap.

**Command:**
```bash
go build -gcflags="-m" ./...
````

**Output Example:**

```
./main.go:10:6: moved to heap: myStruct
```

---

## 📊 2. `pprof` – Memory Profiling

**Purpose:** Profile memory usage and find functions with most allocations.

**Usage with HTTP server:**

1. Import `net/http/pprof`
2. Launch server:

```go
import _ "net/http/pprof"
import "net/http"

go func() {
  log.Println(http.ListenAndServe("localhost:8080", nil))
}()
```

3. Run app, then collect profile:

```bash
go tool pprof http://localhost:8080/debug/pprof/heap
```

**Visualize:**

```bash
go tool pprof -http=:8080 binary heap.prof
```

---

## 🧪 3. Bench-based Profiling

**Purpose:** Profile memory from benchmarks.

**Command:**

```bash
go test -bench=. -benchmem -memprofile=mem.prof
go tool pprof -text ./yourbinary.test mem.prof
```

---

## 📏 4. Manual Memory Inspection

**Purpose:** Check memory stats before and after code blocks.

**Code:**

```go
import "runtime"

var m runtime.MemStats
runtime.ReadMemStats(&m)
fmt.Printf("Alloc = %v MiB\n", m.Alloc/1024/1024)
```

---

## 📉 5. `benchstat`

**Purpose:** Compare memory usage between versions.

**Steps:**

```bash
go test -bench=. -benchmem > old.txt
# Make changes
go test -bench=. -benchmem > new.txt
benchstat old.txt new.txt
```

---

## 🔦 6. allocsnoop (Google tool)

**Purpose:** Real-time memory allocation tracing.

**Repo:** [https://github.com/google/allocsnoop](https://github.com/google/allocsnoop)

---

## 🐞 7. Delve (debugger)

**Purpose:** Inspect memory and data at runtime.

**Command:**

```bash
dlv debug
```

---

## Notes

* For long-running services, prefer `pprof` with HTTP endpoints.
* For short-lived CLI tools, use `go test -benchmem -memprofile`.
* Use `-gcflags="-m"` early to avoid unintended heap usage.
* Use `benchstat` to automate regression detection.

```


