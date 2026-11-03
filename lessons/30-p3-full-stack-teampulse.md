# Lesson 30 — [P3] Full-stack TeamPulse

*Phase 4 · Session 30 of 38 · ~2 hours · Prereqs: Lessons 24, 25, 26, 27, 28, 29*

## Why this matters

This is the phase-defining build: one pass over TeamPulse that makes every layer earn
its keep — the Go contract read properly (24), enforced by zod at the BFF boundary (25),
cached on the client by TanStack Query (26), authenticated through httpOnly cookies with
the token never touching browser JS (27), hardened with timeouts, a real retry policy,
and one normalized error shape (28), and wired per-environment with middleware gating
(29). When you finish, you own the entire seam between your team's Go and your team's
React — which *is* the job. The phase gate at the end of this file is the exam: new Go
endpoint → typed, validated, cached UI in under 45 minutes.

## What you'll be able to do

- Produce the P3 file tree from memory and state what each file guarantees
- Get a clean `pnpm build` with the whole stack wired — no `any`, no direct
  browser→Go calls, no token in any browser-visible payload
- Pass the six acceptance demos: login, guarded route, validated contract, cached
  switch, timeout, flaky retry
- Run the phase-gate drill (add a Go endpoint, surface it end-to-end) against the
  checklist without lesson help

## Principles

P3 introduces no new concepts — that's the point. It's the composition test. Hold the
five-layer stack in your head; every file in the tree belongs to exactly one layer:

```
┌ browser ──────────────────────────────────────────────────────┐
│ L5  client components + TanStack Query   (26)  cache, 3 states│
│     queryKeys · useMutation · invalidate · retry policy       │
├───────────────────────────────────────────────────────────────┤
│ L4  middleware.ts                          (29)  presence gate│
├─ Next.js server ──────────────────────────────────────────────┤
│ L3  app/api/*/route.ts                     (20)  the BFF      │
│     status handling · authHeaders · NextResponse              │
│ L2  lib/upstream.ts + lib/schemas.ts       (25)  the contract │
│     callGo + AbortSignal.timeout (28) + zod.parse             │
│ L1  lib/errors.ts                          (28)  one UiError  │
├───────────────────────────────────────────────────────────────┤
│ Go  backend-mock/main.go                   (24)  the authority│
└───────────────────────────────────────────────────────────────┘
```

Three invariants decide whether P3 passes — recite them before you start, they're the
grading rubric:

1. **No browser-visible payload ever contains `tok-`.** The token's entire life is
   Go → route handler → cookie jar → route handler → Go. (Lesson 27.)
2. **Nothing enters Next.js unvalidated.** Every upstream body passes `callGo` + a zod
   schema; every request body a route handler accepts passes a schema. Interfaces are
   `z.infer` exports — nobody hand-writes a shape. (Lessons 24–25.)
3. **Every failure reaches the UI as a `UiError`.** Timeouts, Go 5xx, network death,
   contract violations, 401s — one shape, `code` + `message` + `retryable`, rendered by
   one component family. (Lesson 28.)

### The target file tree

Your `~/teampulse` should land here (names exact — later phases link to them):

```
teampulse/
├── middleware.ts                  # L4: /dashboard + /tasks gating (29)
├── next.config.ts                 # hardening headers only (29)
├── .env.local                     # GO_API_BASE_URL (never NEXT_PUBLIC_) (29)
├── lib/
│   ├── schemas.ts                 # zod schemas + z.infer types (25)
│   ├── upstream.ts                # callGo, UpstreamError, authHeaders (25/27/28)
│   ├── errors.ts                  # UiError, toUiError, statusFor (28)
│   └── queries.ts                 # typed fetchers: getJson + ApiRequestError (26/28)
├── app/
│   ├── layout.tsx                 # server; wraps <Providers> (26)
│   ├── providers.tsx              # "use client" QueryClientProvider (26)
│   ├── login/page.tsx             # client form → POST /api/login (14/27)
│   ├── dashboard/page.tsx         # client; projects + tasks + add form
│   ├── error.tsx                  # renders UiError, Retry when retryable (21/28)
│   └── api/
│       ├── login/route.ts         # credential exchange + cookie set (27)
│       ├── logout/route.ts        # cookie delete (27)
│       ├── me/route.ts            # UserSchema proxy (25/27)
│       ├── projects/route.ts      # z.array(ProjectSchema) (25)
│       ├── projects/[id]/route.ts # 404 passthrough (25)
│       ├── tasks/route.ts         # GET list + POST 201 (25/26)
│       ├── flaky/route.ts         # resilience demo route (28)
│       └── slow/route.ts          # timeout demo route (28)
```

## Follow along

