# Lesson 38 — Final Review & Survival Glossary

*Phase 5 · Session 38 of 38 · ~2 hours · Prereqs: Lesson(s) 01–37 (yes, all of them)*

## Why this matters

This is the course gate: MISSION.md's definition of done, self-examined. Recall under
spacing — not re-reading — is what moves material from "I saw it in a lesson" to "I used
it on Monday." The 25 questions below deliberately mix all five phases so your brain has
to *retrieve*, not *recognize*. After the quiz you'll be handed your permanent kit: the
`reference/` cheat sheets, plus the honest map of which mission gate is proven by which
artifact you already built.

## What you'll be able to do

- You can answer ≥ 20 of 25 mixed-phase questions cold; anything you miss goes on a redo list tied to its lesson.
- You can say, for each of the five MISSION.md items, the exact demo that proves it on your own work.
- You can navigate the five `reference/` cheat sheets without opening a search engine, and know which one answers which class of question.
- You can state, in one paragraph, the full request path: click → React → `/api` route handler → Go → zod → Query cache → re-render.

## Principles

### How to take this quiz (the learning is in the taking)

1. Paper or blank file. All 25, cold, no repo open, no cheat sheets. ~40 minutes.
2. Grade yourself against the `<details>` answers. Score honestly — a padded score cheats
   only the person who has to onboard at your new team.
3. **Every miss gets a redo assignment**: re-read the cited lesson's "Your turn", redo it,
   then re-answer the question the next day (spaced repetition: today → tomorrow → next
   week). ≥ 20/25 with no misses in your weakest phase is the course bar.
4. The questions are ordered easy→hard, *not* by phase — the interleaving is the point;
   knowing *which tool applies* is half of on-the-job recall.

### Your permanent kit: the reference stack

The `reference/` directory is not course material — it's the toolkit you keep forever.
Learn the routing table:

| Question class | Open this |
|---|---|
| "What does this term mean again?" (hydration, tree-shaking, BFF…) | [glossary.md](../reference/glossary.md) |
| "How do I write this type?" (union, generics, `Pick`/`Omit`, narrowing) | [typescript-cheatsheet.md](../reference/typescript-cheatsheet.md) |
| "How does this component idea work?" (state, props, keys, effects) | [react-cheatsheet.md](../reference/react-cheatsheet.md) |
| "Which file goes where / what runs where?" (App Router conventions, route handlers) | [nextjs-cheatsheet.md](../reference/nextjs-cheatsheet.md) |
| "What the hell does this TS error mean?" | [error-decoding.md](../reference/error-decoding.md) |

Habit to install now: stuck > 20 minutes → *glossary/error-decoding first, primary source
second, human third*. That ordering is what "without hand-holding" means.

### The mission, gate by gate — how you prove it

| MISSION.md item | The proof (you already built it) |
|---|---|
| 1. Read any App Router file; know server vs browser | Lesson 23's 10-second drill; any file in `~/teampulse` — cite the evidence (`"use client"`, hooks, `async` fetch) |
| 2. Add a route handler calling Go, zod-validate, reshape | `app/api/tasks/route.ts` + its Lesson 34 test; Lesson 36's feature endpoint |
| 3. Build a React page with state, forms, fetching | P1 explorer + Lesson 36's feature form/page on the production URL |
| 4. Trace a bug: Network tab → route handler → Go | Lesson 32's injected-bug drill; Lesson 36's adversarial curl test |
| 5. Small, reviewable PR into a Next.js/TS repo | Lesson 37's real PR against your capstone + the foreign-repo onboarding drills |

## Follow along

### The 25-question final

Answer all 25 cold, then grade. Each answer cites the lesson to redo on a miss.

**Foundations (Phase 1)**

**1.** `const` with an object: `const tasks = []; tasks.push(t);` — legal? Why or why not?

<details>
<summary>Answer</summary>

