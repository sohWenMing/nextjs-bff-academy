# Lesson 08 — TypeScript III: generics & utility types

*Phase 1 · Session 08 of 38 · ~2 hours · Prereqs: [Lesson 07](07-typescript-ii-shapes-and-narrowing.md)*

## Why this matters

Open any file in your team's BFF and you will meet `Promise<LoginResponse>`,
`Record<string, Task[]>`, `Partial<Task>`, and `Pick<Project, "id" | "name">`
within the first screen. Generics are TS's way of writing **type-level
functions**: one piece of logic, many shapes. You do not need to *invent*
generics at your level — you need to *read* them fluently and reach for the
built-in utility types instead of hand-copying interfaces. That is exactly
what this lesson drills against the TeamPulse API shapes you built in
[Lesson 07](07-typescript-ii-shapes-and-narrowing.md).

## What you'll be able to do

- You can read any generic signature `f<T>(x: T): Promise<T>` out loud in
  plain English.
- You can use `Partial<T>`, `Required<T>`, `Pick<T,K>`, `Omit<T,K>`,
  `Record<K,V>`, `Awaited<T>` without looking them up.
- You can write a generic fetch helper where the caller supplies the response
  type.
- You can use `import type` and say why teams insist on it.
- You can derive update-payload types from model types instead of duplicating fields.

## Principles

**Generics = parameters for types.** `<T>` declares a placeholder type filled
in at each use site. Read `<T>` as "for some type T, decided by the caller":

```typescript
function firstOr<T>(items: T[], fallback: T): T {
  return items[0] ?? fallback;
}
firstOr([1, 2], 0);        // T inferred as number, returns number
firstOr(["a"], "none");    // T inferred as string
firstOr<Task>([], dummy);  // caller may pin T explicitly when inference can't see it
```

The `Promise<T>` you have been writing since
[Lesson 05](05-async-promises-await-fetch.md) is a generic: "a promise that
delivers a T". Same bracket syntax, same idea, everywhere:
`Array<Task>`, `ApiResult<Project>`, `Record<string, Task>`.

**`Record<K, V>`** is the type of a plain object used as a dictionary: keys of
type `K`, values of type `V`. `Record<string, Task[]>` = object whose keys are
strings and whose values are task arrays — exactly what your
`groupByProject` from Lesson 04's exercises produces. Go programmers: it's
`map[K]V`, and unlike Go, key type is checked.

**Utility types** are the standard library of type transforms — generic types
that compute new types from old ones. The four you'll use daily:

```typescript
interface Task {
  id: string;
  projectId: string;
  title: string;
  status: "todo" | "doing" | "done";
  priority: number;
  dueDate: string;
  assignee: string;
}
Partial<Task>;                  // every field optional — the PATCH/update shape
Pick<Task, "id" | "title">;     // keep only these fields
Omit<Task, "id" | "assignee">;  // drop these fields
Required<Pick<Task, "title" | "priority">>; // the POST body for adding a task, made explicit
```

Why this matters economically: when the Go backend adds a field to `Task`,
you edit the `Task` interface once and every derived type updates itself.
Hand-copied interfaces drift; derived types can't. That is the entire
philosophy — **define each shape once, derive the rest**.

`keyof Task` is the union of a type's property names
(`"id" | "projectId" | "title" | ...`) and `Task["status"]` indexes a property
type (`"todo" | "doing" | "done"`). Both compose with utilities:
`Pick<Task, keyof Pick<Task, "id" | "title">>` is silly, but
`Omit<Task, "id" | "assignee">` is bread and butter — it is precisely the
body shape `POST /api/tasks` wants (well, minus defaults the server fills:
see `main.go` — server sets `status`, `dueDate`, `assignee`).

