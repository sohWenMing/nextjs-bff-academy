# Lesson 32 — Debugging: DevTools, React DevTools, Server Logs

*Phase 5 · Session 32 of 38 · ~2 hours · Prereqs: Lesson(s) [15](15-data-fetching-and-useeffect.md), [20](20-route-handlers-your-first-bff.md), [28](28-resilience-timeouts-retries-error-shapes.md)*

## Why this matters

Mission item #4 is "trace a bug from a browser Network tab → BFF route handler → Go API
response." That sentence is a *decision tree*, and this lesson installs it. On the team,
bugs arrive as "the dashboard shows nothing" — which could be the React component, the
route handler, or the Go backend. Knowing which layer owns a bug in three minutes is the
single highest-leverage debugging skill a BFF owner can have.

## What you'll be able to do

- You can run the three-layer triage tree (UI → BFF → Go) on any "screen is wrong" bug and name the owning layer.
- You can read the Network tab: find a request, read its status, request headers, and response body, and tell a `/api/*` (BFF) request from a direct-to-Go request.
- You can install and use React DevTools: inspect the component tree, read props/state, and take a first Profiler recording to spot the slow component.
- You can read server-side `console.log` output for a route handler and correlate it with a request you just made.
- You can write a reproduction so precise that a teammate can hit the bug on their first try.

## Principles

### The triage tree

Your stack has three layers, and every frontend bug lives in exactly one of them:

```text
Browser (React UI)  →  Next.js BFF (route handlers)  →  Go backend (:8080)
```

The tree, in order — each question has a tool and a binary answer:

1. **Did the right request go out at the correct time?** Open DevTools → Network tab,
   reproduce the bug. No request when there should be one → **UI bug** (a click handler,
   state, or query key problem — Lessons 13–15, 26).
2. **The request went out — what did it return?**
   - Status `500` from your `/api/...` path, or a body that doesn't match what your
     component expects → **BFF bug**. Go to the server terminal (step 3).
   - Status `200` from `/api/...` with the *correct* JSON, but the screen is still wrong →
     **UI bug** again (the component is misreading good data).
   - Status `500`/`502` *and* your route handler's logs show the Go call itself failed →
     **Go bug**. Curl Go directly to confirm (step 4).
3. **Server logs** (the terminal running `pnpm dev`): route-handler `console.log`s print
   here. If your handler logged `go responded 500`, the seam is BFF↔Go, not UI↔BFF. If
   there is *no* log at all for the request, the request never reached the handler —
   back to the Network tab (wrong URL, middleware, or a client cache serving stale data).
4. **Curl the suspect layer directly.** The golden separator: hit
   `http://localhost:8080/api/...` yourself with the token from `/api/login` (password
   `demo`). If curl returns good data but the BFF returns garbage, the BFF is lying. If
   curl *also* fails, the Go backend owns the bug.

The whole skill is: **never debug two layers at once.** Prove one layer innocent, then
descend.

### Network tab: the BFF owner's cockpit

Filter the Network tab by `Fetch/XHR` — that leaves only your `fetch` calls, dropping
images and fonts. Each row is one request. The four columns that matter:

- **Name**: `/api/projects` is your BFF route handler (Lesson 20). A row hitting
  `localhost:8080` directly means something is calling Go from the browser — *a bug in
  itself* (secrets leak, CORS roulette; Lesson 27 says tokens never go client-side).
- **Status**: `200` fine; `401` means the `Authorization` header is missing/expired
  (re-login or cookie bug); `500` is your route handler throwing; `502` in dev often
  means the Go backend isn't running at all.
- **Headers tab of a request**: verify `Authorization: Bearer <token>` is actually
  attached, and check `content-type` on the *response* — `text/html` on a JSON endpoint
  means you hit a 404 page, not your handler.
- **Response / Preview tab**: the literal bytes. When UI and BFF disagree about a field
  name (`dueDate` vs `due_date`), this tab ends the argument.