Two hours, six milestones, checkpoints built in. Work top-down through the stack — a
layer only gets built once the layer below passes its demo. Start Go
(`cd ~/workspace/nextjs-bff-academy/backend-mock && go run main.go`) and keep
`pnpm dev` plus a DevTools Network tab open the whole time.

1. **M0 — baseline (5 min).** Delete `.next`, run `pnpm build`. Expected: route table
   listing `/login`, `/dashboard`, and every `/api/*` route as you understand it —
   `ƒ` (dynamic, server) next to the API routes. If the build is already red, fix that
   before anything else; P3 assumes green.

2. **M1 — contract layer green (15 min).** Confirm `lib/schemas.ts` exports
   `UserSchema`, `ProjectSchema`, `TaskSchema`, `LoginResponseSchema`,
   `CreateTaskSchema`, `FlakySchema`, `SlowSchema` and `z.infer` types for each; confirm
   `lib/upstream.ts` exports `callGo` (with `cache: "no-store"`,
   `signal: AbortSignal.timeout(...)`), `UpstreamError`, `authHeaders`; confirm
   `lib/errors.ts` exports `UiError`, `toUiError`, `statusFor`. Spot-check the contract
   against the source of truth:
   ```bash
   grep -n 'json:"' ../workspace/nextjs-bff-academy/backend-mock/main.go
   ```
   Expected: every tag you see has a schema field; every schema field has a tag
   (`projectId`, `dueDate`, `assignee` especially). Expected green `pnpm tsc --noEmit`.

3. **M2 — auth round-trip (15 min).** Fresh cookies (DevTools → Application → Cookies
   → clear; restart dev server). Sign in at `/login` with username `wen`, password the
   literal word `demo`. Walk the checklist in the Network tab:
   - `POST /api/login` request body has the credentials; response body is **user only**
   - response headers contain `set-cookie: tp_token=tok-...; HttpOnly; Path=/; SameSite=lax`
   - next request to `/dashboard` carries the cookie; its data calls (`/api/me`,
     `/api/projects`, `/api/tasks`) show the cookie in Request Headers
   - search the Network tab (`Cmd+F` in DevTools global search does *not* cover
     headers — use the per-request view): `tok-` appears in no response **body**.
   Expected: all four hold. Then log out → `/api/me` 401s → redirected to `/login`
   (the Lesson 28 redirect inside `getJson`).

4. **M3 — caching behaviour (20 min).** On `/dashboard`, exercise the cache with the
   Network tab open, and expect exactly this:
   - mount: one `projects`, one `me`, one `tasks?p1` — each with visible loading state
     (`isLoading`) and an error branch present in code (`isError` renders `UiError`)
   - switch project filter p1 → p2 → p1: keys `["tasks","p1"]`/`["tasks","p2"]` —
     second visit to p1 renders from cache while revalidating
   - add a task via the form: one `POST` then a refetched `GET` for the active key —
     `invalidateQueries({ queryKey: ["tasks"] })` doing its job; the new task shows
     server-computed `id`, `status: "todo"`, `dueDate` +7 days, `assignee: "u1"`
     (check them against `handleTasks` in `main.go` — Lesson 24's habit: trust the
     source, not your memory)
   - mount two components needing `["projects"]` (e.g. sidebar + filter): **one**
     request. Expected: all four behaviours pass cold, no code changes needed.

5. **M4 — resilience demos (20 min).** These are the acceptance tests, not toys —
   script them and run them in order:
   - **Timeout:** temporarily set `callGo`'s timeout to `1_000`; hit the slow route
     (`/dashboard` debug button or curl `localhost:3000/api/slow` with your cookie).
     Expected: ~1s, HTTP 504, body exactly `{"code":"timeout","message":"The server
     took too long to respond.","retryable":true}`. Restore `5_000`.
   - **Retry:** reload the flaky panel 10 times. Expected: at most ~1 visible failure
     (30% per-attempt × two retries ≈ 3% visible), and the Network tab shows the
     retry spacing growing (1s, 2s).
   - **No 4xx retry:** sign out, trigger a data load. Expected: exactly one request,
     `{"code":"unauthorized","retryable":false}`, UI shows "sign in again", no spinner
     storm.
   - **Contract violation:** rename a schema field to a lie (`assignee` →
     `assignedTo`), load tasks, read the UI ("Unexpected data from the server.", no
     Retry button) *and* the server terminal (full `ZodError` with the path). Revert.
   - **Network death:** stop Go, reload dashboard. Expected: `code: "network"`,
     retryable, Retry button present; restart Go, click Retry, panel recovers without
     page reload.

