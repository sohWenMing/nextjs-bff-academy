# Lesson 21 — Loading, Error & Not-Found States

*Phase 3 · Session 21 of 38 · ~2 hours · Prereqs: Lesson(s) 17, 18, 19, 20*

## Why this matters

In [Lesson 19](19-fetching-in-server-components.md) a throwing server-side fetch gave you a white screen; in [Lesson 20](20-route-handlers-your-first-bff.md) the BFF returned a bare 502. Production UIs need the other half: while Go is thinking, show something; when it fails, show something *you* designed; when the id doesn't exist, show a real 404 — not Next.js's default one. These are **file conventions**: drop `loading.tsx` / `error.tsx` / `not-found.tsx` into a folder and Next.js wires them up. Your team's app has dozens of these files; recognizing them at a glance is mission item #1 again.

## What you'll be able to do

- You can add `loading.tsx`, `error.tsx`, and `not-found.tsx` to any route segment and state what each catches.
- You can explain what a Suspense boundary is at working level and why `loading.tsx` is "Suspense you get for free".
- You can write the required `error.tsx` signature — a client component receiving `{ error, reset }` — from memory.
- You can trigger the app's own 404 page from server code with `notFound()`.
- You can reason about which boundary catches an error thrown two folders deep.

## Principles

### The three convention files, one line each

Inside any route segment folder (any folder with or destined for a `page.tsx`):

| File | Catches | Runs |
|---|---|---|
| `loading.tsx` | while the segment's async work is pending | server |
| `error.tsx` | any error thrown in the segment below | **client** |
| `not-found.tsx` | `notFound()` thrown below, or unmatched URL | server |

They compose with the folder = URL rule from [Lesson 17](17-file-system-routing.md): a convention file covers that segment *and everything nested inside it*, and the nearest one up the tree wins.

### `loading.tsx` is a Suspense boundary you don't write by hand

**Suspense** is React's mechanism for "render a fallback while async work below is pending" (async components + the `Suspense` component are React 19 primitives; you get the wiring free via files). Before it, [Lesson 15](15-data-fetching-and-useeffect.md) forced you to hold a `loading` boolean in state; server components *can't* do that — they have no state. So Next.js says: any `await` in a server component that suspends (is still pending) renders the segment's `loading.tsx` first, then swaps in the real UI when data lands — streamed, no client-side spinner code from you:

```tsx
// app/pulse/loading.tsx
export default function Loading() {
  return <p>Loading projects…</p>;
}
```

That's the whole feature. Two sharp edges:

- The fallback swaps when **the page's data** resolves. A page with two
awaits pays for the slower one unless you wrap the faster parts in your own `<Suspense>` (working-level knowledge; Lesson 33 sharpens it).
- `loading.tsx` only *shows* the fallback — it doesn't make anything faster.
Your friend for that is the 2-second `/api/slow` endpoint, used below.

### `error.tsx` — the one that must be a client component

Errors can arrive at any moment, and recovery ("try again") is interactivity — so `error.tsx` is always a **client component** ([Lesson 18](18-server-vs-client-components.md)). Next.js gives it exactly two props, and the signature is worth memorizing:

```tsx
// app/pulse/error.tsx
"use client";

export default function PulseError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <main>
      <h1>Something broke</h1>
      <p>{error.message}</p>
      <button onClick={() => reset()}>Try again</button>
    </main>
  );
}
```

- `error` is the thrown value. `digest` is a server-side error ID — when the
error came from server rendering, the browser gets only the digest, not the server's message/stack (production security). The full server error prints in your terminal.
- `reset()` re-renders the segment — clears the error and retries. It re-runs
the server render, which re-does the fetch: for a flaky Go endpoint, this *is* the retry button.

Boundary rule: an error propagates up to the **nearest `error.tsx` ancestor**. No `error.tsx` under `app/projects/`? `app/error.tsx` catches it. None at all? Next.js's built-in error screen (dev) / blank-ish 500 (prod). This is why teams put a generic `error.tsx` at the root and specific ones next to data-heavy pages.

### `not-found.tsx` and `notFound()`

Two ways to land on a segment's `not-found.tsx`: a URL that matches no route, or server code that *calls* `notFound()` (from `next/navigation`) — a special throw that Next.js recognizes and routes to the nearest not-found boundary instead of the error boundary:

```tsx
import { notFound } from "next/navigation";

export default async function ProjectPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const res = await fetch(`http://localhost:8080/api/projects/${id}`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (res.status === 404) notFound(); // → not-found.tsx, not error.tsx
  if (!res.ok) throw new Error(`Go said ${res.status}`); // → error.tsx
  const project = await res.json();
  return <h1>{project.name}</h1>;
}
```

Note the two-line distinction that is the whole design lesson: **404 from Go is "this page doesn't exist" (not-found UI); 500/502/network is "something broke" (error UI)**. Mapping upstream status codes to the right convention file is BFF-side judgment you'll do for real on the team.

### Boundary granularity, once

For a folder tree:

```text
app/
  error.tsx            ← catch-all (last resort)
  not-found.tsx        ← default 404 for the whole app
  projects/
    loading.tsx        ← fallback while any /projects/** page awaits
    [id]/
      page.tsx
      error.tsx        ← catches errors ONLY under /projects/[id]
```

An error in `[id]/page.tsx` hits `[id]/error.tsx` (nearest). An error in the *layout* above it isn't caught by it — a segment's `error.tsx` guards its page and children, not its parent layout. When in doubt: put the boundary in the same folder as the failure you're handling.

## Follow along

`pnpm dev` in `~/teampulse`; Go mock on `:8080`.

1. **Slow endpoint, visible loader.** Create `app/slow/page.tsx`:

   ```tsx
   export default async function SlowPage() {
     const res = await fetch("http://localhost:8080/api/slow", {
       cache: "no-store",
     });
     const data = await res.json();
     return (
       <main>
         <h1>Slow demo</h1>
         <p>{data.note}</p>
       </main>
     );
   }
   ```

Visit `/slow`. Expected: **two blank seconds**, then content — a blank screen is your current "loading state". Note `/api/slow` needs no token (`backend-mock/main.go` doesn't call `requireAuth` there).

2. **Fix it with a file.** Create `app/slow/loading.tsx`:

   ```tsx
   export default function SlowLoading() {
     return <main><p>Waiting for Go (2s)...</p></main>;
   }
   ```

Reload `/slow`. Expected: the fallback appears **immediately**, swaps to content at ~2s. You never touched state, effects, or timers.

3. **Break it and catch the crash.** Stop the Go mock (Ctrl-C in its tab).
Reload `/slow`. Expected: error overlay (dev) or the built-in 500 screen — the `fetch` throws (connection refused) with no `error.tsx` in sight.

4. **Add the boundary.** Create `app/slow/error.tsx`:

   ```tsx
   "use client";

   export default function SlowError({
     error,
     reset,
   }: {
     error: Error & { digest?: string };
     reset: () => void;
   }) {
     return (
       <main>
         <h1>Couldn't reach Go</h1>
         <p>{error.message || error.digest}</p>
         <button onClick={() => reset()}>Try again</button>
       </main>
     );
   }
   ```

Reload `/slow`: your UI, not the overlay. Note `error.digest` in the browser console's error log and the *real* message (`fetch failed`) in the dev-server terminal — server error details stay server-side in prod. Restart the Go mock, click **Try again**. Expected: content renders — `reset()` re-ran the server fetch. (If it keeps showing the error after Go is back, hard-reload once; the dev server can cache the failed render.)

5. **Not-found from upstream status.** Recreate the dynamic project page from
Lesson 17/19 as `app/projects/[id]/page.tsx` with the two-line distinction from Principles (404 → `notFound()`, else → `throw`), using the login-then-fetch pattern from [Lesson 19](19-fetching-in-server-components.md) (any username, password `demo`). Create `app/projects/[id]/not-found.tsx`:

   ```tsx
   import Link from "next/link";

   export default function ProjectNotFound() {
     return (
       <main>
         <h1>No such project</h1>
         <Link href="/projects">Back to projects</Link>
       </main>
     );
   }
   ```

Visit `/projects/p1` → real project. Visit `/projects/p9` → your not-found UI (Go returned 404; `notFound()` fired). Visit `/nonsense-url` → Next.js's *built-in* 404 (no custom `not-found.tsx` at root yet). Add `app/not-found.tsx` of your own and revisit `/nonsense-url` — now your component answers unmatched URLs too.

## Your turn

1. Give the whole app a catch-all error boundary: create `app/error.tsx`
("All is lost" + reset) and confirm that with Go down, a segment *without* its own `error.tsx` (e.g. your Lesson 19 `/pulse` page) now shows it.

   <details><summary>Solution</summary>

   ```tsx
   // app/error.tsx
   "use client";

   export default function AppError({
     error,
     reset,
   }: {
     error: Error & { digest?: string };
     reset: () => void;
   }) {
     return (
       <main>
         <h1>All is lost (recoverably)</h1>
         <p>{error.digest}</p>
         <button onClick={() => reset()}>Retry</button>
       </main>
     );
   }
   ```

With Go stopped, `/pulse` shows this instead of the overlay. `/slow` still shows its own `SlowError` — nearest boundary wins.
   </details>

2. Use the flaky endpoint: `/api/flaky` fails ~30% of the time with 500
(`main.go`'s `handleFlaky`, and it requires auth). Build `app/flaky/page.tsx` (login → fetch flaky → render `note`; `throw` on non-ok) plus `app/flaky/loading.tsx` and `app/flaky/error.tsx` with a working retry. Refresh repeatedly: roughly 3 in 10 loads should land on your error UI, and retry should usually recover.

   <details><summary>Answer</summary>
Any implementation that (a) logs in server-side with username `wen` / password `demo`, (b) `throw`s on `!res.ok`, (c) recovers via `reset()`, counts. The load-bearing insight: you cannot "fix" flakiness in the UI — retry is a *policy* (Lesson 28 adds timeouts/retries at the BFF so users see fewer of these screens).
   </details>

3. Harder: in `app/projects/[id]/page.tsx` add
`if (id === "p3") redirect("/projects")` (from `next/navigation`) before the fetch. What does `redirect()` throw have in common with `notFound()`, and why do neither show up as errors in your terminal?

   <details><summary>Answer</summary>
Both throw special Next.js control-flow values, not real failures — Next.js catches them during rendering and performs the navigation/404. They're "signals, not exceptions"; only genuine throws reach `error.tsx` and the logs. (Aside: `p3` is the archived Legacy Portal — gating archived projects away is exactly the kind of rule that belongs in server code where users can't bypass it.)
   </details>

## Checkpoint

1. Which of the three convention files must be a client component, and why?

   <details><summary>Answer</summary>
`error.tsx` — it needs interactivity (`reset()` and usually a retry button) and must survive even when server rendering failed. `loading.tsx` and `not-found.tsx` run on the server.
   </details>

2. An error is thrown in `app/dashboard/stats/page.tsx`. Boundaries exist at
`app/dashboard/error.tsx` and `app/error.tsx`. Which renders? What if you *also* create `app/dashboard/stats/error.tsx`?

   <details><summary>Answer</summary>
`app/dashboard/error.tsx` (nearest ancestor). Add the stats-level one and it wins for errors under that segment — nearest boundary catches first.
   </details>

3. What makes `loading.tsx` appear, given you never reference it from your
page?

   <details><summary>Answer</summary>
The page's async work suspends (pending `await`/fetch); Next.js renders the segment's `loading.tsx` as the Suspense fallback and streams in the real UI when it resolves. It's automatic — a file-convention, not a prop.
   </details>

4. TeamPulse: Go returns 404 for `/api/projects/p9`, 500 (down) for the whole
mock, and ~30%-often 500 for `/api/flaky`. Which convention file should each route's *page* end up in, and what server code produces that?

   <details><summary>Answer</summary>
404 → `not-found.tsx` via `if (res.status === 404) notFound()`. Down → `error.tsx` via `throw` (or the fetch rejecting with `fetch failed`). Flaky 500 → also `error.tsx` via `throw` on `!res.ok`, with `reset()` as the user-facing retry.
   </details>

5. Your `error.tsx` shows `error.digest` but users report "it says nothing
useful" while your *terminal* shows `fetch failed`. Explain the mismatch.

   <details><summary>Answer</summary>
Server-thrown errors are replaced in the browser by a digest reference — full details stay server-side to avoid leaking internals. Log digests to your server logs and write user-facing copy yourself instead of dumping `error.message`.
   </details>

## Primary source

[Next.js Docs — Loading, Error and not-found](https://nextjs.org/docs/app/api-reference/file-conventions) — the file-convention reference defining each file's contract (including `error.tsx`'s client requirement and `reset()` semantics).

## Before next session

Sketch (paper or any drawing app) the file tree you *think* `~/teampulse` will have after the next lesson: every `page.tsx`, `loading.tsx`, `error.tsx`, `not-found.tsx`, and `route.ts` you expect, with URLs beside them. Don't look at the app while doing it — you built the pieces in Lessons 17–20. Lesson 22 is the assembly; bring your sketch to compare.
