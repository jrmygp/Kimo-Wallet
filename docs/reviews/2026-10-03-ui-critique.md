# Kimo Wallet UI Critique — 2026-10-03

**Scope:** `apps/web`, read-only review. No code was changed.

**Method:** The login page (`/auth/login`) was rendered in Chrome at a narrow window and looked at.
Signed-in routes (`/home`, `/wallet/history`, `/wallet/qr`, `/wallet/transfer/[userId]`) send
you to login, so those were reviewed **from source only and not seen rendered**.

---

## Three biggest problems, by user impact

### 1. The home screen is a generic e-wallet layout, and the balance isn't the main thing on it

**What the code shows** (`app/home/page.tsx`)

- The layout is the stock wallet-app arrangement: a coloured header with avatar and name, a 3×2
  grid of round icon tiles, a carousel of stock promo images, then a list.
- The balance is just a second line of white text in the header. It's no bigger than the
  "My Balance" label above it, and it sits next to the user's name.
- Three of the six tiles do nothing when tapped: **Top Up**, **Inbox** and **Settings** are plain
  `<button>`s with no `onClick`.
- Inbox and Settings aren't money actions, so they don't belong next to Transfer and QRIS.
- The carousel images are labelled "Promotion 1/2/3". On a phone they push the user's own
  transactions further down the screen.

**What should change**

- Make the balance the biggest, most prominent text on the screen.
- Put only the three money actions (Top Up, Transfer, QRIS) directly under the balance.
- Move Inbox and Settings to the header as icons.
- Remove or clearly disable tiles that do nothing.
- Move the promos below Latest Transactions, or drop them until real campaigns exist.

### 2. The app still runs on unchanged shadcn and Next.js defaults, so it doesn't look like Kimo

**What the code shows**

- **Colours:** `app/globals.css` keeps the default grey shadcn colour set
  (`--primary: oklch(0.205 0 0)`, `--ring` grey). The `kimo-*` teal colours only appear as
  hard-coded classes on a few elements.
  - Result: the login page has a near-black **Continue** button on teal, while other screens
    use white buttons.
- **Placeholder avatars:** Users without a photo get `https://github.com/shadcn.png`.
  - Every transaction row always shows this avatar (`features/transaction/components/transaction-row.tsx`).
    So the user, every counterparty and every search result have the same face.
  - In a money app, that makes it harder to tell who you're paying.
- **Leftover scaffolding:**
  - The tab title is still "Create Next App" (`app/layout.tsx`).
  - Poppins loads all 9 weights plus italics (`lib/fonts.ts`), but nothing on the pages
    reviewed uses it.
  - `components/layout/Page.tsx` puts a `shadow-lg` / `sm:rounded-2xl` frame around a
    full-height page, so the frame never actually shows.

**What should change**

- Map `--primary`, `--ring` and `--accent` to the kimo teal, so every `Button` and focus ring
  picks it up automatically.
- Use initials (`AvatarFallback`, which already exists) when there's no photo, instead of the
  shadcn placeholder.
- Set real `metadata`.
- Either use Poppins for headings and the balance, or remove it.

### 3. Money is shown as coloured pills, and the send flow looks finished but isn't

**What the code shows**

- **Transaction rows:** `TransactionRow` puts every amount in a `Badge` (a small pill).
  - Outgoing amounts use the `destructive` style, so every normal payment looks like an error.
  - Incoming amounts get a separately hard-coded green.
  - Amounts aren't right-aligned in a fixed-width column.
  - Rows have no "sent / received" label or note.
- **Transfer page** (`app/wallet/transfer/[userId]/page.tsx`):
  - The recipient card overlaps the header using `absolute top-30 w-[80%]`. That won't adapt
    to longer names or error messages.
  - The amount field's only hint is the placeholder `"Rp"`.
  - The note field has no label; the label line is commented out.
  - **Send Now** is enabled with `validationSchema: null`, and submitting only runs
    `console.log`.
  - The button isn't disabled while a request is in flight.
  - There's no idempotency key (a one-time ID that stops a retried payment from being sent
    twice).

**What should change**

- Show amounts as plain right-aligned text with `tabular-nums`. Use a +/− sign and colour
  only to show direction.
- Keep the destructive red for real failures.
- Put the recipient card in normal page flow.
- Show "Rp" as a fixed prefix inside the amount field, show available balance next to it, and
  give the note field a visible label.
- Before Transfer is connected to a real endpoint, it needs all of the following. These are
  required by `docs/CLAUDE.md` §3.7 and §5.3, and they're correctness problems, not just
  styling:
  - a review step;
  - a Yup schema (string amount → integer minor units);
  - an idempotency key created at the review step and stored outside Formik state;
  - a button that stays disabled while the request is in flight.

---

## Smaller signs it was copied from a template

- **Header colour:** Every screen uses the same `bg-kimo-500` header, so no screen has its own
  character.
- **Login page** (seen rendered):
  - The fields sit near the top and **Continue** sits at the bottom, leaving a large empty
    teal gap between them.
  - **Change number** is a big white pill that competes with **Continue**.
  - The terms line has a typo ("you are agree").
  - "T&C" and "Privacy Notice" aren't links.
- **Icons and back buttons:**
  - The app mixes two icon libraries: `react-icons/md` (Material Design) and `lucide-react`.
  - Back buttons differ on every page:
    - icon only on History;
    - icon plus "Back" on Transfer;
    - label hidden on mobile on QR.
- **Hover effects that do nothing on phones:** The menu tiles use
  `hover:bg-white hover:shadow-sm duration-300`. A phone-first app has no hover, so this adds
  nothing.

---

## Not verified

- Home, History, QR and Transfer were not seen rendered. The visual claims about them come from
  reading the code.
- The Next.js dev badge overlapping the login page's bottom-left corner only appears in
  development, so it isn't a real issue.
