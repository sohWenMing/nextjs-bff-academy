# Lesson 07 — TypeScript II: shapes & narrowing

*Phase 1 · Session 07 of 38 · ~2 hours · Prereqs: [Lesson 05](05-async-promises-await-fetch.md), [Lesson 06](06-typescript-i-types-and-inference.md)*

## Why this matters

The BFF's job is transforming *shapes*: Go structs in, UI-shaped objects out.
In TS you express shapes with `interface`/`type`, and handle the world's
messiness — missing fields, either/or responses, unparseable JSON — with
**narrowing**: code the compiler can *follow* to prove a value is safe to use.
Getting `unknown` right here stops you poisoning the codebase with `any`.

## What you'll be able to do

- You can define `interface`/`type` shapes for the mock API's entities cold.
- You can use optional (`?`) properties and read what `string | undefined` means.
- You can write union types and discriminate on a tag field.
- You can narrow with `typeof`, `in`, truthiness, and a custom type-guard function.
- You can explain why `unknown` beats `any` in one sentence and prove it.

## Principles

**Shapes are structural.** TypeScript checks compatibility by *shape*, not
name — an object with the required properties satisfies the type. `type` and
`interface` both name a shape; the differences are academic at your stage
(teams typically use `interface` for extensible object shapes, `type` for
unions). Both erase at build time
([Lesson 06](06-typescript-i-types-and-inference.md)).

The mock API's contract, mirrored in TS — compare with `backend-mock/main.go`:

```typescript
interface User { id: string; name: string; role: string; } // role: "admin" | "member"
interface Project {
  id: string;
  name: string;
  description: string;
  status: "active" | "archived"; // string literal union: only these two strings
  createdAt: string;             // JSON turns Go's time.Time into a string
}
interface Task {
  id: string;
  projectId: string;
  title: string;
  status: "todo" | "doing" | "done";
  priority: number;
  dueDate: string;
  assignee: string;
}
```

The `"active" | "archived"` is a **literal union** — the type system encoding
the API contract: assign `"deletd"` (typo) and the compiler stops you at
keystroke, where Go would catch it at runtime or in review.

**Optional properties.** `?` marks a field that may be absent; its type
becomes `T | undefined` and you must handle both. Use it for genuinely
sometimes-missing fields, never as a modelling shrug.

**Unions** mean "one of these": `string | number`, `null | Task`. The value
starts as all possibilities; **narrowing** refines it to one, with the
compiler tracking your proof:

```typescript
function label(x: string | number): string {
  if (typeof x === "string") {
    return x.toUpperCase();   // here x is narrowed to string
  }
  return x.toFixed(0);        // and here, by exclusion, to number
}
```

`typeof`, `instanceof`, `in` (`"projectId" in obj`), truthiness
(`if (task)` narrows out `null | undefined`), and literal equality are the
tools. The compiler carrying these proofs through branches is *control-flow
analysis* — write code it can follow: early returns, `const` bindings.

**Discriminated unions** — the pattern for "a response is one of several
kinds, tagged by a field". This *is* your team's error-handling idiom:

```typescript
type ApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; error: string };
function render(r: ApiResult<Task[]>) {
  if (!r.ok) {
    console.error(r.error);   // r.error exists here — the tag proved which arm
    return;
  }
  console.log(`${r.data.length} tasks`); // r.data only exists on the ok arm
}
```

One shared field (`ok`, here) with a literal type makes both arms checkable
and prevents you from touching `data` on a failed result.

**Custom type guards.** When narrowing needs logic the compiler can't infer,
write a predicate function — note the `value is T` return type:

```typescript
function isApiError(x: unknown): x is { error: string } {
  return typeof x === "object" && x !== null && "error" in x;
}
```

**`unknown` vs `any`.** `any` opts out of checking entirely — a silent,
spreading hole. `unknown` says "there's data here; prove what it is before
using it." Untrusted entry points (`JSON.parse`, `res.json()` returns `any`)
should re-enter the safe world *through* `unknown`:

