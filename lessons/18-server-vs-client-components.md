# Lesson 18 — Server vs Client Components

*Phase 3 · Session 18 of 38 · ~2 hours · Prereqs: Lesson(s) 10, 13, 15, 17*

## Why this matters

This is the single most important distinction in Next.js, and the one your team's code review will hammer on: **which code runs on the server and which ships to the browser.** A secret in a file that ships to the browser is a security incident; a `useState` in a file that only runs on the server is a crash. Mission item #1 — "read any App Router file and know what runs where" — is exactly this lesson. The BFF you build in Lesson 20 only makes sense once the boundary is obvious to you at a glance.

## What you'll be able to do

- You can state, within 10 seconds of opening any `.tsx` file, whether it runs on the server, the browser, or both — and name the evidence.
- You can explain what `"use client"` does (and the two things it does *not* do).
- You can fix the error "useState only works in Client Components" by moving the boundary, not by scattering `"use client"`.
- You can pass data and UI from a server component into a client component, including via `children`.
- You can predict whether a given `import` line drags code into the browser bundle.

## Principles

### The default: components run on the server

In the App Router, every component runs **on the server, during rendering**, unless you opt it out. A server component is a function React calls on the Node process; the browser receives the *result* — HTML — not your function. This is the opposite of everything you did in Phase 2, where (in the browser) every component you wrote ran on the user's machine.

Practical consequences of running on the server:

- You *can* read files, query databases, and call `http://localhost:8080` with
a secret token — the browser never sees any of it.
- You *cannot* use anything that only exists in a browser: `window`,
`document`, click handlers, `localStorage` — or React's interactive primitives `useState` and `useEffect` ([Lesson 13](13-state-usestate-and-rerenders.md)).

### `"use client"` marks a boundary, not a file type

The string `"use client"` as the very first line of a file (above the imports) tells Next.js: *this component and everything it imports may only run in the browser; ship it as JS.* Three things to get exactly right:

1. **It opts you *out* of server rendering** for that component. React still
renders it on the server once for the initial HTML, but the interactivity — state, events — hydrates and lives in the browser.
2. **It is transitive downward.** If `Page` is a server component and imports
`SearchBar` which has `"use client"`, then `SearchBar` *and every file it imports* become browser code. You do not sprinkle `"use client"` on every child — the boundary at `SearchBar` covers them.
3. **It does not make the file "a client file."** A client component can
*receive* server-rendered UI. See `children` below.

The mental model: `"use client"` draws a line across your component tree. Above the line: server-only, invisible to the browser. Below the line: downloaded, bundled, running on the user's laptop.

### The interactivity rule

Only client components can do interactive things. If you need any of these, that code — and only that code — goes below a `"use client"` boundary:

- event handlers (`onClick`, `onChange` — the controlled forms of [Lesson 14](14-forms-and-controlled-inputs.md))
- `useState`, `useEffect`, and every other hook
- browser APIs (`window`, `navigator`, ...)

You already obey this rule without noticing: the whole P1 explorer page from [Lesson 16](16-p1-teampulse-explorer-page.md) starts with `"use client"` because it has a login form, filter state, and click handlers.

### Crossing the boundary: props and `children`

A server component can render a client component and pass it plain data (strings, numbers, arrays of plain objects — anything that survives `JSON.stringify`). It cannot pass a *function* — functions can't serialize across the wire:

```tsx
// app/dashboard/page.tsx — SERVER component (no "use client")
import { Counter } from "./counter";

export default function DashboardPage() {
  // runs only on the server — the browser never sees this log
  console.log("server-side render");
  return (
    <div>
      <h1>Dashboard</h1>
      <Counter start={0} />        {/* plain data across the line */}
    </div>
  );
}
```

```tsx
// app/dashboard/counter.tsx — CLIENT component
"use client";

import { useState } from "react";

export function Counter({ start }: { start: number }) {
  const [n, setN] = useState(start); // useState requires this file below the line
  return <button onClick={() => setN(n + 1)}>{n}</button>;
}
```

And the reverse trick — server UI *inside* a client component — uses `children`, because children are rendered by the parent (server) and passed in as already-rendered UI:

```tsx
// client shell that receives server-rendered children
"use client";
export function Panel({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useState(true);
  return (
    <section>
      <button onClick={() => setOpen(!open)}>toggle</button>
      {open && children}
    </section>
  );
}
```

