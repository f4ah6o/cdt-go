# Add

`Add` adds two integers.

```go file=add.go symbol=Add
package calc

func Add(a int, b int) int {
	return a + b
}
```

```go test file=add_test.go verifies=Add
package calc

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Fatal("unexpected result")
	}
}
```
