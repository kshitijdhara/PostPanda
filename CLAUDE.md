# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Full-stack development (kills ports, builds all, then watches)
yarn dev

# Dev without worker server
yarn dev:no-workers

# Build all packages (Turborepo)
yarn build

# Lint, format, check (Biome — not ESLint/Prettier)
yarn lint           # lint only
yarn lint:fix       # lint with auto-fix
yarn format         # format with auto-fix
yarn check          # lint + format check
yarn check:fix      # lint + format with auto-fix

# Run tests (from root, targeting a specific workspace)
yarn workspace @saga/app-web run test        # watch mode
yarn workspace @saga/app-web run test:run    # single run

# Clean all build artifacts
yarn clean

# Database (run from root, targeting the database-node workspace)
yarn workspace @saga/database-node run db:migrate         # apply pending migrations
yarn workspace @saga/database-node run db:migrate:create  # create migration without applying
yarn workspace @saga/database-node run db:generate        # regenerate Prisma client
yarn workspace @saga/database-node run db:studio          # open Prisma Studio
```

## Architecture

### Monorepo Structure

**Yarn 4 workspaces + Turborepo**. Three apps depend on shared packages:

```
apps/
  app-web/        # React 19 + Vite frontend (@saga/app-web)
  app-server/     # Express 5 backend (@saga/app-server)
  worker-server/  # Background job processor (@saga/worker-server)

packages/
  node/           # Backend-only packages (database, domain logic)
  web/            # Frontend-only packages (axios client, config, global UI)
  middleware/     # Shared TypeScript types used by both frontend and backend