**Constrained generics.** `<T>` can be limited: `<T extends { id: string }>`
means "any T that at least has an id", letting the body use `x.id` while
staying generic. You'll mostly *read* these in library code (`Record<K extends
string, V>`).

**`import type`.** Two import flavors: values (`import { z } from "zod"`)
and *types only* (`import type { Task } from "./types"`). A type-only import
is guaranteed erased — bundlers/compilers delete it outright, zero runtime
artifact, zero circular-import risk from type files. Team style guides
increasingly enforce `import type` via lint rules; use it whenever you import
nothing but types, which after this lesson will be often.

## Follow along

Same setup as [Lesson 07](07-typescript-ii-shapes-and-narrowing.md):
`~/learn-js/lesson-08`, strict tsconfig, `pnpm add -D typescript tsx`. Create
`types.ts` holding `User`, `Project`, `Task`, and `ApiResult` (copy from
Lesson 07 — this is deliberate spaced repetition).

### 1. Read generics out loud

Create `read-generics.ts`. For each signature, write the English reading as a
comment *before* compiling:

```typescript
declare function getJson<T>(path: string, token?: string): Promise<ApiResult<T>>;
declare function groupBy<T>(items: T[], key: (item: T) => string): Record<string, T[]>;
declare function firstById<T extends { id: string }>(items: T[], id: string): T | undefined;
```

Expected readings: "returns a promise of an ApiResult of whatever T the caller
asks for"; "takes an array of any T and a function that picks a string key
from a T; returns an object mapping strings to arrays of T"; "any T that has
an id; find the first with matching id, or undefined."

### 2. Utility types on the real contract

Create `derive.ts`:

```typescript
import type { Task } from "./types"; // type-only import — erased at build
// What POST /api/tasks actually accepts (see backend-mock/main.go):
type NewTask = Pick<Task, "projectId" | "title" | "priority">;
// What a PATCH-style update might send — everything optional:
type TaskPatch = Partial<Task>;
const draft: NewTask = { projectId: "p1", title: "Write lesson 9", priority: 1 };
const patch: TaskPatch = { status: "doing" }; // partial by design
function describe(x: TaskPatch): string {
  const fields = Object.entries(x).map(([k, v]) => `${k}=${String(v)}`);
  return `patch(${fields.join(", ")})`;
}
console.log(describe(draft)); // patch(projectId=p1, title=Write lesson 9, priority=1)
console.log(describe(patch)); // patch(status=doing)
```

Note `NewTask` is *assignable* to `TaskPatch` (more optional fields accept
fewer present ones) but `Task` is not assignable to `NewTask` until you give
it all three. Run with `pnpm tsx derive.ts` and confirm both lines print.

### 3. Break the contract — feel utilities protect you

Append, one line at a time, running after each:

```typescript
const bad1: NewTask = { projectId: "p1", title: "x" };          // TS2741: priority missing
const bad2: TaskPatch = { status: "in-progress" };              // TS2322: not in the union
const bad3: NewTask = { projectId: "p1", title: "x", priority: 1, status: "todo" }; // TS2353: status not in NewTask
```

`bad3` is the quiet superpower: `Pick`ed payload types *reject* fields the
server doesn't accept, catching "the client sent a field we ignore" bugs at
compile time.

### 4. Generic fetch, final form

Create `client.ts` — the generic helper your P0 CLI will use next session
([Lesson 09](09-p0-teampulse-cli.md)):

```typescript
import type { ApiResult, Task } from "./types";
const API = "http://localhost:8080";
async function api<T>(path: string, init?: RequestInit, token?: string): Promise<ApiResult<T>> {
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
    return { ok: false, error: err instanceof Error ? err.message : "network error" };
  }
}
// usage
const login = await api<{ token: string } >("/api/login", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ username: "wen", password: "demo" }),
});
if (!login.ok) {
  console.error(login.error);
  process.exit(1);
}
const tasks = await api<Task[]>("/api/tasks?projectId=p1", undefined, login.data.token);
if (tasks.ok) {
  for (const t of tasks.data) console.log(`${t.id} ${t.status.padEnd(5)} ${t.title}`);
}
```

Run (backend up): expected — `t1 doing Invoice PDF export`, `t2 todo Retry
payment webhook`, `t5 todo Load test checkout`. Hover `tasks.data` in your
editor: `Task[]`, inferred from the `<Task[]>` you passed. One function, every
endpoint, fully typed.

### 5. Confirm `import type` erases

```bash
pnpm tsc client.ts types.ts --outDir dist --module esnext --target es2022 --moduleResolution bundler
grep -n "types" dist/client.js || echo "no import of types.js at runtime — erased"
```

Expected: `no import of types.js at runtime — erased`. The
`import type { ... }` line vanished from the emitted JS entirely.

## Your turn

1. **Derive everything.** From `Project` (Lesson 07), produce *without
   retyping fields*: `ProjectSummary` (id + name only), `NewProject` (all
   except `createdAt`), `ProjectPatch` (all optional). State which utility
   type each used.

2. **Generic dictionary.** Write `countBy<T>(items: T[], key: (item: T) => string): Record<string, number>`
   and use it to count your tasks by `status`. Expected for the p1 seed data:
   `{ doing: 1, todo: 2 }`.

3. **Transfer.** Using `api<T>` from step 4, add a `createTask(projectId:
   string, title: string, priority: number): Promise<ApiResult<Task>>` that
   POSTs to `/api/tasks` with the `NewTask` payload type as its body shape.
   Verify against the running mock: the returned task has server-filled
   `status: "todo"` and an `assignee`.

<details><summary>Solution 1</summary>

```typescript
import type { Project } from "./types";

