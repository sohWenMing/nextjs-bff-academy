# Lesson 19 — Fetching in Server Components

*Phase 3 · Session 19 of 38 · ~2 hours · Prereqs: Lesson(s) 05, 15, 17, 18*

## Why this matters

In Phase 2 you fought `useEffect` to fetch data in the browser. The App Router deletes most of that fight: a server component is an `async` function, so you `await fetch(...)` *directly in the component* and render the result — no effect, no loading-state dance, no token in the browser. Your team's pages get their data this way, and understanding caching (`revalidate`) here is the difference between "the dashboard shows stale tasks" being a mystery and being a one-line answer.

## What you'll be able to do

- You can write a server component that `await`s a `fetch` to the Go API and renders the JSON — no `useEffect`, no `useState`.
- You can explain, from memory, why the loading/error machinery of [Lesson 15](15-data-fetching-and-useeffect.md) mostly disappears here.
- You can state Next.js 15's default `fetch` caching behavior and change it with `cache` and `next: { revalidate }`.
- You can fetch the same URL twice in one render and explain why only one network request happens.
- You can log server-side and show that the Go token never reaches the browser.

## Principles

### A server component can just... be async

[Lesson 18](18-server-vs-client-components.md) established that server components run on the Node process. Node can do `await` ([Lesson 05](05-async-promises-await-fetch.md)) and can call any HTTP endpoint. So the "data fetching pattern" in the App Router is embarrassingly simple:

```tsx
// app/pulse/page.tsx — no "use client" anywhere
type Project = {
  id: string;
  name: string;
  description: string;
  status: "active" | "archived";
};

async function getProjects(): Promise<Project[]> {
  const login = await fetch("http://localhost:8080/api/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username: "wen", password: "demo" }),
    cache: "no-store",
  });
  const { token } = (await login.json()) as { token: string };

  const res = await fetch("http://localhost:8080/api/projects", {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`Go API said ${res.status}`);
  }
  return res.json();
}

export default async function PulsePage() {
  const projects = await getProjects();
  return (
    <main>
      <h1>Pulse</h1>
      <ul>
        {projects.map((p) => (
          <li key={p.id}>
            {p.name} — {p.status}
          </li>
        ))}
      </ul>
    </main>
  );
}
```

That's the whole page. Compare with the `useEffect` + `useState` + loading boolean + error string you wrote in [Lesson 15](15-data-fetching-and-useeffect.md): here there is no "before the data arrives" state at all, because **the component doesn't exist in the browser until the data is in its hands.** The server awaits, then sends finished HTML. (Loading and error UI do come back — as file conventions in Lesson 21, not as state in your component.)

Note the word `demo` appears in this file — which is safe *only because this file never ships to the browser*. Hold that thought for Lesson 20; it's where the security story gets formal.

### The fetch happens over there, and that's the feature

Three consequences of the fetch running on the server:

1. **No CORS.** [Lesson 15](15-data-fetching-and-useeffect.md)'s
browser-to-`:8080` requests needed the Go server to bless the origin. Server-to-server HTTP has no such rule — CORS is a browser policy.
2. **Credentials can appear in the code** because the code stays on the
server. The Bearer token from `/api/login` exists only inside the Node process and the response HTML contains its *result*, never the token.
3. **One round trip to your users.** The browser fetches one HTML document
from `:3000`; the chatty login-then-projects exchange with `:8080` happens on the machine next to (logically) the database.

### Caching: Next.js wraps `fetch` — and 15 changed the default

Next.js replaces the global `fetch` with one that has a cache in front of it (and dedupes identical in-render calls — see below). Two options matter:

```tsx
// 1. Never cache — the default in Next.js 15. Fresh on every request.
await fetch(url, { cache: "no-store" });

// 2. Cache, but re-check upstream every N seconds.
await fetch(url, { next: { revalidate: 60 } });
```

The vocabulary: **revalidation** = throwing away a cached response after a time window so the next visitor triggers a fresh fetch. In Next 14 the default was "cache forever" (a famous footgun); **Next.js 15 flipped the default to `no-store`** (uncached). If you read a tutorial where data went stale after deploy, it was pre-15.

For TeamPulse right now: task data changes when someone creates a task, so `no-store` is the honest default. When a read is genuinely slow-changing (project descriptions), `next: { revalidate: 3600 }` is how you say "an hour of staleness is fine, a full day is not."

### Request memoization: same URL, one network call

