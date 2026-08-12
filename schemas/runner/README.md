# Runner protocol schemas

Version 1 messages use `runner-message.schema.json`. Receivers negotiate the
highest shared minor version within major version 1; a major mismatch refuses
assignment. Minor-version additions are ignored by consumers and preserved by
generic Go forwarding.

The negotiated limits bound each serialized message, total buffered event data,
and outstanding commands. The compiled defaults are 1 MiB per message and 8 MiB
of buffered data. Protocol errors contain stable codes and resource IDs, never
payload content or lease tokens.
