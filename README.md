# bedrock-jsonfix

`bedrock-jsonfix` is a production-oriented Go library for normalizing **Minecraft Bedrock tolerant JSON-ish** payloads into strict valid JSON.

It is designed for bots, services, webhooks, and CLIs that receive imperfect JSON (comments, trailing commas, weird whitespace/encoding, extra garbage around root objects).

It targets **Go 1.26+** for development and CI.

## Install

```bash
go get github.com/SkaticNET/bedrock-jsonfix@latest
```

## Output contract

- Output is always **strict JSON** (single top-level document).
- Normalized output is newline-terminated (`\n`).
- If `PreserveIfValid` returns the original input, trailing newline behavior is preserved from input.
- Input fixups (comments, trailing commas, junk trimming, encoding cleanup) are only to recover incoming payloads, not to claim Minecraft itself supports those extensions.

A primary sanitization reason is UTF-8 BOM handling: Bedrock tooling and game ingestion commonly fail on BOM-prefixed JSON, so BOM is removed during normalization.

## Quick start

### Example: FixBytes

```go
opt := bedrockjsonfix.DefaultOptions()
opt.Pretty = true
res, err := bedrockjsonfix.FixBytes(input, opt)
if err != nil {
	panic(err)
}
fmt.Println(string(res.Output))
```

### Example: FixReader with context

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

opt := bedrockjsonfix.DefaultOptions()
opt.MaxInputBytes = 2 << 20
res, err := bedrockjsonfix.FixReader(ctx, r, opt)
if err != nil {
	panic(err)
}
_ = res
```

## Options overview

- `Mode`: `ModeStrict`, `ModeBedrock`, `ModeBedrockSafe`
- `Pretty`, `Indent`, `Prefix`
- `PreserveIfValid`
- `MaxInputBytes`, `MaxOutputBytes`
- `AllowCP1252Fallback`
- `AggressiveWhitespace`
- `DropJunkOutsideStrings`
- `TrimToFirstRoot`, `TrimAfterFirstRoot`
- `RootPolicy`, `RootScanMaxCandidates`, `WrongStartMaxOffset`
- `RootValidator`
- `RootScanAttempts`
- `EscapeStringControls`

Use `DefaultOptions()` for safe service defaults.

## Common scenarios

- Bedrock `manifest.json` with comments and trailing commas.
- Windows-1252 text and CRLF newlines.
- Broken payloads containing garbage after top-level JSON.

## Error handling

```go
res, err := bedrockjsonfix.FixBytes(input, bedrockjsonfix.DefaultOptions())
if err != nil {
	if errors.Is(err, bedrockjsonfix.ErrInputTooLarge) {
		// reject payload
	}
	if errors.Is(err, bedrockjsonfix.ErrInvalidJSON) {
		// report invalid content
	}
	var fe *bedrockjsonfix.FixError
	if errors.As(err, &fe) {
		fmt.Println("stable code:", fe.Code)
	}
}
_ = res
```

## Security notes

- This library does **not** execute input.
- It only decodes, sanitizes, and re-encodes JSON.
- Input/output caps reduce DoS risk in public services.

## Examples

- `examples/fixfile`: read file, normalize, write result.
- `examples/http`: safe HTTP handler usage with context timeout and size caps.

## License

MIT.

