# Changelog

## v2.0.0 — 2026-08-25

- **Breaking:** the module path is now
  `github.com/SkaticNET/bedrock-jsonfix/v2`.
- The public API is `Fix([]byte) ([]byte, error)`.
- Bedrock JSON repairs cover BOMs, comments, trailing commas, control bytes,
  and surrounding non-JSON data.
- The fixer preserves JSON member spelling and source ordering while producing
  valid JSON.