A server page can then do `<Panel><ExpensiveServerTable /></Panel>` — the table renders on the server; only `Panel` ships to the browser.

### Where does the code go? Follow the imports

The browser bundle is determined by the import graph below every `"use client"`, not by file names. Two checks that settle any argument:

- Start a production build and read the output. A page importing a huge
library into a client component shows a fat `First Loading JS` number (Lesson 33 makes this precise).
- Put `console.log("I ran on the server")` at the top of a component body.
Server components log in your **terminal** (dev server output), client components log in the **browser console**. During the initial render both run once on the server; after that, only the client one keeps running — on every keystroke, in the browser.

The phrase your team will use: **"it's a server component"** means "no `"use client"` anywhere in its import chain, so none of it ships."

### Why the boundary is the security lesson

Everything below the line is readable by anyone who opens DevTools → Sources — your code, your URLs, your constants. `process.env.SOME_SECRET` read in a server component is invisible; the same read in a client component is a public disclosure (formalized in Lesson 20 with the `NEXT_PUBLIC_` prefix). There is no "mostly private" side of the line.

## Follow along

In `~/teampulse` with `pnpm dev` running (Go backend up on `:8080` from `cd backend-mock && go run main.go`):

1. **Audit the app you already have.** Run:

   ```bash
   grep -rn '"use client"' app/
   ```

Expected: at least your P1 explorer page from [Lesson 16](16-p1-teampulse-explorer-page.md). Open that file: confirm `useState` and `onClick` appear inside it — that's *why* the directive is there.

2. **See both runtimes log.** Create `app/boundary/page.tsx`:

   ```tsx
   import { Clickable } from "./clickable";

   console.log("[server] boundary page body ran");

   export default function BoundaryPage() {
     console.log("[server] BoundaryPage render ran");
     return (
       <main>
         <h1>Boundary demo</h1>
         <Clickable />
       </main>
     );
   }
   ```

Create `app/boundary/clickable.tsx`:

   ```tsx
   "use client";

   import { useState } from "react";

   console.log("[either] Clickable module loaded");

   export function Clickable() {
     const [count, setCount] = useState(0);
     console.log("[either] Clickable render ran");
     return <button onClick={() => setCount(count + 1)}>count: {count}</button>;
   }
   ```

Visit `http://localhost:3000/boundary`. Expected in the **terminal**: `[server]` lines plus `[either]` lines. Expected in the **browser console**: only the `[either]` lines. Click the button three times: the terminal stays silent; the browser console logs `[either] Clickable render ran` three more times. Server components render once per request; client components re-render on every state change, forever, in the browser.

3. **Break it on purpose.** Delete `"use client"` from `clickable.tsx`, save,
reload. Expected: an error overlay saying `useState only works in Client Components. Add the "use client" directive at the top of your file to use it on the client.` Next.js tells you the fix verbatim. Restore the line; confirm it works again.

4. **Move the boundary down, not up.** Your P1 page is one big client
component. Prove the boundary can be finer-grained: create `app/boundary/prose.tsx` *without* the directive that renders a static `<p>` of your choice, and use it inside `boundary/page.tsx` between the `<h1>` and `<Clickable />`. Confirm the terminal logs when you add a top-level `console.log("[server] prose")` to it — static UI stays on the server even next to a client component.

5. **Try to cross with a function (and fail).** In `boundary/page.tsx`
(server), pass a handler to the client component: `<Clickable onBoom={() => console.log("boom")} />`, and in `clickable.tsx` accept `onBoom: () => void` and call it in the button. Expected: a build error along the lines of `Event handlers cannot be passed to Client Component props ... only serializable props are supported`. This is the serialization boundary being honest with you. Revert the prop.

## Your turn

1. Extract the filter control from your P1 explorer page (the input or select
that filters tasks, [Lesson 16](16-p1-teampulse-explorer-page.md)) into its own file as a client component that receives the task list as a prop and renders the filtered result. Leave the surrounding page a server component if you can — if the page still holds token state, keep the page client and note *which hook forces it*.

   <details><summary>Solution</summary>