```

### Package Naming Convention

- `@saga/*-node` — backend-only packages
- `@saga/*-web` — frontend-only packages
- `@saga/*-middleware` — shared types/constants (consumed by both `*-node` and `*-web`)

### Key Packages

- **`@saga/database-node`** — Prisma (PostgreSQL) + Redis client. Schema at `packages/node/database-node/prisma/schema.prisma`. All DB access goes through this package.
- **`@saga/config-node`** / **`@saga/config-web`** — Environment config + Flagsmith (OpenFeature) feature flag client.
- **`@saga/global-web`** — Shared UI primitives (ThemeProvider, ErrorBoundary, AvatarProvider, Toast).
- **`@saga/logger-middleware`** — Structured logger shared across frontend and backend.
- **`@saga/*-middleware`** — Typed API contracts and shared constants for each domain.

### Frontend (`app-web`)

- **React 19**, **React Router v6**, **Vite**, **Vitest**
- Domain-driven: `src/domains/{auth,posts,communities,events,payments,notifications,profile,follow,admin,...}`
- Feature flags via OpenFeature + Flagsmith — check `useBooleanFlagValue('flag_name', false)` before rendering gated features
- Path alias `@` maps to `src/` (configured in Vite + tsconfig)

### Backend (`app-server`)

- **Express 5** + **tsx** (for TypeScript execution in dev)
- Domain logic lives in `packages/node/*-node`; the server wires them together in `src/app.ts`
- Connects to PostgreSQL (Prisma) and Redis on startup; graceful shutdown handles disconnection

### Worker Server (`worker-server`)

- Separate process for background jobs (payments, notifications, scheduled tasks)
- Uses `@saga/scheduler-node` and `@saga/worker-node` packages
- In non-local environments, embedded workers run inside `app-server`; locally they run as a separate process

### Database Schema Key Concepts

- **Records** — generic content model for posts and comments (`recordType: post | comment`), hierarchical via `parentId`/`rootId`
- **Communities** — groups with membership roles and visibility settings
- **Versions** — `UserVersion`, `CommunityVersion`, `RecordVersion` tables track history on every entity change
- **Payments** — Stripe Connect: `ConnectedAccount` → `PostPaymentConfig` → `PaymentTransaction` / `StripeTransfer`
- **Notifications** — delivered via SSE with Redis-backed batching; delivery tracked in `NotificationDeliveries`
- **Events** — first-class entities with RSVPs, community associations, and hierarchical child events

### Linting & Formatting

Uses **Biome** (not ESLint or Prettier). Config at `biome.json`. Single quotes, 2-space indent, 100-char line width.

---

## Workflow Orchestration

### 1. Plan Node Default
- Enter plan mode for ANY non-trivial task (3+ steps or architectural decisions)
- If something goes sideways, STOP and re-plan immediately — don't keep pushing
- Use plan mode for verification steps, not just building
- Write detailed specs upfront to reduce ambiguity

### 2. Subagent Strategy
- Use subagents liberally to keep main context window clean
- Offload research, exploration, and parallel analysis to subagents
- For complex problems, throw more compute at it via subagents
- One task per subagent for focused execution

### 3. Self-Improvement Loop
After ANY correction from the user: update `tasks/lessons.md` with the pattern
- Write rules for yourself that prevent the same mistake
- Ruthlessly iterate on these lessons until mistake rate drops
- Review lessons at session start for relevant context

### 4. Verification Before Done
- Never mark a task complete without proving it works
- Diff behavior between main and your changes when relevant
- Ask yourself: "Would a staff engineer approve this?"
- Run tests, check logs, demonstrate correctness

### 5. Demand Elegance (Balanced)
- For non-trivial changes: pause and ask "is there a more elegant way?"
- If a fix feels hacky: "Knowing everything I know now, implement the elegant solution"
- Skip this for simple, obvious fixes — don't over-engineer
- Challenge your own work before presenting it

### 6. Autonomous Bug Fixing
- When given a bug report: just fix it. Don't ask for hand-holding
- Point at logs, errors, failing tests — then resolve them
- Zero context switching required from the user
- Go fix failing CI tests without being told how

## Task Management
1. **Plan First**: Write plan to `tasks/todo.md` with checkable items
2. **Verify Plan**: Check in before starting implementation
3. **Track Progress**: Mark items complete as you go
4. **Explain Changes**: High-level summary at each step
5. **Document Results**: Add review section to `tasks/todo.md`
6. **Capture Lessons**: Update `tasks/lessons.md` after corrections

## Core Principles
- **Simplicity First**: Make every change as simple as possible. Impact minimal code.
- **No Laziness**: Find root causes. No temporary fixes. Senior developer standards.
- **Minimal Impact**: Changes should only touch what's necessary. Avoid introducing bugs.

---

# Crowd Commissions: Product Strategy & Implementation Plan

## Context

Currently, payments on Saga exist only as a Post feature (`PostPaymentConfig` tied 1:1 to a Post via unique `postId`). This limits the platform to simple donations and deadline-based funding. The goal is to elevate **Crowd Commissions** into a standalone product pillar — a community-driven commissioning system where fans propose, fund, and follow creative work from idea to delivery. This transforms Saga from a social platform with tipping into **the platform where creative communities fund and direct the work they want to see exist**.

---

## 1. Product Vision

### Core Problem
Creative work is collaborative, but existing platforms force isolated funding models. Patreon = subscriptions to vague value. Kickstarter = one-shot campaigns on an island. Ko-fi = tip jar. Fiverr = transactional freelancing. **None answer the question fans actually want to ask: "I have an idea — will you make it for me?"**

Crowd Commissions flips the model: fans pitch to creators, communities collectively fund specific work, creators get paid with accountability, and everyone follows the journey.

### Primary Users
- **Creators** (supply): Illustrators, musicians, animators, writers, game modders — anyone who receives creative requests. JTBD: "Help me understand what my audience wants, give me financial security before I invest 40 hours, let me build deeper fan relationships."
- **Backers/Fans** (demand): Engaged community members who want specific creative output. JTBD: "I want to request something specific, feel ownership over projects I fund, and watch the creative process unfold."

### Why Standalone Pillar
The current `PostPaymentConfig` structurally prevents: multi-milestone disbursement, backer tiers, community governance/voting, commission-specific discovery, lifecycle management beyond active/failed, and creator commission dashboards. **A post is a leaf node; a commission is a long-lived process with its own entity graph.**

### Long-Term Vision
- **Communities** define the audience
- **Commissions** define what the audience wants built (economic engine)
- **Posts** document the journey (updates, WIPs, deliverables)
- **Events** mark milestones (reveals, votes, celebrations)

### Positioning Statement (Geoffrey Moore)
> **For** creative communities and their creators **who** want to collaboratively fund and direct specific creative projects, **Saga Crowd Commissions is** a community-driven commissioning platform **that** lets fans propose, fund, and follow creative work from idea to delivery. **Unlike** Patreon, Kickstarter, or Fiverr, **our product** embeds funding directly into the community that cares about the work, with milestone-based accountability, community voting, and progress transparency built in.

---

## 2. Product Architecture

### Core Domain Objects

| Object | Purpose |
|---|---|
| **Commission** | First-class entity with lifecycle, funding goal, creator, community link |
| **CommissionMilestone** | Ordered delivery checkpoints with per-milestone disbursement |
| **CommissionTier** | Pledge levels with rewards and optional backer caps |
| **CommissionBacker** | Named backer record (replaces anonymous transactions) |
| **CommissionUpdate** | Bridge linking Posts to a Commission (update types: progress, milestone, delivery) |
| **CommissionVote** | Backer governance: milestone approval, revision requests, direction polls |

### Lifecycle
```
DRAFT → PROPOSAL → FUNDING → ACTIVE → DELIVERING → COMPLETED → ARCHIVED

Failure paths: FUNDING → FAILED (goal not met, auto-refund)
               ACTIVE → CANCELLED (creator cancels, partial refund)
               Any → DISPUTED (manual resolution)
```

### Platform Integration (One Ecosystem)
- **Posts**: Commission updates ARE posts (reuses entire Record/media pipeline). Commission page aggregates linked posts chronologically.
- **Communities**: Commissions can be community-associated. Moderators can feature commissions. Community members get prioritized notifications.
- **Events**: Milestone deliveries can trigger events (reveal livestreams, community review sessions).
- **Creator Profiles**: Commission history, completion rate, total earned via commissions.
- **Discovery**: Commission-specific feed with browse/filter/search, plus commission updates in main feed with context badges.
- **Notifications**: New event types — `commission_funded`, `commission_milestone_submitted`, `commission_milestone_approved`, `commission_completed`, `commission_update`, `commission_vote_requested`.
- **Payments**: Same `ConnectedAccount` infrastructure. Checkout extends for commission pledges. Disbursement extends for milestone-based partial payouts.

**Key decision**: Posts keep donation/funding (simple tips). Commissions are the elevated, structured version. A creator can "upgrade" a funding post to a full commission.

---

## 3. UX and Emotional Design

### Engagement Principles (inspired by Duolingo DAU 14.2M→34M growth)

| Principle | Application |
|---|---|
| **Funding Momentum** | Animated progress bars, celebration at 25/50/75/100%, "X backers away from next tier" |
| **Anticipation Loops** | Countdown timers with urgency styling, digest notifications ("80% funded, 3 days left") |
| **Progress Transparency** | Visual timeline with milestone nodes, WIP "dots" between milestones, "Your Impact" backer view |
| **Social Proof** | Backer avatars on cards, "Backed by X from [Community]" badges, creator completion rate |
| **Community Participation** | Comments on milestone submissions, direction polls, milestone approval voting |
| **Creator-Fan Interaction** | Q&A before funding, WIP posts, side-by-side "promised vs. delivered", special reveal formatting |

### Habit-Forming Loops
1. **Discovery Loop**: Browse → Find project → Back it → Check updates → Discover more from same creator/community
2. **Funding Momentum**: Nears goal → Platform notifies community → Social pressure → Funded! → Creator starts
3. **Milestone Anticipation**: WIP posted → Backers discuss → Submitted → Review/vote → Approved! → Payout → Next milestone
4. **Community Participation**: See commission → Join community to back → Vote/comment → See completed work → Propose next

---

## 4. Monetization Strategy

### Commission Revenue
| Stream | Structure | Notes |
|---|---|---|
| Platform fee | 5% on pledge (at milestone disbursement) | Matches existing `platformFeeBps: 500`. Only earned when work delivered. |
| Milestone escrow fee | 1-2% on multi-milestone commissions | Waived for single-milestone. Covers escrow management. |
| Premium creator tools | $9.99/month subscription | Priority placement, analytics, custom branding, unlimited milestones |
| Promoted commissions | CPM-based boost in discovery | Pay-to-promote like promoted posts |

### Platform-Wide (holistic)
- Donations & funding on posts (existing 5%)
- Community premium tools ($4.99/month)
- Event features (future)

**Principle**: Fees on successful outcomes only. Backers never pay extra. No paywall on core features.

---

## 5. Network Effects & Flywheel

```
Creators join → Create commissions → Communities form around projects
     ↑                                              ↓
More creators see    ←    Successful commissions attract backers
successful payouts         who join communities and follow creators
     ↑                                              ↓
Platform grows       ←    Backers propose new commissions to creators
```

**Cold start**: Seed with existing creators who have active funding posts. Auto-suggest commission upgrades for successful funding posts.

**Incentives**:
- Creators: Completion rate badges, leaderboard, early feature access
- Backers: Patron badges (5+ backed), early deliverable access, ability to propose commissions

---

## 6. Go-To-Market Strategy

### Phase 1: Commissions as Enhanced Feature
- **Positioning**: "Fund the art you want to see exist"
- **Acquisition**: Personal outreach to top 20 creators with active funding posts. 0% fee for first 3 commissions.
- **Targets**: 10 commissions/month, 5 funded, 50+ unique backers, $200-500 avg size

### Phase 2: Full Product Launch
- **Creator verticals** (priority order):
  1. Digital illustrators/artists (commission culture already exists)
  2. Musicians/producers (demo → mix → master → release milestones)
  3. Indie animators (high cost, long timelines, perfect for milestones)
  4. Writers (chapter-by-chapter delivery)
  5. Game modders (community-funded mods with feature voting)
- **Launch**: "Creator Spotlight" series documenting a commission journey publicly

### Market Sizing
- **TAM**: Online creative commissions ~$15B
- **SAM**: English-speaking digital creators via platforms ~$3B
- **SOM**: Year 1 — 1,000 active commission creators, $500K GMV, $25K platform revenue

---

## 7. Product Roadmap

### Phase 1: MVP (Weeks 1-6)
- Commission model decoupled from Post (new first-class entity)
- Single-milestone commissions (funded → delivering → completed)
- Funding with deadline/goal (reuses existing escrow/disbursement patterns)
- Commission detail page, creator management page
- Backer list, commission update posts
- New notification types

### Phase 2: Core Product (Weeks 7-14)
- Multi-milestone with per-milestone disbursement
- Backer tiers with rewards
- Milestone submission + backer approval flow
- Commission discovery feed
- Creator commission dashboard
- Revision request flow

### Phase 3: Platform Integration (Weeks 15-22)
- Community-featured commissions (moderator curation)
- Backer-proposed commissions ("I want you to create X")
- Direction voting on in-progress commissions
- Commission-linked events
- "Upgrade to commission" for funding posts
- Premium creator tools

### Phase 4: Marketplace (Weeks 23-30)
- Full marketplace with categories, trending, search
- Creator storefront pages
- Promoted commissions
- Cross-community discovery
- Recommendation engine

---

## 8. Metrics & Success Criteria

### North Star: Monthly Gross Commission Volume (GCV)

| Metric | Phase 1 | Phase 2 | Phase 4 |
|---|---|---|---|
| Commissions created/month | 10 | 50 | 500 |
| Funding success rate | 40% | 55% | 65% |
| Monthly GCV | $5K | $50K | $500K |
| Platform revenue/month | $250 | $2.5K | $25K |
| Unique backers/month | 50 | 500 | 5,000 |
| Commission completion rate | 60% | 75% | 85% |
| Backer retention (MoM) | 30% | 45% | 60% |

### Funnel
Page views → Pledge initiated → Checkout completed → Milestones delivered → Creator creates next → Backer backs another

---

## 9. Technical System Design

### New Database Models (Prisma)

```prisma
enum CommissionStatus {
  draft
  proposal
  funding
  active
  delivering
  completed
  archived
  cancelled
  failed
  disputed
}

enum MilestoneStatus {
  pending
  in_progress
  submitted
  revision_requested
  approved
  paid
}

model Commission {
  id                 String   @id @default(cuid())
  title              String
  description        String   @db.Text
  creatorId          String
  communityId        String?
  status             CommissionStatus @default(draft)
  fundingGoalCents   Int
  currentFundingCents Int     @default(0)
  fundingDeadlineAt  DateTime
  backerCount        Int      @default(0)
  platformFeeBps     Int      @default(500)
  connectedAccountId String
  category           String?
  tags               Json?    @db.JsonB
  visibility         Int      @default(0) // 0=public, 1=community, 2=invite
  pitchPostId        String?  // optional link to pitch post
  version            Int      @default(1)
  createdAt          DateTime @default(now())
  updatedAt          DateTime @updatedAt
  // relations: creator, community, connectedAccount, milestones, tiers, backers, updates, votes
}

model CommissionMilestone {
  id                   String   @id @default(cuid())
  commissionId         String
  title                String
  description          String?  @db.Text
  orderIndex           Int
  fundingPercentageBps Int      // e.g. 3333 = 33.33% of total
  status               MilestoneStatus @default(pending)
  deliverableUrl       String?
  dueDate              DateTime?
  submittedAt          DateTime?
  approvedAt           DateTime?
  paidAt               DateTime?
  createdAt            DateTime @default(now())
  updatedAt            DateTime @updatedAt
}

model CommissionTier {
  id                String @id @default(cuid())
  commissionId      String
  title             String
  description       String? @db.Text
  minAmountCents    Int
  maxBackers        Int?
  rewards           Json?   @db.JsonB
  currentBackerCount Int    @default(0)
  orderIndex        Int
  createdAt         DateTime @default(now())
  updatedAt         DateTime @updatedAt
}

model CommissionBacker {
  id              String  @id @default(cuid())
  commissionId    String
  userId          String
  tierId          String?
  amountCents     Int
  status          String  // pledged, confirmed, refunded, refund_failed
  stripeSessionId String?
  stripePaymentIntentId String?
  stripeChargeId  String?
  pledgedAt       DateTime?
  confirmedAt     DateTime?
  refundedAt      DateTime?
  createdAt       DateTime @default(now())
  updatedAt       DateTime @updatedAt
  @@unique([commissionId, userId])
}

model CommissionUpdate {
  commissionId String
  postId       String
  updateType   String  // progress, milestone_submission, final_delivery, general
  milestoneId  String?
  createdAt    DateTime @default(now())
  @@id([commissionId, postId])
}

model CommissionVote {
  id           String  @id @default(cuid())
  commissionId String
  milestoneId  String?
  userId       String
  voteType     String  // approve, request_revision, poll_option
  createdAt    DateTime @default(now())
  @@unique([commissionId, milestoneId, userId])
}
```

### Payment Flow Changes

**Commission Pledge** (extends `CheckoutService` pattern):
- `POST /api/commissions/:id/pledge` → validates status=funding, deadline not passed
- Creates `CommissionBacker` (status: pledged) → Stripe Checkout session (escrow, no transfer_data)
- Webhook `checkout.session.completed` → marks backer `confirmed`, increments `currentFundingCents`

**Milestone Disbursement** (extends `DisbursementService` pattern):
- `MilestoneDisbursementWorker` (5 min interval) checks milestones with status=approved
- Calculates: `commission.currentFundingCents * milestone.fundingPercentageBps / 10000`
- Deducts platform fee → creates `StripeTransfer` → marks milestone `paid`
- Same Redis lock pattern as existing `DisbursementWorker`

**Funding Deadline** (extends `DisbursementWorker` pattern):
- `CommissionFundingWorker` (5 min interval) checks commissions past deadline
- Goal met → status=active, notify creator
- Goal not met → status=failed, refund all backers (reuses batch refund pattern)

### New Package Structure
```
packages/node/commissions-node/          # Backend: service, store, routes, workers
packages/middleware/commissions-middleware/  # Shared types
apps/app-web/src/domains/commissions/    # Frontend domain
  api/commissionsApi.ts
  ui/CommissionCard/, CommissionDetail/, CommissionCreate/,
     MilestoneTimeline/, BackerList/, PledgeButton/, CommissionFeed/
```

### API Routes (`/api/commissions`)
```
POST   /                              Create commission (draft)
GET    /                              Discovery feed (browse/filter/search)
GET    /:id                           Commission detail
PATCH  /:id                           Update commission
POST   /:id/publish                   Draft → proposal/funding
POST   /:id/pledge                    Create pledge checkout
GET    /:id/backers                   List backers
POST   /:id/milestones/:mid/submit    Submit milestone deliverable
POST   /:id/milestones/:mid/vote      Vote on milestone
GET    /:id/updates                   Get linked update posts
POST   /:id/updates                   Link post as update
GET    /me/created                    Creator's commissions
GET    /me/backed                     Backed commissions
```

### Fraud Prevention
- Rate limiting on pledges (existing `rateLimiter.ts` pattern)
- Creator identity via Stripe Connect (already handled)
- Minimum backer participation for milestone approval
- Dispute flow with admin review
- Auto-refund if no first milestone within 30 days of funding

### Critical Files to Modify/Reference
- `packages/node/database-node/prisma/schema.prisma` — extend with new models
- `packages/node/payments-node/src/services/DisbursementService.ts` — pattern for milestone disbursement
- `packages/node/payments-node/src/services/CheckoutService.ts` — pattern for commission pledge checkout
- `packages/node/payments-node/src/workers/DisbursementWorker.ts` — template for new workers
- `packages/middleware/payments-middleware/src/types/` — pattern for commissions-middleware types
- `packages/node/notification-node/src/` — extend with new event types

---

## 10. Competitive Positioning

| | **Saga Commissions** | Patreon | Kickstarter | Ko-fi | Fiverr |
|---|---|---|---|---|---|
| Community-driven funding | **Yes** | No | Partial | No | No |
| Milestone delivery | **Yes** | No | Stretch goals | No | Yes |
| Backer governance | **Yes** | Polls only | No | No | No |
| Integrated social platform | **Yes** | Limited | No | Limited | No |
| Escrow protection | **Yes** | No | Limited | No | Dispute-based |
| Progress transparency | **Timeline** | Monthly posts | Updates | None | Order status |

**New category**: Not crowdfunding (Kickstarter), not subscriptions (Patreon), not freelance marketplace (Fiverr). **Community-Commissioned Creative Work** — the audience already exists in the community, the value is a specific deliverable with milestones, and the relationship lives in the same social platform.

---

## Verification Plan
1. Schema validation: `yarn workspace @saga/database-node run db:generate` after Prisma changes
2. Build check: `yarn build` — ensure new packages compile
3. API testing: Manual test commission CRUD, pledge flow, milestone submission
4. Payment flow: Test with Stripe test mode — pledge → escrow → milestone approve → disbursement
5. Notification testing: Verify SSE events fire for each commission lifecycle transition
6. Frontend: Visual review of commission card, detail page, milestone timeline, pledge button
7. Integration: Create commission from community, link update posts, verify cross-feature navigation
