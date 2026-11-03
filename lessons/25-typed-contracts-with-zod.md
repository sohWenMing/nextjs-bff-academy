# Lesson 25 — Typed contracts with zod

*Phase 4 · Session 25 of 38 · ~2 hours · Prereqs: Lessons 07, 08, 20, 24*

## Why this matters

In Lesson 24 you mirrored the Go structs as TS interfaces. Here's the uncomfortable
truth: an `interface` is a rumour. It convinces the compiler and vanishes at runtime —
if Go renames `dueDate` to `due`, your app doesn't crash at the boundary, it crashes
three screens later with `undefined is not a function`, in production. zod puts a guard
at the BFF boundary so a lying upstream fails loudly, on the server, in one place.
This is MISSION item 2: a route handler that calls Go, validates with zod, and reshapes
for the UI.

## What you'll be able to do

- Write zod schemas for the TeamPulse `User`, `Project`, and `Task` shapes from memory
- Derive TS types from schemas with `z.infer` so there is exactly one source of truth
- Parse every upstream response inside a route handler, before it reaches the UI
- Explain the difference between `parse` (throws) and `safeParse` (returns a result)
- Make a deliberate decision about *how loudly* a validation failure should fail

## Principles

### A schema is a runtime type checker

zod (v3) lets you describe a shape as a *value* — real code that runs — instead of an
`interface`, which is a comment the compiler deletes. `pnpm add zod` in `~/teampulse`,
then:

```ts
// lib/schemas.ts
import { z } from "zod";

export const TaskSchema = z.object({
  id: z.string(),
  projectId: z.string(),
  title: z.string(),
  status: z.enum(["todo", "doing", "done"]),
  priority: z.number().int(),
  dueDate: z.string(),
  assignee: z.string(),
});
```

`TaskSchema` is an object you can *call*: `TaskSchema.parse(someUnknown)` either returns
the data, correctly typed, or throws a `ZodError` naming the exact path that broke
(`dueDate: expected string, received undefined`).

### The boundary is the route handler

Where do you parse? **Wherever untrusted bytes enter your trust zone.** For the BFF
that's inside `app/api/*/route.ts`, right after `res.json()`, per
[Lesson 20](20-route-handlers-your-first-bff.md). The UI never talks to Go directly, so if
the route handler parses, the UI is guaranteed clean data:

```
Go JSON ──▶ res.json() ──▶ schema.parse() ──▶ NextResponse.json() ──▶ UI
  unknown      unknown        typed or THROWS        typed            typed
```

TypeScript's `unknown` (Lesson 07) is the type for "we don't know yet". `res.json()`
returns `unknown`-ish data; `parse` is the narrowing gate that turns it into a type.

### One source of truth: z.infer

Never hand-write both an interface *and* a schema for the same shape — they drift.
Write the schema; mint the type from it:

```ts
export type Task = z.infer<typeof TaskSchema>;
//       ^ now `Task` is a real TS type, identical to the Lesson 24 interface,
//         but guaranteed to match what parse() actually accepts.
```

Delete the hand-written interfaces from `lib/api-types.ts` and re-export types from
`lib/schemas.ts`. `lib/api-types.ts` survives only for things zod can't check, like
JSDoc.

### parse throws — that's the point

`parse` failing inside a route handler makes that route return a 500 with a stack trace
in the server terminal. That is *correct behaviour* for a contract violation: the
upstream broke its promise and you want to know immediately, not have it papered over.
Reserve `safeParse` for cases where you have a real fallback:

```ts
const result = LoginResponseSchema.safeParse(await res.json());
if (!result.success) {
  // result.error.issues is a structured array — log it with full detail
  console.error("login contract broken", result.error.issues);
  return Response.json({ error: "upstream contract violation" }, { status: 502 });
}
```

Note the status: **502 Bad Gateway**, not 500. You are a proxy; an upstream that
misbehaves is the upstream's fault. This distinction pays off in Lesson 28.

### Trust status codes before bodies

Parse the body only after you've checked `res.ok`. A 401 from Go carries
`{"error": "..."}`, not a `Task` — parsing it with `TaskSchema` produces a confusing
"contract violation" when the real story is "not logged in". The order in every
upstream call is fixed: `fetch` → check status → parse shape → return.

## Follow along

Work in `~/teampulse` (the app from Lessons 16, 22). Start the Go mock in its usual tab:
`cd ~/workspace/nextjs-bff-academy/backend-mock && go run main.go`.

1. Install zod:
   ```bash
   cd ~/teampulse && pnpm add zod
   ```
   Expected: `+ zod 3.x` in the output.

