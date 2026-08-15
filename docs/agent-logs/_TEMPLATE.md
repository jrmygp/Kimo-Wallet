<!--
HOW TO USE THIS FILE

1. Open `docs/agent-logs/YYYY-MM-DD.md` for TODAY. One file per day — never a second file for
   the same date.
2. If it does not exist, create it starting with the day header below.
3. If it DOES exist, read every entry in it first, then APPEND the entry block below at the
   bottom, numbered as the next Entry N.
4. Never edit, reorder, or delete an entry someone else wrote. See `docs/CLAUDE.md` §2.2.
5. Prepend one line to `docs/agent-logs/INDEX.md` — one line per entry, newest first.
-->

<!-- ===== DAY HEADER — only when creating the file for a new date ===== -->

# Agent Change Log — YYYY-MM-DD

Append-only. Add new sessions at the BOTTOM as the next `## Entry N`.
Never edit, reorder, or delete an existing entry — see `docs/CLAUDE.md` §2.2.

<!-- ===== ENTRY BLOCK — copy from here down, for every session ===== -->

---

## Entry N — <short title>

- **time:** <HH:MM local, or "unknown">
- **agent:** <model / agent name>
- **task:** <one line: what you were asked to do>
- **status:** completed | partial | blocked
- **scope:** apps/web, apps/transaction-service, packages/contracts
- **touches_money:** yes | no

### What changed
- <file:line> — <what and why, one line each>

### Why
<the reasoning / the decision made, including alternatives rejected>

### Invariants affected
<money, auth, contracts, migrations, or "none">

### Verification
<exact commands run + real result. "Not run" is an acceptable answer; a fabricated pass is not.>

### Money checklist (delete if touches_money: no)
- [ ] What happens if this request arrives twice?
- [ ] What happens if this runs concurrently with itself on the same wallet?
- [ ] What happens if the process dies halfway through?
- [ ] What happens if the event is delivered three times?
- [ ] Is the ledger still balanced afterwards?
- [ ] Is there a test that proves the duplicate is rejected?

### Known gaps / TODO for the next agent
- <the thing you did not finish, and where to pick it up>

### Do NOT do
- <traps you found: things that look wrong but are intentional, things that will break>
