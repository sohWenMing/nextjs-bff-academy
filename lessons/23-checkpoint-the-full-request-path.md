# Lesson 23 — Checkpoint: the full request path

*Phase 3 · Session 23 of 38 · ~2 hours · Prereqs: Lesson(s) 17, 18, 19, 20, 21, 22*

## Why this matters

Phase 3's phase gate is the sentence: *"given any `.tsx` file, you state within 10 seconds whether it runs on the server, the browser, or both — and how you know."* Today is not new material — it's the audit. You will trace one real click end-to-end through your own P2 app, draw the architecture from memory, and take an integrated quiz across [Lesson 17](17-file-system-routing.md)–[Lesson 22](22-p2-teampulse-bff.md). Everything in Phase 4 (typed contracts, caching, auth, retries) lands on top of this path being *obvious* to you. If a hop is fuzzy today, it will be invisible in Lesson 25's stack traces.

## What you'll be able to do

- You can narrate, unaided: browser click → client component → route handler → `lib/go.ts` → Go :8080 → response → `router.refresh()` → server re-render → updated HTML.
- You can draw the P2 architecture (two servers, one boundary, one env file) on paper from memory, secrets and runtimes labeled.
- You can classify every file in `~/teampulse/app/` as server-only, browser, or both — and name the evidence for each.
- You can score ≥80% on the integrated quiz cold; below that, you know exactly which lesson to revisit.
- You can debug a broken hop by checking that hop's contract first, not by googling the symptom.

## Principles

### The path, canonically (this is the thing you're certifying today)

The click that creates a task in your P2 app touches **nine** checkpoints. Read until it's a film strip in your head:

```text
1 BROWSER  user clicks Add            (create-task.tsx — "use client")
2 BROWSER  onSubmit → fetch POST /api/tasks   same-origin :3000, JSON body
3 SERVER   Next.js matches app/api/tasks/route.ts → calls exported POST()
4 SERVER   handler validates body      → 400 path: NextResponse, done
5 SERVER   createTask() in lib/go.ts   → POST /api/login to Go (demo) → token
6 SERVER   POST /api/tasks + Bearer    to Go :8080 → 201 + created task JSON
7 SERVER   handler returns NextResponse.json(task, {status: 201})
8 BROWSER  await resolves, res.ok      → setTitle(""), router.refresh()
9 SERVER   server components re-run    (page.tsx awaits getTasks via lib/go)
         → streamed HTML → browser paints the new task
```

Every arrow crosses a *contract*: browser↔Next.js speaks same-origin HTTP + your error shape; Next.js↔Go speaks Go's contract (`backend-mock/main.go`, Bearer auth, `{"error":...}`). The BFF's job is that neither side sees the other's mess.

### The debugging corollary (mission item #4, rehearsal)

Each hop has exactly one observable that tells you which side is wrong:

| Symptom | First check | Tool |
|---|---|---|
| Click does nothing | `onSubmit` ran? console in the client component | browser console |
| POST /api/tasks 400/500 | your handler's response body + dev terminal | DevTools Network |
| Handler ran, 502 `go down` | lib/go error (`go 401...` vs `fetch failed` = Go not running) | dev terminal |
| Go returned 201, UI unchanged | `router.refresh()` called? fetch cached? revalidate set? | Network tab |
| Page 404s on refresh | route exists? `page.tsx` vs `route.ts` at that path? | file tree |

The professional habit: **locate the hop before fixing anything.** The white screen at hop 3 and the white screen at hop 9 need opposite fixes.

### Boundaries, once more with evidence

Your 10-second classifier, with the evidence you cite:

- **Server-only**: no `"use client"` anywhere in its import chain; may be
`async`, import `lib/go`, touch `process.env`. Evidence: grep imports.
- **Browser**: `"use client"` at line 1 (or imports something with it —
transitive!). Evidence: the directive, or an import that drags it in.
- **Both**: client components render on the server once for initial HTML,
then live in the browser forever. Evidence: same file, two lifetimes — that's why hooks must survive hydration, not just first render.
- **`route.ts`**: server, always, no render at all. Evidence: exports HTTP
verbs, sits under `app/api/**`.

### Why we draw it

Diagramming forces the distinctions quizzes let you fake. Your diagram must have: two boxes (Next.js :3000, Go :8080), one arrow between them labeled with *who initiates it* (only Next.js), the browser touching only :3000, the `"use client"` boundary drawn as a line through the component tree, `.env.local`/`GO_BASE_URL`/`demo` physically inside the :3000 box, and one labeled `loading.tsx`/`error.tsx` pair showing which boundary catches what.

