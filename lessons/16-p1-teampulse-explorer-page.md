# Lesson 16 — [P1] TeamPulse Explorer Page

*Phase 2 · Session 16 of 38 · ~2 hours · Prereqs: Lessons 10–15 (all of Phase 2)*

## Why this matters

This is the phase's payoff: one client-side React page — login form → project
list → task list → live text filter — talking directly to the mock Go API at
`http://localhost:8080`. Everything from this week assembles here. Lesson 22
rips out the browser-side fetching; this UI survives it.

## What you'll be able to do

- You can build a multi-component interactive page from a spec — every piece
  of state in the component that needs it, dependent fetches chained in event
  handlers (login → projects, select → tasks) without a single `useEffect`.
- You can handle every state an async UI has — loading, error (with Retry),
  empty, success — per request, not globally.
- You can keep an auth token in React state and attach it via
  `Authorization: Bearer <token>` on subsequent calls.
- You can diff your architecture sketch from last night against a reference
  design — and defend either version.

## Principles

Today adds no new concepts — only the *composition decisions* a project
session makes for you. Four rules while building:

**1. State lives where the event happens; derived data is computed, never
stored.** The reference design has exactly **six** pieces of state, all in
`Explorer`, because it owns every event:

| State | Why it's state |
|---|---|
| `token: string \| null` | login's outcome; gates the whole view |
| `loginError: string \| null` | failed-login message |
| `projects: Project[] \| null` | loaded after login (`null` = not yet) |
| `selectedProjectId: string \| null` | the user's selection |
| `tasks: Task[] \| null` | loaded per selected project |
| `query: string` | the filter text |

Everything else — the *filtered* list, done-counts, in-flight labels — is
derived in render. Count your sketch's states: more than seven means you
stored something derivable; fewer means something is missing a memory.

**2. Dependent requests are chained in one async handler, or triggered by the
next event.** `GET /api/projects` needs the token, which exists only after
login; tasks load when the user **clicks a project**. No effect watching `token`.

**3. One request = one loading/error budget.** A failed projects fetch must
not blank the task area; Retry retries *that* request, not the page.

**4. The token is React state — on purpose, for now.** Reload loses the
session; deliberately awkward so Lesson 27's cookie session lands as relief.
Anything in `useState` is readable by page scripts — never ship this for real
secrets. Note it, build it anyway.

### Target UX (your acceptance criteria)

```
[logged out]
  TeamPulse heading
  Login form (username, password, submit)   ← password is the literal word demo
  Login error shown in red when the API says 401

[logged in]
  Header: "Hi, Wen" + Logout button
  Projects: list of buttons, selected one highlighted
    ├─ loading → "Loading projects…"   error → message + Retry   empty → "No projects yet."
  Tasks for selected project: rows [STATUS] title — P<n> @assignee
    ├─ loading / error+Retry / empty ("No tasks match…")
  Filter input narrowing tasks by title substring, case-insensitive
    └─ "X of Y tasks" count line
```

## Follow along

Verify first: Terminal 1 runs `cd backend-mock && go run main.go` **with the
CORS patch from Lesson 15 step 2** — skip it and every fetch fails red.
Terminal 2: `cd ~/teampulse && pnpm dev`. Keep `lib/types.ts` and
`components/login-form.tsx`; stop importing `lib/data.ts` — data is the API's.

### 1. The page shell + Explorer

Replace `app/page.tsx` — this is the reference design; type it, don't paste:

```tsx
"use client";

import { useState } from "react";
import { LoginForm } from "@/components/login-form";
import { ProjectPicker } from "@/components/project-picker";
import { TaskBoard } from "@/components/task-board";
import type { Project, Task } from "@/lib/types";

const API = "http://localhost:8080";

export default function Home() {
  const [token, setToken] = useState<string | null>(null); const [loginError, setLoginError] = useState<string | null>(null);
  async function handleLogin(username: string, password: string) {
    setLoginError(null);
    try {
      const res = await fetch(`${API}/api/login`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ username, password }), });
      const body = (await res.json()) as { token?: string; error?: string };
      if (!res.ok || !body.token) {
        setLoginError(body.error ?? `HTTP ${res.status}`);
        return;
      }
      setToken(body.token);
    } catch {
      setLoginError("Backend unreachable — is 'go run main.go' running?");
    }
  }
  if (!token) {
    return (
      <main>
        <h1>TeamPulse</h1>
        <LoginForm onLogin={handleLogin} />
        {loginError && <p style={{ color: "crimson" }}>{loginError}</p>}
      </main>
    );
  }

  return (
    <main>
      <h1>TeamPulse</h1>
      <Explorer token={token} onLogout={() => setToken(null)} />
    </main>
  );
}

function Explorer({ token, onLogout }: { token: string; onLogout: () => void }) {
  const [projects, setProjects] = useState<Project[] | null>(null); const [projectsError, setProjectsError] = useState<string | null>(null);
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null); const [tasks, setTasks] = useState<Task[] | null>(null);
  const [tasksError, setTasksError] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  async function loadProjects() {
    setProjectsError(null); setProjects(null);
    try {
      const res = await fetch(`${API}/api/projects`, { headers: { Authorization: `Bearer ${token}` }, });
      if (!res.ok) { const b = (await res.json()) as { error?: string }; throw new Error(b.error ?? `HTTP ${res.status}`); }
      setProjects((await res.json()) as Project[]);
    } catch (err) {
      setProjectsError(err instanceof Error ? err.message : "projects failed");
    }
  }
  async function handleSelectProject(id: string) {
    setSelectedProjectId(id); setTasks(null);
    setTasksError(null); setQuery("");
    try {
      const res = await fetch(
        `${API}/api/tasks?projectId=${encodeURIComponent(id)}`,
        { headers: { Authorization: `Bearer ${token}` } },
      );
      if (!res.ok) { const b = (await res.json()) as { error?: string }; throw new Error(b.error ?? `HTTP ${res.status}`); }
      setTasks((await res.json()) as Task[]);
    } catch (err) {
      setTasksError(err instanceof Error ? err.message : "tasks failed");
    }
  }
  const visible =
    tasks === null
      ? null
      : tasks.filter((t) => t.title.toLowerCase().includes(query.toLowerCase()));

  return (
    <>
      <header>
        <p>Hi, Wen</p>
        <button onClick={onLogout}>Logout</button>
      </header>
      {projectsError && (
        <p style={{ color: "crimson" }}>
          {projectsError} <button onClick={loadProjects}>Retry</button>
        </p>
      )}
      {projects === null && !projectsError && (
        <button onClick={loadProjects}>Load projects</button>
      )}
      {projects !== null && (
        <ProjectPicker
          projects={projects}
          selectedId={selectedProjectId}
          onSelect={handleSelectProject}
        />
      )}
      <TaskBoard
        visible={visible}
        total={tasks?.length ?? 0}
        query={query}
        onQueryChange={setQuery}
        error={tasksError}
        onRetry={() => selectedProjectId && handleSelectProject(selectedProjectId)}
      />
    </>
  );
}
```

Two design notes for the rubber duck:

- `Explorer` is a **second component in the same file**, rendered only after
  the `if (!token)` early return — hooks always run with a valid token. Logout
  unmounts it and all six states die with it: Lesson 13's component-owned
  state does session cleanup for free.
- `visible` is derived (Lesson 12); clearing `query` on project switch is a
  deliberate UX call — filters are per-view.

### 2. The two child components

`components/project-picker.tsx` — props-only, no state, no directive of its
own (its parent is the client boundary):

```tsx
import type { Project } from "@/lib/types";

export function ProjectPicker({
  projects,
  selectedId,
  onSelect,
}: {
  projects: Project[];
  selectedId: string | null;
  onSelect: (id: string) => void;
}) {
  if (projects.length === 0) return <p>No projects yet.</p>;
  return (
    <nav aria-label="Projects">
      <h2>Projects</h2>
      <ul style={{ listStyle: "none", padding: 0 }}>
        {projects.map((p) => (
          <li key={p.id}>
            <button
              onClick={() => onSelect(p.id)}
              style={{ fontWeight: p.id === selectedId ? 700 : 400 }}
            >
              {p.name}
              {p.status === "archived" && " [ARCHIVED]"}
            </button>
          </li>
        ))}
      </ul>
    </nav>
  );
}
```