2. Create `lib/schemas.ts` with `UserSchema`, `ProjectSchema`, `TaskSchema`,
   `ApiErrorSchema`, `LoginResponseSchema`, and `z.infer` types for each. Compare field
   names against `backend-mock/main.go` — `projectId` not `ProjectID`, dates as
   `z.string()`. Expected content:

   ```ts
   import { z } from "zod";

   export const UserSchema = z.object({
     id: z.string(),
     name: z.string(),
     role: z.string(),
   });

   export const ProjectSchema = z.object({
     id: z.string(),
     name: z.string(),
     description: z.string(),
     status: z.enum(["active", "archived"]),
     createdAt: z.string(),
   });

   export const TaskSchema = z.object({
     id: z.string(),
     projectId: z.string(),
     title: z.string(),
     status: z.enum(["todo", "doing", "done"]),
     priority: z.number().int(),
     dueDate: z.string(),
     assignee: z.string(),
   });

   export const ApiErrorSchema = z.object({ error: z.string() });

   export const LoginResponseSchema = z.object({
     token: z.stri…ings().startsWith("tok-"),
     user: UserSchema,
   });

   export type User = z.infer<typeof UserSchema>;
   export type Project = z.infer<typeof ProjectSchema>;
   export type Task = z.infer<typeof TaskSchema>;
   export type LoginResponse = z.infer<typeof LoginResponseSchema>;
   ```

3. Create the boundary helper `lib/upstream.ts`. This is the one function through which
   all Go traffic flows:

   ```ts
   import { z } from "zod";

   const GO_API = process.env.GO_API_BASE_URL ?? "http://localhost:8080";

   export class UpstreamError extends Error {
     constructor(
       public status: number,
       message: string,
     ) {
       super(message);
     }
   }

   export async function callGo<T>(
     path: string,
     schema: z.ZodType<T>,
     init?: RequestInit,
   ): Promise<T> {
     const res = await fetch(`${GO_API}${path}`, {
       ...init,
       cache: "no-store", // upstream data must never come from Next's fetch cache
     });

     const body: unknown = await res.json();

     if (!res.ok) {
       // error path: still validate, so we fail on garbage too
       const err = ApiErrorSchema.safeParse(body);
       throw new UpstreamError(res.status, err.success ? err.data.error : "upstream error");
     }

     // throws ZodError on contract violation — deliberate, see Principles
     return schema.parse(body);
   }
   ```

   Expected: `pnpm tsc --noEmit` is clean.

4. Rewrite `app/api/projects/route.ts` to use it:

   ```ts
   import { NextRequest, NextResponse } from "next/server";
   import { z } from "zod";
   import { callGo } from "@/lib/upstream";
   import { ProjectSchema } from "@/lib/schemas";

   export async function GET(req: NextRequest) {
     const token = req.cookies.get("tp_token")?.value;
     if (!token) {
       return NextResponse.json({ error: "not logged in" }, { status: 401 });
     }
     try {
       const projects = await callGo(
         "/api/projects",
         z.array(ProjectSchema),
         { headers: { Authorization: `Bearer ${token}` } },
       );
       return NextResponse.json(projects);
     } catch (err) {
       if (err instanceof UpstreamError) {
         return NextResponse.json({ error: err.message }, { status: err.status });
       }
       throw err; // ZodError: let it crash loudly, Lesson 28 catches it prettily
     }
   }
   ```

   (The cookie read is a preview of Lesson 27 — hard-code a real token from
   `curl -s -X POST http://localhost:8080/api/login -H 'Content-Type: application/json' -d
   '{"username":"wen","password":"demo"}'` if you want to test before then.)

5. Run `pnpm dev`, then in another tab:
   ```bash
   curl -s localhost:3000/api/projects -H "Cookie: tp_token=<your-token>"
   ```
   Expected: the three projects, identical to hitting Go directly.

6. Prove the guard works. Temporarily change `ProjectSchema`'s `status` to
   `z.enum(["open", "closed"])` and re-run the curl. Expected: the dev terminal shows a
   `ZodError` with `status: Invalid enum value. Received 'active'` — the contract
   violation surfaced on the server, in one place, naming the field. Revert the schema.

## Your turn

1. Write `app/api/projects/[id]/route.ts` (dynamic segment per
   [Lesson 17](17-file-system-routing.md)) using `callGo` + `ProjectSchema`. Go returns
   404 with `{"error":"project not found"}` for unknown ids — your route must pass that
   404 through, not crash.

   <details><summary>Solution</summary>

   ```ts
   import { NextRequest, NextResponse } from "next/server";
   import { callGo, UpstreamError } from "@/lib/upstream";
   import { ProjectSchema } from "@/lib/schemas";

   export async function GET(
     _req: NextRequest,
     { params }: { params: Promise<{ id: string }> },
   ) {
     const { id } = await params; // Next 15: params is a Promise
     const token = _req.cookies.get("tp_token")?.value;
     if (!token) {
       return NextResponse.json({ error: "not logged in" }, { status: 401 });
     }
     try {
       const project = await callGo(`/api/projects/${id}`, ProjectSchema, {
         headers: { Authorization: `Bearer ${token}` },
       });
       return NextResponse.json(project);
     } catch (err) {
       if (err instanceof UpstreamError) {
         return NextResponse.json({ error: err.message }, { status: err.status });
       }
       throw err;
     }
   }
   ```
   `curl -si localhost:3000/api/projects/p999 -H "Cookie: tp_token=<token>"` → `404` +
   `{"error":"project not found"}` — the Go status rode straight through.
   </details>

