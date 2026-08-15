# Kimo Wallet — Product Requirements Document

## 1. Overview

**Kimo Wallet** is a mobile-first digital wallet built as an installable Progressive Web App (PWA). The project is designed as a production-style portfolio application that demonstrates modern frontend engineering with Next.js and TypeScript and backend engineering with Go microservices, gRPC, Protocol Buffers, event-driven architecture, and production-oriented reliability patterns.

Kimo Wallet will simulate core wallet experiences such as managing a balance, topping up, transferring money, paying merchants through QR codes, and viewing transaction activity.

> **Tagline:** Your money, in motion.

## 2. Goals

### Product goals

- Provide a polished, mobile-first wallet experience.
- Make common money actions fast and simple.
- Provide clear transaction status and history.
- Support installable PWA behavior.
- Provide safe read-only offline access to cached information.

### Engineering goals

- Showcase Next.js + TypeScript expertise.
- Showcase independently deployable Go microservices.
- Use gRPC + Protocol Buffers for internal service communication.
- Use event-driven processing for asynchronous workflows.
- Demonstrate financial transaction consistency and idempotency.
- Demonstrate observability, testing, containerization, and CI/CD.
- Keep the architecture understandable and maintainable for a solo project.

## 3. Non-Goals

The first version will **not**:

- Process real money.
- Integrate with real banks or payment providers.
- Implement actual QRIS settlement.
- Store real payment-card information.
- Support real-world financial compliance/KYC.
- Allow financial mutations while offline.
- Start with a large number of microservices or Kubernetes.

All payments and balances are simulated.

## 4. Target User

The primary user is a consumer who wants to:

- Store and view a digital wallet balance.
- Send money to another Kimo user.
- Top up their wallet.
- Pay a simulated merchant.
- Scan a payment QR code.
- Review previous transactions.
- Receive payment notifications.

## 5. Platform

### Primary

- Mobile web / PWA
- Responsive layout optimized for approximately 360–430px mobile widths
- Desktop support is secondary

### PWA capabilities

- Installable on supported devices
- App-like navigation
- Service worker
- Cached read-only data
- Web push notifications
- Offline state handling

## 6. Core Navigation

The application should use a simple mobile navigation model:

- **Home**
- **Activity**
- **Scan**
- **Profile**

Quick actions on Home:

- Top Up
- Send
- Pay
- Scan QR

## 7. Functional Requirements

### 7.1 Authentication

Users must be able to:

- Register an account.
- Log in.
- Log out.
- Refresh an authenticated session.
- View their profile.
- Manage their wallet PIN.
- View active/recent devices or sessions.

Security requirements:

- Passwords must never be stored in plaintext.
- PINs must be securely hashed.
- Authentication tokens must have expiration and refresh behavior.
- Repeated failed PIN attempts should trigger temporary protection.

### 7.2 Home

The Home screen should display:

- Available wallet balance.
- Hide/show balance control.
- Wallet identifier.
- Quick actions.
- Recent transactions.
- Connection/offline status where relevant.

Example:

```text
Good morning 👋

Rp 4.250.000
Available balance

[ Top Up ] [ Send ]

[ Pay ] [ Scan ]

Recent Transactions
Coffee Shop       -Rp25.000
John Doe         -Rp100.000
Top Up           +Rp500.000
```

### 7.3 Wallet

Users must be able to:

- View wallet balance.
- View wallet details.
- View wallet status.
- View supported currency.
- View balance changes through transaction history.

The system must protect the balance from concurrent update errors.

### 7.4 Top Up

Users can initiate a simulated top-up using:

- Bank transfer
- Virtual account
- Debit/card simulation

Top-up lifecycle:

```text
CREATED
   ↓
PENDING
   ↓
PROCESSING
   ↓
SUCCESS

or

FAILED
```

The system must be able to safely retry asynchronous processing without creating duplicate balance credits.

### 7.5 Transfer

Users can send money to another Kimo user.

Flow:

```text
Select recipient
       ↓
Enter amount
       ↓
Add optional note
       ↓
Review
       ↓
Enter PIN
       ↓
Create transaction
       ↓
Process
       ↓
Success / Failed
```

Requirements:

- Validate recipient.
- Validate amount.
- Validate available balance.
- Require transaction confirmation.
- Protect against duplicate requests.
- Record transaction state.
- Record an auditable ledger entry.

### 7.6 Merchant Payment

Users can pay simulated merchants.

Payment flow:

```text
Merchant
   ↓
Amount
   ↓
Review
   ↓
PIN
   ↓
Processing
   ↓
Success
```

### 7.7 QR Payment

Users can scan a Kimo payment QR code.

The first version can use a Kimo-specific QR format, for example:

```text
kimo://pay?merchant=coffee-shop&amount=25000
```

Requirements:

- Open camera/scanner.
- Parse QR payload.
- Validate payload.
- Display merchant and amount.
- Require user confirmation.
- Process payment through the transaction system.