```typescript
async function getJson<T>(res: Response): Promise<T> {
  const data: unknown = await res.json(); // block the any from escaping
  return data as T; // deliberate, reviewed assertion — we harden this in Lesson 25
}
```

`as T` is a **type assertion**: *you* telling the compiler "trust me" —
erased at runtime, checking nothing. The escape hatch; needing many means
your model of the API is wrong.

## Follow along

In `~/learn-js/lesson-07`, reuse the tsconfig from
[Lesson 06](06-typescript-i-types-and-inference.md) verbatim;
`pnpm add -D typescript tsx` if starting fresh. Make sure the mock backend is
up (`cd backend-mock && go run main.go`).

### 1. Model the contract

Create `shapes.ts` — type the three interfaces above **from the README's data
model, from memory**, then diff against `backend-mock/main.go`'s structs:
identical field names, status strings as literal unions, times as `string`.

### 2. The compiler catches contract drift

Append to `shapes.ts`:

```typescript
const p: Project = {
  id: "p4",
  name: "New Thing",
  description: "x",
  status: "actve",
  createdAt: "2026-11-01T00:00:00Z",
};
```

Run `pnpm tsx shapes.ts`. Expected: TS2322 — `'actve'` is not assignable to
`"active" | "archived"`. Fix the spelling: this typo reaching production is
exactly the bug class the team pays you to prevent.

### 3. Discriminated union: normalize API results

Create `result.ts`:

```typescript
type ApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; error: string };
async function fetchSafe<T>(url: string, init?: RequestInit): Promise<ApiResult<T>> {
  try {
    const res = await fetch(url, init);
    if (!res.ok) {
      const body = (await res.json().catch(() => null)) as { error?: string } | null;
      return { ok: false, error: body?.error ?? `HTTP ${res.status}` };
    }
    return { ok: true, data: (await res.json()) as T };
  } catch (err) {
    return { ok: false, error: err instanceof Error ? err.message : "network error" };
  }
}
const login = await fetchSafe<{ token: string; user: { name: string } }>(
  "http://localhost:8080/api/login",
  {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username: "wen", password: "demo" }),
  },
);
if (!login.ok) {
  console.error("login failed:", login.error);
  process.exit(1);
}
console.log("welcome,", login.data.user.name); // data accessible only after the tag check
```

Run it: `pnpm tsx result.ts` → `welcome, Wen`. Kill the Go server, rerun:
`login failed: fetch failed` — same code path, no escaping exceptions. A
BFF-shaped function.

### 4. `unknown` at the boundary

Create `guard.ts`:

```typescript
function isTask(x: unknown): x is { id: string; title: string } {
  return typeof x === "object" && x !== null && "id" in x && "title" in x &&
    typeof (x as { id: unknown }).id === "string";
}
const candidates: unknown[] = [{ id: "t1", title: "ok" }, "not an object", { id: 42 }, null];
for (const c of candidates) {
  console.log(isTask(c) ? `task: ${c.title}` : `rejected: ${JSON.stringify(c)}`);
}
// task: ok / rejected: "not an object" / rejected: {"id":42} / rejected: null
```

Every `unknown` used the `isTask` proof first. Try printing `c.title`
*without* the guard — the compiler refuses. That refusal is the feature.

## Your turn

1. **Contract mirror.** Write `LoginResponse` (`{ token, user }`, reusing
   `User`) and `LoginRequest`. Annotate a login function returning
   `Promise<ApiResult<LoginResponse>>` with them. (Extend `result.ts`.)

2. **Narrowing drill.** Without running, predict which lines error, then
   check:

```typescript
function f(x: string | number | null) {
  if (x) {
    return x.length;          // A
  }
  return x.toString();        // B
}
```

3. **Transfer.** `GET /api/projects/p999` returns 404 with
   `{"error":"project not found"}`. Using your `fetchSafe`, write the snippet
   that logs in (password `demo`), fetches that project with the bearer
   token, and prints `project: <name>` or `server said: <error>`.

