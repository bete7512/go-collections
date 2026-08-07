# go-collections

A comprehensive collection of 50+ programming exercises in Go, covering strings, collections, algorithms, and data structures. Each exercise is self-contained with a specification, implementation template, and test suite.

## Purpose

This collection is designed to help developers:
- Master Go fundamentals (strings, slices, maps, interfaces)
- Practice common algorithms (sorting, searching, deduplication)
- Implement classic data structures (stack, queue, linked list, BST)
- Understand Go patterns (pointer receivers, table-driven tests, error handling)
- Handle edge cases properly (Unicode strings, empty collections, boundary conditions)

## Prerequisites

- Go 1.26.4 or later
- Basic familiarity with Go syntax

## Getting Started

1. Clone the repository:
```bash
git clone https://github.com/bete7512/go-collections.git
cd go-collections
```

2. Navigate to any exercise directory:
```bash
cd 01-reverse-string
```

3. Read the exercise specification in README.md

4. Implement the required function in main.go

5. Run the test suite:
```bash
go test
```

6. (Optional) Run the demo:
```bash
go run main.go
```

## Exercise Structure

Each exercise directory contains:
- **README.md** - Problem specification, expected behavior, edge cases, hints
- **main.go** - Skeleton code and/or demo with test data
- **main_test.go** - Table-driven test suite

## Exercise List

### String Manipulation (01-09)
| # | Title | Difficulty | Key Concepts |
|---|-------|-----------|--------------|
| 1 | Reverse String | Easy | Runes, UTF-8 safety, Unicode |
| 2 | Count Vowels | Easy | String iteration, case-insensitivity |
| 3 | Palindrome | Easy | Two-pointer technique, punctuation/case handling |
| 4 | Longest Word | Easy | String splitting, comparison |
| 5 | Capitalize Words | Easy | Title-case transformation |
| 6 | Truncate String | Easy | String slicing, ellipsis |
| 7 | Split String | Easy | Custom splitting logic |
| 8 | Join Strings | Easy | Array joining |
| 9 | Dedup Chars | Medium | Character deduplication, order preservation |

### Numeric Operations (10-12)
| # | Title | Difficulty | Key Concepts |
|---|-------|-----------|--------------|
| 10 | Min, Max, Sum | Easy | Single-pass computation, error handling |
| 11 | Reverse Slice | Easy | In-place reversal, swap |
| 12 | Rotate Slice | Medium | Array rotation, circular indexing |

### Slice Operations (13-22)
| # | Title | Difficulty | Key Concepts |
|---|-------|-----------|--------------|
| 13 | Remove At | Medium | Slice mutation, O(n) removal |
| 14 | Remove At Fast | Medium | Swap-with-last optimization, O(1) removal |
| 15 | Chunk Slice | Medium | Partitioning, batch processing |
| 16 | Flatten Slice | Medium | Nested slices, recursion |
| 17 | Filter Slice | Medium | Predicate filtering, dynamic sizing |
| 18 | Map Slice | Medium | Transformation, type conversion |
| 19 | Zip Slices | Medium | Parallel iteration, struct composition |
| 20 | Binary Search | Medium | Divide-and-conquer, sorted arrays |
| 21 | Word Frequency | Medium | String splitting, maps, counting |
| 22 | First Unique Char | Medium | Character tracking, order preservation |

### Hash/Map Operations (23-31)
| # | Title | Difficulty | Key Concepts |
|---|-------|-----------|--------------|
| 23 | Anagram Check | Medium | Character frequency, sorting |
| 24 | Group By First | Medium | Map grouping, slices as values |
| 25 | Invert Map | Medium | Bidirectional mapping |
| 26 | Set Type | Medium | Map-based sets, membership testing |
| 27 | Set Operations | Medium | Union, intersection, difference |
| 28 | Top N Words | Medium | Sorting by count, heap concepts |
| 29 | Dedup Structs | Hard | Struct equality, pointer semantics |
| 30 | Merge Maps | Medium | Map merging, conflict handling |
| 31 | Sorted Map Iter | Hard | Ordering guarantees, key sorting |