### 7.8 Activity

Users can:

- View transaction history.
- Filter by transaction type.
- Filter by status.
- View transaction details.
- Search transactions.
- View transaction ID and timestamps.

Suggested transaction types:

- TOP_UP
- TRANSFER_IN
- TRANSFER_OUT
- MERCHANT_PAYMENT
- WITHDRAWAL
- REFUND

### 7.9 Notifications

Users should receive notifications for important events:

- Money received.
- Transfer completed.
- Payment completed.
- Top-up completed.
- Transaction failed.
- New login/device.

Notifications should be generated asynchronously from domain events.

### 7.10 Profile

Users can manage:

- Name
- Email
- Phone number
- PIN
- Security settings
- Notification preferences
- Sessions/devices
- Logout

## 8. Transaction and Ledger Requirements

The financial core must not rely only on directly mutating a wallet balance.

The system should maintain an auditable ledger.

Example:

```text
TOP_UP          +Rp1.000.000
TRANSFER_OUT    -Rp100.000
PAYMENT         -Rp25.000
---------------------------
Balance          Rp875.000
```

The implementation should support a consistent balance model and immutable transaction/ledger history.

### Idempotency

Financial mutation APIs must accept an idempotency key.

Example:

```http
POST /transactions/transfer
Idempotency-Key: 7f4a8...
```

Repeated requests with the same key must not create duplicate financial effects.

### Concurrency

The system must prevent scenarios where concurrent requests spend the same available balance.

Example:

```text
Balance = Rp100.000

Request A → Pay Rp80.000
Request B → Pay Rp80.000

Only one transaction may succeed.
```

The backend should use appropriate database transaction and locking/consistency mechanisms.

## 9. Transaction States

A common transaction lifecycle should be:

```text
CREATED
  ↓
PENDING
  ↓
PROCESSING
  ↓
COMPLETED
```

Failure path:

```text
PROCESSING
  ↓
FAILED
```

Where compensation is required:

```text
PROCESSING
  ↓
FAILED
  ↓
ROLLBACK / COMPENSATION
```

## 10. Offline Behavior

The PWA may cache:

- User profile.
- Last known balance.
- Recent transactions.

Offline users must **not** be able to perform:

- Transfers.
- Payments.
- Top-ups.
- Withdrawals.

The UI should clearly communicate:

> You're offline. You can view cached information, but payments and transfers require an internet connection.

## 11. Non-Functional Requirements

### Performance

- Fast initial mobile load.
- Minimize JavaScript sent to the browser.
- Use appropriate Server/Client Component boundaries.
- Optimize images and assets.
- Avoid unnecessary client-side fetching.

### Reliability

- Idempotent financial mutations.
- Safe retries.
- Explicit transaction states.
- No duplicate financial effects.
- Graceful service failure handling.

### Security

- Secure password/PIN storage.
- Authentication and authorization.
- Rate limiting.
- Input validation.
- Secure headers.
- Audit logging.
- No sensitive secrets committed to source control.

### Accessibility

- Keyboard accessibility where applicable.
- Semantic HTML.
- Screen-reader-friendly controls.
- Sufficient touch target sizes.
- Clear error and status states.

## 12. MVP Scope

The first milestone should include:

- Registration/login.
- PIN setup.
- Wallet.
- Balance.
- Top-up simulation.
- User-to-user transfer.
- Transaction history.
- Transaction details.
- Basic notifications.
- Mobile-first PWA.
- Go microservices.
- gRPC + Protobuf.
- PostgreSQL.
- Docker Compose.

## 13. Post-MVP

Potential future features:

- Merchant accounts.
- Recurring payments.
- Bill payments.
- Refunds.
- Scheduled transfers.
- Multiple currencies.
- Spending limits.
- Virtual cards.
- Passkeys/WebAuthn.
- Web push.
- Advanced fraud/risk simulation.
- Multi-device session management.

## 14. Success Criteria

Kimo Wallet is successful as a portfolio project when:

1. A user can complete the main wallet flows entirely from a mobile UI.
2. Financial mutations are protected against duplicate requests.
3. Concurrent balance updates are handled safely.
4. Services communicate internally using gRPC/Protobuf.
5. Asynchronous workflows use events.
6. Transaction history is auditable.
7. The application can run locally through Docker Compose.
8. Tests cover important business logic.
9. Logs and metrics make failures diagnosable.
10. The repository clearly documents the architecture and engineering decisions.

## 15. Portfolio Positioning

Kimo Wallet should be presented as a production-style distributed fintech system rather than simply a wallet UI.

Suggested portfolio description:

> Kimo Wallet is a mobile-first e-wallet PWA built with Next.js and TypeScript, backed by Go microservices using gRPC, Protocol Buffers, PostgreSQL, and event-driven processing. The platform focuses on reliable financial transaction processing, idempotency, concurrency control, ledger-based accounting, and production observability.
