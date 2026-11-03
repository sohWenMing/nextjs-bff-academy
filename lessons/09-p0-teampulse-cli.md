# Lesson 09 — [P0] TeamPulse CLI

*Phase 1 · Session 09 of 38 · ~2 hours · Prereqs: Lessons 01–08, all of them*

## Why this matters

This is the Phase 1 consolidation project: you build `teampulse-cli`, a
TypeScript command-line client for the TeamPulse mock Go API, entirely from
your own typing. Logging in, carrying a bearer token, listing projects and
tasks, and creating a task is the exact request lifecycle your BFF route
handlers will perform in Phase 3, minus the browser. If this works and you can
explain every line, you have the foundation the rest of the course stands on.

## What you'll be able to do

- You can scaffold a strict-TypeScript project (tsconfig, tsx, scripts) from
  memory in under five minutes.
- You can build and explain a generic `request<T>()` client that logs in,
  attaches the token, and normalizes errors into `ApiResult<T>`.
- You can parse `process.argv` into commands via a discriminated union and
  dispatch on it.
- You can add a new endpoint call to the CLI (the phase-gate task) unaided.

## Principles

**The architecture** — three files, one job each, imports flowing one
direction ([Lesson 04](04-modules-and-bundler-mental-model.md) made real):
`src/types.ts` holds shapes + `ApiResult`, `src/api.ts` holds `login()` + the
generic `request<T>()`, `src/index.ts` parses argv, dispatches, and prints.

**Login is the spine.** Every data command calls `login()` first (the mock
accepts any username; the password is literally `demo` — see
`backend-mock/main.go`'s demo rule), keeps the token in a local variable, and
passes it to `request<T>()` as a bearer header. In Phase 3 the BFF does the
same dance server-side with a cookie; the mechanics transfer directly.

**Parse argv into a discriminated union, then narrow.** Command-line input is
untrusted strings; the pattern (from
[Lesson 07](07-typescript-ii-shapes-and-narrowing.md)) is a `parseArgs`
function returning a tagged union which `main()` narrows with
`switch (cmd.kind)`, so the compiler guarantees each branch only touches
fields that exist for that command.

**Two stated decisions.** Errors set `process.exitCode = 1` — a CLI that
prints an error but exits 0 lies to shell scripts and CI. And we re-login on
every command instead of caching a token: the mock's tokens die on restart, so
caching buys nothing but complexity. Simplicity is a feature.

## Follow along

### 0. Start the backend (keep it running all session)

```bash
cd ~/workspace/nextjs-bff-academy/backend-mock && go run main.go
# ... TeamPulse mock backend listening on http://localhost:8080
```

### 1. Scaffold

```bash
mkdir -p ~/teampulse-cli/src && cd ~/teampulse-cli
pnpm init                 # accept defaults
pnpm add -D typescript tsx
```

Create `tsconfig.json` (same strict settings that carried Lessons 06–08):

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "outDir": "dist"
  }
}
```

Edit `package.json`: add `"type": "module",` at the top level (marks the
project ESM — required for `pnpm build` output to run under Node) and set
scripts to:

```json
"scripts": {
  "start": "tsx src/index.ts",
  "check": "tsc --noEmit",
  "build": "tsc"
}
```

### 2. `src/types.ts` — the contract, typed

Type this from the README's data model **from memory**, then check against
`backend-mock/main.go`:

```typescript
export interface User { id: string; name: string; role: string; }
export interface Project {
  id: string;
  name: string;
  description: string;
  status: "active" | "archived";
  createdAt: string;
}
export interface Task {
  id: string;
  projectId: string;
  title: string;
  status: "todo" | "doing" | "done";
  priority: number;
  dueDate: string;
  assignee: string;
}
export interface LoginResponse { token: string; user: User; }
export type NewTask = Pick<Task, "projectId" | "title" | "priority">;
export type ApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; error: string };
```

### 3. `src/api.ts` — one generic, one login

```typescript
import type { ApiResult, LoginResponse, NewTask, Project, Task, User } from "./types";
const API = process.env.TEAMPULSE_API ?? "http://localhost:8080";

async function request<T>(
  path: string,
  init?: RequestInit,
  token?: string,
): Promise<ApiResult<T>> {
  try {
    const res = await fetch(`${API}${path}`, {
      ...init,
      headers: {
        ...(init?.headers ?? {}),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
    });
    if (!res.ok) {
      const body = (await res.json().catch(() => null)) as { error?: string } | null;
      return { ok: false, error: body?.error ?? `HTTP ${res.status}` };
    }
    return { ok: true, data: (await res.json()) as T };
  } catch (err) {
    const why = err instanceof Error ? err.message : "network error";
    return { ok: false, error: `is the mock backend running? (${why})` };
  }
}

export async function login(username = "wen"): Promise<ApiResult<string>> {
  const res = await request<LoginResponse>("/api/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password: "demo" }),
  });
  if (!res.ok) return res; // same ApiResult shape, error forwarded
  return { ok: true, data: res.data.token };
}

export const getProjects = (token: string) =>
  request<Project[]>("/api/projects", undefined, token);