## Follow along

Everything is already built (Lesson 22). Today you interrogate it.

1. **Live trace with two windows open.** Dev server + Go running; DevTools
open on `/projects/p2`, Network tab recording, browser console visible, dev-server terminal visible. Create a task. Narrate out loud, checking each against what you actually observe:

   - click → expect no page navigation (hop 1–2: SPA form submit)
   - Network tab → one `POST /api/tasks`, Status 201, same-origin `:3000`,
request body `{projectId,title,priority}` (hop 2–3)
   - handler ran → your dev terminal prints nothing unless you added a
`console.log`; add one inside `POST()` in `route.ts` and retry to prove the server executed it (hop 3–7)
   - Go saw it → in `backend-mock`, no logs by default; instead restart Go
mid-test later (step 4). For now verify hop 6 by result: the response JSON has Go's random `t<number>` id and `dueDate` +7 days — only Go mints those (hop 6–7)
   - `router.refresh()` → Network tab shows a document/rsc request to
`/projects/p2` and the new task appears with no manual refresh (hop 8–9)

2. **Prove the boundary, not the vibes.** In a terminal:

   ```bash
   cd ~/teampulse
   grep -rln '"use client"' app/ ; echo --- ; grep -rln 'lib/go' app/ ; echo --- ; grep -rln 'route.ts' app/ | cat; find app -name 'route.ts'
   ```

Build the classifier table for every file under `app/` on paper: file → server / browser / both / route-handler → *evidence* (the grep hit or absence that proves it). Anything you can't classify in 10 seconds is your personal Lesson 18 homework.

3. **Draw from memory (25 minutes, no peeking).** Paper or whiteboard.
Required elements (the list from Principles). Then compare against the real tree:

   ```bash
   find app lib -type f | sort
   ```

Score yourself: missing arrow direction, secrets on the wrong side of a box, a browser→Go arrow anywhere, missing boundary files = draw it again.

4. **Break the path on purpose (bisect drill).** Do each row, predict the
symptom *before* acting, confirm, restore:

   - Kill Go → create task → expected: your POST returns 502 JSON (hop 5–6
failed in the lib), form shows the `msg` error, page still renders (server render also failed → check your `error.tsx`).
   - Restore Go. In `lib/go.ts`, change login username to `""` → create task
→ expected: still 200 login (Go accepts any username) — proving which field actually gates auth (the password). Change password to wrong → `go login failed: 401` → your 502/500 path.
   - In `create-task.tsx`, comment out `router.refresh()` → create task →
expected: 201 success but *UI unchanged until manual reload* — that symptom alone isolates hop 8.
   - Visit `/api/tasks` (GET) in the browser → JSON, no HTML: same folder,
