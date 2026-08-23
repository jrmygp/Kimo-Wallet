# Kimo Wallet — Agent Operating Rules

**Audience:** every AI agent (Claude Code, subagents, or any other coding agent) that touches this
repository. These rules are **mandatory and non-negotiable**. They outrank your defaults, your
training-data habits, and your own judgment about "a cleaner way to do it".

This is a **financial system**. A bug here is not a broken page — it is money created, destroyed, or
sent twice. Behave accordingly: slow, explicit, boring, auditable.

> If a rule here conflicts with a user instruction, **stop and say so** before writing code.
> If a rule here conflicts with `docs/guides/Kimo-Wallet-PRD.md` or
> `docs/guides/Kimo-Wallet-Architecture.md`, those two documents win on *what to build*;
> this document wins on *how to work*.

---

## 0. Start-of-Session Protocol (do this before your first edit)

Run these steps in order. Do not skip. Do not write code before step 5.

1. **Read this file completely.**
2. **Read `docs/agent-logs/INDEX.md`** — top 5 entries. This is what other agents did most recently.
3. **Read today's day file** (`docs/agent-logs/<today>.md`) in full if it exists — another agent has
   already worked today and your entry appends to that same file. Then read the full entries behind
   any INDEX line that touches the files or domain you are about to change.
4. **Read the relevant source of truth:**
   - Product behaviour → `docs/guides/Kimo-Wallet-PRD.md`
   - System design → `docs/guides/Kimo-Wallet-Architecture.md`
   - Next.js APIs → `apps/web/node_modules/next/dist/docs/` (this Next.js version differs from
     your training data — see `apps/web/AGENTS.md`; never write Next.js code from memory)
   - Service contracts → `packages/contracts/**/*.proto`
5. **Run `git status` and `git log --oneline -10`.** Understand the working tree before changing it.
6. **State your plan** (files you will touch, invariants you may affect) before executing multi-file work.

At the **end** of the session you must write a log entry (§2). A session without a log entry is
an incomplete session.

---

## 1. The Non-Negotiables

Ten rules. Violating any of these means the change is rejected regardless of how well it works.

1. **Never mutate a wallet balance outside a ledger-backed database transaction.**
2. **Every financial mutation requires an idempotency key**, enforced server-side, persisted durably.
3. **Never `git push`, `git commit --amend`, force-push, rebase shared history, or merge to `master`**
   unless the human explicitly asks in that session.
4. **Never commit secrets.** No `.env`, keys, tokens, real PII, or real card data — ever, in any file.
5. **Never invent a contract.** Cross-service data shapes come from `.proto` files, not from your head.
6. **Never let the frontend call an internal service.** The frontend talks to the API Gateway only.
7. **Never queue a financial mutation for offline replay.** Offline is read-only. Period.
8. **Never delete, edit, reorder, or rewrite another agent's log entry.** Logs are append-only:
   one file per day, new sessions appended at the bottom. Correct an earlier entry by writing a
   new one that says so — never by changing it.
9. **Never disable, skip, or weaken a test to make a build pass.** Fix the code or report the failure.
10. **Never claim work is done without running the relevant build/lint/test and pasting real
    output — and for any change that alters the UI, never without seeing it rendered.**
    A green build is not evidence that a visual change worked. See §5.6.

---

## 2. The Change Log (mandatory for every agent)

Agents work in isolation and cannot see each other's context. The log is the only shared memory.
**Write the log or the next agent will redo, undo, or break your work.**

### 2.1 Where — one file per DAY, not per session

```text
docs/agent-logs/
├── INDEX.md          # newest-first pointer list — READ THIS FIRST, prepend here
├── _TEMPLATE.md      # copy this for a new entry
├── 2026-08-15.md     # ALL sessions from 15 Aug live in this one file
└── 2026-08-16.md     # ALL sessions from 16 Aug live in this one file
```

**The filename is the date and nothing else: `docs/agent-logs/YYYY-MM-DD.md`.**

