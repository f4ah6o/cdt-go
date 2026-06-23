# cdt-go

`cdt` keeps **code**, **docs**, and **tests** as three projections of one graph.

```text
        code
       /    \
      /      \
     / graph  \
    /          \
 docs -------- test
```

The first implementation is a Go-focused MVP:

- extract a Graph IR from Markdown fenced blocks and `// cdt:*` markers
- render Go code, Go tests, or generated Markdown from the Graph IR
- run `go test` from the graph in a temporary workspace
- report concept-level coverage across docs/code/tests

## Install

```sh
go install github.com/f4ah6o/cdt-go/cmd/cdt@latest
```

## Markdown input

````md
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
````

## Go marker input

```go
package calc

// cdt:doc start concept=Add
// # Add
//
// `Add` adds two integers.
// cdt:doc end

// cdt:code start concept=Add symbol=Add file=add.go
func Add(a int, b int) int {
	return a + b
}
// cdt:code end
```

```go
package calc

import "testing"

// cdt:test start concept=Add symbol=TestAdd file=add_test.go
func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Fatal("unexpected result")
	}
}
// cdt:test end
```

## Commands

```sh
cdt graph ./docs -o .cdt/graph.json
cdt check
cdt check --strict
cdt coverage
cdt render code -o generated
cdt render docs -o generated-docs
cdt test
```

`cdt check --strict` fails when a concept lacks any of docs, code, or tests.

## Graph IR

The MVP schema uses five node kinds:

- `file`
- `doc`
- `code`
- `test`
- `concept`

And these edge kinds:

- `contains`
- `describes`
- `implements`
- `verifies`
- `renders_to`
- `derived_from`

Documentation uses `describes`, implementation code uses `implements`, and tests use `verifies` to connect back to concepts.

LLM linking can be added later as a candidate graph generator. The deterministic extractor, validator, renderer, and test runner stay authoritative.