The component that calls `useState` for the filter *must* have `"use client"`. The page itself stays client for now because the login token lives in page state (`useState`) — that's the hook forcing it. The key win: `TaskFilter` and its imports are the only browser JS; move the filter's `useState` inside `TaskFilter`:

   ```tsx
   // app/explorer/task-filter.tsx
   "use client";

   import { useState } from "react";

   type Task = { id: string; title: string; status: string };

   export function TaskFilter({ tasks }: { tasks: Task[] }) {
     const [query, setQuery] = useState("");
     const shown = tasks.filter((t) =>
       t.title.toLowerCase().includes(query.toLowerCase()),
     );
     return (
       <div>
         <input
           value={query}
           onChange={(e) => setQuery(e.target.value)}
           placeholder="filter tasks"
         />
         <ul>
           {shown.map((t) => (
             <li key={t.id}>{t.title}</li>
           ))}
         </ul>
       </div>
     );
   }
   ```

Plain-object arrays are serializable, so passing `tasks` across the boundary is legal.
   </details>

2. Without opening the files, predict server-vs-browser for each, then check
by grepping for `"use client"` and by scanning imports: (a) `app/layout.tsx` of a stock create-next-app, (b) any file your P1 page imports that contains `onClick`, (c) a file with no directive that *imports* a file with the directive.

   <details><summary>Answer</summary>
(a) Stock `app/layout.tsx` has no directive — server component (its font setup runs on the server). (b) `onClick` requires interactivity → client. (c) Server *component* still, but the imported client file (and its import graph) ships to the browser — "imports drag code below the line."
   </details>

3. Harder: a teammate writes a server component that reads
`localStorage.getItem("token")` and gets `ReferenceError: localStorage is not defined`. Fix it in two different ways and say when you'd pick each.

   <details><summary>Solution</summary>
Way 1: make that component a client component (`"use client"`) — pick this when the value genuinely belongs to *this browser tab*. Way 2 (better, and the course's direction): stop using `localStorage` for tokens entirely and read the token on the server — from a cookie or route handler — which is where Lesson 20–22 and eventually Lesson 27 take us. Browser storage is readable by any script on the page; a server-held session is the team-grade answer.
   </details>

## Checkpoint

1. You open a `.tsx` file with no `"use client"` anywhere in it — but it
imports a component that has `"use client"`. What ships to the browser?

   <details><summary>Answer</summary>
The client component and everything *it* imports. The outer component's own code stays server-only; boundaries are per-import, transitive downward.
   </details>

2. Name two things `"use client"` does and one thing people wrongly assume it
does.

   <details><summary>Answer</summary>
Does: (1) marks the component + its import graph as browser code (bundled and shipped); (2) enables hooks/events (interactivity hydrates in the browser). Wrongly assumed: that the component *only* runs in the browser — React still renders it server-side once for the initial HTML. (Also wrong: that it affects files that import *it*.)
   </details>

3. Why can't a server component pass `onSubmit={() => ...}` to a client
component, and what are the two sanctioned alternatives?

   <details><summary>Answer</summary>
Props crossing the boundary must be serializable (sent over the wire as JSON); functions aren't. Alternatives: put the handler inside the client component itself, or pass server-rendered UI as `children` so the server code never has to reference a browser function.
   </details>

4. TeamPulse: your P1 explorer stores the Go API's Bearer token in React state
in a `"use client"` file. Where does that token live at runtime, and why is that uncomfortable even though the mock password is just `demo`?

   <details><summary>Answer</summary>
In the browser's JS memory (and any devtools inspection of React state). Anything the browser holds, the user — or any injected script — can read. With real credentials that's token theft; the fix direction is "tokens are read on the server only," which Lessons 20 and 27 implement.
   </details>

5. Your page's JS bundle is suspiciously large. What's the first question to
ask about the component tree?

   <details><summary>Answer</summary>
"How much of this tree is below a `"use client"` line?" Bundle size is the cost of the boundary — find the client components pulling in heavy imports and move static UI above the line (or pass it as `children`).
   </details>

## Primary source

[Next.js Docs — Server and Client Components](https://nextjs.org/docs/app/building-your-application/rendering/server-components) — defines the two runtimes, the `"use client"` boundary, and the `children` escape hatch with the exact diagrams worth memorizing.

## Before next session

In `app/boundary/page.tsx`, above the imports, add the two lines:

```tsx
const res = await fetch("http://localhost:8080/api/health");
const health = await res.json();
```

and render `{JSON.stringify(health)}` in the page. Save and reload `/boundary`, then answer in one sentence: which console (terminal or browser) showed any errors, and what does that tell you about *where the fetch ran*? Bring your sentence — Lesson 19 starts from it.
