// Package bedrockjsonfix normalizes Minecraft Bedrock tolerant JSON-ish input
// into strict JSON with safety limits and a structured cleanup report.
//
// ModeStrict accepts strict JSON only and only performs formatting/canonicalization.
// ModeBedrock (and ModeBedrockSafe alias) enables tolerant parsing features such as
// comment stripping, trailing-comma removal, root trimming, and policy-controlled
// root-scan fallback.
//
// The API is designed for bots and services: use DefaultOptions and enforce
// MaxInputBytes/MaxOutputBytes to protect against oversized payloads.
package bedrockjsonfix