6. **M5 — config & gating (10 min).** Middleware: sign out, navigate directly to
   `/dashboard` → redirected before render. Spoofed cookie (`curl --cookie
   "tp_token=fake"` to `/dashboard`) → passes middleware, data calls 401 — prove you can
   explain why that's correct, not a bug. Env: rename `.env.local`'s `GO_API_BASE_URL`
   line, restart, confirm the `network` UiError appears (misconfig fails *loudly but
   safely*); restore. Global-search the served JS (DevTools Sources, Cmd+F): `8080`
   absent, `tok-` absent.

7. **M6 — green build & self-review (15 min).**
   ```bash
   pnpm tsc --noEmit && pnpm build
   ```
   Expected: clean compile, clean build. Then `git diff` review against the three
   invariants; grep your own app as a hostile reviewer would:
   ```bash
   grep -rn "any" app lib --include='*.ts' --include='*.tsx' | grep -v "z.infer"  # zero hits
   grep -rn "8080" app lib --include='*.tsx'                                        # zero hits (env only)
   grep -rn "localStorage" app lib                                                   # zero hits
   ```
   Commit as `P3: typed, validated, cached, authenticated, resilient TeamPulse`.

### The phase gate (the exam)

From the syllabus: *"you can add a new Go endpoint (in `main.go`) and surface it in the
UI — typed, validated, cached — in under 45 minutes."* Set a timer. Steps, which you
should now be able to do without re-reading any lesson:

1. In `backend-mock/main.go`, after `handleSlow`, add:
   ```go
   func handleStats(w http.ResponseWriter, r *http.Request) {
     if _, ok := requireAuth(w, r); !ok {
       return
     }
     writeJSON(w, http.StatusOK, map[string]any{
       "projects": len(projects),
       "tasks":    len(tasks),
     })
   }
   ```
   and register `mux.HandleFunc("/api/stats", handleStats)` in `main`. Restart Go.
2. Read it like Lesson 24: auth required, 200 shape `{projects: number, tasks: number}`,
   failure shape `{error}`.
3. `StatsSchema = z.object({ projects: z.number().int(), tasks: z.number().int() })` in
   `lib/schemas.ts` + `export type Stats = z.infer<typeof StatsSchema>`.
4. `app/api/stats/route.ts`: `callGo("/api/stats", StatsSchema, { headers:
   authHeaders(req) })`, catch → `toUiError`/`statusFor`.
5. `fetchStats` in `lib/queries.ts`; a `<StatsPanel/>` on the dashboard with
   `useQuery({ queryKey: ["stats"], queryFn: fetchStats })`, three states, and
   `invalidateQueries({ queryKey: ["stats"] })` added to the add-task `onSuccess`.
6. Verify: numbers match `curl` against Go; add a task → both caches refresh; sign out
   → stats panel shows unauthorized, no retry storm.

If you finished inside 45 minutes with the timer honest, Phase 4 is passed — write
"passed, N minutes" in `~/teampulse/PROGRESS.md`. If not, find the layer that slowed
you down (it's always one of: hand-writing types, forgetting the cookie, or fumbling
v4-style query syntax) and redo that lesson's "Your turn" first.

## Your turn

1. **Hardening pass (the one new thing):** add an auth-aware `error.tsx` so a crashed
   subtree shows the `UiError` UI, not React's default. Given `error.tsx` receives a
   plain `Error`, how do you recover the `code`, and what do you show when it's missing?

   <details><summary>Solution</summary>

   ```tsx
   "use client";

   import { ApiRequestError } from "@/lib/queries";

   export default function ErrorBoundary({
     error,
     reset,
   }: {
     error: Error & { digest?: string };
     reset: () => void;
   }) {
     const isApi = error instanceof ApiRequestError;
     return (
       <main>
         <h2>{isApi ? error.message : "Something broke."}</h2>
         {isApi && error.retryable && <button onClick={reset}>Retry</button>}
         {error.digest && <pre>{error.digest}</pre>}
       </main>
     );
   }
   ```
   `instanceof` survives the client bundle because `ApiRequestError` is imported, not
   ambient. Non-`ApiRequestError` errors (real bugs) show a calm generic message plus
   Next's `digest`, which keys into the server log — the user sees nothing sensitive,
   you see everything.
   </details>

2. **Adversarial pass:** three bugs were introduced into a classmate's P3. Name the
   layer (L1–L5, from the stack diagram) each belongs to and the demo from this lesson
   that would have caught it:
   (a) login handler: `NextResponse.json({ token: data.token, user: data.user })`;
   (b) `useQuery({ queryKey: ["tasks"], queryFn: () => fetchTasks(projectId) })`;
   (c) a schema with `priority: z.string()`.

   <details><summary>Solution</summary>
   (a) L3 (login route) — violates invariant 1; M2's "no `tok-` in any response body"
   check catches it. (b) L5 (query key) — parameter missing from key; M3's p1→p2→p1
   switch shows stale data. (c) L2 (contract) — M4's contract-violation-style check;
   the moment a real `number` arrives, `parse` throws and the boundary fails loudly.
   The point of the demos: each is a tripwire for exactly one class of mistake.
   </details>