export const getTasks = (token: string, projectId?: string) =>
  request<Task[]>(`/api/tasks${projectId ? `?projectId=${projectId}` : ""}`, undefined, token);

export const addTask = (token: string, body: NewTask) =>
  request<Task>("/api/tasks", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  }, token);

export const getMe = (token: string) => request<User>("/api/me", undefined, token);
```

Review before typing (then type it — do not paste): one `request<T>` for every
endpoint, so token injection, `ok` checking, and Go's `{"error":...}`
normalization happen once
([Lesson 08](08-typescript-iii-generics-and-utility-types.md)); `login` maps
`LoginResponse → token` but forwards the failure arm as-is; and
`process.env.TEAMPULSE_API ?? ...` repoints the CLI without editing code.

### 4. `src/index.ts` — argv → discriminated union → dispatch

```typescript
import { addTask, getMe, getProjects, getTasks, login } from "./api";
import type { NewTask } from "./types";

type Command =
  | { kind: "projects" }
  | { kind: "tasks"; projectId?: string }
  | { kind: "add"; projectId: string; title: string; priority: number }
  | { kind: "me" }
  | { kind: "bad"; reason: string };

function parseArgs(argv: string[]): Command {
  const [command, ...rest] = argv;
  switch (command) {
    case undefined:
    case "help":
      return { kind: "bad", reason: "usage: teampulse <projects|tasks [p1]|add <p1> <title> <1-3>|me>" };
    case "projects":
      return { kind: "projects" };
    case "tasks":
      return { kind: "tasks", projectId: rest[0] };
    case "me":
      return { kind: "me" };
    case "add": {
      const [projectId, title, priorityRaw] = rest;
      const priority = Number(priorityRaw);
      if (!projectId || !title || !Number.isInteger(priority) || priority < 1 || priority > 3) {
        return { kind: "bad", reason: "add needs <projectId> <title> <priority 1-3>" };
      }
      return { kind: "add", projectId, title, priority };
    }
    default:
      return { kind: "bad", reason: `unknown command: ${command}` };
  }
}

function fail(error: string): void {
  console.error("api error:", error);
  process.exitCode = 1;
}

async function main(): Promise<void> {
  const cmd = parseArgs(process.argv.slice(2));

  if (cmd.kind === "bad") {
    console.error(cmd.reason);
    process.exitCode = 1;
    return;
  }

  const session = await login("wen");
  if (!session.ok) {
    console.error("login:", session.error);
    process.exitCode = 1;
    return;
  }
  const token = session.data;

  switch (cmd.kind) {
    case "projects": {
      const res = await getProjects(token);
      if (!res.ok) return fail(res.error);
      for (const p of res.data) console.log(`${p.id}  ${p.status.padEnd(8)} ${p.name}`);
      return;
    }
    case "tasks": {
      const res = await getTasks(token, cmd.projectId);
      if (!res.ok) return fail(res.error);
      if (res.data.length === 0) {
        console.log(cmd.projectId ? `no tasks for ${cmd.projectId}` : "no tasks at all");
        return;
      }
      for (const t of res.data) {
        console.log(`${t.id}  ${t.status.padEnd(5)} P${t.priority}  ${t.title}  (${t.projectId})`);
      }
      return;
    }
    case "add": {
      const body: NewTask = { projectId: cmd.projectId, title: cmd.title, priority: cmd.priority };
      const res = await addTask(token, body);
      if (!res.ok) return fail(res.error);
      console.log(`created ${res.data.id}: "${res.data.title}" in ${res.data.projectId} (assigned to ${res.data.assignee})`);
      return;
    }
    case "me": {
      const res = await getMe(token);
      if (!res.ok) return fail(res.error);
      console.log(`${res.data.name} (${res.data.role}) — id ${res.data.id}`);
      return;
    }
  }
}

main();
```

### 5. Run the whole thing

Predict each output before running it:

```bash
cd ~/teampulse-cli
pnpm check                    # type-check: silence is success
pnpm start projects
# p1  active   Billing Service
# p2  active   Mobile App
# p3  archived Legacy Portal
pnpm start tasks p1
pnpm start tasks p999         # no tasks for p999 (API returns [], not 404)
pnpm start add p1 "Pass the phase gate" 1
pnpm start tasks p1           # your new task is there, status todo
pnpm start me
pnpm start bogus; echo $?     # unknown command: bogus — exit code 1
```

### 6. Kill the backend — resilience check

Stop the Go server, run `pnpm start projects`: expected `login: is the mock
backend running? (fetch failed)` with exit 1 — no stack trace, no unhandled
rejection. Restart the server and confirm recovery. This clean failure was
three deliberate decisions: `try/catch` around `fetch`, `ApiResult` instead of
throwing, `exitCode` in the CLI. Any one missing = ugly failure.

## Your turn

1. **Sort command.** Add an optional trailing `--sort=priority` flag to
   `tasks`: parse it in `parseArgs` (add a `sort?: "priority"` field to the
   `tasks` arm) and order output by priority ascending without mutating
   ([Lesson 03](03-arrays-objects-immutability.md)).

2. **Table it.** `projects` prints columns that drift with name lengths.
   Write `table(rows: string[][])` that computes each column's width from its
   content and prints aligned rows; use it for `projects` and `tasks`.

3. **The phase gate.** Under 15 minutes, timed: add a `whoami` command that
   calls `GET /api/me` and prints `name — role`. (The syllabus gate names the
   endpoint `/api/whoami`; the mock implements it as `/api/me` — same thing,
   different label; see `backend-mock/README.md`.) Type-check with `pnpm
   check` before stopping the clock. Over the limit? Redo it — the point is
   fluency, not the ticket.

<details><summary>Solution 1</summary>

```typescript
// Command union arm: | { kind: "tasks"; projectId?: string; sort?: "priority" }
// parseArgs, tasks arm:
case "tasks": {
  const sort = rest.includes("--sort=priority") ? ("priority" as const) : undefined;
  return { kind: "tasks", projectId: rest[0], sort };
}
// tasks branch of main():
const rows = cmd.sort === "priority"
  ? [...res.data].sort((a, b) => a.priority - b.priority)
  : res.data;
