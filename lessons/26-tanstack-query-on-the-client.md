# Lesson 26 — TanStack Query on the client

*Phase 4 · Session 26 of 38 · ~2 hours · Prereqs: Lessons 13, 15, 20, 25*

## Why this matters

Your BFF now hands the UI validated, typed JSON (Lesson 25). But the UI side of the
seam is probably still the hand-rolled `useState` + `useEffect` fetch from Lesson 15 —
which you wrote to *understand* fetching, not to keep. Every component that needs data
re-implements loading state, error state, caching, deduplication, and (hardest of all)
"refetch after a mutation". TanStack Query does all of that. Your team's codebase uses
it; MISSION item 3 names it explicitly.

## What you'll be able to do

- Explain the difference between client state and server state, and why `useState` is
  the wrong tool for the latter
- Set up a `QueryClientProvider` in a Next 15 App Router app without breaking the
  server/client boundary
- Fetch from a BFF route with `useQuery` and handle all three states: `isLoading`,
  `isError`, `error`
- Reason about cache identity using `queryKey`, including parameterised keys
- Invalidate the right cache entries after a mutation with `useMutation` +
  `invalidateQueries`

## Principles

### Server state ≠ client state

Client state is data the browser owns and that nobody else mutates: the open/closed
flag of a dropdown, text in a form box. Server state is data that *lives elsewhere* and
can change without you noticing: the task list your teammate just reassigned. `useState`
handles the first beautifully and the second terribly — it has no concept of "this data
is stale after 30 seconds" or "two components want the same fetch". TanStack Query is a
cache for server state: it tracks who wants what, when it was last fetched, and when to
refetch.

### queryKey is the cache's primary key

Every piece of cached server state is filed under a `queryKey` — an array of
serialisable values that *fully identifies* the request, like a map key in Go. Order
and contents matter:

```ts
["tasks", "p1"]          // tasks for project p1
["tasks", "p2"]          // different data — different cache entry
["tasks"]                // all tasks — yet another entry
["projects"]             // unrelated
```

Two components asking for `["tasks", "p1"]` at the same time trigger **one** network
call and share the result. Forget the `projectId` into the key and you'll serve p1's
tasks on p2's screen — the single most common TanStack bug. If a value changes what the
server returns, it belongs in the key.

### The v5 API (this is not the old API)