- If today's file **does not exist**, create it with the day header (§2.4) and `## Entry 1`.
- If today's file **already exists**, another agent has worked today. **Read every entry in it
  first** — that is your most recent context — then **append** your entry at the bottom as the next
  `## Entry N`.
- **Never create a second file for the same date.** No `2026-08-15-transfer.md`, no `-part2`,
  no `-v2`. One date, one file.

### 2.2 Append-only — the hard rule

**You may only ADD to a day file. You may never remove or rewrite what another session wrote.**

- Do not delete, edit, reword, reorder, summarize, condense, "clean up", "consolidate", or
  "tidy" an existing entry — not even one that is now outdated, wrong, or contradicted by your work.
- If an earlier entry is wrong or you reversed its decision, **say so in your own entry**
  ("Entry 2 stated X; this session reverted it because Y"). The original stays exactly as written.
  The mistake and its correction are both part of the history.
- Do not rewrite the whole file to "reformat" it. Append only.
- Never overwrite a day file with the Write tool if it already has content. Append.

An outdated entry is still useful — it tells the next agent what was tried and why it changed.
A deleted entry is lost context, and lost context is how two agents ship the same bug twice.

### 2.3 When

- **Before you finish your turn**, always — even for a one-line change.
- **Immediately after** any change to money handling, contracts, migrations, or auth. Do not batch it.
- If you were **blocked** and shipped nothing, still log it: what you tried, why you stopped.
- One entry per session/task. If your single session did two unrelated things, that is still one
  entry — list both under "What changed".

### 2.4 Day file header (only when creating the file for a new date)

```markdown
# Agent Change Log — 2026-08-15

Append-only. Add new sessions at the BOTTOM as the next `## Entry N`.
Never edit, reorder, or delete an existing entry — see `docs/CLAUDE.md` §2.2.
```

### 2.5 Entry template (copy verbatim, fill every field)

Note there is **no YAML frontmatter** — a day file holds many entries, so the metadata is a list
inside each entry.

```markdown
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

### Known gaps / TODO for the next agent
- <the thing you did not finish, and where to pick it up>

