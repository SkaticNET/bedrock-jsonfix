# Leading zeros in Bedrock JSON

The Windows client 1.26.52.3 accepts numbers such as `026`, `-001.25` and
`001e+02`. Strict JSON rejects them, so `Fix` previously rejected these files.

The client tokenizer at `0x14e0344d0` starts a number at a minus sign or digit
and consumes digits, decimal points, signs and exponent markers. Its integer
decoder at `0x14e035400` accumulates digits in base ten without rejecting leading
zeros. Numbers containing a decimal point or exponent go to `0x14e035d30`,
which calls `0x14028bf70` with the format string `%lf` at `0x14efaea58`.

The inspected executable has SHA-256
`4a92bfa3ce2428b40ee517b7c1125d9f6f79274382de03846351992663b2e2e4`.
This is static client evidence; the client was not executed.

`Fix` removes redundant zeros from the integer part and validates the result
with `encoding/json`. It preserves the sign, fractional digits and exponent
text. It does not parse numbers into floating-point values, change text inside
strings or accept every spelling supported by the client's numeric conversion.
For example, `1.` remains an error.

Regression cases are in `bedrockjsonfix/numbers_test.go`, including version
arrays, decimal `08`, precise numeric text, quoted numbers, other repairs and
malformed numbers.