Two pro moves: **Preserve log** (keeps requests across navigation, so you can see the
request that *caused* a redirect) and **Copy as cURL** (right-click a request → reproduce
it in your terminal without the browser — the fastest way to prove "the browser isn't
sending the cookie").

### Server logs: the BFF↔Go seam

`console.log` inside `app/api/**/route.ts` prints to the *terminal running `pnpm dev`*,
not the browser. Prefix every BFF log with the route and outcome so the seam is visible
at a glance:

```ts
// app/api/tasks/route.ts
const res = await fetch(`${GO_URL}/api/tasks`, {
  headers: { Authorization: `Bearer ${token}` },
  signal: AbortSignal.timeout(3000),
});
console.log("[bff /api/tasks] go:", res.status);
```

Expected (terminal): `[bff /api/tasks] go: 200`. If you see `go: 500`, open
`backend-mock/main.go` — the mock's `/api/flaky` fails ~30% of the time on purpose
(Lesson 28), and correlating "30% of my requests show the empty state" with "flaky" is
exactly the debugging muscle the team will pay you for.

### React DevTools: Components and Profiler at working level

Install the "React Developer Tools" browser extension. Two tabs become available:

- **Components**: a tree of your rendered components. Click one; the side panel shows its
  **props** (what came down from the parent) and **state** (its `useState`/hook values).
  The #1 use: a list renders empty — click the list component, look at its props. If
  `tasks` is `[]`, the data never arrived (go up the stack / to Network). If `tasks` has
  8 items and the list is still empty, the render logic is wrong (go to keys/filters,
  Lesson 12). Also try the **⚛︎ Highlight updates** button, reproduce the interaction, and
  watch which boxes flash — those are the components re-rendering. If clicking one task
  flashes the entire page, you've found a re-render storm (Lesson 33).
- **Profiler**: press record, interact, stop. You get a per-component flamechart of
  *render time per commit*. Working-level use only: find the widest (slowest) bar. You
  do not need to micro-optimize — out of scope per MISSION.md — but you must know that
  one component takes 80% of the commit time.

### Source maps: why your stack trace line makes sense

Browser "Sources" and server stack traces show *your original TS/TSX lines*, not the
compiled JS, because Next.js emits **source maps** — sidecar files that map bundled code
back to source. Working-level consequences: (1) in dev you can set real breakpoints in
`page.tsx`; (2) a server-side exception's stack trace points at `route.ts:31`, trust it;
(3) if line numbers look insane, a source-map-less production build is the reason —
reproduce in `pnpm dev` first, always.

### Reproduction discipline

The difference between "can debug" and "can debug on a team" is the bug report. A
repro is done when a teammate who has never seen the bug hits it on their first try:

> **Steps**: log in as any user → open `/projects/p1` → click "Mark done" on the first
> task. **Expected**: task moves to Done column. **Actual**: spinner forever; Network
> shows `POST /api/tasks` → `500`; server log shows `go: 401`.
> **Frequency**: 1 in 3 attempts. **Environment**: `pnpm dev`, Go mock on :8080, Chrome.

"1 in 3 attempts" is a clue by itself — non-determinism points straight at the flaky
endpoint or a race, not at styling.

## Follow along

Work in `~/teampulse` with the Go mock running (`cd backend-mock && go run main.go`) and
`pnpm dev` in another tab.

1. **Prove the layers are separate.** Open your TeamPulse dashboard, DevTools → Network →
   `Fetch/XHR`, and reload. Expected: rows for your `/api/...` routes only (e.g.
   `/api/projects`, `/api/tasks`) — and *zero* rows addressed to `localhost:8080`. If you
   see a `:8080` row, you have a leaked direct call; find it (`grep -rn 8080 src/app`
   should only match route handlers and lib/server code).

2. **Inject a BFF bug.** In `app/api/tasks/route.ts`, change the zod schema field
   `title: z.string()` to `titlez: z.string()` (Lesson 25). Reload the page. Expected:
   Network shows `/api/tasks` → `500` (or your normalized error shape with a 500 status),
   the UI shows its error state, and the server terminal logs a zod failure. Read the
   *first* server log line — it names the field. Revert the schema; confirm `200`.

3. **Inject a UI bug.** With the BFF healthy, break only the component: in the task list
   component, render `tasks.filter(() => false)` temporarily. Reload. Expected: Network
   shows `/api/tasks` → `200` with a full JSON array in the Response tab, yet the list is
   empty. This is the pattern-match to burn in: **good request + good response + wrong
   screen = UI bug.** Revert.

4. **Meet the flaky endpoint.** Add a temporary button that fetches `/api/flaky` through
   your BFF (route handler forwarding `Authorization`, Lesson 27), click it ~10 times,
   and watch both the Network tab statuses and the server terminal. Expected: roughly
   3 of 10 are `500` at the Go layer with `go: 500` in the server log — the mock's
   built-in ~30% failure rate. Your retry policy from Lesson 28 should mask some of
   these; count how many user-visible errors remain.

5. **React DevTools once, properly.** With the dashboard open: Components tab → click
   the task-list component → confirm `tasks` prop is a non-empty array. Then enable
   Highlight updates, click a filter, and note which components flash. Then Profiler →
   Record → click the same filter → Stop. Expected: a flamechart; identify the widest
   bar and its ms cost. Write both numbers down — Lesson 33 reuses them.

6. **Write the repro.** Take the flaky-endpoint behavior from step 4 and write the
   five-part bug report (steps/expected/actual/frequency/environment) in your scratch
   notes as if filing it for a teammate.

## Your turn

1. A teammate says: "Projects load, but tasks are always empty, no error shown." Using
   only the Network tab and one curl, decide which layer owns this in under 3 minutes.
   What do you look at first, and what observation is decisive?
2. Your login form shows "Something went wrong" on every submit. The Network tab shows
   `POST /api/auth/login` → `401`. Server log shows `go: 401`. Curl the Go mock's login
   directly. What is the most likely cause, and what exact curl proves it?
3. Using React DevTools only (no code changes), how would you determine whether a
   sluggish filter interaction is (a) one slow component or (b) the whole tree
   re-rendering?

<details>
<summary>Solution</summary>

1. Network tab first. Decisive observation: is there a `/api/tasks` request at all? No
   request → UI bug (query disabled/bad key, Lesson 26). Request with `200` → curl
   `http://localhost:8080/api/tasks?projectId=...` with a token from
   `POST /api/login` (password `demo`). Curl returns tasks but the BFF response body is
   empty → BFF bug (reshape/validation). Curl also returns `[]` → Go owns it.

2. Most likely your route handler forwards the wrong credentials — a wrong password,
   missing `Content-Type: application/json`, or a JSON body shaped differently from what
   Go expects. Proof:

   ```bash
   curl -si -X POST http://localhost:8080/api/login \
     -H 'Content-Type: application/json' \
     -d '{"username":"wen","password":"d-mo"}'
   ```

   `200` here + `401` in the browser → the BFF is mangling the request (BFF bug). `401`
   here too → you changed the Go mock or the contract (Go bug).

3. Profiler: record the interaction, stop, look at the flamechart — one dominant wide bar
   = (a) one slow component. Components tab with Highlight updates: if every box on the
   page flashes for one small click = (b) whole-tree re-render.

</details>

## Checkpoint

1. A request appears in the Network tab addressed directly to `localhost:8080` from the
   browser. Why is that not merely a style problem?

<details>
<summary>Answer</summary>

It bypasses the BFF: the token is in browser-visible JS (Lesson 27 says the BFF holds the
session), request shaping/validation/retries are skipped, and CORS becomes a variable.
All Go traffic belongs in `app/api/**/route.ts` (Lesson 20).

</details>

2. Network tab shows your `/api/projects` route returning `200`, but the Response tab
   contains an HTML page, not JSON. What probably happened?

<details>
<summary>Answer</summary>

The request didn't hit your route handler — a wrong URL or missing file hit a page/404
route that returns HTML with a 200 in some dev setups. Check the path matches
`app/api/projects/route.ts` exactly and no middleware rewrote it (Lesson 29).

</details>

3. Where do `console.log`s from a route handler appear, and what does "no log for a
   request I can see in the Network tab" mean?

<details>
<summary>Answer</summary>

They appear in the terminal running `pnpm dev` (server side). No log means the request
never executed the handler — wrong URL, middleware short-circuit, or the browser/TanStack
Query cache answered without a network request at all.

</details>

4. Which React DevTools tab answers "what data does this component have?" and which
   answers "what is slow?"

<details>
<summary>Answer</summary>

Components tab for props/state of a selected component; Profiler for per-component render
cost per commit.

</details>

5. Your "tasks sometimes vanish" bug reproduces ~3 times in 10 against the mock Go API at
   `http://localhost:8080`. What does the ~30% frequency point at, and how do you confirm
   without touching the UI?

<details>
<summary>Answer</summary>

`GET /api/flaky` (or a route forwarding to it) — the mock fails ~30% of those with 500
(Lesson 28). Confirm by curling `/api/flaky` with a valid token repeatedly, or by the
`go: 500` lines in the server log for those requests.

</details>

## Primary source

[Chrome DevTools — Monitor Network Activity](https://developer.chrome.com/docs/devtools/network).
The canonical reference for the panel that separates UI bugs from BFF bugs, including
filters, the request detail panes, and Copy as cURL.

## Before next session

Keep React DevTools installed. In your Profiler notes from step 5, add your dashboard's
route and the widest component's ms cost — Lesson 33 turns "working" into "fast enough",
and we'll start from your numbers.