type ProjectSummary = Pick<Project, "id" | "name">;       // keep two
type NewProject = Omit<Project, "createdAt">;             // drop one
type ProjectPatch = Partial<Project>;                     // all optional
```

`Pick` / `Omit` / `Partial` respectively. Adding a field to `Project`
automatically appears in `ProjectPatch` and disappears from no one's face
manually.
</details>

<details><summary>Solution 2</summary>

```typescript
function countBy<T>(items: T[], key: (item: T) => string): Record<string, number> {
  const out: Record<string, number> = {};
  for (const item of items) {
    const k = key(item);
    out[k] = (out[k] ?? 0) + 1;
  }
  return out;
}

const byStatus = countBy(tasks.data, (t) => t.status); // { doing: 1, todo: 2 }
```

`T` is inferred as `Task` from the argument — you never wrote `Task` in the
function. That's generics doing their job invisibly.
</details>

<details><summary>Solution 3</summary>

```typescript
type NewTask = Pick<Task, "projectId" | "title" | "priority">;

async function createTask(
  projectId: string,
  title: string,
  priority: number,
  token: string,
): Promise<ApiResult<Task>> {
  const body: NewTask = { projectId, title, priority };
  return api<Task>("/api/tasks", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  }, token);
}

const created = await createTask("p1", "Type everything", 2, login.data.token);
if (created.ok) console.log(created.data);
// { id: 't<random>', projectId: 'p1', title: 'Type everything', status: 'todo',
//   priority: 2, dueDate: '<+7 days>', assignee: 'u1' }
```

The compiler-checked body plus the server-filled defaults = the exact
client/server contract discipline Lesson 25's zod will make runtime-safe too.
</details>

## Checkpoint

1. Read aloud: `function cache<T>(key: string, value: T): Record<string, T>`.
<details><summary>Answer</summary>
"For some type T chosen by the caller: takes a string key and a value of
that T, and returns an object mapping strings to T values." (This signature
is a toy, but the reading is the skill.)
</details>

2. Difference between `Pick<Task, "id" | "title">` and
   `Omit<Task, "projectId" | ...>` — and when each reads better?
<details><summary>Answer</summary>
Pick lists what survives; Omit lists what's dropped. Use Pick when keeping a
few fields (payload types), Omit when removing a few from many (drop
server-owned fields like `id`). Both derive from the single source type, so
both survive schema drift.
</details>

3. What does `Partial<NewTask>` do to `NewTask`'s fields, and what request
   kind does that model?
<details><summary>Answer</summary>
Makes every field optional (`field?: T`). Models a PATCH/partial update where
the client may send any subset.
</details>

4. Why `import type { Task }` instead of `import { Task }`?
<details><summary>Answer</summary>
It declares the import carries zero runtime meaning: compilers/bundlers erase
it unconditionally, preventing accidental runtime edges (and circular imports)
between purely-typed modules.
</details>

5. The mock API returns `GET /api/tasks` as a JSON array. Write the exact
   call to your generic `api` helper to fetch all tasks typed, with token
   `"tok-abc"`.
<details><summary>Answer</summary>

```typescript
const all = await api<Task[]>("/api/tasks", undefined, "tok-abc");
```
</details>

## Primary source

[TypeScript Handbook — Generics](https://www.typescriptlang.org/docs/handbook/2/generics.html)
The official walkthrough from first `T` to constrained generics and inference
guidance — written so that after it you can also *read* the utility types'
own declarations (`type Partial<T> = { [P in keyof T]?: T[P] }`) without
fear. Phase 4's zod + TanStack Query code leans on every page of this chapter.

## Before next session

Copy `client.ts`, `types.ts`, and your `createTask` into a folder
`~/learn-js/p0-draft/` and make `pnpm tsx client.ts` print projects and tasks
end to end against the mock backend. Tomorrow is
[Lesson 09](09-p0-teampulse-cli.md) — the P0 project: you turn these pieces
into a real TypeScript CLI, `teampulse-cli`, and pass the Phase 1 gate.
