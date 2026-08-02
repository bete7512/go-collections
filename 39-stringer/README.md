# 39 · Implement Stringer

## Problem

Make `fmt` print your type the way you want. In Go you don't register formatters or override `toString` — you satisfy `fmt.Stringer`, a one-method interface, and every `Println`, `Sprintf("%v")`, and `%s` in the ecosystem starts using it automatically. The satisfaction is *implicit*: no declaration, no keyword; having the method **is** being a Stringer. This is the interface machinery of Tier 4 arriving early, in its friendliest form — and it comes with a famous self-inflicted crash you're required to trigger once.

## Contract (what the tests enforce)

```go
type Temp struct{ C float64 }

func (t Temp) String() string   // e.g. "21.5°C"
```

- **Format: the number, then `°C`, nothing else.** The number renders in Go's default float style (`%g` — what `fmt.Sprint(21.5)` gives you): no trailing zeros, no forced decimal point.
  - `Temp{21.5}` → `"21.5°C"`
  - `Temp{21}` → `"21°C"` (not `"21.0°C"`)
  - `Temp{0}` → `"0°C"`
  - `Temp{-40}` → `"-40°C"`
- **Value receiver, and that's load-bearing.** With `func (t Temp) String()`, both `Temp` and `*Temp` satisfy `fmt.Stringer` (a pointer's method set includes the value-receiver methods). The tests print through both and expect the custom format from each. Had you chosen a pointer receiver, printing a plain `Temp` value would fall back to `{21.5}` — try it after the tests pass and watch the difference.
- `fmt.Sprint(t)`, `fmt.Sprintf("%v", t)`, and `fmt.Sprintf("%s", t)` must all produce the custom format — that's `fmt` detecting the interface, not you calling `String()` yourself.
- Elements inside composites format too: `fmt.Sprint([]Temp{{1}, {2}})` → `"[1°C 2°C]"` — for free.
- Compile-time check required in your code: `var _ fmt.Stringer = Temp{}`.

## The trap you must trigger once (not tested — it crashes)

Write the body as:

```go
func (t Temp) String() string {
    return fmt.Sprintf("%v°C", t)   // ← %v on t, the Temp itself
}
```

`%v` sees a `fmt.Stringer` → calls `String()` → which calls `Sprintf("%v", t)` → which calls `String()`… The stack overflows and the runtime kills the process. Run it, read the repeating stack trace, then fix it by formatting the **field**, not the struct: `Sprintf("%g°C", t.C)` (or `strconv.FormatFloat`). Every Go engineer does this exactly once; do yours here, cheaply.

## Worked examples

```go
t := Temp{21.5}
fmt.Println(t)               // 21.5°C   (Println uses Stringer)
fmt.Sprintf("temp: %v", t)   // "temp: 21.5°C"
fmt.Sprintf("%s outside", t) // "21.5°C outside"
fmt.Sprint(&t)               // "21.5°C" (pointer works too — method set rule)
```

Without `String()`, those would print `{21.5}` and `&{21.5}` — the default struct rendering.

## Edge cases the tests cover

- Whole-number temperature (no trailing `.0` — the `%g` pin).
- Zero, negative, and fractional values, including `-273.15`.
- All three fmt paths: `Sprint`, `%v`, `%s`.
- Printing through a pointer (`&t`) — the method-set rule made visible.
- A `[]Temp` rendering each element with the custom format.
- The compile-time `var _ fmt.Stringer` assertion (in *your* file — the test file has its own).
