# bedrock-jsonfix

`bedrock-jsonfix` converts one Bedrock-style JSON object or array into strict,
valid UTF-8 JSON without rebuilding its object structure.

```go
fixed, err := bedrockjsonfix.Fix(input)
```

## Install

```bash
go get github.com/SkaticNET/bedrock-jsonfix/v2@v2.0.0
```

```go
import "github.com/SkaticNET/bedrock-jsonfix/v2/bedrockjsonfix"
```

## What it fixes

Bedrock resource-pack JSON is often JSON-with-comments rather than strict JSON.
Mojang's own sample UI files use line and block comments, which Go's strict
`encoding/json` rejects. `Fix` handles a leading UTF-8 BOM, line and block
comments, trailing commas, stray bytes around the first complete JSON object or
array, raw control bytes, and numbers with redundant leading zeros (`026` becomes
`26`). Fractions and exponents keep their original spelling and precision.
It preserves quoted content, member spelling,
source member order, and duplicate keys exactly as written.

It is deliberately tolerant of common authoring noise around a document, so
the recovery behavior is broader than a strict JSON reader.

It does not merge documents, normalize keys, infer schemas, or change values.
The result is always one valid JSON document or an error. Invalid UTF-8 bytes
are recovered as Windows-1252; valid UTF-8 sequences are preserved. Input must
be no larger than 128 MiB.

## Resource-pack example

The official [Bedrock Samples resource pack](https://github.com/Mojang/bedrock-samples/tree/main/resource_pack)
includes a strict [pack manifest](https://github.com/Mojang/bedrock-samples/blob/main/resource_pack/manifest.json)
alongside comment-bearing resource JSON such as
[item_texture.json](https://github.com/Mojang/bedrock-samples/blob/main/resource_pack/textures/item_texture.json)
and [manifest_validation_screen.json](https://github.com/Mojang/bedrock-samples/blob/main/resource_pack/ui/manifest_validation_screen.json).
Use the fixer at the file-read boundary before a strict decoder consumes the
resource JSON.

## File-to-file use

The runnable example includes a deliberately messy
[`hello.json`](examples/fixfile/testdata/hello.json):

```jsonc
discarded prefix /* Resource-pack authoring comment */
{
  "name": "hello",
  "items": ["first",],
}
ignored trailer
```

Copy it to a working directory, then write the repaired document to a separate
file:

```bash
cp examples/fixfile/testdata/hello.json ./hello.json
go run ./examples/fixfile hello.json hello.cleaned.json
```

`hello.cleaned.json` is valid JSON and retains the document's permitted
whitespace and member order:

```json
{
  "name": "hello",
  "items": ["first"]
}
```

Pass `--report` when a machine-readable result summary is useful:

```bash
go run ./examples/fixfile --report hello.report.json hello.json hello.cleaned.json
```

The optional `hello.report.json` is strict JSON as well:

```json
{
  "input_path": "hello.json",
  "output_path": "hello.cleaned.json",
  "input_bytes": 117,
  "output_bytes": 43,
  "changed": true,
  "valid_json": true
}
```

See [`examples/fixfile`](examples/fixfile) for the complete command.

## Scope

This package repairs JSON syntax for machines that require valid JSON. It is not
a resource-pack merger, a JSON formatter, or a replacement for Bedrock's own
resource loading and UI-combination behavior.