3. **Teach-back:** write, in `~/teampulse/README.md`, a 10-line onboarding note for a
   new teammate: the five layers, the three invariants, and where each lives. If a
   sentence can't be written without looking something up, that's tomorrow's warm-up.

   <details><summary>Solution</summary>
   No fixed text — self-grade: a reader who has never seen the repo should be able to
   answer "where do I add a new endpoint?" (schemas → upstream route → queries →
   component) and "why can't the token be in React state?" (invariant 1 → Lesson 27's
   XSS reasoning) from your note alone.
   </details>

## Checkpoint

1. Which file holds the single upstream-call implementation, and what are the four
   things it does on every call?

   <details><summary>Answer</summary>
   `lib/upstream.ts`'s `callGo`: (1) prefixes `GO_API_BASE_URL` (env, server-side —
   Lesson 29), (2) sends `cache: "no-store"` and `signal: AbortSignal.timeout(...)`
   (Lesson 28), (3) checks `res.ok` first, turning error bodies into `UpstreamError`
   (Lesson 24/25), (4) validates success bodies with the passed zod schema — throwing
   `ZodError` on contract violations.
   </details>

2. Where is the Go token at these four moments: just after login, mid-day during a
   task fetch, inside your route handler when forwarding, and in DevTools Console
   running `document.cookie`?

   <details><summary>Answer</summary>
   Just after login: Go's response → route handler → `Set-Cookie`; browser cookie jar
   (httpOnly). Mid-day: attached automatically by the browser to same-origin `/api/*`
   requests only. In the handler: read once by `authHeaders`, written into Go's
   `Authorization: Bearer` header, never serialized to the client. In Console: nowhere
   — `document.cookie` cannot see httpOnly cookies (Lesson 27).
   </details>

3. (Go API question) Your stats-panel drill added `GET /api/stats` requiring auth. You
   skipped the cookie. What exact bytes does the UI end up rendering, and trace them:
   Go → … → component.

   <details><summary>Answer</summary>
   Go's `requireAuth` → 401 `{"error":"missing or invalid token"}`. Actually your BFF
   short-circuits first: `authHeaders(req)` sees no cookie and throws
   `UpstreamError(401)` without calling Go — route catch → `toUiError` →
   `{"code":"unauthorized","message":"not logged in","retryable":false}` with status
   401 → client `getJson` throws `ApiRequestError` → TanStack `isError` → panel renders
   "not logged in", no Retry button. (If the cookie existed but was invalid, the same
   path with Go's message instead.)
   </details>

4. A user leaves `/dashboard` open overnight, returns, clicks a filter, and TanStack
   fetches — but Go restarts wiped its `tokens` map, so everything 401s. Walk the
   recovery and say which layer decides each step.

   <details><summary>Answer</summary>
   Query fetches, route handler forwards dead token, Go's `requireAuth` rejects →
   `unauthorized` UiError → `getJson` sees 401 → redirects to `/login` (L5 fetcher
   policy) → user re-authenticates, fresh cookie, cache repopulates from clean mount.
   Middleware (L4) also helps on the *next* full navigation. Nobody retries a 401
   (L5 policy, Lesson 28).
   </details>

5. Name the three invariants cold. Then say which one the `grep -rn "localStorage"`
   check hunts, and why that grep is even in the review list.

   <details><summary>Answer</summary>
   1) No browser-visible payload contains the token. 2) Nothing enters Next.js
   unvalidated. 3) Every failure reaches the UI as a `UiError`. The localStorage grep
   hunts invariant 1 — P1-era code stored the token there (Lesson 16); leftover
   `localStorage.setItem("token", ...)` means a token leaked to JS somewhere, even if
   the happy path looks clean now. Grep is a tripwire for dead-but-dangerous code.
   </details>

## Primary source

[nextjs.org — Route handlers](https://nextjs.org/docs/app/api-reference/file-conventions/route) —
the official `route.ts` reference: exported HTTP method signatures, `NextRequest`/
`NextResponse` (including `cookies`), and the dynamic-ness rules — the one page to
keep open during M2 and M6.

## Before next session

Phase 5 begins with error decoding. Tonight, run `pnpm build` with one deliberate
type error (add `const x: number = fetchProjects();` in any client file), paste the
compiler's first three lines into a note verbatim, and circle the one phrase you
couldn't parse — Lesson 31 opens with exactly that kind of output.
