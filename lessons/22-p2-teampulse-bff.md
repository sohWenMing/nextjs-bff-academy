# Lesson 22 — [P2] TeamPulse BFF

*Phase 3 · Session 22 of 38 · ~2 hours · Prereqs: Lesson(s) 17, 18, 19, 20, 21*

## Why this matters

This is the session the MISSION's definition-of-done #2 and #4 are really about: **every byte between your app and Go flows through your route handlers.** Today you finish the refactor that [Lesson 20](20-route-handlers-your-first-bff.md) started — the browser stops knowing `localhost:8080` exists, the login credential (`demo`) and Bearer token become Node-process-only, lists server-render, and the one mutation the app has (create task) goes through a POST route handler. When you leave this session you have *built the thing you're being hired to own*.

## What you'll be able to do

- You can refactor an app so all external traffic to the Go API passes through `app/api/**/route.ts` handlers, and prove it with grep + DevTools.
- You can centralize upstream access in one server-only module (`lib/go.ts`) that both route handlers and server pages import.
- You can convert client-fetched lists into server-rendered pages with `loading`/`error`/`not-found` boundaries from [Lesson 21](21-loading-error-not-found-states.md).
- You can write a small client component (create-task form) that POSTs to your own route handler and refreshes server-rendered data with `router.refresh()`.
- You can run a 6-point acceptance checklist against P2 before calling it done.

## Principles

### Target architecture (memorize the shape, not the keystrokes)

```text
Browser (DevTools sees ONLY :3000)
  │  GET /            server-rendered project list   (Lesson 19 pattern)
  │  GET /projects/p1 server-rendered project + tasks (Lesson 21 boundaries)
  │  POST /api/tasks  from the create-task form       (the one browser-made API call)
  ▼
Next.js :3000
  app/api/login/route.ts          POST  → Go /api/login
  app/api/projects/route.ts       GET   → Go /api/projects
  app/api/projects/[id]/route.ts  GET   → Go /api/projects/{id}
  app/api/tasks/route.ts          GET+POST → Go /api/tasks
        │  all import ONE file:
        ▼
  lib/go.ts   ← the ONLY module that knows GO_BASE_URL; server-only
        │  fetch + Bearer token (login with password demo happens HERE)
        ▼
Go mock :8080                     (invisible to the browser)
```

Three rules the whole project hangs on:

1. **One upstream door.** `GO_BASE_URL` appears exactly once in the codebase
(`lib/go.ts`), and its value lives only in `.env.local`. Grep for `8080` in `app/` and `lib/`: zero hits expected — the number exists only in the env file.
2. **Route handlers are the *external* Go surface.** The browser, curl, and
any future mobile client hit `/api/*`. Server components may import `lib/go.ts` functions *directly* — that's not "bypassing the BFF", it's the same server process calling a function instead of looping back through HTTP. (Never `fetch("http://localhost:3000/...")` from inside Next.js; Lesson 20's rule.)
3. **Credentials never cross to the browser.** `demo`, tokens, and
`GO_BASE_URL` have no `NEXT_PUBLIC_` prefix and sit above every `"use client"` line ([Lesson 18](18-server-vs-client-components.md)).

### Why `lib/go.ts` instead of copy-pasting handlers

Lesson 20's handlers each re-logged-in. Fine for learning; a real BFF keeps the Go client in one module: one place to add timeouts (Lesson 28), one place to swap to real credentials (Lesson 27), one place where zod validation will land (Lesson 25). Route handlers shrink to: *parse request → call `lib/go` → map to status codes*. That division — **handlers do HTTP, lib does Go** — is the exact layering in your team's codebase.

### One mutation pattern: POST + `router.refresh()`

Server-rendered data (Lesson 19) has no client cache to update. The App-Router-native loop for a write: client form → `fetch("/api/tasks", {method: "POST"})` → on success, `router.refresh()` (from `next/navigation`) re-runs the server components on screen, re-fetching from Go through your lib. No TanStack Query yet — that's Lesson 26's upgrade, and you'll feel the difference then.

## Follow along

`pnpm dev` in `~/teampulse`, Go mock up (`cd backend-mock && go run main.go`), `.env.local` containing `GO_BASE_URL=http://localhost:8080` (restart dev server after creating it). Compare against the file-tree sketch you drew in the prep.

1. **The one Go client.** Create `lib/go.ts` (any folder name except a
route-shape; `lib/` is convention, server-imported only):

   ```ts
   const GO = process.env.GO_BASE_URL;

   async function goLogin(): Promise<string> {
     const res = await fetch(`${GO}/api/login`, {
       method: "POST",
       headers: { "Content-Type": "application/json" },
       body: JSON.stringify({ username: "bff", password: "demo" }),
       cache: "no-store",
     });
     if (!res.ok) throw new Error(`go login failed: ${res.status}`);
     const data = (await res.json()) as { token: string };
     return data.token;
   }

   async function go<T>(path: string, init: RequestInit = {}): Promise<T> {
     const token = await goLogin();
     const res = await fetch(`${GO}${path}`, {
       ...init,
       headers: { Authorization: `Bearer ${token}`, ...init.headers },
       cache: "no-store",
     });
     if (!res.ok) {
       throw new GoError(res.status, await res.text());
     }
     return res.json();
   }

   export class GoError extends Error {
     constructor(readonly status: number, body: string) {
       super(`go ${status}: ${body}`);
     }
   }

   export type Project = {
     id: string;
     name: string;
     description: string;
     status: "active" | "archived";
     createdAt: string;
   };

   export type Task = {
     id: string;
     projectId: string;
     title: string;
     status: "todo" | "doing" | "done";
     priority: number;
     dueDate: string;
     assignee: string;
   };

   export const getProjects = () => go<Project[]>("/api/projects");
   export const getProject = (id: string) => go<Project>(`/api/projects/${id}`);
   export const getTasks = (projectId?: string) =>
     go<Task[]>(`/api/tasks${projectId ? `?projectId=${projectId}` : ""}`);
   export const createTask = (t: { projectId: string; title: string; priority: number }) =>
     go<Task>("/api/tasks", {
       method: "POST",
       headers: { "Content-Type": "application/json" },
       body: JSON.stringify(t),
     });
   ```

The `GoError` type carries the upstream status so handlers can map it ([Lesson 07](07-typescript-ii-shapes-and-narrowing.md)). `readonly` on a constructor param is TS shorthand for "field + assign" — read it, don't gloss over it.

2. **Shrink the handlers to HTTP glue.** Replace Lesson 20's
`app/api/projects/route.ts` body:

   ```ts
   import { NextResponse } from "next/server";
   import { getProjects, GoError } from "@/lib/go";

   export async function GET() {
     try {
       return NextResponse.json(await getProjects());
     } catch (err) {
       if (err instanceof GoError && err.status === 401) {
         return NextResponse.json({ error: "unauthorized" }, { status: 401 });
       }
       return NextResponse.json({ error: "upstream unavailable" }, { status: 502 });
     }
   }
   ```

(`@/lib/...` works because create-next-app sets the `@/*` path alias to the project root; if yours isn't configured, use `../../lib/go` and note the import depth.) Expected: `curl -s localhost:3000/api/projects` → the three projects, unchanged from Lesson 20 — now through one lib.

3. **The rest of the handlers.** Create the trio — verify each with curl
immediately after writing it:

   ```ts
   // app/api/projects/[id]/route.ts
   import { NextResponse } from "next/server";
   import { getProject, GoError } from "@/lib/go";

   export async function GET(
     _req: Request,
     ctx: { params: Promise<{ id: string }> },
   ) {
     const { id } = await ctx.params;
     try {
       return NextResponse.json(await getProject(id));
     } catch (err) {
       if (err instanceof GoError && err.status === 404) {
         return NextResponse.json({ error: "project not found" }, { status: 404 });
       }
       return NextResponse.json({ error: "upstream unavailable" }, { status: 502 });
     }
   }
   ```

   ```ts
   // app/api/tasks/route.ts
   import { NextResponse } from "next/server";
   import { getTasks, createTask, GoError } from "@/lib/go";

   export async function GET(req: Request) {
     const projectId = new URL(req.url).searchParams.get("projectId") ?? undefined;
     try {
       return NextResponse.json(await getTasks(projectId));
     } catch (err) {
       const st = err instanceof GoError ? err.status : 502;
       return NextResponse.json({ error: "upstream unavailable" }, { status: st >= 400 ? st : 502 });
     }
   }

   export async function POST(req: Request) {
     let body: { projectId?: string; title?: string; priority?: number };
     try {
       body = await req.json();
     } catch {
       return NextResponse.json({ error: "invalid JSON body" }, { status: 400 });
     }
     if (!body.projectId || !body.title) {
       return NextResponse.json(
         { error: "projectId and title are required" },
         { status: 400 },
       );
     }
     try {
       const task = await createTask({
         projectId: body.projectId,
         title: body.title,
         priority: body.priority ?? 3,
       });
       return NextResponse.json(task, { status: 201 });
     } catch (err) {
       const st = err instanceof GoError ? err.status : 502;
       return NextResponse.json({ error: "upstream unavailable" }, { status: st >= 400 ? st : 502 });
     }
   }
   ```

   ```bash
   curl -s 'localhost:3000/api/tasks?projectId=p2'   # t3, t4
   curl -s -X POST localhost:3000/api/tasks -H 'Content-Type: application/json' \
     -d '{"projectId":"p2","title":"P2 checkpoint","priority":1}'   # 201 + task
   curl -s localhost:3000/api/projects/p9            # 404 + your error shape
   ```

(Keep `app/api/login/route.ts` from Lesson 20 or add a thin one now if you want it in the public surface; nothing in the app needs it yet — Lesson 27 does.)

4. **Server-render the dashboard.** Replace `app/page.tsx` with the Lesson 19
pattern — but importing from the lib, no fetch, no URLs, no secrets:

   ```tsx
   import Link from "next/link";
   import { getProjects } from "@/lib/go";

   export default async function HomePage() {
     const projects = await getProjects();
     return (
       <main>
         <h1>TeamPulse</h1>
         <ul>
           {projects.map((p) => (
             <li key={p.id}>
               <Link href={`/projects/${p.id}`}>{p.name}</Link>{" "}
               <em>({p.status})</em>
             </li>
           ))}
         </ul>
       </main>
     );
   }
   ```

Add `app/loading.tsx` (`<p>Loading TeamPulse…</p>`) and `app/error.tsx` (client, `error`+`reset`, per Lesson 21). Kill Go, reload `/`, restart Go, hit your retry button. Expected: graceful both ways.

5. **Project detail with real 404s.** `app/projects/[id]/page.tsx`:

   ```tsx
   import { notFound } from "next/navigation";
   import { getProject, getTasks, GoError } from "@/lib/go";

   export default async function ProjectDetail({
     params,
   }: {
     params: Promise<{ id: string }>;
   }) {
     const { id } = await params;
     let project, tasks;
     try {
       project = await getProject(id);
       tasks = await getTasks(id);
     } catch (err) {
       if (err instanceof GoError && err.status === 404) notFound();
       throw err;
     }
     return (
       <main>
         <h1>{project.name}</h1>
         <p>{project.description}</p>
         <h2>Tasks</h2>
         <ul>
           {tasks.map((t) => (
             <li key={t.id}>
               {t.title} — {t.status} (P{t.priority}, {t.assignee})
             </li>
           ))}
         </ul>
       </main>
     );
   }
   ```

Plus `app/projects/[id]/not-found.tsx` from Lesson 21. `/projects/p9` → your 404 UI. `/projects/p1` → three tasks (t1, t2, t5 + any you created).

6. **The only browser-side API call.** Create the client form component and
mount it on the detail page:

   ```tsx
   // app/projects/[id]/create-task.tsx
   "use client";

   import { useState } from "react";
   import { useRouter } from "next/navigation";

   export function CreateTask({ projectId }: { projectId: string }) {
     const [title, setTitle] = useState("");
     const [busy, setBusy] = useState(false);
     const [msg, setMsg] = useState<string | null>(null);
     const router = useRouter();

     async function submit(e: React.FormEvent) {
       e.preventDefault();
       setBusy(true);
       setMsg(null);
       const res = await fetch("/api/tasks", {
         method: "POST",
         headers: { "Content-Type": "application/json" },
         body: JSON.stringify({ projectId, title, priority: 3 }),
       });
       setBusy(false);
       if (!res.ok) {
         const data = (await res.json().catch(() => null)) as { error?: string } | null;
         setMsg(data?.error ?? `failed (${res.status})`);
         return;
       }
       setTitle("");
       router.refresh(); // re-run server components → re-fetch through the BFF
     }

     return (
       <form onSubmit={submit}>
         <input value={title} onChange={(e) => setTitle(e.target.value)}
           placeholder="new task title" />
         <button disabled={busy || !title}>Add</button>
         {msg && <p role="alert">{msg}</p>}
       </form>
     );
   }
   ```

Add `import { CreateTask } from "./create-task";` and `<CreateTask projectId={id} />` under the task list in the page. Add a task in the UI. Expected: **the new task appears without a manual refresh** — POST → your handler → lib → Go 201 → `router.refresh()` → page re-ran → list re-fetched. Open DevTools → Network during this: the *only* XHR is a same-origin `POST /api/tasks`. Zero requests to `:8080`.

7. **Retire the P1 client fetches.** Your [Lesson
16](16-p1-teampulse-explorer-page.md) explorer still calls Go directly from the browser. Delete its direct-`:8080` fetch code (the page can keep its layout; gut the data path) — or delete the page outright. P2's rule admits no exceptions.

## Your turn

1. Run the acceptance checklist; every line must be ✅ before Lesson 23:

   - [ ] `grep -rn "8080" app/ lib/` → **zero matches** (only `.env.local` knows the port)
   - [ ] `grep -rn "demo" app/ lib/` → only `lib/go.ts`
   - [ ] `grep -rn "NEXT_PUBLIC" . --include="*.ts" --include="*.tsx" | grep -v node_modules` → zero matches
   - [ ] `curl -s localhost:3000/api/tasks?projectId=p3` → `[]` (empty array, 200)
   - [ ] Browser Network tab while browsing + creating a task → no `:8080`, no `tok-` string in any response body
   - [ ] With Go stopped: `/`, `/projects/p1`, and the form all show *your* error UI, not a white screen or stack trace

   <details><summary>Solution</summary>
If a box is unchecked: 8080 leaks mean a page bypassed `lib/go.ts`; a `demo` outside `lib/go.ts` is a copy-paste leftover (delete it — that string existing in a browser-imported file is the lesson-20 landmine); white screens mean a segment lacks its `error.tsx` (nearest-boundary rule, Lesson 21). Fix and re-run all six.
   </details>

2. Add `GET /api/health` as a *passthrough to Go's* `/api/health` (no auth)
via a new exported function in `lib/go.ts` (note: `go()` adds an auth header and Go's health needs none — it works anyway, but a separate no-auth path is cleaner; decide and defend your choice in one sentence).

   <details><summary>Solution</summary>

   ```ts
   // add to lib/go.ts — no login, no Authorization header
   export async function getHealth() {
     const res = await fetch(`${GO}/api/health`, { cache: "no-store" });
     if (!res.ok) throw new GoError(res.status, await res.text());
     return res.json();
   }
   ```

   ```ts
   // app/api/health/route.ts
   import { NextResponse } from "next/server";
   import { getHealth } from "@/lib/go";

   export async function GET() {
     try {
       return NextResponse.json(await getHealth());
     } catch {
       return NextResponse.json({ error: "go down" }, { status: 502 });
     }
   }
   ```

Defense for the separate path: auth on a health check means "Go is up but I can't log in" and "Go is down" become the same signal — keep liveness probes dependency-free.
   </details>

3. Harder: `/api/tasks` POST currently trusts any body. Make your handler
reject `priority: 99` with a 400 *without* asking Go (Go happily accepts it — check `main.go`: no priority validation exists). Where does the fix go, and what will eventually replace your hand-rolled checks?

   <details><summary>Answer</summary>
In `app/api/tasks/route.ts`'s `POST`, after JSON parse: `if (body.priority !== undefined && ![1,2,3].includes(body.priority)) return NextResponse.json({ error: "priority must be 1-3" }, { status: 400 });` — HTTP concerns (status codes, request parsing) live in the handler; the lib stays a pure Go client. Hand-rolled `if`s are exactly what zod schemas replace in [Lesson 25](25-typed-contracts-with-zod.md), where the same schema will validate *Go's responses* too.
   </details>

## Checkpoint

1. Name the two mechanisms by which data gets onto `/projects/p1`'s first
paint, and the one by which the newly-created task appears after submit.

   <details><summary>Answer</summary>
First paint: server component (`ProjectDetail`) awaits `getProject` + `getTasks` from `lib/go.ts` during server render → HTML. After submit: `router.refresh()` re-runs those same server components. Both are server renders; there is no client data cache in P2.
   </details>

2. Why is it acceptable for `ProjectDetail` to import `lib/go.ts` directly
instead of fetching `/api/projects/p1`, and what would be unacceptable?

   <details><summary>Answer</summary>
Same process, function call — no network, no secret exposure; the route handlers exist for *external* callers. Unacceptable: the **browser** fetching `:8080`, or Next.js `fetch`ing its own `localhost:3000` (pointless self-HTTP).
   </details>

3. Your `POST /api/tasks` returns 201 with Go's created task, but the page's
list doesn't change. DevTools shows the POST succeeded and `router.refresh()` fired. One likely cause inside your page component?

   <details><summary>Answer</summary>
The page's fetch was cached — e.g. `next: { revalidate }` on the tasks fetch in `lib/go.ts` (default 15 is no-store, so someone added caching). refresh() re-runs the component, but a revalidate window serves the stale Go response.
   </details>

4. Where, in the whole P2 codebase, may the string `demo` appear, and why is
everywhere-else a bug *class* rather than a style nit?

   <details><summary>Answer</summary>
Only `lib/go.ts` (server-only, imported solely by route handlers/server pages). Anywhere else risks crossing a `"use client"` boundary — and below that line it's shipped, greppable browser JS: a secret leak, not lint.
   </details>

5. Trace: user clicks Add. List the four systems the request touches, in
order, and the status code each hands back on success.

   <details><summary>Answer</summary>
Browser → `POST /api/tasks` on Next.js :3000 (validates, 400 possible) → `lib/go.ts` logs in (`/api/login`, 200 + token) → `POST /api/tasks` on Go :8080 → 201 + task → handler `NextResponse.json(task, { status: 201 })` → browser `router.refresh()` → server re-render (200 HTML).
   </details>

## Primary source

[Next.js Docs — `useRouter` (`router.refresh()`)](https://nextjs.org/docs/app/api-reference/functions/use-router)
— P2's novel mechanism is the post-write refresh loop (Lesson 20 already
covered the route-handler contract); this is the authoritative statement of
what `refresh()` actually re-runs and when the UI updates.

## Before next session

Close DevTools. Open `~/teampulse` in your head instead: on paper, trace *your finger* through one click — cursor on the Add button → click → … → updated list on screen — writing every hop (component? boundary? handler? Go?) and marking, at each hop, what runs on the server vs the browser. Bring it to Lesson 23; we'll audit it line by line.