### Do NOT do
- <traps you found: things that look wrong but are intentional, things that will break>
```

`N` is the next number after the last entry already in the file. Do not renumber existing entries.

### 2.6 INDEX.md

After writing your entry, **prepend** one line to the top of the list in `docs/agent-logs/INDEX.md`:

```markdown
- 2026-08-15 — Entry 3 — [Transfer idempotency keys](2026-08-15.md) — transaction-service — money:yes — completed
```

One line **per entry**, not per file — so a day with three sessions has three lines all pointing at
the same day file, distinguished by entry number. Newest first. Never reorder, edit, or remove an
existing line.

---

## 3. Money Safety — Preventing Double Disbursement

This section is the reason this document exists. Read it every session where `touches_money: yes`.

### 3.1 The threat model

Money moves twice because of, in order of real-world frequency:

| Cause | Defence |
|---|---|
| User double-taps "Send" | Client-side single-flight + idempotency key minted once per intent |
| Client/network retry on timeout | Server-side idempotency record |
| Gateway or proxy retry | Same key propagated end-to-end, never regenerated |
| Kafka at-least-once redelivery | Idempotent consumers keyed on event id |
| Concurrent requests on one wallet | Row lock / atomic conditional update |
| Retry after a partial failure | Explicit transaction state machine + compensation, never blind re-run |
| Manual "just re-run the job" | Every job must be safe to run twice |

Assume **every** message is delivered at least twice and every request may be retried. Design so
that the second delivery is a no-op.

### 3.2 Idempotency — rules

1. Every mutating financial endpoint accepts `Idempotency-Key` (HTTP header) or an
   `idempotency_key` field (gRPC). Missing key on a financial mutation → reject `400`. Never
   auto-generate one server-side to "be helpful".
2. The key is **minted by the client once per user intent**, not per HTTP attempt. A retry reuses it.
3. Store `(idempotency_key, user_id)` **unique-constrained** in the database, in the *same
   transaction* as the financial effect. Not in Redis alone — Redis may be a fast-path cache in
   front of it, never the authority.
4. On a repeat key: return the **stored original result** with the original transaction id.
   Do not re-execute. Do not return a new transaction id.
5. On a repeat key with a **different request body**: reject `409 Conflict`. Never silently
   process the new body.
6. Idempotency records are retained at minimum as long as any client may retry (≥ 24h; prefer
   keeping them with the transaction record permanently).
7. Reserve the key **before** the effect, commit both **atomically**. A key written after a
   successful transfer is a race, not a defence.

### 3.3 Ledger — rules

1. The ledger is **append-only**. No `UPDATE`, no `DELETE` on ledger rows. Ever. Corrections are
   new compensating entries.
2. Every money movement is **double-entry**: a DEBIT and a matching CREDIT, equal in amount,
   sharing one `transaction_id`. They are written in one database transaction or neither is written.
3. A wallet's balance must always be reconcilable to `SUM(credits) - SUM(debits)`. If you add a
   cached/materialised balance column, it is updated only inside the same transaction that writes
   the ledger entries.
4. Money is **integer minor units** (`int64` rupiah / cents). **Never** `float`, `float64`,
   `double`, or JS `number` for an amount in storage or arithmetic. Serialize as string or integer
   in JSON. A `float` amount anywhere is an automatic rejection.
5. Amounts are always accompanied by a currency. No bare amounts crossing a service boundary.
6. Negative or zero amounts are rejected at the boundary; direction is expressed by
   `direction` / entry type, never by a negative amount.

### 3.4 Concurrency — rules

1. Debiting a wallet requires either `SELECT ... FOR UPDATE` on the wallet row, or an atomic
   conditional update (`UPDATE ... WHERE balance >= :amount` and check rows-affected). Read-then-write
   without one of these is a race, no matter how narrow the window looks.
2. Lock wallets in a **deterministic order** (e.g. ascending wallet id) in any multi-wallet
   operation, to avoid deadlock on A→B / B→A transfers.
3. Never hold a database transaction open across a network call (gRPC, Kafka publish, HTTP).
   Use the **transactional outbox** pattern: write the event row in the same transaction, publish
   from the outbox afterwards.
4. Never validate balance in the API Gateway or the frontend as an authority. Those are UX hints.
   The only authoritative check is inside the locked transaction.
5. `DEFAULT` isolation is not automatically enough — state the isolation level you rely on in a
   comment above the transaction and in your log entry.

### 3.5 Transaction state machine — rules

1. Legal transitions only: `CREATED → PENDING → PROCESSING → COMPLETED | FAILED`, and
   `FAILED → COMPENSATED` where a reversal is required. Encode this in one place; reject anything else.
2. Transitions are **guarded and atomic**: `UPDATE ... WHERE id = :id AND status = :expected_status`,
   then check rows-affected. Never `UPDATE ... WHERE id = :id` alone.
3. `COMPLETED` and `FAILED` are **terminal**. A terminal transaction is never re-processed —
   a redelivered event for a terminal transaction is acknowledged and dropped.
4. A failure after a debit must produce an explicit compensating ledger entry, never a silent
   rollback of history.

### 3.6 Event consumers — rules

1. Every consumer is idempotent: dedupe on `event_id` (or `transaction_id` + event type) in a
   processed-events table, checked in the same transaction as the effect.
2. Events are named as **facts in the past tense** (`transaction.completed`), never as commands.
3. Retries use bounded, backed-off attempts and then a dead-letter queue. Never an infinite
   tight retry loop. Never `for { retry() }`.
4. A consumer must tolerate **out-of-order** delivery. Do not assume `created` arrives before `completed`.

### 3.7 Frontend money rules

1. The "Send" / "Pay" / "Top Up" submit button is **disabled from the moment the request is
   in-flight** until a terminal response, in addition to server-side idempotency. Both, not either.
2. Generate the idempotency key **when the user reaches the review/confirm step**, store it with
   the pending intent, and reuse it for every retry of that intent. Regenerating it on retry
   defeats the entire mechanism.
3. Never auto-retry a mutation. Retry only on explicit user action, reusing the same key.
4. Never optimistically update a displayed balance from a client-side computation. Show the
   server's balance, or show a pending state.
5. On timeout / network error, the correct message is **"we're confirming your transaction"** with a
   status poll — never "failed", which invites the user to send again.
6. Offline: mutations are blocked in the UI with a clear message. Never buffer them.

### 3.8 Money-change checklist

Anything with `touches_money: yes` must answer all of these in the log entry:

- [ ] What happens if this request arrives twice?
- [ ] What happens if this runs concurrently with itself on the same wallet?
- [ ] What happens if the process dies halfway through?
- [ ] What happens if the event is delivered three times?
- [ ] Is the ledger still balanced afterwards?
- [ ] Is there a test that proves the duplicate is rejected?

---

## 4. Architecture Boundaries

1. **A service owns its data.** Never query, join to, or migrate another service's tables. Cross
   boundaries via gRPC or events only.
2. **The API Gateway holds no domain logic.** Auth, validation of shape, rate limiting, protocol
   translation, aggregation — nothing else. No balance arithmetic, no transaction rules.
3. **The frontend never talks to internal services.** Gateway only.
4. **Contracts are `.proto` files under `packages/contracts/`**, versioned by directory
   (`user/v1`). Generated code lives with the consuming service (`apps/*/gen/`) and is never
   hand-edited. Changing a published contract in a breaking way requires a new version directory.
5. **Redis is never authoritative** for balances, ledger, or the durable idempotency record.
6. **Do not add a new service, database, message broker, or infrastructure dependency** without the
   human explicitly asking. Prefer boring. Prefer fewer moving parts.
7. **Do not add a runtime dependency** (npm/Go module) without stating why the stdlib or an
   existing dependency is insufficient. Log it in your entry.

---

## 5. Code Standards

### 5.1 Universal

- **Match the surrounding code.** Its naming, its comment density, its file layout, its error style.
  Consistency with the existing codebase outranks your preferred style.
- **Small, single-purpose changes.** Do not opportunistically refactor unrelated code while fixing
  a bug. If you spot something, note it under "Known gaps" in your log.
- **No dead code, no commented-out code, no `TODO` without a name and reason.**
- **No placeholder implementations presented as complete.** If a function is stubbed, it says so
  and your log says so.
- **Errors are handled, wrapped with context, and never swallowed.** No empty catch blocks, no
  ignored error returns, no `_ = err`.
- **Never log secrets, tokens, PINs, passwords, full card numbers, or full account identifiers.**
  Log transaction ids and trace ids instead.
- **Validate at the boundary.** Every external input is untrusted: type, range, currency, ownership.

### 5.2 Frontend (`apps/web` — Next.js + TypeScript)

- **Read `node_modules/next/dist/docs/` before writing Next.js code.** This version has breaking
  changes from your training data. Do not guess at APIs.
- Keep `apps/web/AGENTS.md`'s generated block intact — it is regenerated by `next dev`.
- **Server Components by default.** `"use client"` only where interactivity genuinely requires it,
  and as deep in the tree as possible.
- **`strict: true` TypeScript.** No `any`, no `as unknown as`, no `@ts-ignore` / `@ts-expect-error`
  without a comment explaining the exact reason. No non-null `!` on data from the network.
- **Feature-first layout**: `features/<domain>/` holds that domain's components, hooks, types, and
  API access. Shared primitives go in `components/` and `lib/`. Do not scatter a feature across the tree.
- **Validate API responses at the edge with a Yup schema.** A network response is not a typed object
  just because you annotated it. Use the same Yup schemas here as for forms — one validation
  library across the frontend, no second one bolted on.
- Never put secrets in `NEXT_PUBLIC_*`. Anything `NEXT_PUBLIC_` is published to the world.
- Accessibility is a requirement, not a polish step: semantic HTML, labelled controls, keyboard
  reachable, ≥44px touch targets, visible error and status states.
- Mobile-first: design for 360–430px, then widen.

### 5.3 Forms & Validation (Formik + Yup — mandatory, no exceptions)

The form stack is **Formik** for form state and **Yup** for schemas, per
`docs/guides/Kimo-Wallet-Architecture.md` §15.

1. **Do not introduce another form or validation library.** No React Hook Form, no Zod, no Valibot,
   no ad-hoc `useState` form handling for anything with more than one field or any validation.
   If you believe Formik cannot do what is needed, say so and stop — do not silently reach for a
   second library. Mixed form stacks are how a codebase becomes unmaintainable.
2. **Every form has a Yup schema.** No inline `if (!value) setError(...)` validation. The schema is
   the single declaration of what a valid submission is.
3. **Schema location**: `features/<domain>/schemas/<name>.schema.ts`, exported alongside its inferred
   type (`yup.InferType<typeof schema>`). Never redeclare a form's shape as a hand-written
   `interface` next to its schema — infer it, so the two cannot drift.
4. **Wire the schema in**, never re-validate by hand: `validationSchema` on `<Formik>` /
   `useFormik`. Cross-field rules use `yup.ref` / `.test()` inside the schema.
5. **Client validation is UX, never a security or financial control.** The server re-validates
   everything: amount, currency, balance, ownership, PIN. A Yup schema stops a typo; it does not
   stop an attacker. See §1 rule 6 and §6.
6. **Accessibility**: bind errors to inputs via `aria-describedby` and `aria-invalid`, show them
   after `touched`, and move focus to the first error on failed submit. A red border alone is not
   an error message.

#### Money forms (Top Up, Send, Pay, QR confirm) — additional rules

7. **Amount fields are strings in form state and integer minor units on submit.** Parse and convert
   once, in one shared helper, validated by the schema. `yup.number()` on a rupiah field invites
   float rounding — use a string field with a digits-only `.matches()` plus a `.test()` for the
   min/max range, then convert to `int64`-safe integer minor units at the boundary.
   Re-read §3.3 rule 4: **a float amount anywhere is an automatic rejection.**
8. **Formik's `isSubmitting` is a UX guard, not the duplicate-payment defence.** The submit button
   is disabled while `isSubmitting`, *and* the request carries an idempotency key, *and* the server
   enforces it. All three. Never argue that one of them makes another redundant.
9. **The idempotency key is not Formik state.** Mint it when the user reaches the review/confirm
   step, hold it with the pending intent outside the form, and reuse it for every retry of that
   intent. A key stored in form state dies on remount and re-mints on retry — which is exactly the
   double-disbursement bug this whole document exists to prevent. See §3.7 rule 2.
10. **Never `resetForm()` or navigate away on a network error or timeout.** The intent, and its key,
    must survive so the retry is the *same* payment. Reset only after a terminal server response.
11. **Never call `setSubmitting(false)` in a way that re-enables the button while the request is
    still in flight**, including inside `finally` blocks that race the response.
12. **Never auto-submit a form** — no submit-on-mount, no submit-on-change, no retry-on-error
    inside `onSubmit`. A money form submits only on an explicit user action.
13. **Offline: block submission in the UI** with the §10 PRD message. Do not let Formik queue,
    buffer, or persist a pending money submission for later replay. See §1 rule 7.

### 5.4 Backend (Go services)

- Standard layout per the architecture doc: `cmd/`, `internal/`, `migrations/`, `gen/`.
- **GORM is the mandatory ORM for every Go service — no exceptions.** Do not reach for
  raw `database/sql`/`pgx` queries, a second ORM (sqlx, ent, sqlc, etc.), or a hand-rolled
  query builder. If GORM genuinely can't do something a service needs, say so and stop —
  do not silently reach for a second data-access library. `user-service` is the reference
  implementation (`internal/storage/postgres/`).
  - **Never use GORM's `AutoMigrate`.** Schema changes are explicit, forward-only SQL
    files under `migrations/`, applied by the service's own migration runner at startup —
    never inferred from Go struct tags at runtime. `AutoMigrate` lets a real database get
    altered with no reviewable diff, which is exactly what the forward-only-migrations
    rule below exists to prevent.
  - Open the connection with `TranslateError: true` so driver-specific errors (e.g. a
    unique-constraint violation) unwrap cleanly via `errors.As` into a typed driver error
    (e.g. `*pgconn.PgError`), instead of surfacing as an opaque `database/sql` error.
  - GORM's default logger writes plain-text lines straight to stdout and can mislabel an
    expected outcome (e.g. a duplicate-key conflict) as an "ERROR". Silence it
    (`logger.Default.LogMode(logger.Silent)`) and rely on the service's own structured
    `slog` logging instead — otherwise it bypasses the structured-JSON-logs rule below.
  - Keep GORM row-mapping structs (the ones carrying `gorm:"..."` tags) separate from a
    service's domain types. Domain types stay ORM-tag-free; the storage layer maps
    between the two.
- `context.Context` is the first parameter on anything doing I/O, and it is actually honoured
  (timeouts, cancellation) — not accepted and ignored.
- Wrap errors with `fmt.Errorf("...: %w", err)`. Use typed/sentinel errors at domain boundaries so
  callers can distinguish validation failure, business-rule failure, and infrastructure failure.
- `gofmt` / `go vet` clean. No panics in request paths.
- Migrations are **forward-only and additive**. Never edit a migration that has been applied; write a
  new one. Never write a destructive migration without explicit human approval in that session.
- Structured JSON logs including `service`, `trace_id`, and `transaction_id` where relevant.
- Propagate trace context across gRPC and event boundaries.

### 5.5 Testing

- Every money path needs tests for: the happy path, the **duplicate request**, the **concurrent
  request**, and the **insufficient balance** case. A money change with no duplicate-request test
  is incomplete.
- Every event consumer needs a **redelivery** test.
- Every Yup schema on a money form needs unit tests for its boundary cases: zero, negative,
  non-numeric, above balance, and above any max. The schema is a contract — test it directly,
  not only through the rendered form.
- Test behaviour and invariants, not implementation details.
- Tests must be deterministic — no reliance on wall-clock timing, ordering, or sleeps.
- Run the tests. Paste real output into the log. Never describe a test result you did not observe.

### 5.6 UI Verification — look at it before you call it done

**A passing build is not evidence that a UI change works.** `tsc`, `eslint` and `next build` check
that code *compiles*. They cannot see that a style was overridden, a class did nothing, an element
rendered at zero height, or that the page looks exactly as it did before you touched it. Every
CSS bug in §5.6.3 below was found in this repo **after** a completely clean build.

This rule exists because "done" was repeatedly reported for changes that did nothing at all. That
wastes far more time and tokens than verifying would have: the user reports it, the agent
re-reads the file, re-derives the context, and re-does the work — often several times.

#### 5.6.1 The requirement

If your change touches **anything the user can see** — a component, a class name, a layout, a
style, a page, a conditional render — you must **observe the rendered result** before reporting
completion. Not the diff. Not the build log. The rendered UI.

Acceptable evidence, in order of preference:

1. **A screenshot of the actual page** via the browser tools (`claude-in-chrome`), navigated to the
   affected route, in the state that triggers the change (error state, sheet open, list scrolled).
2. **A DOM/computed-style read** of the specific element, proving the class landed and resolved —
   e.g. reading the element's `class` attribute and its computed `border-color` / `height`.
3. **A reproduction of the resolved CSS** outside the browser when the question is purely one of
   class merging — e.g. running `cn()` in node to prove which class survives `tailwind-merge`.

Evidence that does **not** count, on its own:

- "`next build` succeeded."
- "TypeScript and ESLint pass."
- "The class is present in the file."
- "This should work" / "this will now render correctly."
- Any sentence describing an appearance you did not actually look at.

#### 5.6.2 When you cannot verify

Sometimes the browser tooling is unavailable, the extension is not connected, or no dev server is
reachable. That is a normal situation and there is exactly one correct response:

> **Say so explicitly, in the same message, and do not use the word "done".**

State what you changed, state the mechanism you expect to fix it, state that you could **not**
visually confirm it, and tell the user precisely what to look at. "I verified the build; I could
not verify the rendering — please check whether the border is now red on the Number field" is an
honest, useful report. "Done, the border is now red" — when you never saw it — is not, and it is
the specific failure this rule forbids.

Never substitute confidence for observation. If you did not see it, do not describe it.

#### 5.6.3 Repo-specific CSS and interaction traps (all found here, all build-clean)

Check these before assuming a style or an interaction will work. Each one silently does nothing:

1. **Variant specificity beats plain classes.** A built-in `data-[side=bottom]:h-auto` in a shadcn
   component outranks your `h-[80vh]`, because a class+attribute selector has higher specificity.
   Override using the *same* variant (`data-[side=bottom]:h-[80vh]`) so `tailwind-merge` recognises
   the conflict and drops one.
2. **Percentage heights need a definite-height ancestor.** `h-full` / `min-h-full` resolve against
   an ancestor's `height`, **not** its `min-height`. A parent with `min-h-screen` gives descendants
   nothing to resolve against; it must be `h-screen`.
3. **Border colour without border width paints nothing.** Tailwind preflight sets `border-width: 0`,
   so `aria-invalid:border-destructive` is invisible unless the base classes also declare a width
   (`border border-transparent`, the pattern `Button` uses).
4. **Conflicting shorthand in one string.** `overflow-hidden overflow-x-auto` sets the same property
   twice; in a raw `className` string (not passed through `cn()`), the winner is decided by
   stylesheet order, not by your intent.
5. **Flex children will not shrink by default.** A `flex-1 overflow-y-auto` scroll area also needs
   `min-h-0`, or `min-height: auto` makes it grow to fit its content and never scroll.
6. **A `size`/variant prop that TypeScript accepts may have no CSS behind it.** `Avatar` accepted
   `size="xl"` while the root element had no `data-[size=xl]` rule at all. Types are not styles.
7. **Flex items also refuse to shrink *below* their declared width by default — the other
   direction of #5.** A row of fixed-width cards (`w-96`) inside `overflow-x-auto` needs
   `shrink-0` on each card, or `flex-shrink: 1` squeezes them all to fit instead of overflowing —
   which looks like "the carousel isn't scrollable" but is actually "there's nothing to scroll to."
8. **An IntersectionObserver-based infinite-scroll sentinel fires on a *transition* into view, not
   on "currently intersecting."** A fast or programmatic scroll that lands exactly at the maximum
   scrollable position in one motion never produces that transition, so the load silently stalls
   even though more data exists — reproduced with `react-infinite-scroll-component` after 3–6 pages
   loaded correctly. A plain `scroll` listener on the real scroll container, checked independently
   of the library's own trigger, is the reliable fix — see `app/wallet/history/page.tsx`.

#### 5.6.4 Verifying an interactive change

For anything with state — a sheet, a dropdown, a form error, a scroll container — verifying the
default render is not enough. Drive it into the state you changed:

- Opened the sheet, and **scrolled the list**?
- Submitted the form, and **seen the error style** on the invalid field?
- Actually **scrolled the carousel** horizontally?
- Checked the state you did *not* change still behaves (no regression)?

A change that only works before you interact with it is not done.

---

## 6. Security

- Passwords and PINs: hashed with a memory-hard KDF (argon2id/bcrypt), never plaintext, never
  reversible, never logged, never returned by an API.
- Short-lived access tokens with refresh; protect and rotate refresh tokens.
- **Authorize every request against the resource owner.** Authentication ≠ authorization. A user
  must never be able to read or move another user's wallet by changing an id in a request.
- Rate-limit auth attempts, PIN attempts, and financial mutations. Lock out on repeated PIN failure.
- No secrets in source control, in logs, in error messages, or in client bundles.
- Financial mutations require explicit authorization and PIN confirmation where the PRD says so.
- If you find a security problem while doing something else: **stop, report it, do not silently fix
  and bury it in an unrelated diff.**

---

## 7. Git & Workflow

- Work on a branch. **Never commit directly to `master`** unless the human says to.
- Commit only when asked. Never `push`, `amend`, force-push, rebase shared history, or
  `reset --hard` without an explicit request in that session.
- Never commit build output (`.next/`, `dist/`, `gen/` where generated at build time),
  `node_modules/`, or `.env*`.
- Commit messages: imperative, scoped, explaining *why*. `fix(transaction): reject replayed
  idempotency key with differing body`.
- Do not `git add -A` blindly — stage the files you actually changed.

---

## 8. Communication Standards

- **Report honestly.** If tests fail, say so and paste the output. If you skipped something, say
  which and why. Do not describe partial work as complete.
- **Do not fabricate** command output, test results, file contents, or another agent's findings.
- **Flag, don't silently decide.** When you hit a genuine fork (two valid designs with different
  consequences), state the tradeoff and your recommendation. When it's a routine judgment call,
  just make it and note it.
- **Stay in scope.** Deliver what was asked. Extra unrequested changes to a financial codebase are
  a liability, not a bonus.
- If you cannot satisfy a rule in this document, **say which rule and why** rather than working
  around it quietly.

---

## 9. Definition of Done

A task is done when **all** of these are true:

1. The change matches the PRD/architecture intent, or the deviation is documented and justified.
2. It compiles: `npm run build` (web) / `go build ./...` (services).
3. It lints: `npm run lint` / `go vet ./...`.
4. Relevant tests pass, and new tests cover the new behaviour — including duplicate/concurrent
   cases if money is involved.
5. No secrets, no `any`, no swallowed errors, no float money, no skipped tests, no form or
   validation library other than Formik + Yup.
6. §3.8 checklist answered if `touches_money: yes`.
7. **Your entry is appended to `docs/agent-logs/<today>.md`** (one file per day — created if this
   is the day's first session, appended to otherwise, never a second file for the same date, never
   overwriting an earlier entry), **and a line is prepended to `INDEX.md`.**
8. **If the change is visible to the user, you have seen it rendered** — screenshot or computed
   style, in the state that triggers it (§5.6). If you could not, you said so explicitly and did
   not call it done.
9. The user has been told plainly what was done, what was verified, and what was not.

---

## 10. Quick Reference

```text
BEFORE:  read this file → read docs/agent-logs/INDEX.md → read today's day file in full
         → read PRD/architecture → read node_modules/next/dist/docs (frontend)
         → git status → state plan

DURING:  money = idempotency key + row lock + double-entry ledger + guarded state transition
         integers only, never floats
         forms = Formik + Yup only; schema per form; key lives outside form state
         Go services = GORM only, never AutoMigrate — migrations are explicit SQL files
         services own their data, contracts come from .proto
         one purpose per change, match surrounding style

AFTER:   build + lint + test, paste real output
         UI change? LOOK AT IT RENDERED — screenshot/computed style, in the state
           that triggers it. A green build proves nothing visual. Cannot see it?
           say so and do NOT say "done".
         APPEND '## Entry N' to docs/agent-logs/YYYY-MM-DD.md (one file per day,
           never a second file for today, never touch an existing entry)
         prepend a line to docs/agent-logs/INDEX.md
         report honestly, including what you did not do
```

**When in doubt about money: stop and ask. A blocked task costs a message. A double disbursement
costs trust.**
