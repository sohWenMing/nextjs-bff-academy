# Next.js BFF Academy

A self-paced, 38-session course (2 hours/day, ~7 weeks) that takes you from
**zero TypeScript / zero React** to **operating a Next.js frontend + Backend-for-Frontend
on top of a Go backend** — the exact stack your team is shipping.

Built for working programmers who know how to code but have never touched the
JS/TS/React ecosystem.

## The setup

- **2 hours a day, ~6 days a week.** One lesson = one session. Never do two at once.
- You learn by **building one app that grows all course long**: *TeamPulse*, a team
  dashboard talking to a **real Go backend** (`backend-mock/`, stdlib-only Go, runs
  with `go run main.go` on `:8080`).
- Every lesson: principles → guided build → solo exercises → checkpoint questions.
  If you can't answer the checkpoints cold, you redo the session. That's the whole
  method.

## Start here

1. Read [MISSION.md](MISSION.md) — what "done" looks like.
2. Read [SYLLABUS.md](SYLLABUS.md) — the 38 sessions and the phase gates.
3. Install: [Node.js LTS](https://nodejs.org) (`brew install node`),
   `corepack enable` (gives you pnpm), and [Go](https://go.dev/dl) (`brew install go`).
4. Smoke-test the backend so you trust it:
   ```bash
   cd backend-mock && go run main.go
   curl -s localhost:8080/api/health
   ```
5. Open `lessons/01-toolchain-node-pnpm-your-first-scripts.md`. Go.

## The five projects (each builds on the last)

| Project | Phase | What you end up with |
|---|---|---|
| **P0** TeamPulse CLI | 1 | A TypeScript command-line client that logs into the Go API and lists projects/tasks |
| **P1** Explorer page | 2 | A React page: login form, project list, task list with filter — all client-side |
| **P2** First BFF | 3 | All Go traffic moved into route handlers; server-rendered pages; no secrets in the browser |
| **P3** Real engineering | 4 | zod-validated typed contracts, TanStack Query caching, token auth through the BFF, timeouts/retries |
| **P4** Capstone | 5 | Polished, tested, deployed-to-Vercel TeamPulse + a self-written feature |

## Reference material (skim now, return forever)

- [Glossary](reference/glossary.md) — every term, in plain language, in course order
- [TypeScript cheat sheet](reference/typescript-cheatsheet.md) — the 20% of TS you use daily
- [React cheat sheet](reference/react-cheatsheet.md) — component/state/render model
- [Next.js cheat sheet](reference/nextjs-cheatsheet.md) — file conventions, server vs client, route handlers
- [Error decoding guide](reference/error-decoding.md) — "what the hell does this TS error mean"

## Rules of engagement

1. Type every line yourself. Copy-paste is how you fool yourself.
2. Never skip a checkpoint. Recall now beats recognition later.
3. When stuck >20 minutes on one error, look it up in
   [error-decoding.md](reference/error-decoding.md), then the primary source the
   lesson links, then ask your team (or your AI agent).
4. Project sessions are sacred — that's where the transfer happens.

## Daily rhythm (the 2 hours)

- 0:00–0:35 — read the Principles section, typing every code block
- 0:35–1:15 — Follow along, verbatim, expecting each stated output
- 1:15–1:50 — Your turn, solo, no peeking at solutions until you've attempted
- 1:50–2:00 — Checkpoint, cold recall; mark anything shaky for tomorrow's warm-up

## Why this curriculum, in this order

JS before TS (types are notation, not a new language) → React before App Router
(the mental model transfers) → App Router before BFF (you must know what runs
where before you place secrets) → BFF before polish (it's the part of the job
that's actually yours). Node, ESLint, tests, and deployment arrive exactly when
they become load-bearing, not before.
