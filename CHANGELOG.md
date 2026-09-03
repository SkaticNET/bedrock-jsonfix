# Changelog

## v2.0.1 — 2026-09-03

- Fixed legacy Bedrock JSON containing Windows-1252 bytes, including raw `§`
  formatting codes, by converting invalid bytes to UTF-8 before JSON
  validation.
- Fixed invalid UTF-8 JSON bypassing normalization and later becoming Unicode
  replacement characters.
- Preserved already-valid UTF-8 sequences in mixed legacy files.
- No public API changes.

## v2.0.0 — 2026-08-25

- **Breaking:** the module path is now
  `github.com/SkaticNET/bedrock-jsonfix/v2`.
- The public API is `Fix([]byte) ([]byte, error)`.
- Bedrock JSON repairs cover BOMs, comments, trailing commas, control bytes,
  and surrounding non-JSON data.
- The fixer preserves JSON member spelling and source ordering while producing
  valid JSON.
