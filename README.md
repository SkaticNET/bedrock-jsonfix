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

Bedrock resource-pack JSON commonly contains authoring syntax that Go's strict
`encoding/json` rejects. `Fix` handles a leading UTF-8 BOM, line and block
comments, trailing commas, stray bytes around the first complete JSON object or
array, and raw control bytes. It preserves quoted content, member spelling,
source member order, and duplicate keys exactly as written.

It does not merge documents, normalize keys, infer schemas, or change values.
The result is always one valid JSON document or an error. Input must be valid
UTF-8 after the BOM is removed and be no larger than 128 MiB.

## Resource-pack example

The official [Bedrock Samples resource pack](https://github.com/Mojang/bedrock-samples/tree/main/resource_pack)
includes JSON in familiar forms: a [pack manifest](https://github.com/Mojang/bedrock-samples/blob/main/resource_pack/manifest.json),
texture metadata, and JSON-UI files such as [manifest_validation_screen.json](https://github.com/Mojang/bedrock-samples/blob/main/resource_pack/ui/manifest_validation_screen.json).
Use the fixer at the read boundary before passing a document to a strict JSON
decoder:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/SkaticNET/bedrock-jsonfix/v2/bedrockjsonfix"
)

func main() {
	input := []byte(`{
	  // Resource-pack UI authoring comment
	  "namespace": "sample_ui",
	  "title": { "type": "label", },
	}`)

	fixed, err := bedrockjsonfix.Fix(input)
	if err != nil {
		panic(err)
	}

	var document map[string]json.RawMessage
	if err := json.Unmarshal(fixed, &document); err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", fixed)
}
```

The output is strict JSON:

```json
{"namespace":"sample_ui","title":{"type":"label"}}
```

## File-to-file use

The included example writes repaired JSON to a separate path:

```bash
go run ./examples/fixfile input.json output.json
```

See [`examples/fixfile`](examples/fixfile) for the complete program.

## Scope

This package repairs JSON syntax for machines that require valid JSON. It is not
a resource-pack merger, a JSON formatter, or a replacement for Bedrock's own
resource loading and UI-combination behavior.
