# Lesson 36 — P4 Capstone: Ship TeamPulse

*Phase 5 · Session 36 of 38 · ~2 hours (budget 3 for the full send-off) · Prereqs: Lesson(s) [30](30-p3-full-stack-teampulse.md), [32](32-debugging-devtools-network-server-logs.md), [34](34-minimal-testing-teams-actually-use.md), [35](35-deploying-to-vercel.md)*

## Why this matters

This is MISSION.md's graduation ceremony. P0–P3 built the parts; P4 assembles the *job*:
take a working app, pick a feature the way a ticket arrives (scope it, write acceptance
criteria, build, test, deploy publicly), and prove every definition-of-done item on
yourself, not on a grader. When this lesson is done you have a public URL, a passing test
suite, and a feature you specified and shipped end-to-end — the exact artifacts you'd
point at in a review conversation.

## What you'll be able to do

- You can scope ONE feature from a menu into acceptance criteria before writing code, and refuse scope creep.
- You can implement a feature across all three layers — Go-contract → route handler (zod-validated) → typed UI — the way Lessons 24–28 taught.
- You can back the feature with at least one test of each kind from Lesson 34 (unit / route-handler / component).
- You can ship it: green `pnpm build && pnpm lint && pnpm test`, then a public Vercel production deployment that actually works.
- You can walk the definition-of-done checklist and demonstrate every MISSION.md item live.

## Principles

### How real tickets arrive

A ticket is not "add a task form." It is a behavior + boundaries + proof:

> **As a** user, I can do X. **Given** state, **when** action, **then** observable result.
> Out of scope: … . Done when: tests + deploy + no console errors.

Write that paragraph *before* code, on paper or in a scratch file. If you can't write the
"then" line, you don't understand the ticket yet — and if you can write it, the
implementation is transcription. This is also the exact PR description format that makes
you reviewable (Lesson 37).

### The feature menu — pick exactly one

All three work against the mock Go API (`http://localhost:8080`, endpoints and data model
in `backend-mock/README.md`), reuse your P3 stack (zod schemas in the handler, TanStack
Query on the client, session cookie held by the BFF), and are sized for one focused
sprint. Pick one. Writing the other two later is homework for your own portfolio, not
scope for today.

---

**Option A — Task creation form with optimistic update.**

A "New task" form (project select, title, priority) that appears in the list immediately
on submit, before the server confirms.

Acceptance criteria:
1. Given a logged-in session, when I submit a valid form, then a `POST /api/tasks` goes
   through **my route handler** (visible in the Network tab as `/api/tasks`, never
   `:8080` directly), which validates the body with zod *before* calling Go, forwards
   `Authorization` from the session, and returns the created task with 201.
2. The new task appears in the list **optimistically** — TanStack Query's `onMutate`
   inserts a placeholder immediately, `onSuccess` reconciles with the server copy (real
   `id`), and `onError` rolls the list back and shows the error state.
3. Invalid submissions (empty title, priority outside 1–3, no project) are blocked
   client-side *and* rejected by the handler's zod schema with 400 — proving the boundary
   doesn't trust the client (Lesson 25's whole point).
4. One route-handler test (happy + 400 + 401) and one component test (form renders and
   validates), per Lesson 34.

---

**Option B — Project detail page at `/projects/[id]`.**

Clicking a project opens its page: project metadata plus only its tasks, with 404 and
loading/error states.

Acceptance criteria:
1. `/projects/[id]` renders server-side (App Router dynamic segment, Lesson 17), showing
   name, description, status, and tasks fetched via `GET /api/projects/{id}` and
   `GET /api/tasks?projectId={id}` **through route handlers** with zod validation.
2. Unknown id → your `not-found.tsx` boundary (Lesson 21) renders, driven by the handler
   translating Go's 404 into *your* 404 — normalize the error shape, don't forward raw.
3. Slow-network demo: while tasks load, `loading.tsx` shows (test it against the mock's
   `/api/slow`, which takes 2s — Lesson 28).
4. Deep-link works: pasting the URL into a fresh tab (no client cache) renders the page —
   proof it's server-rendered, not client-state-only.
5. One route-handler test per endpoint (including the 404 path) and one component test
   (task list renders N rows given props).

---

**Option C — Weekly summary route: `GET /api/summary/weekly`.**

A BFF-only feature (plus one UI card): your route handler aggregates across Go endpoints
into one reshaped response the UI couldn't easily compose itself.

Acceptance criteria:
1. `GET /api/summary/weekly` (session required) fans out to Go's `/api/projects` and
   `/api/tasks` **in parallel** (`Promise.all`), with the timeout policy from Lesson 28 on
   every call, and returns exactly:

   ```json
   { "generatedAt": "...", "projects": [ { "id": "p1", "name": "...", "total": 4, "done": 1, "overdue": 2 } ] }
   ```

   Reshaping/aggregation happens in the BFF — the client receives nothing it must compute.
2. Every Go response passes zod before use; any Go failure produces your single
   normalized error shape with a sane status, never a stack trace.
3. One dashboard card (Server Component, no new `"use client"` if avoidable — Lesson 33)
   consumes it via TanStack Query with `refetchInterval`.
