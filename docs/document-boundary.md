# The first completed document

`Fix` selects one JSON object or array. Bytes after its closing delimiter do not
belong to that document. Version 2.0.3 applies this rule during comment removal
as well: `{"a":1}/*` returns `{"a":1}`. Previously, the unfinished comment caused
the entire input to fail before document selection.

The comment pass now stops when the first object or array closes outside
strings and comments. The existing matching-delimiter check and final strict
JSON validation still run. An unfinished comment before or inside the selected
document remains an error, and an invalid first candidate cannot be replaced
by a later valid candidate. The 128 MiB whole-input check runs before selection.
Member spelling, duplicate members, their source order and quoted text remain
unchanged.

A static comparison with Minecraft Windows client 1.26.52.3 motivated the fix.
In that client's Music reader, constructor instruction `0x14e033682`
(`66 c7 87 b8 00 00 00 01 00`) enables comment skipping and disables the
root-type check. Range parser `0x14e036ae0` copies the parsed root before reading
the next token. An unfinished trailing block comment becomes token `0xd` in
`0x14e0344d0`; the range parser exits its comment loop and returns the saved
successful root status at `0x14e036cb0`.

This is static evidence for the completed-root/trailing-comment case, not a
complete emulation of the client's parser. The executable SHA-256 is
`4a92bfa3ce2428b40ee517b7c1125d9f6f79274382de03846351992663b2e2e4`.
Native client execution was **NOT_RUN**. The existing
[Music reader analysis](https://github.com/SkaticNET/bedrock-pack-core/blob/1e495e8/research/results/music-first-source-1.26.52.md)
describes the separate partial-root failure paths.
