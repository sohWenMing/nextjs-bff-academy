# Lesson 33 — Performance at Working Level

*Phase 5 · Session 33 of 38 · ~2 hours · Prereqs: Lesson(s) [13](13-state-usestate-and-rerenders.md), [18](18-server-vs-client-components.md), [26](26-tanstack-query-on-the-client.md)*

## Why this matters

MISSION.md explicitly puts performance *micro*-optimization out of scope — but working-level
performance is in scope, because the five footguns in this lesson are how real Next.js teams
get paged: a login token in the JS bundle, a page that hydrates wrong, a dashboard that
re-renders 200 components per keystroke. You own the BFF boundary; several of these footguns
live exactly at that boundary, and all of them are cheap to fix if you spot them on day one.

## What you'll be able to do

- You can state, in one sentence, what `"use client"` costs and where the boundary should sit.
- You can use `next/image` and `next/font` correctly and say what each prevents.
- You can look at any `.tsx` file and defend why it is (or isn't) a Server Component — default is server.
- You can name the five common footguns — hydration mismatch, unnecessary client fetch, prop drilling, re-render storms, shipping secrets — and give the fix for each.
- You can find your own dashboard's JS bundle cost per route and tell whether it grew.

## Principles

### `"use client"` is a bundle bill, not a keyword

Every Server Component (Lesson 18) renders on the server and ships **zero JavaScript** for
itself — the browser gets HTML plus the *boundary* components. The moment you write
`"use client"` at the top of a file, that component **and every component it imports**
gets bundled, sent to the browser, parsed, and executed. The bill is paid by your users on
slow phones, not by you on your M-series laptop.

The correct move is the **smallest possible client island**: keep the page a Server
Component, fetch data with plain `fetch` on the server (Lesson 19), and push the
interactive leaf — the button, the form, the filter — into a small `"use client"` child.

```tsx
// app/projects/page.tsx — server component: no "use client", no JS shipped for this file
import { TaskFilter } from "@/components/task-filter"; // the client island

export default async function ProjectsPage() {
  const res = await fetch(`${process.env.GO_URL}/api/projects`, {
    headers: { Authorization: `Bearer ${await getToken()}` },
    cache: "no-store",
  });
  const projects = await res.json();
  return (
    <main>
      <h1>Projects</h1>
      <TaskFilter projects={projects} /> {/* only this subtree is JS in the browser */}
    </main>
  );
}
```

Rule of thumb your team will recognize: **`"use client"` at the top of a leaf, never at the
top of a page** — unless the whole page is genuinely one interactive widget.

### `next/image` and `next/font`: free wins

`<img>` problems: no size means layout jump when it loads (Cumulative Layout Shift —
content visibly jumps), full original bytes sent to every visitor, no lazy loading by
default. `next/image` fixes all three:

```tsx
import Image from "next/image";

<Image src={avatarUrl} alt="Wen" width={48} height={48} />
```

You must give `width`/`height` (or `sizes`) so Next can reserve the space and pick the
right derivative. Remote hosts must be allowlisted in `next.config.ts` under
`images.remotePatterns` (Lesson 29) or the build refuses them.

`next/font` problems solved: fonts fetched at runtime from a CDN add a network round-trip
and cause **FOUT** (Flash of Unstyled Text — you see one font, then a swap). `next/font`
downloads fonts at *build time* and self-hosts them, with zero layout shift:

```ts
// app/layout.tsx
import { Inter } from "next/font/google";
const inter = Inter({ subsets: ["latin"], display: "swap" });

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={inter.className}>
      <body>{children}</body>
    </html>
  );
}
```

### Server components by default — the performance architecture

The mental model from Lesson 18, now with a performance frame: **data fetching belongs on
the server unless interactivity demands otherwise.** A Server Component `await`s the BFF
or the Go backend directly, renders to HTML, and the browser paints content without ever
downloading that logic. Client-side fetching (Lesson 15) is for data that changes *with
user interaction* — which is why TanStack Query (Lesson 26) lives on the client for
task mutations and refetches, not for first paint.

### The five footguns

1. **Hydration mismatch.** Hydration is React attaching event handlers to server-rendered
   HTML, expecting the first client render to *exactly match* it. It doesn't, React logs
   `Hydration failed because the initial UI does not match` in the console, and throws
   away/re-renders the subtree (slow, sometimes visibly flickering). Classic causes:
   `Date.now()`/`Math.random()`/`new Date().toLocaleString()` computed in a component,
   reading `localStorage` or `window` during render, or invalid nesting (`<div>` inside
   `<p>`). Fix: compute such values on the server and pass them as props, or render them
   only after mount inside a client component (`const [mounted, setMounted] =
   useState(false); useEffect(() => setMounted(true), []); return mounted ? <Real/> :
   <Placeholder/>`).
2. **Unnecessary client fetch.** The page fetches `/api/projects` with `useEffect`/Query on
   mount even though nothing about it depends on interaction — so first paint waits for
   HTML *plus* a JS round-trip *plus* an extra BFF call. Fix: fetch in the Server
   Component (Lesson 19) and hand the data to a client island, or keep Query but seed it
   with server-fetched data.
3. **Prop drilling.** Passing `token`/`user`/`projects` down five layers of components that
   don't use them, just so a leaf can. Cost: every intermediate component must accept and
   forward props, changes ripple through files, and re-renders cascade. Fix at working
   level: colocate (split the tree so the leaf is closer), or React Context for truly
   global things like the session. Redux is explicitly out of scope per MISSION.md.
4. **Re-render storms.** One `useState` at the top of a big page → every keystroke in one
   input re-renders the whole tree (React DevTools Highlight updates from Lesson 32 shows
   the whole page flashing). Root cause is the render rule from Lesson 13: state change =
   re-render *that component and everything below it*. Fixes, in order of value: move
   state down to the smallest component that needs it; extract stable sections into
   separate components (so they're siblings of the state, not children); for lists, pass
   primitives instead of fresh object/array literals into children. `memo`/`useCallback`
   are last resorts, not first moves — reach for them only when the Profiler proves a
   specific component is the wide bar.
5. **Shipping secrets.** Anything in a client component — or any string prefixed
   `NEXT_PUBLIC_` — ends up in the JS bundle, readable by anyone with View Source. The
   Go API key, the session token, and internal URLs belong in route handlers (Lesson 20)
   and httpOnly cookies (Lesson 27). Check tonight: `grep -r "NEXT_PUBLIC_" .env*` and ask
   for each one "would I paste this in a public GitHub gist?" — if not, it must not be
   `NEXT_PUBLIC_`.

## Follow along

Work in `~/teampulse` (Lesson [30](30-p3-full-stack-teampulse.md)), Go mock on `:8080`.

1. **Measure the baseline.** Stop the dev server and build:

   ```bash
   pnpm build
   ```

   Expected: the route table, with a `First-Load JS` column per route. Write down the
   number for your main dashboard route (e.g. `○ /projects  105 kB`). This is the bundle
   bill from every `"use client"` reachable from that route.

2. **See the bill grow.** Temporarily add `"use client";` as the very first line of your
   projects *page* file (and fix any `async`/`await` it breaks by moving one fetch to a
   quick `useEffect`). Run `pnpm build` again. Expected: the same route's JS is several kB
   heavier — that delta is exactly what `"use client"` at the top of a page costs. Revert.

3. **Hydration mismatch, on purpose.** In a client component, render
   `{new Date().toLocaleTimeString()}` directly. Reload. Expected: a console error
   `Hydration failed because the initial UI does not match...` (plus the visible
   mismatch: the server's time ≠ the browser's time). Fix it with the `mounted` flag
   pattern and confirm the console is clean.

4. **Find a storm.** Components tab → Highlight updates → type in your task filter input
   (or equivalent). Expected: if page-level state re-renders everything, watch most of the
   page flash. Move that state down to just the input's component, repeat. Expected: one
   box flashes. Compare against your Profiler notes from Lesson 32 — the wide bar should
   have shrunk or disappeared.

5. **Secret audit.**

   ```bash
   grep -rn "NEXT_PUBLIC_" .env* src/ || true
   grep -rn "demo\|api[_-]key\|token" src/components src/app --include="*.tsx" -i | grep -v "app/api" || true
   ```

   Expected: no `NEXT_PUBLIC_` entry holds anything you wouldn't publish, and no client
   component file (anything under `src/components` or a page without a matching route
   handler) contains credential strings. Anything it prints is a footgun #5 to fix now.

6. **Swap in `next/image`.** If your dashboard shows any avatar/logo via `<img>`, convert
   it to `next/image` with explicit `width`/`height` (add `remotePatterns` in
   `next.config.ts` if the host is remote). Expected: `pnpm build` passes, and in DevTools
   → Network (throttled to Slow 4G) the image arrives at a sensible byte size and doesn't
   shift the layout when it lands.

## Your turn

1. Your dashboard route shows 180 kB First-Load JS and you expected ~110 kB. Name the
   first two things you'd check and the tool for each.
2. A teammate writes a route handler that reads `process.env.NEXT_PUBLIC_GO_URL` and says
   "it works in dev." Later someone notes the internal Go URL is visible in the browser
   bundle. What went wrong, and what's the fix?
3. The app's header shows "Welcome back" only after a visible 200ms flash of nothing, and
   the console sometimes shows a hydration error. Diagnose and fix using footgun #1.

<details>
<summary>Solution</summary>

1. (a) Which files carry `"use client"` reachable from that route — search for
   `use client` and check the import graph from the page; tool: editor search +
   `pnpm build` output comparing route JS before/after lifting one boundary. (b) A heavy
   dependency pulled into a client component (date library, icon set) — tool:
   `pnpm build` with `ANALYZE=true` if the team has `@next/bundle-analyzer`, otherwise the
   React DevTools Profiler plus removing imports one at a time.

2. `NEXT_PUBLIC_` values are inlined into the JS bundle at **build time** — anything in
   them is public. The Go URL should be a plain (non-`NEXT_PUBLIC_`) env var, readable
   only server-side; the route handler keeps working in dev and prod because route
   handlers *are* the server. Lesson 35 formalizes this.

3. The header reads `localStorage`/session during first render, so server HTML and client
   first-render differ (hydration error) and the text pops in late. Fix: fetch the user
   server-side and pass `name` as a prop (preferred), or gate the client-only text behind
   a `mounted` flag with a stable placeholder so first client render matches the server.

</details>

## Checkpoint

1. One sentence: what exactly does `"use client"` cost the user?

<details>
<summary>Answer</summary>

That component and everything it imports is downloaded, parsed, and executed as JS in the
browser — bytes that a Server Component would have spent zero of.

</details>

2. Why must `next/image` receive `width` and `height` (or `sizes`)?

<details>
<summary>Answer</summary>

So Next can reserve the layout space before the bytes arrive (preventing layout shift)
and select/serve a correctly sized derivative.

</details>

3. Your page component contains `const now = new Date().toLocaleDateString()`. Why is this
   a hydration hazard, and name two fixes?

<details>
<summary>Answer</summary>

The server renders with the server clock, the client's first render uses the browser
clock — any mismatch (even across midnight or timezones) breaks hydration. Fixes: compute
on the server and pass as a prop; or render it client-only after mount with a placeholder.

</details>

4. Which of the five footguns is "one input keystroke re-renders 200 rows," and what is
   the first fix before reaching for `memo`?

<details>
<summary>Answer</summary>

Re-render storm (footgun #4). First fix: move the state down to the smallest component
that needs it (and extract stable sections as siblings), verified with React DevTools
Highlight updates.

</details>

5. You want the mock Go URL `http://localhost:8080` used by your route handlers, plus a
   client-readable app name. Write the two env var names — and justify the prefix choice.
   (Bonus: what is the login password in every curl you'd test with?)

<details>
<summary>Answer</summary>

`GO_URL=http://localhost:8080` (no prefix — read only in route handlers, never bundled)
and `NEXT_PUBLIC_APP_NAME=TeamPulse` (the UI legitimately displays it). Password for curl
tests against the mock: `demo` (`backend-mock/README.md`).

</details>

## Primary source

[Next.js docs — Server Components](https://nextjs.org/docs/app/building-your-application/rendering/server-components).
The canonical statement of the default-server model, the `"use client"` boundary rules,
and the import graph that decides what gets bundled — everything in the "bundle bill"
section traces to this page.

## Before next session

Run `pnpm build`, note the First-Load JS for your two busiest routes, and pick the single
heaviest client component you can find. Lesson 34 adds tests; keep the build numbers
handy — you'll re-check them after Lesson 35's deploy too.