4. Unit test the pure aggregation function (feed it fixture projects/tasks, assert counts
   — including an overdue task with a past `dueDate`), plus a route-handler test against
   the live mock.

---

### The polish pass (before the feature)

Thirty minutes on P3 hygiene; skip it and the capstone feels like a bug hunt:

- **States**: every page has loading, error, and empty states wired (Lesson 21). Empty
  task list should say "No tasks yet", not render a blank div.
- **One error shape**: every route handler returns your normalized `{ error: string }`
  (Lesson 28) — grep for stray `NextResponse.json({ message: ... })` variants.
- **No direct-Go calls from the browser**: Lesson 32's `grep -rn 8080 src/` check.
- **Env audit**: every `process.env.X` justified (Lesson 35).
- **Dead code**: delete `console.log`s that aren't BFF seam logs; delete commented blocks.

## Follow along

Work in `~/teampulse`. Go mock running; Vercel project connected (Lesson 35).
The steps below are feature-agnostic; bracketed notes point at the option-specific bits.

1. **Branch like it's a ticket.**

   ```bash
   git checkout -b feat/task-creation   # or feat/project-detail, feat/weekly-summary
   ```

   Write the ticket paragraph + acceptance criteria (copy from the menu, add your own
   "out of scope" line — e.g. Option A: *no editing, no deleting, no due dates in the
   form*). Commit it as `docs: feature spec`. Yes, a docs commit before code.

2. **Baseline check.** `pnpm lint && pnpm test && pnpm build` — all green before you
   touch anything. Expected: zero lint errors, suite passes with the Go mock up, build
   prints the route table. A red baseline discovered *after* your changes is a afternoon
   lost to blame-shifting.

3. **Contract first (bottom-up).** Open `backend-mock/main.go` and re-read the endpoint(s)
   your feature consumes or produces (Lesson 24). For [Option A: the `POST /api/tasks`
   handler — body fields `projectId`, `title`, `priority`, response 201 + created task
   including the generated `id` and any server-set fields]. Update your zod schema and
   inferred types in `src/lib/` *first*, so the UI can type-check against the truth.

4. **Route handler.** Add/extend `app/api/.../route.ts`:
   - validate inbound body with zod before any `fetch` to Go ([A]: 400 on bad body);
   - forward `Authorization: Bearer <token>` pulled from the session cookie (Lesson 27) —
     never a token from the request body;
   - `AbortSignal.timeout(...)` on every Go call (Lesson 28);
   - normalize upstream failures into your one error shape;
   - seam log: `console.log("[bff /api/tasks] go:", res.status)` (Lesson 32).
   Verify manually before writing tests: log in via the UI with password `demo`, exercise
   the flow, and read both the Network tab and the dev terminal. Expected: your seam log
   prints for every interaction; no `:8080` rows in the Network tab.

5. **UI.** [A: the form as a small `"use client"` island with controlled inputs (Lesson
   14), a `useMutation` with `onMutate` optimistic insert + rollback (Lesson 26); B: the
   server-rendered page with `params`, `not-found.tsx`, `loading.tsx` (Lessons 17, 21);
   C: the aggregation function in `src/lib/summary.ts` — pure, unit-testable — called by
   the handler, and a Server Component card reading the route through Query].
   Keep the client island small — Lesson 33's bundle bill is still running.

6. **Tests.** Extend the Lesson 34 files: [A: `route.test.ts` for POST — 201 with an `id`,
   400 on empty title, 401 without cookie — plus a component test for the form's
   validation; B: handler tests incl. 404-unknown-project, component test rendering
   fixtures; C: unit test on the pure aggregator incl. an overdue case, plus handler
   test]. Run `pnpm test` after each; expected: green after each addition, not "green
   eventually."

7. **Definition-of-done walk.** Tick every box below against the *running* app before
   deploying.

   - [ ] **Mission 1 — read what runs where.** For every file you touched today, you can
         say server / client / both and cite the evidence (`"use client"`, `async` +
         `fetch`, hook usage).
   - [ ] **Mission 2 — BFF route handler.** Your feature's Go calls all live in
         `app/api/**/route.ts`, zod-validate responses, and reshape for the UI.
   - [ ] **Mission 3 — React page.** State, form(s), and fetching follow the course
         patterns (controlled inputs, Query for server state).
   - [ ] **Mission 4 — trace a bug.** Break one thing on purpose (wrong schema field),
         localize it to a layer in <3 minutes using Network + server logs (Lesson 32),
         then fix it.
   - [ ] **Mission 5 — small reviewable PR.** The branch contains: spec commit, feature
         commits, test commits — each commit one coherent thing; `pnpm lint/test/build`
         green; PR description = your ticket paragraph + "out of scope" + test evidence.
   - [ ] **Course extras.** No secrets in `NEXT_PUBLIC_*` or client files (Lesson 33's
         grep); loading/error/empty states visible when you throttle Network to Slow 4G;
         no console errors on the golden path.