If two components in the same render fetch the *same URL with the same options*, Next.js makes **one** network request and shares the result — you saw the trick in [Lesson 15](15-data-fetching-and-useeffect.md) with a hand-rolled `Map` of promises; Next.js does it for you, keyed on (URL + options). This is why you can call `getProjects()` in a page *and* in a layout without doubling your traffic to Go. It is per-request, not global — it's not a cache that outlives the render.

### What this is *not* yet

- This is not your final architecture: by Lesson 22, browser pages won't be
the ones holding Go URLs and tokens — route handlers will. Today's pattern (direct fetch in server components) is still legitimate and still used for first-paint data; tomorrow you learn the layer that wraps it.
- This is not client-side navigation data-fetching. When the user clicks and
a *client* component needs fresh data later, that's TanStack Query's job (Lesson 26).
- Error handling here is `throw`, and Lesson 21 gives you the file
conventions that catch it.

## Follow along

`pnpm dev` in `~/teampulse`; Go mock running on `:8080`.

1. **Build the page from Principles.** Create `app/pulse/page.tsx` with the
full `PulsePage` code above (type it). Visit `http://localhost:3000/pulse`. Expected: `Billing Service — active`, `Mobile App — active`, `Legacy Portal — archived`. In your **terminal** (not browser), watch for the dev server's compile line; in DevTools → Network, the only XHR-ish traffic for the page is to `:3000` — **zero requests to `:8080` appear in the browser Network tab**. That absence is the lesson.

2. **Prove the token stayed server-side.** DevTools → Network → the `pulse`
document → Response. Search (`Cmd+F` in the response body) for `tok-`. Expected: no match. The token existed in Node memory and never crossed the wire to the browser.

3. **Make it fail visibly.** Change the projects URL to
`http://localhost:8080/api/projects-nope`. Reload `/pulse`. Expected: a red error overlay naming your `throw new Error("Go API said 404")` — or an HTML-error `SyntaxError` from `.json()`. Either way: an `async` server component that throws brings down *the render*, and the dev overlay shows the server stack trace. Revert the URL.

4. **See memoization.** Above `export default async function PulsePage`, add
a second component and use it in the page:

   ```tsx
   async function ProjectCount() {
     const projects = await getProjects();
     return <p>{projects.length} projects</p>;
   }
   ```

Render `<ProjectCount />` under the `<h1>`. Now count log lines in the terminal after one reload — add this inside `getProjects`, before the login fetch:

   ```tsx
   console.log("hitting Go for projects");
   ```

Expected: it prints **once** per request even though two components called `getProjects()` — memoized identical fetches. (The login fetch also runs once, for the same reason.)

5. **Feel revalidate.** First restart the Go mock so its data is pristine
(`Ctrl-C` then `go run main.go` in `backend-mock/`). Change the projects fetch to `{ next: { revalidate: 10 } }` and reload `/pulse` once. Create a task via curl so the dataset changes (any project works):

   ```bash
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \ -H 'Content-Type: application/json' \ -d '{"username":"wen","password":"demo"}' | sed 's/.*"token":"[^"]*".*/\1/') curl -s -X POST http://localhost:8080/api/tasks \ -H "Authorization: Bearer $TOKEN" \ -H 'Content-Type: application/json' \ -d '{"projectId":"p1","title":"Revalidate test","priority":1}'
```

   (Projects themselves don't change, so instead point `getProjects`' second
   fetch at `/api/tasks` with the same filter to see a visible change: swap
   the URL to `http://localhost:8080/api/tasks?projectId=p1`, render
   `t.title`s, keep `revalidate: 10`.) Reload immediately: the *new* task may
   not appear (cached within the 10s window). Wait 11 seconds, reload: it
   appears. That gap is a revalidation window you can *see*.

## Your turn

1. Convert your P1 task list ([Lesson 16](16-p1-teampulse-explorer-page.md)) into
   a *server* page at `app/pulse/tasks/page.tsx` that fetches
   `/api/tasks?projectId=p1` and renders `title`, `status`, `assignee`. No
   hooks allowed in the file.

   <details><summary>Solution</summary>

   ```tsx
// app/pulse/tasks/page.tsx type Task = { id: string; projectId: string; title: string; status: "todo" | "doing" | "done"; priority: number; assignee: string; };

async function getTasks(): Promise<Task[]> { const login = await fetch("http://localhost:8080/api/login", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ username: "wen", password: "demo" }), cache: "no-store", }); const { token } = (await login.json()) as { token: string }; const res = await fetch("http://localhost:8080/api/tasks?projectId=p1", { headers: { Authorization: `Bearer ${token}` }, cache: "no-store", }); if (!res.ok) throw new Error(`Go API said ${res.status}`); return res.json(); }

export default async function TasksPage() { const tasks = await getTasks(); return ( <main> <h1>p1 tasks</h1> <ul> {tasks.map((t) => ( <li key={t.id}> {t.title} — {t.status} ({t.assignee}) </li> ))} </ul> </main> ); }
   ```

   Expected: three p1 tasks (`t1`, `t2`, `t5`) plus any you created in
   Follow-along step 5. The old client explorer still works; both patterns can
   coexist until Lesson 22 consolidates.
   </details>