```

`as const` pins the literal type `"priority"`; `[...res.data]` copies before
sorting.
</details>

<details><summary>Solution 2</summary>

```typescript
function table(rows: string[][]): void {
  if (rows.length === 0) return;
  const widths = rows[0].map((_, col) =>
    Math.max(...rows.map((r) => (r[col] ?? "").length)),
  );
  for (const row of rows) {
    console.log(row.map((cell, i) => (cell ?? "").padEnd(widths[i])).join("  "));
  }
}

// usage in the projects branch:
table(res.data.map((p) => [p.id, p.status, p.name]));
```

`noUncheckedIndexedAccess` made the `?? ""` guards necessary — `widths[i]`
really could be missing if a row had extra cells.
</details>

<details><summary>Solution 3</summary>

Three additions, each pattern-matched from existing code:

```typescript
// Command union:      | { kind: "whoami" }
// parseArgs:          case "whoami": return { kind: "whoami" };
// main() dispatch:
case "whoami": {
  const res = await getMe(token);
  if (!res.ok) return fail(res.error);
  console.log(`${res.data.name} — ${res.data.role}`);
  return;
}
```

That is what "can operate" looks like at this layer. If any step felt
foreign, redo [Lesson 07](07-typescript-ii-shapes-and-narrowing.md) first.
</details>

## Checkpoint

1. Why does `request<T>` check `res.ok` itself, and return `ApiResult<T>`
   instead of throwing?
<details><summary>Answer</summary>
Checked once, so no endpoint can forget it (fetch doesn't reject on HTTP
errors). `ApiResult<T>` makes failure a *value* callers must handle at compile
time; a thrown error is an invisible hole in the signature.
</details>

2. Where does the token come from, where does it live, and why is that fine
   for a CLI but not a browser?
<details><summary>Answer</summary>
From `POST /api/login` (password `demo`); a local variable for one process,
re-fetched per command. Fine on your own machine; bearer tokens in browser JS
are an XSS theft target — Phase 3 uses an httpOnly cookie instead.
</details>

3. What does `parseArgs` return for `["add", "p1", "x", "9"]`, and why didn't
   the types catch it?
<details><summary>Answer</summary>
`{ kind: "bad", reason: "add needs <projectId> <title> <priority 1-3>" }` —
9 fails the range guard. `NewTask.priority: number` doesn't constrain ranges,
and types are erased at runtime ([Lesson 06](06-typescript-i-types-and-inference.md)),
so argv validation must be real code.
</details>

4. `pnpm check` passes but `pnpm build && node dist/index.js projects` fails
   to resolve `./types`. What broke?
<details><summary>Answer</summary>
ESM extensions: `tsc` emits `from "./types"` extension-less and Node ESM
requires `"./types.js"`; `moduleResolution: "bundler"` only guides the
type-checker, Node enforces its own rules at runtime. Fix: write `.js`
extensions in imports, or run via tsx.
</details>

5. Against the mock API: `pnpm start add p999 "ghost task" 2` — what happens,
   and should it? Check `main.go` before answering.
<details><summary>Answer</summary>
It succeeds (201) — `handleTasks` only checks `projectId`/`title` are
non-empty, never that the project exists. A real bug worth a ticket: the Go
backend (or a BFF pre-check of `GET /api/projects/p999`) should enforce
referential integrity — Lesson 24's contract-reading skill, early.
</details>

## Primary source

[Node.js — `process` reference](https://nodejs.org/api/process.html)
Your CLI's whole interface to the outside world — `argv`, `env`, `exitCode` —
with exact semantics and gotchas (like `exit` truncating pending output vs
`exitCode` exiting gracefully).

## Before next session

Commit the CLI (`.gitignore` with `node_modules/` and `dist/` first, then
`git init && git add -A && git commit -m "P0: TeamPulse CLI"`). Then skim the
rest of `backend-mock/README.md` — from Lesson 10 the same API gains a
browser front-end and JSX finally appears: read ahead the first page of the
official Next.js "Getting Started: Installation" guide and bring any term you
could not parse to session 10.
