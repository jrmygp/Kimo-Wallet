# Kimo Wallet UI/UX Review (UI/UX Pro Max) — 2026-10-03

**Scope:** `apps/web`, read-only review. No code was changed.

**Method:** Read every page under `app/`, the shared components under `components/`, the
transaction and wallet feature components, `app/globals.css` and `lib/fonts.ts`. Rendered
`/auth/login` in Chrome at a phone-width window and looked at it. Signed-in routes (`/home`,
`/wallet/qr`, `/wallet/history`, `/wallet/transfer/[userId]`) redirect to login without a
session, so those were reviewed **from source only and not seen rendered**.

**Companion document:** [2026-10-03-ui-critique.md](2026-10-03-ui-critique.md) is an earlier
review from the same day using a different method (anti-ui-slop). The two overlap in places.
This one goes further on the token system, typography and component patterns.

**Status:** These are recommendations only. None have been approved or put into code.

> **About the skill's design-system output:** the UI/UX Pro Max `--design-system` search for
> "fintech e-wallet" returned a dark gold/purple "Trust & Authority" landing-page theme. That
> doesn't fit a consumer e-wallet with an existing teal brand, so it was **not** used. Only the
> narrower suggestions were kept: the IBM Plex Sans / Plus Jakarta Sans type pairings and the
> teal + blue palette.

---

## 1. Current visual/design problems

### Contrast and accessibility (the most serious)

- **White text on `kimo-500` (`#14B8A6`) is about 2.5:1 contrast; normal text needs 4.5:1.**
  This affects every teal screen: login, register, QR, the home header, the transfer header, and
  labels such as "Country", "Number" and "My Balance". On the rendered login screen, the T&C line
  and "Lost or inactive number?" are hard to read.
- Error labels use raw `text-red-500` on teal (`app/auth/login/page.tsx:90`). That's red on
  green, which has low contrast and is hard to tell apart for colour-blind users.
- Login and register errors appear in an empty `FieldDescription` slot, so the layout jumps when
  an error shows. The Country field also resizes itself (`w-30` → `w-56`) when it's invalid.
- Sheet close buttons are bare 20px icons with no padding and no accessible name
  (`app/home/page.tsx:141`, `components/country-code-select.tsx:96`). They're well under the
  44px touch-target minimum.

### No real design tokens

- `app/globals.css` is still the default shadcn grayscale theme: `--primary` is near-black,
  `--ring` is grey, and the chart colours are grey. The teal ramp is a separate `kimo-*` scale
  that no semantic token points to.
- So every component styles itself:
  - `Button`'s default variant is hard-coded white with `text-[#2C2C2C]` and turns teal-600 on
    hover (`components/ui/button.tsx:11`).
  - The transfer page uses `bg-[#F5F5F5]`. Labels use `text-black`.
  - Incoming money uses Tailwind's green; outgoing money uses `destructive`.
- `.dark` is defined but never switched on, and it's grayscale anyway.

### Hierarchy and composition

- The balance, the most important number in the app, is `text-2xl` regular weight under a
  `text-xl` bold "My Balance" label. The label carries more visual weight than the money.
- Login and register are teal from edge to edge with a white button. There's no surface for the
  form to sit on; only the logo and the CTA compete for attention. This is the most "generic
  template" signal in the app.
- On the transfer screen, the card is positioned with `absolute top-30 w-[80%]`. The content
  below it doesn't move around it, so it will overlap or clip on short screens or when an error
  appears.
- `components/layout/Page.tsx` wraps every page in `max-w-3xl h-screen shadow-lg sm:rounded-2xl`.
  On desktop that's a 768px "phone" with a shadow and nothing behind it: too wide to read as a
  phone, too bare to read as a web app. It also uses `h-screen` rather than `dvh`, which breaks
  on mobile Safari when the toolbar is visible.
- Spacing has no consistent scale: `gap-8`, `gap-6`, `gap-4`, `gap-2`, with padding of `px-4`,
  `px-10` and `px-2 sm:px-0`.