2. Your `PulsePage` and `TasksPage` both log in separately. Is that one login
   per page load or two? Then: with `revalidate: 3600` on the *login* fetch,
   what bug would you create 61 minutes after the Go mock restarts?

   <details><summary>Answer</summary>
   One per page (memoization is per-request; visiting two pages in sequence
   logs in twice). Caching login for an hour breaks when the Go mock restarts:
   its token map is in-memory, so the cached token becomes invalid — every
   `/api/projects` call then 401s ("missing or invalid token") for up to an
   hour while the UI keeps using a dead cached token. Slow-changing *data* is
   a safe revalidate target; *credentials* are not.
   </details>

3. Harder: render the health-check timestamp. Server-fetch
   `http://localhost:8080/api/health` (no auth needed — see
   `backend-mock/main.go`) in a small server component and print `time`. Then
   explain why refreshing shows a newer timestamp than a page that fetched
   health two minutes ago, and which single option guarantees that behavior.

   <details><summary>Answer</summary>
   `{ cache: "no-store" }` (also the Next 15 default). Each request re-runs
   the component and re-fetches, so `time` (which the Go handler stamps with
   `time.Now().UTC()`) is always fresh. Without it — say with `force-cache` —
   the first response would be reused indefinitely and the timestamp would
   freeze at first fetch.
   </details>

## Checkpoint

1. Write from memory the skeleton of a server component that fetches
   `http://localhost:8080/api/me` (assume you already have a `token` string)
   and renders `name`.

   <details><summary>Answer</summary>

   ```tsx
export default async function MePage() { const res = await fetch("http://localhost:8080/api/me", { headers: { Authorization: `Bearer ${token}` }, cache: "no-store", }); if (!res.ok) throw new Error(`Go API said ${res.status}`); const me = await res.json(); return <h1>{me.name}</h1>; }
   ```
   </details>

2. Why does this pattern need no `loading` state, and when does it need
   `useState` again?

   <details><summary>Answer</summary>
   The component only renders after the await completes — the browser never
   sees a no-data version. You need `useState` when *the browser* must react
   after load (typing, clicking), i.e., in client components below
   `"use client"`.
   </details>

3. In Next.js 15, does `await fetch(url)` in a server component hit the
   network on every request by default, and what two option shapes change
   that?

   <details><summary>Answer</summary>
   Yes — the 15 default is uncached (`no-store` behavior).
   `cache: "force-cache"` makes it cache until revalidated;
   `next: { revalidate: N }` caches with an N-second freshness window.
   </details>

4. TeamPulse: two sibling server components both call
   `getProjects()` with `cache: "no-store"`. How many HTTP requests to
   `:8080` fire in one page render, and what mechanism is responsible?

   <details><summary>Answer</summary>
   One. Request memoization dedupes identical fetches (same URL + same
   options) within a single render pass — per-request, not a shared cache.
   (`no-store` disables the *cache*, not *memoization*.)
   </details>

5. A teammate reports: "the projects list updated in Go but the page shows
   old data for ten minutes." Name the likely culprit and the one-line fix.

   <details><summary>Answer</summary>
   A `next: { revalidate: 600 }` (or `force-cache`) on that fetch — the
   server is serving a still-fresh cached response. Fix: drop to `no-store`
   (or shorten the revalidate window to match how stale is acceptable).
   </details>

## Primary source

[Next.js Docs — Fetching, caching, and revalidating](https://nextjs.org/docs/app/building-your-application/caching)
— the official reference for the `fetch` wrapper's `cache` / `next.revalidate`
options and the Next 15 default flip this lesson depends on.

## Before next session

Create the one-line file `app/api/health/route.ts`:

```ts
export async function GET() { return Response.json({ ok: true, where: "next.js" }); }
```

Visit `http://localhost:3000/api/health` and note what you see. Bring the
answer to this question: that URL returned JSON from *Next.js*, not Go — what
file made it exist, and how is it different from a `page.tsx`?
