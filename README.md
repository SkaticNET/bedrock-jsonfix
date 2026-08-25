# bedrock-jsonfix

```go
fixed, err := bedrockjsonfix.Fix(input)
```

`Fix` converts a single Bedrock-style JSON object or array into valid UTF-8 JSON.
It keeps JSON member names and values intact while removing a BOM, comments,
trailing commas, and bytes outside the first complete document.