`components/task-board.tsx` — the row-level view; all state and fetching stay
in `Explorer`. It takes the nullable `visible` plus a separate `total`: one
state, two honest outputs:

```tsx
interface TaskRowData {
  id: string;
  title: string;
  status: "todo" | "doing" | "done";
  assignee: string;
  priority: number;
}

export function TaskBoard({
  visible,
  total,
  query,
  onQueryChange,
  error,
  onRetry,
}: {
  visible: TaskRowData[] | null; // null = not loaded yet
  total: number;
  query: string;
  onQueryChange: (q: string) => void;
  error: string | null;
  onRetry: () => void;
}) {
  if (error) {
    return (
      <p style={{ color: "crimson" }}>
        {error} <button onClick={onRetry}>Retry</button>
      </p>
    );
  }
  if (visible === null) return <p>Select a project to see tasks.</p>;
  return (
    <section>
      <h2>Tasks</h2>
      <input
        value={query}
        onChange={(e) => onQueryChange(e.target.value)}
        placeholder="Filter by title…"
      />
      <p>
        {visible.length} of {total} tasks
      </p>
      {visible.length === 0 ? (
        <p>No tasks match{query ? ` “${query}”` : ""}.</p>
      ) : (
        <ul style={{ listStyle: "none", padding: 0 }}>
          {visible.map((t) => (
            <li key={t.id}>
              [{t.status.toUpperCase()}] {t.title} — P{t.priority} @{t.assignee}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
```

### 3. Click every acceptance path (the rubric)

Devtools Network tab open the whole time:

1. Empty username → submit → Lesson 14's validation message; **no `login`
   entry in Network**.
