# TypeScript cheatsheet — the 20% you use daily

All examples use TeamPulse domain types mirroring `backend-mock` (Go structs: `User{ID,Name,Role}`, `Project{ID,Name,Description,Status,CreatedAt}`, `Task{ID,ProjectID,Title,Status,Assignee,Priority,DueDate}`). Password for `/api/login` is the literal `demo`.

## 1. Annotations

Annotate function params and exported boundaries; let **inference** handle the rest (`const` is never reassigned, `let` is).

```ts
function taskLabel(priority: number, title: string): string {
  return `P${priority} — ${title}`;
}
const projects = await res.json();      // inference (but really: run through zod, §9)
let filter: "todo" | "doing" | "done" = "todo";   // literal union type
```

## 2. `interface` vs `type`

Both describe object shapes. Use `type` by default; `interface` only if your team uses it or you need declaration merging. They interop fine.

```ts
interface Project {
  id: string;
  name: string;
  description: string;
  status: "active" | "archived";
  createdAt: string;          // Go time.Time arrives as an ISO string over JSON
}

type TaskStatus = "todo" | "doing" | "done";   // unions must be `type`
```

## 3. Optionals & undefined

`?` means "may be missing". Missing and `undefined` are the same thing in JSON land.

```ts
interface Task {
  id: string;
  assignee?: string;                    // Go may omit it
}
const name = task.assignee ?? "Unassigned";     // nullish coalescing
const initial = task.assignee?.charAt(0);        // optional chaining → string | undefined
```

Never write `task.assignee.charAt(0)` on an optional — that's TS2532.

## 4. Unions

A value is *one of* several types. You must narrow (checks) before using member-specific stuff.

```ts
type ApiResult<T> = T | null;
function show(t: Task | null) {
  if (t === null) return "no task selected";
  return t.title;                     // narrowed: t is Task here
}
```

## 5. Discriminated unions

The pattern for modeling "either success or failure" from the BFF. One shared literal field (the discriminant) lets TS narrow the whole object.

```ts
type FetchState<T> =
  | { kind: "loading" }
  | { kind: "error"; message: string }
  | { kind: "ok"; data: T };

function view(s: FetchState<Project[]>) {
  switch (s.kind) {
    case "loading": return "Loading…";
    case "error":   return s.message;   // TS knows .message exists
    case "ok":      return s.data.map(p => p.name);
  }
}
```

## 6. Generics as a consumer

You rarely *write* `<T extends ...>` bounds; you *read* them: `Promise<Task[]>` is "a Promise of an array of Task". Read inside-out.

```ts
async function login(username: string, password: string): Promise<{ token: string; user: User }> {
  const res = await fetch("http://localhost:8080/api/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  if (!res.ok) throw new Error(`login failed: ${res.status}`);
  return res.json();                     // res.json() is Promise<any> — validate in real code
}
```

## 7. Utility types

| Type | Meaning | TeamPulse use |
|---|---|---|
| `Partial<Task>` | every field optional | payload for a future PATCH |
| `Pick<Task, "id" \| "title">` | keep these fields | slim list row |
| `Omit<Task, "id">` | drop these fields | `POST /api/tasks` body (Go assigns id) |
| `Record<string, Task[]>` | object with `string` keys, `Task[]` values | tasks grouped by `projectId` |
| `ReturnType<typeof getTasks>` | the return type of a function | reuse without re-typing |

```ts
type NewTask = Omit<Task, "id" | "status" | "assignee" | "dueDate">;
// { projectId: string; title: string; priority: number } — matches POST /api/tasks body
```

## 8. `import type`

Import that exists only at compile time and is fully erased from the bundle. Use it for type-only imports — ESLint teams require it.

```ts
import { z } from "zod";                       // runtime value
import type { Task, Project } from "@/types";  // erased at build
```

## 9. zod: `parse` + `z.infer` — one source of truth

Trust nothing from the Go backend. Schema first, type for free.

```ts
import { z } from "zod";

const ProjectSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string(),
  status: z.enum(["active", "archived"]),
  createdAt: z.string(),
});
type Project = z.infer<typeof ProjectSchema>;   // static type from runtime schema

const projects = z.array(ProjectSchema).parse(await res.json()); // throws if Go lies
```

## 10. `unknown`, not `any`

`any` opts out of checking entirely (a lie you'll pay for at runtime). `unknown` says "I don't know yet" and forces you to check before use. `JSON.parse` and `res.json()` give you `any` — immediately route it through zod as `unknown`.

```ts
function isTask(v: unknown): v is Task {           // type predicate
  return typeof v === "object" && v !== null && "id" in v && "priority" in v;
}
```

## 11. Strict-mode gotchas (`strict: true` is non-negotiable)

| Gotcha | Symptom | Fix |
|---|---|---|
| `noImplicitAny` | TS7006 "implicit any" on a param | annotate it: `(task: Task)` |
| `strictNullChecks` | TS2532 "possibly undefined" | `?.`, `??`, or an `if` guard first |
| `exactOptionalPropertyTypes` (off by default) | `undefined` not assignable to `assignee?: string` | omit the key instead of setting `undefined` |
| array `.find` | returns `T \| undefined` | `const p = projects.find(...); if (!p) return;` |
| `Object.entries` | values typed `{}` / union | assert shape or zod-parse the object |
| `catch (e)` | `e` is `unknown` (TS 4.4+) | `e instanceof Error ? e.message : String(e)` |

## 12. Quick coercion patterns

```ts
const priority = Number(searchParams.get("priority") ?? "1");   // string | null → number
const statuses: TaskStatus[] = ["todo", "doing"];               // array of literals
const byId: Record<string, Project> = Object.fromEntries(projects.map(p => [p.id, p]));
type LoginBody = { username: string; password: "demo" };       // literal type = only that value
```

---
*Related lessons: [06 types & inference](../lessons/06-typescript-i-types-and-inference.md) · [07 shapes & narrowing](../lessons/07-typescript-ii-shapes-and-narrowing.md) · [08 generics & utility types](../lessons/08-typescript-iii-generics-and-utility-types.md) · [25 typed contracts with zod](../lessons/25-typed-contracts-with-zod.md) · [31 decoding errors](../lessons/31-decoding-typescript-and-eslint-errors.md)*