`route.ts` not `page.tsx` ([Lesson 17](17-file-system-routing.md)'s conflict rule).

5. **Rubber-duck the gate.** Say out loud, to anything that tolerates you:
"This file runs on the server because ____; this one ships to the browser because ____; the token is safe because ____." Three blanks, filled without notes, and the phase gate is passed.

## Your turn

1. Write the narration of hop sequence 1–9 **from memory** in ≤10 numbered
lines, each line tagged `[browser]` or `[server]`. Then re-read the Principles film strip and mark your own omissions.

   <details><summary>Solution</summary>
Self-grade against the nine checkpoints; typical omissions: forgetting login is a *separate* Go call per request (hop 5), forgetting the response is JSON from *your* handler not Go's raw response (hop 7 — same body today, but your status-code policy), and forgetting refresh() causes a *server re-render*, not a client re-fetch from a cache (hop 9).
   </details>

2. A teammate pastes this file with no other context. Classify it and list
every violation in P2's rules:

   ```tsx
   "use client";
   import { useState } from "react";

   const GO = "http://localhost:8080";
   const PW = "demo";

   export function TaskList() {
     const [tasks, setTasks] = useState<Task[]>([]);
     async function load() {
       const l = await fetch(`${GO}/api/login`, { method: "POST", body: JSON.stringify({ username: "x", password: PW }) });
       const { token } = await l.json();
       const r = await fetch(`${GO}/api/tasks`, { headers: { Authorization: `Bearer ${token}` } });
       setTasks(await r.json());
     }
     return <button onClick={load}>load</button>;
   }
   ```

   <details><summary>Answer</summary>
Classification: browser code (`"use client"` — everything in it ships). Violations: (1) Go URL hardcoded in browser bundle — the BFF's purpose inverted; (2) `demo` shipped as JS — credential leak, Lesson 20's landmine made real; (3) token held in browser; (4) cross-origin call to :8080 will die on CORS anyway; (5) no loading/error UI for the button (Lesson 21); (6) `Task` type undefined as written. Fix: it's already written — as the server-rendered Lesson 22 pages plus `GET /api/tasks`. The right review comment is "delete this file; the list comes from the server".
   </details>

3. Harder — contract reasoning: Go returns `{"error":"project not found"}`
with 404 for `/api/projects/p9`, yet visiting `/projects/p9` in the browser shows your `not-found.tsx`, not your `error.tsx`. Name the exact two lines in `app/projects/[id]/page.tsx` that make that happen and what each maps upstream status → convention file.

   <details><summary>Answer</summary>
`if (err instanceof GoError && err.status === 404) notFound();` — maps Go's 404 to the nearest `not-found.tsx` ([Lesson 21](21-loading-error-not-found-states.md)); and `throw err;` — every other failure (500, 502, `fetch failed`) propagates to the nearest `error.tsx`. The BFF-side status mapping (`route.ts` turning GoError statuses into NextResponse statuses) is the sibling contract for API consumers.
   </details>

## Checkpoint

1. Cold, from memory: the 9-hop task-creation path — one line per hop,
server/browser tagged.

   <details><summary>Answer</summary>
Browser click (client comp) → browser fetch POST /api/tasks → server route.ts POST() runs → handler validates (400 exit) → lib/go logs in to Go → Go creates task, 201 → handler 201 JSON back → browser sees ok, calls router.refresh() → server components re-run + fetch tasks, streamed HTML updates the list. Tags: 1–2 browser, 3–7 server, 8 browser, 9 server → paint.
   </details>

2. A file under `app/` has no `"use client"` but imports
`./some-client-comp.tsx`. Give its classification and its *browser payload*: what, if anything, ships?

   <details><summary>Answer</summary>
The file itself: server-only. Its imported client component (and that component's import graph): ships to the browser. Boundaries are per-import and transitive downward — the file's own code never does.
   </details>

3. Which single observable distinguishes "handler never ran" from "handler
ran, Go unreachable", and where do you look for it?

   <details><summary>Answer</summary>
The dev-server terminal: `lib/go.ts` throws `fetch failed` only when it actually tried (handler ran); no server log/line at all + no Network hit under `/api/` means the request never reached the handler (wrong URL / missing `route.ts`).
   </details>

4. TeamPulse: someone renames `app/api/tasks/route.ts` to
`app/api/tasks/route_backup.tsx`. Which UI features break, which keep working, and why?

   <details><summary>Answer</summary>
Break: the create-task form (POST /api/tasks 404s — NextResponse 404 from no matching handler; and the form shows its `failed (404)` msg) and anything curling /api/tasks. Keep working: server-rendered lists — `page.tsx` imports `lib/go.ts` directly, never through the HTTP surface. Proof that pages and route handlers are *parallel* consumers of the lib.
   </details>

5. Why can `loading.tsx` cover a whole segment without knowing which fetch
is pending, but `error.tsx` must be a client component while `loading.tsx` need not be?

   <details><summary>Answer</summary>
loading.tsx is the Suspense fallback Next.js renders while *any* async work in the segment is unresolved — awareness of specifics isn't required. error.tsx carries interactive recovery (`reset()`), and error delivery can interrupt interactivity at any time — so it runs in the browser (client component) and receives `{error, reset}`.
   </details>

## Primary source

[Next.js Docs — Request Lifecycle](https://nextjs.org/docs/app/building-your-application/rendering) (Routing & Rendering section) — the official description of what happens per request in the App Router: where handlers execute, when components render on the server, and how client components hydrate — the ground truth your diagram gets graded against.

## Before next session

Phase 4 begins by reading Go like a consumer. Open `backend-mock/main.go` (it's ~240 lines) and do a 10-minute skim with three questions in mind: which handlers check auth, where do the in-memory slices `projects`/`tasks`/`tokens` live, and what does `writeErr` guarantee about error shape? Bring one question you'd ask the Go team about this API — [Lesson 24](24-reading-your-go-backend.md) starts from exactly that list.
