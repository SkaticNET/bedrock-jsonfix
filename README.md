# bedrock-jsonfix

`bedrock-jsonfix` turns a Bedrock-style JSON object or array into valid UTF-8 JSON.

```go
fixed, err := bedrockjsonfix.Fix(input)
```

## Install

```bash
go get github.com/SkaticNET/bedrock-jsonfix@v1.2.2
```

## Behavior

`Fix` preserves JSON member names and values. It removes a leading UTF-8 BOM,
line and block comments, trailing commas, non-JSON bytes around the first complete
object or array, and malformed control bytes outside strings. Raw control bytes
inside strings are escaped.

The returned bytes are one valid JSON document. Input must be valid UTF-8 after
the BOM is removed and must not exceed 128 MiB. Incomplete or structurally invalid
JSON returns an error.

## Example

```go
input := []byte(`{
  // pack metadata
  "name": "Example",
  "modules": ["resources",],
}`)

fixed, err := bedrockjsonfix.Fix(input)
if err != nil {
	panic(err)
}
fmt.Println(string(fixed))
```

See [`examples/fixfile`](examples/fixfile) for a complete file-to-file program.