2. Password `wrong` → red `invalid credentials` (the backend's own message).
3. Any username + demo password → signed-in view: `Hi, Wen`, Logout, `Load
   projects`.
4. `Load projects` → project buttons render from `GET /api/projects`.
5. Click a project → task rows render, count line correct, button bold.
6. Filter `load` → only matching rows, count reflects it; `zzz` → `No tasks
   match “zzz”.`, count `0 of N`.
7. Switch projects → query cleared; no stale rows from the old project.
8. **Error paths:** stop the Go server → `Failed to fetch` with a working
   Retry after restart. Logout → login form; log in again → clean slate.
9. Token discipline: the tasks request carries
   `Authorization: Bearer tok-...`; the login POST payload shows
   `{"username":"wen","password":"demo"}`.

If any box fails, fix it now — this rubric *is* P1.

### 4. The double-request audit

Click between two projects fast: one `tasks` request per *click* is fine;
typing in the filter must produce **zero** requests — a request there means
you wired fetching to the wrong place. Previous project's rows flashing = the
Lesson 15 race; `setTasks(null)` on select fixes it cheaply.

## Your turn

Solo, 30–40 minutes.

1. **Harden projects loading:** replace the `Load projects` button with the
   *one* sanctioned fetch-on-mount shape (Lesson 15): `useEffect` +
   `AbortController` + cleanup + `aborted` check — the only effect allowed in
   P1. StrictMode doubles the request in dev: why tolerable, what guards?

2. **Priority filter:** a second control (All / urgent / low buttons or a
   `<select>`) narrowing by `priority` (`urgent` ≤ 2, `low` = 3). Composes
   with the text filter, stays derived (one new state, zero new fetches),
   resets on project switch; count line reflects both.

3. **Meet `/api/flaky`:** a `Ping /api/flaky` button in the header —
   authorized GET (~30% deliberate 500s, see `backend-mock/README.md`),
   loading/error *widget-local* (Rule 3), showing the backend's `upstream
   hiccup...` message, Retry re-pings — one red click in three, never a
   blank page.

<details><summary>Solution</summary>

1. In `Explorer`:

   ```tsx
   useEffect(() => {
     const controller = new AbortController();
     (async () => {
       try {
         const res = await fetch(`${API}/api/projects`, { headers: { Authorization: `Bearer ${token}` }, signal: controller.signal, });
         if (!res.ok) { const b = (await res.json()) as { error?: string }; throw new Error(b.error ?? `HTTP ${res.status}`); }
         setProjects((await res.json()) as Project[]);
       } catch (err) {
         if (controller.signal.aborted) return;
         setProjectsError(err instanceof Error ? err.message : "projects failed");
       }
     })();
     return () => controller.abort();
   }, [token]);
   ```

   (Add `useEffect` to the react import.) The doubled dev requests are
StrictMode's fake remount; the abort means only the live mount's result lands.

2. One more state in `Explorer` — `const [priority, setPriority] =
   useState<"all" | "urgent" | "low">("all")` — reset in `handleSelectProject`,
   chained into `visible`:

   ```tsx
   .filter((t) =>
     priority === "all" ? true : priority === "urgent" ? t.priority <= 2 : t.priority === 3,
   )
   ```

   Buttons call `setPriority`; bold the active one like `ProjectPicker`.

3. `components/flaky-ping.tsx`, mirroring `HealthWidget`'s discriminated
   union, pointed at `…/api/flaky` with the token header, plus
   `{state.kind === "error" && <button onClick={ping}>Retry</button>}` beside
   the message; render `<FlakyPing token={token} />` in the header. The ~30%
   failures are the Go handler's deliberate `rand.Int` check — Lesson 28
   builds retry policy out of it.

</details>

## Checkpoint

1. Your P1 page has at most one `useEffect`. Explain the chain from a
   project button click to rows rendering — name every `setState`.
2. Why does `setToken(null)` erase the task list with no cleanup code for `tasks`?
3. The filter box triggers zero network requests. Give the two-line reason,
   and name the design mistake it would be if it *did* fetch.
4. The projects area renders three distinct states — which? Why does
   `tasks === null` structurally prevent stale rows from the previous project?
5. `GET /api/tasks?projectId=p1` returned 401 right after a successful login.
   Give two distinct causes diagnosable from devtools Network alone — one
   about the request, one about the backend's memory.

<details><summary>Answers</summary>

1. `onClick` → `onSelect(p.id)` → `handleSelectProject`: `setSelectedProjectId(id)`
   (bolds the button), `setTasks(null)` + `setTasksError(null)` +
   `setQuery("")`, then `await` fetch with the auth header → `setTasks(rows)`.
   The batched re-render recomputes `visible`; `TaskBoard` renders it.
2. `Explorer` renders only while `token` is truthy; `setToken(null)` makes
   `Home` return the login form, so `Explorer` **unmounts** and its state dies
   with the instance. A fresh login gets fresh initial state.
3. `visible` is computed from `tasks` + `query` during render; typing only
   updates `query`. Fetching instead would mean treating **derived data as
   server data** — re-requesting instead of recomputing.
4. `null` + no error → loading/`Load projects`; error → message + Retry;
   array → picker (`[]` → "No projects yet."). Every load path sets
   `tasks = null` before awaiting, so old rows can't display.
5. (Request) The GET went out with a missing/malformed `Authorization`
   header — visible in Network → Request headers. (Backend memory) The Go
   server restarted between login and tasks: tokens are in-memory, the old
   token no longer exists — same 401, perfect headers; fix = re-login.

</details>

## Primary source

[Next.js — Client Components](https://nextjs.org/docs/app/building-your-application/rendering/client-components)
— what `"use client"` marks: a boundary, not an annotation (Lesson 18 formalizes).

## Before next session

Phase 3 starts: the App Router proper. 5-minute prep — in `~/teampulse`,
create `app/about/page.tsx` with a trivial component, visit `/about`, and
note: (a) the error when `"use client"` is removed while a `useState`
remains — paste it, (b) the URL-to-folder mapping. Lesson 17 builds on both.