Legal. `const` freezes the *binding* (you can't reassign `tasks`), not the object's
contents (Lesson 02–03). React's immutability rule is a *convention on top*: you still
must create new arrays with `map`/spread so re-renders detect changes.

</details>

**2.** Why must a state update be `setTasks([...tasks, newTask])` and never
`tasks.push(newTask); setTasks(tasks)`?

<details>
<summary>Answer</summary>

React compares state by reference; the mutated array is the same object, so React skips
the re-render and the UI lies (Lesson 03, Lesson 13).

</details>

**3.** What does `unknown` give you that `any` doesn't, and where did the course insist you use it?

<details>
<summary>Answer</summary>

`unknown` forces a check before use (compiler refuses operations); `any` disables all
checks. At every boundary: `catch (e)` and `JSON.parse`/Go responses, validated with zod
(Lesson 07, Lesson 25).

</details>

**4.** Rewrite mentally: what does `Pick<Task, "id" | "title">` produce, and when would a route handler use it?

<details>
<summary>Answer</summary>

A type with only `id` and `title` from `Task`. Handlers reshape Go's full response down to
what the UI needs — return the `Pick`ed shape, not the whole record (Lesson 08, Lesson 20).

</details>

**5.** In your P0 CLI you called `GET /api/whoami`… sorry — the mock API's actual auth-echo
endpoint is `GET /api/me` at `http://localhost:8080`. What header must it receive, where
does the value come from, and what does the body contain?

<details>
<summary>Answer</summary>

`Authorization: Bearer <token>`, token from `POST /api/login` (any username, password
`demo`). Body: `{ "id", "name", "role" }` — you always land as `u1`, "Wen", admin
(Lesson 09, Lesson 24, `backend-mock/README.md`).

</details>

**React (Phase 2)**

**6.** "A component is a function of its props." What is a *render*, and why must it be pure?

<details>
<summary>Answer</summary>

A render is calling the function to get the UI for current state+props. Impure renders
(read clocks, mutate outer state) make output unpredictable — React may re-run them any
time (Lesson 11, Lesson 13).

</details>

**7.** Why is `key={index}` dangerous when the list supports deletion?

<details>
<summary>Answer</summary>

After deletion, index-based keys shift onto different items; React matches old DOM to the
"wrong" component — state/checkbox values teleport to the wrong rows. Keys must be stable
identities (Lesson 12).

</details>

**8.** The controlled-input contract, in two sentences.

<details>
<summary>Answer</summary>

The input's `value` comes from state; `onChange` calls the setter, which re-renders with
the new value. State is the single source of truth for what's typed (Lesson 14).

</details>

**9.** Why is fetching in `useEffect` with `[fetchData]` (a function defined in the
component) a footgun, and what does the course's data-fetching stack use instead for
server state?

<details>
<summary>Answer</summary>

The function is a new reference each render → effect dependencies churn (or you lie with
`[]` and stale-close over state). TanStack Query on the client / plain `fetch` in Server
Components replace hand-rolled effects (Lesson 15, Lesson 19, Lesson 26).

</details>

**App Router (Phase 3)**

**10.** You see a `.tsx` file with `export default async function` and a top-level
`await fetch(...)`, no `"use client"`. Server or browser? What's the evidence?

<details>
<summary>Answer</summary>

Server. Evidence: `"use client"` absent (server is the default), `async` component with
top-level await (impossible in a component), and browser code can't be async-default in
that position (Lesson 18–19).

</details>

**11.** Where do route-handler `console.log`s appear, and what does a Network request with
*no* matching server log mean?

<details>
<summary>Answer</summary>

The `pnpm dev` terminal. No log → the request never ran that handler: wrong URL,
middleware short-circuit, or client cache answered it (Lesson 20, Lesson 32).

</details>

**12.** Which file makes a route show a spinner while its Server Component awaits data,
and which renders when its handler throws?

<details>
<summary>Answer</summary>

`loading.tsx` (Suspense boundary) and `error.tsx` (must be a client component); 404 →
`not-found.tsx` (Lesson 21).

</details>

**13.** Your P2 rule: browser JS must never contain what two things, and where do they live instead?

<details>
<summary>Answer</summary>

Secrets (API keys, passwords like `demo`) and session tokens. They live in route handlers
server-side and httpOnly cookies (Lesson 20, Lesson 27).

</details>

**BFF engineering (Phase 4)**

**14.** The three things your route handler must do to *every* Go response before the UI
sees it.

<details>
<summary>Answer</summary>

Validate with zod against the mirrored contract (400/502 your error shape on failure),
reshape to what the UI needs, normalize errors into the single `{ error: string }` shape
(Lesson 25, Lesson 28).

</details>

**15.** Why is `GO_URL=http://localhost:8080` correct locally and *guaranteed* wrong in
prod, and why doesn't CORS apply to that fetch?

<details>
<summary>Answer</summary>

In prod the handler runs on the deploy host where no Go mock listens on its localhost —
prod `GO_URL` targets the real service (Lesson 35). CORS is a *browser* policy; server→
server fetches aren't subject to it (Lesson 29).

</details>

**16.** What does `AbortSignal.timeout(3000)` on the Go `fetch` prevent, and what must the
handler return when it fires?

<details>
<summary>Answer</summary>

An upstream hang tying up request handling forever (the mock's `/api/slow` takes 2s —
imagine production stalling). Fires as an `AbortError` → catch → your normalized 5xx shape
so the UI shows its error state (Lesson 28).

</details>

**17.** TanStack Query: after your POST creates a task, what single call makes every open
list show it, and what does it actually do?

<details>
<summary>Answer</summary>

`queryClient.invalidateQueries({ queryKey: ["tasks"] })` — marks matching cached queries
stale and refetches them; server state's "cache bust" (Lesson 26).

</details>

**18.** The session token model: who holds the token, what does the browser send on each
request, and what does the handler forward to Go?

<details>
<summary>Answer</summary>

The BFF holds the token (httpOnly cookie or server session); the browser sends only the
cookie automatically; the handler re-attaches `Authorization: Bearer <token>` when calling
Go (Lesson 27).

</details>

**Operating (Phase 5)**

**19.** Nine TypeScript errors in one file. First move, and the reason?

<details>
<summary>Answer</summary>

Fix the first one and recompile — cascades come from one bad type; later errors are noise
(Lesson 31).

</details>

**20.** Network tab shows `/api/tasks` → 200 with correct JSON, but the screen is wrong.
Which layer owns the bug?

<details>
<summary>Answer</summary>

UI — good request, good response, bad render (component misreading data; check props in
React DevTools) (Lesson 32).

</details>

**21.** Which two facts does the Components tab of React DevTools give you about a
selected component, and what does a whole-page flash under "Highlight updates" mean?

<details>
<summary>Answer</summary>

Its current props and state. Whole-page flash = a re-render storm — state is too high in
the tree; push it down (Lesson 32, Lesson 33).

</details>

**22.** Why is a `NEXT_PUBLIC_` var "a build artifact, not configuration", and what does
that imply about secrecy and about changing it?

<details>
<summary>Answer</summary>

It's pasted into the JS bundle at build time: public to anyone, and immutable until the
next build (Lesson 33, Lesson 35).

</details>

**23.** Your route-handler test hits the real mock Go API. Give one thing this catches
that a mocked-`fetch` test misses, and one risk it introduces.

<details>
<summary>Answer</summary>

Catches real contract drift/connectivity (Go up/down, actual field shapes). Risk: flakiness
from shared state — the mock's in-memory data resets on restart, so never depend on
cross-run data (Lesson 34).

</details>

**24.** A teammate's PR adds a new `/api/weekly` endpoint with zero tests. Using the
course's minimum, what should the review ask for — name the test kinds and the two
statuses worth pinning?

<details>
<summary>Answer</summary>

One route-handler test hitting the real backend: happy path `200` + reshaped body shape,
and one negative (`401` without session or the upstream-500 path). A unit test if pure
aggregation logic was added (Lesson 34, Lesson 36).

</details>

**25.** The final synthesis, one paragraph with no omissions: a user clicks "New task",
the form submits, and the new task appears. Name every station the data passes through,
from `onClick` to re-render.

<details>
<summary>Answer</summary>

onClick → form submit handler (client component) → TanStack Query `useMutation` →
`onMutate` optimistic insert + re-render → `POST /api/tasks` (Next.js **route handler**,
server) → zod-validates the inbound body → reads session cookie → `fetch` to
`http://localhost:8080/api/tasks` with `Authorization: Bearer` (from the session, not the
client) → Go returns 201 + created task → handler zod-validates the *response*, reshapes,
normalizes errors → returns JSON → `onSuccess` reconciles/invalidates the `["tasks"]`
query → Query refetches/replaces cache → task list re-renders with the server truth
(Lessons 14, 20, 25–28 — the Lesson 23 full-request-path diagram, now with blood in it).

</details>

## Your turn

1. Score the 25. If any phase scored ≤ 1/5, redo that phase's final project (P0–P4) — not by rereading, by rebuilding the one thing it consolidated. If total < 20, schedule the misses for tomorrow/tomorrow+7days and re-test before moving on.
2. Write your **Day-1 README** for TeamPulse (half a page): what it is, where each piece runs in prod (Lesson 35's honesty statement), how to run it (`go run main.go`, `pnpm dev`, `pnpm test`), and the demo login (username: anything, password: `demo`). This is the README you'd hand the team — it doubles as Lesson 37's artifact.
3. Book the next milestone: pick the real team repo you'll onboard, block 90 minutes this week, and run Lesson 37's package.json → routes → one-slice drill on it.

<details>
<summary>Solution</summary>

No single answer — but item 1 is only "done" when you can answer every previously-missed
question *the day after* re-studying it, unprompted. That's what spaced repetition buys.

</details>

## Checkpoint

1. Route a question to the right cheat sheet, without looking: "What's the difference between `Omit` and `Pick`?" and "Which files can only be client components?"

<details>
<summary>Answer</summary>

`typescript-cheatsheet.md` (utility types) and `nextjs-cheatsheet.md` (file conventions —
`error.tsx`/`loading.tsx` must be client, etc.).

</details>

2. Name the two `reference/` files you open *before* Googling a stuck error, and the rule that triggers them.

<details>
<summary>Answer</summary>

`error-decoding.md` (and `glossary.md` if the error's vocabulary is unfamiliar) — after
~20 minutes stuck on one error, per README's rules of engagement.

</details>

3. MISSION item #4 (trace a bug end-to-end): name the three tools that make the trace, in the order you'd use them.

<details>
<summary>Answer</summary>

Network tab (UI vs BFF), server runtime logs / `pnpm dev` terminal (BFF vs Go), direct
curl to `http://localhost:8080` (proves the Go layer's innocence or guilt) — Lesson 32's
triage tree.

</details>

4. Which course artifact is your *permanent* proof for mission item #5 (small reviewable PR), and why does it count?

<details>
<summary>Answer</summary>

The Lesson 37 PR against your capstone (spec commit → story commits → tests →
preview link), plus the foreign-repo onboarding drills — it exercises the full loop on a
repo you didn't write, which is the mission's wording.

</details>

5. What is the one sentence you'd tell a colleague starting this course tomorrow about *how* to study it?

<details>
<summary>Answer</summary>

Anything honest that includes: type every line yourself, never skip checkpoints, and treat
project sessions as the course (the lessons are the syllabus; P0–P4 are the education) —
the two SYLLABUS ground rules, internalized.

</details>

## Primary source

[react.dev — Learn React](https://react.dev/learn) — the permanent bookmark for a
graduate: the only doc that explains the render model you were trained on, in the order
the course teaches it, and the place every "wait, how does state work again?" resolves
correctly.

## Before next session

There is no next session — this is the send-off. Do three things in the next 24 hours:
fix every quiz miss's redo, finish the Day-1 README, and put the 90-minute real-repo
onboarding in your calendar. Then go be the person who owns the seam between TS and Go.
