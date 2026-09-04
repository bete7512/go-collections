# 83 · Read a CSV and print the third column

## The idea in one paragraph

CSV looks like "split on commas" until the first real export lands: `"Doe, John",41,"Paris, FR"` is a **three**-field record, and `strings.Split(line, ",")` shreds it into five. Quoting is what makes CSV a format rather than a convention — quoted fields may contain commas, escaped quotes (`""`), even **newlines**, which means a record isn't a line and you cannot parse it line-by-line at all. `encoding/csv` handles all of it. The second lesson is in the signature: the function takes an **`io.Reader`, not a path**. Files, network bodies, and test strings all satisfy `io.Reader`, so the parser is testable with `strings.NewReader` and zero filesystem — the same decoupling every well-designed Go API uses.

## Vocabulary

| term | meaning |
|---|---|
| record / field | CSV's row and cell; a record can span multiple physical lines if a field is quoted |
| `csv.NewReader(r)` | wraps any `io.Reader`; handles quoting, CRLF, embedded newlines |
| `ReadAll()` vs repeated `Read()` | all-in-memory vs streaming one record at a time — streaming wins for big files; note the trade-off in a comment |
| `FieldsPerRecord` | default (0): every record must have the same field count as the first, ragged rows are errors; `-1`: variable-length records allowed |
| `LazyQuotes` | leave it off — strict quoting catches malformed input |
| `""` escaping | inside a quoted field, a doubled quote is one literal quote character |

---

## The function

```go
func ThirdColumn(r io.Reader) ([]string, error)
```

Parse CSV from `r` and return the third field (index 2) of every qualifying data record, in order.

## Pinned contract (every decision the tests assert)

- **The first record is a header and is always skipped** — whatever it contains, however many fields it has.
- **`FieldsPerRecord = -1`: variable-length records are allowed**, and any *data* record with **fewer than 3 fields is skipped silently** — no panic, no error, no placeholder entry. Records with more than 3 fields contribute their index-2 field like any other.
- Quoting per the format: `"Paris, FR"` is one field; `"say ""hey"""` yields `say "hey"`; a quoted field containing `\n` stays one field with the newline preserved in the returned string.
- Fields are returned **verbatim** — no trimming; `  spaced  ` keeps its spaces; an empty third field (`a,b,` or `,,`) contributes `""`.
- Blank lines between records are ignored (csv.Reader's behavior); CRLF input parses identically to LF; a missing trailing newline changes nothing.
- **Empty input → zero results, nil error.** Header-only input → zero results, nil error. (Nothing to return is not an error.)
- **Malformed CSV** (e.g. a bare/unclosed quote) → non-nil error and zero results — the parser's error must reach the caller, not be swallowed.

## Worked examples

**The spec's example** (quoted commas — the reason `strings.Split` can never work):

```
name,age,city
Alice,30,Berlin
"Doe, John",41,"Paris, FR"

→ (["Berlin", "Paris, FR"], nil)      header dropped; "Paris, FR" is ONE field
```

**Embedded newline** (legal CSV — a record is not a line):

```
h1,h2,h3
a,b,"line one
line two"

→ (["line one\nline two"], nil)       one record, one result
```

**Short rows skipped, wide rows fine:**

```
h1,h2,h3
just,two
x,y,z
p,q,r,s,t

→ (["z", "r"], nil)                   "just,two" has no third field → skipped
```

## Edge cases the tests hit

- Quoted comma inside a field (mandatory per spec).
- `""` quote escaping inside a quoted field.
- A field containing a newline inside quotes.
- Data rows with 1 and 2 fields (skipped), and with 4+ (index 2 used).
- Empty input; header-only input; input that is only blank lines after the header.
- Trailing newline present and absent; CRLF endings.
- Empty third field → `""` in the result.
- Whitespace kept verbatim.
- An unclosed quote → error surfaced, zero results.

## What the tests cover

Everything is driven from `strings.NewReader` — no filesystem access anywhere. Exact-slice assertions (`slices.Equal`) on all content cases, including the two headline ones (quoted comma, embedded newline); the skip policy for short rows; nil-error-with-empty-result for empty and header-only inputs; a non-nil error with zero results for malformed input.

**Done when:** the quoted-comma and embedded-newline tests pass, and the function signature still takes an `io.Reader`.