### Algorithms (32-42)
| # | Title | Difficulty | Key Concepts |
|---|-------|-----------|--------------|
| 32 | Two Sum | Medium | Hash lookup, index pairing |
| 33 | Memoize | Hard | Function caching, closure state |
| 34 | Rune Counts | Medium | Unicode counting, frequency analysis |
| 35 | Tx Totals | Easy | Struct slicing, aggregation |
| 36 | Point Distance | Medium | Geometry, math operations |
| 37 | Rectangle | Medium | Struct methods, area/perimeter |
| 38 | Counter Pointer | Medium | Pointer receivers, state mutation |
| 39 | Stringer | Medium | fmt.Stringer interface, custom string representation |
| 40 | Sort By Age | Easy | sort.Slice, custom comparator |
| 41 | Sort Age+Name | Medium | Multi-field sorting |
| 42 | Sort Interface | Hard | sort.Interface implementation (Len, Less, Swap) |

### Data Structures (43-53)
| # | Title | Difficulty | Key Concepts |
|---|-------|-----------|--------------|
| 43 | Stack | Medium | LIFO, pointer receivers, zero values |
| 44 | Queue | Medium | FIFO, circular buffer concepts |
| 45 | Linked List | Hard | Node pointers, traversal, insertion |
| 46 | Reverse List | Hard | In-place list reversal, pointer manipulation |
| 47 | Binary Search Tree | Hard | Tree insertion, recursive structure |
| 48 | Inorder Traversal | Hard | Tree traversal, recursion |
| 49 | Matrix Transpose | Medium | 2D slices, index manipulation |
| 50 | Ring Buffer | Hard | Circular buffer, modulo arithmetic |
| 51 | Shape Interface | Medium | Interface polymorphism, type assertion |
| 52 | Total Area | Medium | Interface methods, iteration |
| 53 | Type Switch | Medium | Type switching, reflection-free dispatch |

## Difficulty Levels

- **Easy**: 1-2 hours, fundamental concepts
- **Medium**: 2-4 hours, combines multiple concepts
- **Hard**: 4+ hours, complex data structures or algorithm design

## Testing Approach

All exercises use table-driven tests:

```go
func TestExercise(t *testing.T) {
    tests := []struct {
        name     string
        input    Type
        expected Type
    }{
        {"case1", input1, expected1},
        {"case2", input2, expected2},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := FunctionUnderTest(tt.input)
            if got != tt.expected {
                t.Errorf("got %v, want %v", got, tt.expected)
            }
        })
    }
}
```

Run all tests in a directory:
```bash
go test -v
```

Run a specific test:
```bash
go test -run TestExercise
```

## Common Patterns

### Unicode-Safe String Handling
```go
// Convert to runes for proper Unicode manipulation
runes := []rune(s)
// Do work
return string(runes)
```

### Pointer Receivers for Mutation
```go
func (s *Stack) Push(v int) {
    s.items = append(s.items, v)
}
```

### Two-Value Returns for Safety
```go
func (s *Stack) Pop() (int, bool) {
    if len(s.items) == 0 {
        return 0, false
    }
    // ...
    return val, true
}
```

### Table-Driven Tests with Subtests
```go
for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) {
        // test logic
    })
}
```

## Tips for Success

1. **Read the spec carefully** - each README specifies edge cases and what makes a solution complete
2. **Implement the basic case first** - then handle edge cases
3. **Use table-driven tests** - they make patterns obvious
4. **Check your assumptions** - especially about nil vs empty, negative numbers, and Unicode
5. **Single-pass algorithms** - when specified, avoid making multiple passes over data
6. **Pointer vs value semantics** - understand when to use pointer receivers

## License

This repository is open source and available under the MIT License.

---

**Last Updated:** August 2026