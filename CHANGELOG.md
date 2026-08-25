# Changelog

## v1.2.2 — 2026-08-25

- Reduced the public API to `Fix([]byte) ([]byte, error)`.
- Kept the Bedrock JSON repairs for BOMs, comments, trailing commas, control
  bytes, and surrounding non-JSON data.
- Removed configurable parser modes, reports, reader helpers, CLI tools, and
  duplicate examples.
- Preserved JSON member spelling and source ordering while producing valid JSON.
