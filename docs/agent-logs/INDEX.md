# Agent Change Log — Index

**Newest first.** Every agent working in this repo must read the top entries here *before* editing,
and prepend a line here *after* editing. See `docs/CLAUDE.md` §2 for the rules and entry template.

Line format:

```markdown
- YYYY-MM-DD — Entry N — [Short title](YYYY-MM-DD.md) — <area> — money:yes|no — completed|partial|blocked
```

Rules:

- **Prepend.** Never reorder, edit, or delete an existing line.
- **One line per entry, not per file.** Logs are one file per day, so several lines will point at
  the same day file — the entry number tells them apart.
- `money:yes` means the change touched balances, ledger, transactions, idempotency, or payment flows.

## Entries

- 2026-08-22 — Entry 3 — [JWT access tokens + api-gateway auth middleware](2026-08-22.md) — apps/user-service, apps/api-gateway — money:no — completed
- 2026-08-22 — Entry 2 — [Standardised response envelope: {statusCode, message, data}](2026-08-22.md) — apps/api-gateway — money:no — completed
- 2026-08-22 — Entry 1 — [Fix Login: unimplemented RPC + two repository bugs](2026-08-22.md) — apps/user-service — money:no — completed
- 2026-08-21 — Entry 4 — [API Gateway (HTTP → gRPC, wraps user-service's Register)](2026-08-21.md) — apps/api-gateway — money:no — completed
- 2026-08-21 — Entry 3 — [Auto-load .env via joho/godotenv](2026-08-21.md) — apps/user-service — money:no — completed
- 2026-08-21 — Entry 2 — [Switch user-service persistence from raw pgx to GORM](2026-08-21.md) — apps/user-service — money:no — completed
- 2026-08-21 — Entry 1 — [user-service bootstrap + Register (phone + full name only)](2026-08-21.md) — apps/user-service, packages/contracts — money:no — completed
- 2026-08-17 — Entry 1 — [Transaction History page with verified infinite scroll](2026-08-17.md) — apps/web — money:no — completed
- 2026-08-15 — Entry 4 — [Mandatory UI verification before reporting done](2026-08-15.md) — docs — money:no — completed
- 2026-08-15 — Entry 3 — [Change log switched to one file per day](2026-08-15.md) — docs — money:no — completed
- 2026-08-15 — Entry 2 — [Formik + Yup form rules added to ruleset](2026-08-15.md) — docs — money:no — completed
- 2026-08-15 — Entry 1 — [Agent rules and change-log protocol established](2026-08-15.md) — docs — money:no — completed