8. **Ship.** Push the branch → watch the **preview deploy** pass (Lesson 35's ritual) →
   confirm the feature works on the preview URL with `GO_URL` pointing at your reachable
   mock → merge → confirm the **production** URL. Expected: public URL, working login
   (`demo`), feature demonstrable in under a minute.

9. **The demo sentence.** Record or rehearse 60 seconds: "This is TeamPulse. I built the
   BFF on top of a Go API, here's my feature, here's where its tests run, here's the prod
   deploy, and here's why `GO_URL` in prod points at the real backend." That sentence is
   the interview version of this lesson.

## Your turn

1. Finish your chosen option to its full acceptance-criteria list — every numbered item
   demonstrable on the production URL. (This *is* the exercise; the two items below are
   the extras that separate "finished" from "senior-finished.")
2. **Adversarial pass.** Try to break your own feature the way QA does: submit the form
   with `"title": ""` and with `priority: 999` directly via curl against your *local*
   dev server's BFF route (not Go). Expected: zod returns 400 and Go never logs a write —
   proof the boundary holds even when the browser is bypassed.
3. **Regression walk.** Click every page of TeamPulse once on the production URL with
   DevTools console + Network open. List every error and every unexpected request. Expected
   by the end of this item: zero of each, or a written decision for each survivor.

<details>
<summary>Solution</summary>

1. Self-check against the menu: re-read each acceptance criterion and demo it, live, on
   the deployed URL. Any criterion you can't demo in one click is unfinished work, not a
   test flake.
2. Example for Option A:

   ```bash
   TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
     -H 'Content-Type: application/json' \
     -d '{"username":"wen","password":"d-mo"}' | jq -r .token)
   curl -si -X POST http://localhost:3000/api/tasks \
     -H "Content-Type: application/json" \
     -b "tp_token=$TOKEN" \
     -d '{"projectId":"p1","title":"","priority":999}'
   ```

   Expected: `HTTP/1.1 400` with your `{ "error": ... }` shape and **no** new task in
   `GET http://localhost:8080/api/tasks?projectId=p1`. (The mock's own password is
   `demo`; the token is only for Go — your BFF authenticates via its session cookie.)
3. Common survivors and their verdicts: a favicon 404 (fix — free win), TanStack Query
   devtools request (remove the devtools package from the prod bundle), a hydration
   warning on one page (Lesson 33 footgun #1 — fix before calling it done), third-party
   analytics you never added (audit `next.config.ts`).

</details>

## Checkpoint

1. Your optimistic update shows the new task, then the server response arrives with a
   *different* `id`. What were the two client-side states, and which TanStack Query
   mutation callbacks manage each?

<details>
<summary>Answer</summary>

Optimistic placeholder (from `onMutate`, temporary id) and server truth. `onSuccess`
reconciles — replace the placeholder with the server object (real `id`), typically by
invalidating/refetching the list query or writing the response into the cache; `onError`
rolls back to the snapshot taken in `onMutate` (Lesson 26).

</details>

2. Why must the zod validation of the task-creation body live in the route handler even
   though the form already validates?

<details>
<summary>Answer</summary>

Client validation is a UX convenience, not a security control — your curl proved the
browser can be bypassed entirely. The BFF is the last place that can enforce the contract
with Go, so it validates everything (Lesson 25: trust nothing, including your own UI).

</details>

3. Which layer owns the 404 for `/projects/does-not-exist` in Option B, and why not the
   browser or React Router "just passing it through"?

<details>
<summary>Answer</summary>

The route handler: it sees Go's 404 and decides the BFF's own status (404 with your
normalized error shape), which Next renders via `not-found.tsx`. The browser can't invent
it, and raw pass-through leaks Go's error body/behavior into your contract (Lesson 28's
normalization rule).

</details>

4. For Option C: why does the aggregation function get a unit test *in addition to* the
   route-handler test, when the handler test already hits the real mock API?

<details>
<summary>Answer</summary>

The handler test proves wiring (status, shape, Go connectivity) but can only feed it the
mock's fixed dataset — no overdue tasks or edge counts unless you create them. A unit
test on the pure function pins the *math* (done/overdue counting, empty lists) with
arbitrary fixtures, in milliseconds, without the Go process. Different failure classes.

</details>

5. You deploy and log in on the production URL with `demo` — the page loads, but every
   task card is missing and Network shows your `/api/tasks` handler returning 500 while
   the mock is confirmed up. Name your first two moves in order (Lesson 32's tree).

<details>
<summary>Answer</summary>

(1) Vercel runtime logs for that route — read the seam log: did the handler run, and what
did Go return? (2) Check `GO_URL` in Vercel's env (is it pointed at a reachable backend,
or still at `localhost:8080`?) — the Lesson 35 "prod broke, dev worked" audit.

</details>

## Primary source

[Vercel — Next.js on Vercel](https://nextjs.org/docs/app/deploying) (Next.js's official
deployment guide). One authoritative walk of build settings, env scopes, and what a
production deployment actually contains — every step-8 claim above is checkable here.

## Before next session

Don't merge-delete your branch yet. Lesson 37 reuses your capstone branch to rehearse the
"smallest reviewable PR" workflow and the day-one questions you'll ask your team — and
your `docs: feature spec` commit is the artifact that makes that rehearsal real.