2. Route handlers must validate *incoming* bodies too — the browser is also untrusted.
   Add `CreateTaskSchema` (`projectId`, `title` non-empty, `priority` 1–3) and a
   `POST /api/tasks` route that rejects bad input with 400 **before** calling Go.

   <details><summary>Solution</summary>

   ```ts
   // add to lib/schemas.ts
   export const CreateTaskSchema = z.object({
     projectId: z.string().min(1),
     title: z.string().min(1),
     priority: z.number().int().min(1).max(3),
   });
   export type CreateTask = z.infer<typeof CreateTaskSchema>;
   ```

   ```ts
   // app/api/tasks/route.ts (POST half)
   export async function POST(req: NextRequest) {
     const token = req.cookies.get("tp_token")?.value;
     if (!token) return NextResponse.json({ error: "not logged in" }, { status: 401 });

     const parsed = CreateTaskSchema.safeParse(await req.json());
     if (!parsed.success) {
       return NextResponse.json(
         { error: parsed.error.issues[0].message },
         { status: 400 },
       );
     }
     try {
       const task = await callGo("/api/tasks", TaskSchema, {
         method: "POST",
         headers: {
           Authorization: `Bearer ${token}`,
           "Content-Type": "application/json",
         },
         body: JSON.stringify(parsed.data),
       });
       return NextResponse.json(task, { status: 201 }); // 201, per Lesson 24
     } catch (err) {
       if (err instanceof UpstreamError) {
         return NextResponse.json({ error: err.message }, { status: err.status });
       }
       throw err;
     }
   }
   ```
   </details>

3. Deliberately break the contract *upstream*: stop Go, add a field `Temperature int` to
   nothing — instead, in `lib/schemas.ts`, make `assignee` required-but-absent by
   renaming it to `assignedTo` in the schema only. What do you see in the browser vs the
   terminal, and why is that the right split?

   <details><summary>Solution</summary>
   The browser shows your `error.tsx` fallback (Lesson 21) or a generic failed state —
   the UI learns only "it failed". The server terminal shows the full `ZodError`:
   `assignedTo: expected string, received undefined`. That's the right split: users see
   calm, developers see forensic detail, and the failure is at the boundary — no
   mysterious downstream `undefined` crashes.
   </details>

## Checkpoint

1. Why can't `interface Task {...}` protect you from Go renaming a field, and what does
   zod do differently?

   <details><summary>Answer</summary>
   Types are erased at build time (Lesson 06) — an interface exists only while the
   compiler runs; at runtime nothing checks incoming JSON. A zod schema is a value that
   executes at runtime: `.parse()` actually inspects the object and throws on mismatch.
   </details>

2. In a route handler, why check `res.ok` *before* `schema.parse(body)`?

   <details><summary>Answer</summary>
   Because error responses carry `{"error": "..."}`, not the success shape. Parsing an
   error body with `TaskSchema` reports a bogus "contract violation" and hides the real
   cause (e.g. 401 = your token is dead). Status first, shape second.
   </details>

3. `parse` vs `safeParse` — when is each the right call?

   <details><summary>Answer</summary>
   `parse` throws: use when a contract violation is a bug you want to surface loudly
   (upstream success responses in your BFF). `safeParse` returns
   `{success, data | error}`: use when you have a real fallback path, e.g. parsing an
   error body, or validating user input you plan to bounce back as a 400.
   </details>

4. (Go API question) Go's `POST /api/tasks` requires `projectId` and `title`, else it
   returns 400. With the Lesson 25 `POST` route, when does Go's 400 ever fire — and is
   that good?

   <details><summary>Answer</summary>
   Rarely — `CreateTaskSchema.safeParse` rejects missing `projectId`/`title` with a 400
   before Go is called. That's good: cheap local validation, one consistent error shape,
   no wasted round-trip. Go's check remains as the authority; yours mirrors the contract
   from `main.go` (Lesson 24), it doesn't replace Go's.
   </details>

5. Where is the *one* place all zod parsing for upstream responses happens in the
   pattern from this lesson, and why does centralising there matter?

   <details><summary>Answer</summary>
   Inside `callGo` in `lib/upstream.ts` — every upstream response passes through it, so
   one function enforces the whole contract. A new endpoint costs a schema and a path,
   not a new copy of the fetch/check/parse dance, and you can't forget to validate.
   </details>

## Primary source

[zod — Basic usage](https://zod.dev/?id=basic-usage) — the official docs; schemas,
`parse`/`safeParse`, and `z.infer` are all there with the exact semantics you need for
v3.

## Before next session

Run `pnpm add @tanstack/react-query` in `~/teampulse` (don't configure it yet), then
open <https://tanstack.com/query/latest/docs/framework/react/overview> and read the
"Ready to learn all about TanStack Query?" motivation section only. Note the one
sentence that explains why `useState` + `useEffect` fetching is considered a bug factory.