Version 5 unified everything into option objects. If you see old snippets with
`useQuery(key, fn)` or `useMutation(fn, opts)` — position-based — they are v4 and will
mislead you. The only shapes you'll write, all from
[the v5 docs](https://tanstack.com/query/latest/docs/framework/react/reference/useQuery):

```ts
const { data, isLoading, isError, error } = useQuery({
  queryKey: ["tasks", projectId],
  queryFn: () => fetchTasks(projectId),
});

const mutation = useMutation({
  mutationFn: (input: CreateTask) => postTask(input),
  onSuccess: () => queryClient.invalidateQueries({ queryKey: ["tasks"] }),
});
```

Also note: for mutations the loading flag is `isPending` (v4 called it `isLoading`).

### Three states, always

`isLoading` means "no data yet, first fetch running". `isError` means "the fetch threw —
read `error`". You handle all three, every time; a query without an error branch is a
crash waiting for the flaky endpoint of Lesson 28. `queryFn` must **throw** on failure —
checking `res.ok` and throwing is what hands the error to `error`. A `queryFn` that
returns `{error: ...}` instead of throwing makes TanStack think the fetch succeeded.

### Invalidation, not manual updates

After `POST /api/tasks` succeeds, resist the urge to splice the new task into local
state. `invalidateQueries({ queryKey: ["tasks"] })` marks every cache entry starting
with `tasks` as stale and refetches the ones in use. You traded micro-optimisation for
correctness-by-default: the UI shows what the server actually has. Given your mock
backend generates the task `id` and `assignee` server-side, letting the refetch bring
the authoritative row is simply less code that can lie.

### The provider is a client component

`QueryClient` holds mutable state and `QueryClientProvider` is context — both are
browser things, so the provider file starts with `"use client"` (Lesson 18). The root
`app/layout.tsx` stays a Server Component and just renders `<Providers>` around
children.

## Follow along

Working in `~/teampulse` with the Go mock running (`go run main.go` in `backend-mock/`)
and the zod boundary from Lesson 25 in place. Assume a `tp_token` cookie is set; if not,
log in per Lesson 27's preview or temporarily hard-code a token in the route handlers.

1. Install:
   ```bash
   cd ~/teampulse && pnpm add @tanstack/react-query
   ```
   Expected: `+ @tanstack/react-query 5.x`. The version matters — v5, not v4.

2. Create `app/providers.tsx`:

   ```tsx
   "use client";

   import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
   import { useState } from "react";

   export function Providers({ children }: { children: React.ReactNode }) {
     // useState keeps ONE client across re-renders without a module-level singleton
     const [client] = useState(() => new QueryClient());
     return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
   }
   ```

3. Wire it into `app/layout.tsx` — the layout itself stays a Server Component:

   ```tsx
   import { Providers } from "./providers";

   export default function RootLayout({ children }: { children: React.ReactNode }) {
     return (
       <html lang="en">
         <body>
           <Providers>{children}</Providers>
         </body>
       </html>
     );
   }
   ```
   Expected: `pnpm dev` still boots with no hydration warnings.

4. Create the fetcher layer `lib/queries.ts` — plain functions, testable, throwing:

   ```ts
   import type { Project, Task, CreateTask } from "@/lib/schemas";

   async function getJson<T>(url: string, init?: RequestInit): Promise<T> {
     const res = await fetch(url, init);
     if (!res.ok) {
       const body = (await res.json().catch(() => null)) as { error?: string } | null;
       throw new Error(body?.error ?? `request failed: ${res.status}`);
     }
     return res.json() as Promise<T>;
   }

   export const fetchProjects = () => getJson<Project[]>("/api/projects");

   export const fetchTasks = (projectId: string) =>
     getJson<Task[]>(`/api/tasks?projectId=${projectId}`);

   export const createTask = (input: CreateTask) =>
     getJson<Task>("/api/tasks", {
       method: "POST",
       headers: { "Content-Type": "application/json" },
       body: JSON.stringify(input),
     });
   ```

5. Rewrite the tasks page as a Client Component using all three states:

   ```tsx
   "use client";

   import { useQuery } from "@tanstack/react-query";
   import { useState } from "react";
   import { fetchTasks } from "@/lib/queries";

   export default function TasksPage() {
     const [projectId, setProjectId] = useState("p1");

     const { data, isLoading, isError, error } = useQuery({
       queryKey: ["tasks", projectId],
       queryFn: () => fetchTasks(projectId),
     });

     return (
       <main>
         <select
           value={projectId}
           onChange={(e) => setProjectId(e.target.value)}
         >
           <option value="p1">p1 Billing Service</option>
           <option value="p2">p2 Mobile App</option>
         </select>

         {isLoading && <p>Loading tasks…</p>}
         {isError && <p role="alert">Failed: {error.message}</p>}
         {data && (
           <ul>
             {data.map((t) => (
               <li key={t.id}>
                 {t.title} — {t.status} (P{t.priority})
               </li>
             ))}
           </ul>
         )}
       </main>
     );
   }
   ```
   Expected on `http://localhost:3000/tasks`: "Loading tasks…" flashes, then t1/t2/t5 for
   p1. Switch the dropdown to p2: t3/t4 — different key, different cache slot.

6. Watch the cache work. Open DevTools → Network, filter `tasks`, flip p1 → p2 → p1.
   Expected: requests fire, but flip back and forth *again* within a few seconds and the
   second visit to each renders instantly from cache before any request lands (TanStack
   shows stale data while refetching in the background — the `data` from the earlier
   fetch is still there while `isFetching` quietly runs).

7. Add the mutation with invalidation:

   ```tsx
   "use client";

   import { useMutation, useQueryClient } from "@tanstack/react-query";
   import { useState } from "react";
   import { createTask } from "@/lib/queries";

   function AddTaskForm({ projectId }: { projectId: string }) {
     const queryClient = useQueryClient();
     const [title, setTitle] = useState("");

     const addTask = useMutation({
       mutationFn: createTask,
       onSuccess: () => {
         queryClient.invalidateQueries({ queryKey: ["tasks"] });
         setTitle("");
       },
     });

     return (
       <form
         onSubmit={(e) => {
           e.preventDefault();
           addTask.mutate({ projectId, title, priority: 2 });
         }}
       >
         <input value={title} onChange={(e) => setTitle(e.target.value)} />
         <button disabled={addTask.isPending}>
           {addTask.isPending ? "Adding…" : "Add task"}
         </button>
         {addTask.isError && <p role="alert">{addTask.error.message}</p>}
       </form>
     );
   }
   ```
   Expected: submit "Write docs" → the new task appears under p1 with **no** manual
   state splice — `invalidateQueries` marked `["tasks", "p1"]` stale and TanStack
   refetched it. Check the Network tab: one `POST /api/tasks`, one `GET /api/tasks?...`.

## Your turn

1. Render the project list on the same page with `useQuery({ queryKey: ["projects"],
   queryFn: fetchProjects })` and feed the `<select>` options from `data` instead of
   hard-coded `<option>`s. Handle its loading and error states too.

   <details><summary>Solution</summary>

   ```tsx
   const { data: projects, isLoading: projectsLoading } = useQuery({
     queryKey: ["projects"],
     queryFn: fetchProjects,
   });

   // ...
   {projectsLoading && <option disabled>Loading…</option>}
   {projects?.map((p) => (
     <option key={p.id} value={p.id}>
       {p.name}
     </option>
   ))}
   ```
   Two independent queries coexist on one page; each tracks its own three states.
   </details>

2. Someone claims: "My dropdown for p2 shows p1's tasks." Given the code above is
   correct, what is the most likely single-line cause?

   <details><summary>Answer / Solution</summary>
   The `queryKey` omits the parameter — `queryKey: ["tasks"]` regardless of `projectId`.
   Both dropdowns then share one cache slot. Fix: `queryKey: ["tasks", projectId]`.
   Rule: everything the server filters on must be in the key.
   </details>

3. Make `createTask` optimistic-free but *observable*: set
   `staleTime: 30_000` on the tasks query. Predict what changes in the Network tab when
   you navigate away and back within 30 seconds, then verify.

   <details><summary>Solution</summary>
   With `staleTime: 30_000`, data is considered fresh for 30s — remounting the page
   within that window renders from cache with **no** network request at all (versus the
   default `staleTime: 0`, which fires a background refetch on every mount). Trade-off:
   fresher UI versus more requests; for a team dashboard 30s is usually honest enough.
   </details>

## Checkpoint

1. Give one example of client state and one of server state from TeamPulse, and say why
   the second is wrong for `useState`.

   <details><summary>Answer</summary>
   Client: the selected `projectId` in the dropdown, the half-typed task title. Server:
   the task list — it lives in Go, can change under you, is needed by several
   components, and has a staleness lifetime. `useState` gives you none of cache,
   dedupe, staleness, or refetch machinery.
   </details>

2. Write the queryKey for "tasks for project p2" and explain why `["tasks"]` for that
   same fetch is a bug.

   <details><summary>Answer</summary>
   `["tasks", "p2"]`. `["tasks"]` makes every project's tasks collide in one cache slot —
   switch projects and you render the previous project's data from cache.
   </details>

3. What must a `queryFn` do on a failed request, and what happens if it returns an
   error object instead?

   <details><summary>Answer</summary>
   It must `throw`. TanStack Query only enters `isError` state when the promise rejects;
   a resolved `{error: ...}` is "successful" data, so `isError` stays false and your
   error UI never renders. That's why `getJson` checks `res.ok` and throws.
   </details>

4. (Go API question) After `POST /api/tasks` succeeds, the Go server assigns the task a
   random `id`, `status: "todo"`, a `dueDate` 7 days out, and `assignee` from the token.
   Why does `invalidateQueries` beat splicing `response.data` into the list yourself?

   <details><summary>Answer</summary>
   The list you splice is your *guess* at the server's row; the invalidation refetch
   shows the real one — including the fields only Go computes (id, dueDate, assignee).
   One line of code, no drift between what you rendered and what Go stored.
   </details>

5. Why does `providers.tsx` need `"use client"` while `layout.tsx` doesn't?

   <details><summary>Answer</summary>
   `QueryClient` is a mutable browser-side cache and context is client-only — they must
   run in the browser, so the file that uses them is a Client Component. The layout is
   pure JSX composition and can keep rendering on the server (Lesson 18's boundary
   rules).
   </details>

## Primary source

[TanStack Query v5 — useQuery reference](https://tanstack.com/query/latest/docs/framework/react/reference/useQuery) —
the official v5 API page: exact option names, every returned state flag, and the v5
object signature, so you never accidentally write v4 code.

## Before next session

Open DevTools on your running TeamPulse, Application → Cookies → `http://localhost:3000`,
and note every cookie listed. Then read the Lesson 27 opener in
`lessons/27-auth-tokens-and-cookies-through-the-bff.md` up to "Principles" and write down
one question you expect it to answer about where the Go token should live.
