# Syllabus — 38 sessions × 2 hours (~7 weeks at 6 days/week)

Every session is one file in `lessons/`. Sessions marked **[P]** are project
sessions that extend the running "TeamPulse" app — never skip them, they are
where learning actually happens.

Two ground rules:
1. **Type everything yourself.** Do not copy-paste code from the lesson without
   typing it at least once. Typing is where syntax becomes memory.
2. **The checkpoint questions are the contract.** If you can't answer them from
   memory, redo the "Your turn" section before moving on.

---

## Phase 1 — JavaScript & TypeScript foundations (Sessions 1–9)
*Goal: read and write modern TS comfortably at the level a BFF needs. No UI yet.*

| # | Lesson | Core principle you must own |
|---|---|---|
| 01 | Toolchain: Node, pnpm, your first scripts | What `package.json` is; how JS runs outside a browser |
| 02 | Modern JS syntax | `let`/`const`, arrow functions, template literals, destructuring, spread/rest |
| 03 | Arrays, objects, immutability | `map`/`filter`/`reduce`; why you never mutate — the root of every React rule |
| 04 | Modules & the bundler mental model | `import`/`export`; what a bundler actually does to your files |
| 05 | Async: Promises, async/await, fetch | The event loop at working level; why `fetch` returns a Promise; `try/catch` |
| 06 | TypeScript I: types & inference | Types are erased at build time; inference does 90% of your typing |
| 07 | TypeScript II: shapes & narrowing | `interface`/`type`, optionals, unions, discriminated unions, `unknown` not `any` |
| 08 | TypeScript III: generics & utility types | Reading `Promise<T>`, `Record<K,V>`; `Partial`/`Pick`/`Omit`; `import type` |
| 09 | **[P0] TeamPulse CLI** | Consolidation: a TS command-line client for the mock Go API (login → list projects/tasks) |

**Phase gate:** you can add a second call to the authenticated `GET /api/me`
endpoint in the CLI, type its
response properly, and explain every line, in under 15 minutes.

## Phase 2 — React core (Sessions 10–16)
*Goal: the React mental model, learned inside a real Next.js app from day one.*

| # | Lesson | Core principle you must own |
|---|---|---|
| 10 | Your first Next.js app & JSX | JSX is a function call; the anatomy of `create-next-app` output |
| 11 | Components & props | A component is a function of its props; composition over configuration |
| 12 | Conditional rendering & lists | `&&`/ternaries, `.map()` + `key`; why keys must be stable |
| 13 | State: `useState` & re-renders | Render is a pure function of state+props; state updates are async; immutability again |
| 14 | Forms & controlled inputs | The input's value IS state; `onChange` re-renders it back |
| 15 | Data fetching & the truth about `useEffect` | Most `useEffect` is an anti-pattern; client-side `fetch` with proper loading/error state |
| 16 | **[P1] TeamPulse explorer page** | Consolidation: a client-side React page — login form, project list, task list, task filter |

**Phase gate:** you can explain to a rubber duck why clicking a button updates
the screen, in terms of state → re-render → render function.

## Phase 3 — The App Router (Sessions 17–23)
*Goal: know what runs where (server vs browser) — the single most important Next.js skill.*

| # | Lesson | Core principle you must own |
|---|---|---|
| 17 | File-system routing | Folders are URLs: `page`/`layout`/`link`, dynamic `[id]` segments |
| 18 | Server vs Client Components | `"use client"` marks a boundary; server components never ship to the browser |
| 19 | Fetching in Server Components | `fetch` directly in the component; caching & `revalidate` in one page |
| 20 | Route Handlers: your first BFF | `app/api/*/route.ts` is a server function; secrets live here, never in the browser |
| 21 | Loading, error & not-found states | `loading.tsx`/`error.tsx`/`not-found.tsx`; Suspense at working level |
| 22 | **[P2] TeamPulse BFF** | Consolidation: all Go calls go through route handlers; server-rendered pages; API key stays server-side |
| 23 | Checkpoint: the full request path | Draw/trace one click from browser → route handler → Go → response → re-render |

**Phase gate:** given any `.tsx` file in the repo, you state within 10 seconds
whether it runs on the server, the browser, or both — and how you know.

## Phase 4 — BFF engineering against a Go backend (Sessions 24–31)
*Goal: the part of the job that is actually yours — the seam between TS and Go.*

| # | Lesson | Core principle you must own |
|---|---|---|
| 24 | Reading your Go backend | The API is a contract: read `backend-mock/main.go` like a consumer; status-code semantics |
| 25 | Typed contracts with zod | Mirror Go structs in TS; validate at the BFF boundary — trust nothing from upstream |
| 26 | TanStack Query on the client | Server state ≠ client state; caching, `isLoading`/`isError`, invalidation |
| 27 | Auth: tokens & cookies through the BFF | The BFF holds the session; forward `Authorization` headers; never leak tokens to JS |
| 28 | Resilience: timeouts, retries, error shapes | `AbortSignal.timeout`, retry policy, normalising Go's `{"error":...}` into one UI shape |
| 29 | Middleware & configuration | `middleware.ts`, `next.config.ts`, env vars per environment, CORS in one page |
| 30 | **[P3] Full-stack TeamPulse** | Consolidation: typed BFF + zod + TanStack Query + auth + resilience, against the mock Go API |

**Phase gate:** you can add a new Go endpoint (in `main.go`) and surface it in
the UI — typed, validated, cached — in under 45 minutes.

## Phase 5 — Operating & shipping (Sessions 31–38)
*Goal: day-1-on-the-team skills: debug, decode, deploy, contribute.*

| # | Lesson | Core principle you must own |
|---|---|---|
| 31 | Decoding TypeScript & ESLint errors | Compiler errors are honest; read the first line, ignore the rest; common TS jargon |
| 32 | Debugging: DevTools, React DevTools, server logs | Browser Network tab tells you UI-vs-BFF; server logs tell you BFF-vs-Go; bisect fast |
| 33 | Performance at working level | Bundle size is the cost of `"use client"`; `next/image`, `next/font`; five common footguns |
| 34 | Minimal testing that teams actually use | One vitest unit test, one route-handler test, one component test — enough to extend theirs |
| 35 | Deploying to Vercel | Build vs runtime env vars; preview deploys; why it broke in prod but not dev |
| 36 | **[P4] Capstone: ship TeamPulse** | Consolidation: polish P3, add one feature of your choice, deploy publicly |
| 37 | The team-codebase playbook | Onboarding a foreign repo; the smallest-reviewable-PR workflow; the 10 questions to ask your team on day one |
| 38 | Final review & survival glossary | Spaced-repetition quiz across all 5 phases; your personal cheat-sheet stack |

**Course gate (= MISSION definition of done):** trace a bug end-to-end and ship
a small PR to a foreign Next.js repo.