### Things that read as placeholder or unfinished

- The page title is "Create Next App" (`app/layout.tsx:10`), so that's what every browser tab
  shows.
- The fallback avatar everywhere is `https://github.com/shadcn.png`, so every user and every
  transaction shows the same stranger's face.
- The home transactions are hard-coded, with four identical "Jane Doe" rows. Amounts appear in
  red/green `Badge` pills, which makes a ledger look like a list of status tags.
- The home menu has six equal tiles. Top Up, Inbox and Settings are buttons that do nothing, and
  Inbox and Settings get the same visual weight as Transfer.
- The promo carousel uses fixed `w-96` cards with `px-10` padding: 384px cards on a 360px screen,
  with no position indicator and no way to pause.
- Copy issues: "you are agree", alt text "kimo-logo", and "Request from QRIS" (it means "Receive
  via QRIS").

### Mixed visual language

- There are three icon sets: `react-icons/md` (Material, mostly filled), `lucide-react` (outline)
  and `country-flag-icons`. Filled and outline icons sit next to each other, for example in the
  sheet headers.
- `lib/fonts.ts` loads Poppins with all 9 weights in both normal and italic (18 font files), but
  `font-heading` is never used. Inter does all the work.

### Money UX

- `formatTransactionAmount` (`features/transaction/format.ts:16`) outputs `- Rp 50.000`, with a
  hyphen as the minus sign and a space after it. That isn't how Indonesian e-wallets show amounts.
- The balance shows `${currency} ${balance.toLocaleString("id")}` ("IDR 150.000"), but the hidden
  state shows "Rp ••••••". The visible and hidden states use different currency labels.
- Only the balance uses `tabular-nums`; the other amounts don't.

---

## 2. Proposed design direction

**"Calm, confident fintech":** white and neutral surfaces, with teal kept for brand moments and
primary actions only. This is the direction GoPay, Jago, Wise, Revolut and Cash App have
converged on. Brand colour fills small, deliberate areas and isn't used as a full-screen
background.

- **At most one teal hero surface per screen:** the balance card on Home, and the header of flow
  screens. Everything else sits on a near-white canvas (`#F6F8F8`) with white cards.
- **The balance is the hero.** 32–40px, semibold, tabular figures, with a smaller and lighter
  currency prefix and the eye toggle inline. The primary actions (Top Up, Transfer,
  Receive/QRIS, History) sit inside or directly under the balance card as a row of four, not a
  3×2 grid of six.
- **Flow screens** (Transfer, Top Up, Pay) follow the same pattern every time:
  1. App bar with a back button and title.
  2. Recipient or context card.
  3. Large amount input.
  4. Review sheet.
  5. Result screen.

  One primary CTA, pinned to the bottom above the safe area.
- **Auth screens:** a white screen with a small teal brand mark, a clear heading, the form, and a
  pinned CTA. Optionally keep a teal band at the top only.
- **Desktop:** keep the app as a 420–480px column centred on a soft neutral or subtly branded
  background, and drop the 768px card. Or commit fully to a two-column layout from 1024px up.
  Don't stay in between.

---

## 3. Typography and colour system

### Typography

| Role | Recommendation |
|---|---|
| Family | **Plus Jakarta Sans** as the single family: geometric and friendly, and designed in Jakarta, which suits an IDR wallet. It replaces both Poppins and Inter. Alternative: IBM Plex Sans if you want something more serious and bank-like. |
| Weights | Load 400, 500, 600 and 700 only. No italics. |
| Numbers | `font-variant-numeric: tabular-nums` on every amount, balance, time and ID, either through a `.num` utility or inside the `<Money>` component. |

A type scale on 4px steps:

| Token | Size / line height | Weight | Use |
|---|---|---|---|
| `display` | 36/44 | 700 | Balance, amount being entered |
| `title-lg` | 22/28 | 600 | Screen titles |
| `title` | 18/24 | 600 | Section headings, card titles |
| `body` | 16/24 | 400 | Default text; 16px also stops iOS zooming into inputs |
| `body-sm` | 14/20 | 400–500 | Secondary text, list subtitles |
| `label` | 13/16 | 500 | Field labels, menu labels |
| `caption` | 12/16 | 400 | Timestamps, legal text (never smaller) |

### Colour system (replaces the default shadcn grayscale tokens)

Keep the teal, but **move the brand primary down to teal-700 (`#0F766E`)** for text and for fills
behind white text. White on `#0F766E` is about 5.5:1, which passes. Keep teal-500 for decoration
only: illustrations, the balance card gradient, and focus rings on white.

| Semantic token | Light | Notes |
|---|---|---|
| `--primary` | `#0F766E` (teal-700) | Primary buttons, links, active nav |
| `--primary-hover` | `#115E59` (teal-800) | |
| `--primary-foreground` | `#FFFFFF` | |
| `--brand-surface` | gradient `#0F766E → #0D9488` | Balance card and flow headers only |
| `--background` | `#F6F8F8` | App canvas, slightly cool |
| `--card` | `#FFFFFF` | |
| `--foreground` | `#0F1F1E` | Near-black with a hint of teal |
| `--muted-foreground` | `#5B6B6A` | Passes 4.5:1 on white |
| `--border` | `#E3E9E8` | |
| `--ring` | `#14B8A6` | Teal-500 is fine as a ring on white |
| `--success` / `--money-in` | `#047857` on `#ECFDF5` | Green text for incoming amounts |
| `--danger` | `#DC2626` | **Errors and destructive actions only, never outgoing money** |
| `--warning` / `--pending` | `#B45309` on `#FFFBEB` | "We're confirming your transaction" (`docs/CLAUDE.md` §3.7.5) |
| `--info` (accent) | `#0369A1` | Links and promotional chips, sparingly |

**Key rule: outgoing money is not red.** Show it in the normal foreground colour with a `−` sign,
and show incoming money in green with `+`. Red means an error or a failed payment. If every
payment the user makes looks like an error, it's alarming, and it competes with real failure
states.

**Dark mode:** design one properly from these tokens (for example `#2DD4BF` as primary on
`#0B1514`), or remove the unused `.dark` block until you're ready.

### Radius, elevation, spacing

- **Radius:** `--radius: 14px`. Cards 16px, inputs and buttons 12px, sheets 24px at the top,
  chips fully rounded.
- **Elevation:** two levels only.
  - `e1` for cards: `0 1px 2px rgb(16 24 40 / .06)`.
  - `e2` for sheets and floating CTAs.

  Drop the ad-hoc `shadow-lg`.
- **Spacing:** an 8-point scale. 16px screen gutter, 24px between sections, 12px inside lists.

---

## 4. Recommended component patterns

| Component | Pattern |
|---|---|
| **`<Money amount currency direction? size>`** | One component for every amount. Takes integer minor units, formats with `Intl.NumberFormat("id-ID")` as `Rp 150.000`, applies `+` or `−` (U+2212), tabular figures, and a smaller currency prefix. Fixes the IDR/Rp mismatch and the scattered `toLocaleString` calls. It only formats; it does no arithmetic (`docs/CLAUDE.md` §3.3.4). |
| **`BalanceCard`** | Teal gradient; label in `body-sm` at 80% white; amount at `display` size; eye toggle with a 44px hit area; the four quick actions directly beneath. Shows a skeleton while loading, never "IDR undefined". |
| **`QuickAction`** | 48px icon tile plus a 13px label, four per row. Real actions only; Inbox and Settings move to the app bar or a profile screen. |
| **`AppBar`** | Sticky, 56px tall: a 44px back button with an accessible name, a centred title and an optional trailing action. Replaces the three back-button styles currently in use. |
| **`ListItem` / `TransactionRow`** | Leading icon (direction arrow or initials avatar), title and subtitle, trailing `<Money>` as text rather than a badge, a 64px row, grouped under date headers ("Today", "Yesterday", "12 Sep"). A status chip (Pending / Failed) only when the status isn't Completed. |
| **`Avatar`** | Initials fallback with a deterministic colour per user. Remove the `shadcn.png` fallback entirely. |
| **`AmountInput`** | Large, centred, `display`-size input with an `Rp` prefix and live thousands separators, a "Balance Rp X" hint underneath, and quick-pick chips (50k / 100k / 200k). The value stays a string in Formik state (`docs/CLAUDE.md` §5.3.7). |
| **`BottomSheet`** | Drag handle, title, labelled 44px close button, safe-area padding. Used for country select, recipient search and the review/confirm step. |
| **`StickyFooterCTA`** | Primary button pinned above the safe area, full width, 52px tall. Its loading state keeps the label width ("Sending…" with a spinner). |
| **`ResultScreen`** | Success, Pending and Failed variants: icon, amount, recipient, transaction ID, and Done/Share buttons. Pending carries the "we're confirming your transaction" copy. |
| **`EmptyState` / `Skeleton`** | History, search with no result, and the first-run home screen. |
| **`Field`** | Label always visible, persistent helper text, and the error in the same reserved slot (with a min-height) so nothing jumps. Error shown with an icon and red text; `aria-describedby` and `aria-invalid` wired up. |

**One icon set:** Lucide, since it's already installed and shadcn uses it. 1.75px stroke, 20px in
lists, 24px in the app bar. Then remove `react-icons`.

---

## 5. Making it feel like a production e-wallet, not AI-generated

1. **Stop flooding screens with brand colour.** A fully teal login screen with a white pill
   button is the clearest "template" signal. Production wallets are mostly white, with brand
   colour in two or three deliberate places.
2. **Make money typography excellent:** tabular figures, a real `−` sign, a de-emphasised currency
   prefix, and amounts aligned to a consistent right edge. This is the biggest perceived-quality
   lever in a wallet.
3. **Remove every placeholder:** "Create Next App", `shadcn.png`, the duplicate Jane Doe rows,
   the dead Top Up / Inbox / Settings buttons, and the copy typos. Add a favicon, a PWA manifest
   icon and a `theme-color`.
4. **Design the states, not just the happy path:** balance and history skeletons, empty history,
   recipient not found, the offline banner (PRD §10), Pending, Failed and session expired. Real
   apps spend most of their design effort here.
5. **Use local, specific language.** "KimoID", "Saldo", "QRIS" and "Transfer ke" make the app
   feel built for its market. Pick English or Bahasa and use it consistently, or add proper i18n.
6. **Use restrained, meaningful motion:**
   - Sheets rise in 250ms with deceleration and exit in about 180ms.
   - Pressed controls scale to 0.97.
   - The balance crossfades when toggled.
   - A short check animation on success.
   - Everything respects `prefers-reduced-motion`.

   Remove the blanket `transition-all duration-300`.
7. **Make it trustworthy at a glance.** Show the recipient's name, masked phone number and KimoID
   on review. Add a "Secured by PIN" note on confirm, and put the transaction ID on the receipt
   with a copy button.
8. **Lean into the PWA:** a `dvh` layout, `env(safe-area-inset-*)` padding, `theme-color` set to
   teal-700 for the status bar, and no fake phone frame on desktop.
9. **Add a bottom tab bar** (Home · History · Scan/QRIS · Profile), with a raised centre Scan
   button as in Indonesian wallets. It gives the app a home base, replaces the menu grid, and
   makes Settings findable.

---

## Out-of-scope issues noticed (rules in `docs/CLAUDE.md`, not design)

- **The Transfer form is a money form with `validationSchema: null`**, and its submit handler only
  does `console.log(values)` (`app/wallet/transfer/[userId]/page.tsx:22-25`). There's also a
  commented-out `FieldLabel`, and the note input has no label. It's a stub, so it isn't a live
  risk yet, but it breaks §5.3. Any redesign of that screen should be built to the money-form
  rules (Yup schema, string → integer minor units, idempotency key held outside Formik), not on
  top of the stub.
- `localStorage` is read during render on the login and register pages. That will throw during
  server rendering if those pages are ever pre-rendered.