<details><summary>Solution 1</summary>

```typescript
interface LoginRequest { username: string; password: string; }
interface LoginResponse { token: string; user: User; }
async function login(username: string, password: string): Promise<ApiResult<LoginResponse>> {
  const body: LoginRequest = { username, password };
  return fetchSafe<LoginResponse>("http://localhost:8080/api/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}
```

`const body: LoginRequest = ...` gives autocomplete on the request shape —
misspell `usernme` and the compiler fires before the server 400s on you.
</details>

<details><summary>Solution 2</summary>

A errors: inside `if (x)`, `x` narrows to `string | number` (null and falsy
values excluded) and `.length` doesn't exist on `number`. B is **legal**: the
falsy arm keeps `0`, `""`, `null`, and `toString()` exists on all three.
Fix A: `typeof x === "string" ? x.length : String(x).length`.
</details>

<details><summary>Solution 3</summary>

```typescript
const login = await fetchSafe<LoginResponse>("http://localhost:8080/api/login", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ username: "wen", password: "demo" } satisfies LoginRequest),
});
if (!login.ok) {
  console.error("login failed:", login.error);
  process.exit(1);
}
const proj = await fetchSafe<Project>("http://localhost:8080/api/projects/p999", {
  headers: { Authorization: `Bearer ${login.data.token}` },
});
if (!proj.ok) {
  console.log(`server said: ${proj.error}`); // server said: project not found
} else {
  console.log(`project: ${proj.data.name}`);
}
```

`satisfies LoginRequest` checks the literal against the type while keeping
inference — the modern alternative to annotations.
</details>

## Checkpoint

1. `interface` vs `type` — what actually matters at your level?
<details><summary>Answer</summary>
Both name a shape; both erase at compile time. Convention: `interface` for
object shapes, `type` for unions. Structural checking makes the name
irrelevant to compatibility.
</details>

2. What does `assignee?: string` make the property's type, and what must
   callers do?
<details><summary>Answer</summary>
`string | undefined`; every use must handle the undefined case (`??`, a
check, a default) — strict mode forces it.
</details>

3. Why is a discriminated union safer than `{ data?: T; error?: string }`?
<details><summary>Answer</summary>
The optional-fields type admits nonsense states (both set, neither set) and
lets you touch `data` when undefined. The tag field makes arms exclusive; the
compiler exposes an arm's fields only after you check the tag.
</details>

4. One sentence: `any` vs `unknown`.
<details><summary>Answer</summary>
`any` disables all checks on a value (a silent hole); `unknown` keeps the
value but forces you to narrow/prove its shape before use.
</details>

5. The mock API's `POST /api/tasks` returns 201 with the created task. Using
   your `ApiResult<Task>`, write the signature of a function that adds a task
   (takes projectId, title, priority), and name the two narrowing steps it
   needs inside.
<details><summary>Answer</summary>

```typescript
async function addTask(projectId: string, title: string, priority: number): Promise<ApiResult<Task>>
```

Steps: check `res.ok` (else build the `ok:false` arm from the `{"error":...}`
body), then narrow the parsed JSON into `Task` (`unknown` + assertion now,
zod in Lesson 25).
</details>

## Primary source

[TypeScript Handbook — Narrowing](https://www.typescriptlang.org/docs/handbook/2/narrowing.html)
The official catalog of every narrowing technique — `typeof`, `in`, equality,
`instanceof`, type predicates — with exact compiler behavior for each. When
the compiler "should know" a value is safe and doesn't, the answer is here.

## Before next session

Add to `result.ts` a `getJson` helper typed `(res: Response) => Promise<unknown>`
and route every `res.json()` through it. Feel the friction of narrowing
everywhere — then come to
[Lesson 08](08-typescript-iii-generics-and-utility-types.md), where
`fetchSafe<T>` gets demystified.
