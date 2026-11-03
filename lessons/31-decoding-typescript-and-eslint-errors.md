# Lesson 31 — Decoding TypeScript & ESLint Errors

*Phase 5 · Session 31 of 38 · ~2 hours · Prereqs: Lesson(s) [06](06-typescript-i-types-and-inference.md), [07](07-typescript-ii-shapes-and-narrowing.md), [25](25-typed-contracts-with-zod.md)*

## Why this matters

On day one with the team, the thing that will slow you down is not writing TypeScript —
it's staring at a 40-line compiler error and not knowing which line matters. The compiler
is honest but verbose: one real mistake generates a cascade of follow-on errors, and
ESLint adds its own vocabulary on top. Reading errors fast is a *learned pattern-matching
skill*, and this lesson installs the 12 patterns that cover most of what you'll see in a
Next.js BFF codebase.

## What you'll be able to do

- You can take any TypeScript error and, within 30 seconds, identify the **first line** as the real one and discard the cascade.
- You can name and fix the 12 most common `TS2xxx` errors without looking them up.
- You can explain "assignability", "excess property check", and "union narrowing" to a rubber duck in one sentence each.
- You can tell a TypeScript error (build fails) from an ESLint error (lint fails), and fix the three most common ESLint rules in React code.
- You can add the correct one-line ESLint disable comment when your team's config genuinely gets it wrong — and only then.

## Principles

### The compiler is honest, the cascade is noise

The TypeScript compiler checks your file top-to-bottom, expression by expression. When
one type is wrong, everything that *uses* that value also becomes "wrong" — so one bug
prints ten errors. Rule zero: **read the first error in the file, fix it, recompile, and
watch most of the rest vanish.** Chasing the last error in a cascade wastes hours.

A TypeScript error has four parts:

```text
src/lib/tasks.ts:14:7 - error TS2322: Type 'string' is not assignable to type 'number'.

14   const priority: number = task.dueDate;
     ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
```

1. **Location** — `src/lib/tasks.ts:14:7` (file : line : column).
2. **Error code** — `TS2322`. This code is stable and searchable. Google or your AI agent
   with `"TS2322"` plus your message; someone has answered it on Stack Overflow.
3. **Message** — the actual claim the compiler is making.
4. **The squiggle** — the `~~~~` underlines the *exact expression* it's complaining about.
   The bug is at the squiggle, not necessarily where the value came from.

### Three jargon terms you must own

**Assignability** — "type A is *assignable* to type B" means every value of type A is a
legal value where B is expected. A `Dog` is assignable to `Animal` (it has everything
`Animal` requires); a `Cat` with four legs is *not* assignable to a `Dog` type missing the
`bark` method. Structural typing: shapes matter, names don't. If two types have the same
required properties, they are interchangeable.

**Excess property check** — when you pass an *object literal written right there in the
code* to something expecting a specific shape, TypeScript rejects properties that aren't
in the target type. It assumes an unexpected property is a typo. This check only fires on
literals: assign the same object to a `const` first and the check disappears (which is
why a helper function sometimes "fixes" a TS2322 on an object literal — usually it's
hiding a real bug instead).

**Union narrowing** — a union type like `"todo" | "doing" | "done"` means "one of these".
Narrowing is TypeScript watching your `if`/`switch` checks and shrinking the union inside
each branch: after `if (status === "todo") return;`, the compiler knows the remaining
union is `"doing" | "done"`. Most "not assignable" errors on unions mean the compiler saw
a path where you *didn't* narrow — often a missing `default:` in a `switch`.

### The 12 errors you'll actually meet

Fix recipes assume the TeamPulse codebase from Lessons 25–30.

1. **TS2322 — `Type 'X' is not assignable to type 'Y'`.** The workhorse. Read left to
   right: what you *gave* is on the left, what was *wanted* is on the right. Usual cause:
   the Go API returned a field as `string` where your type says `number`, or a field can
   be `null`. Fix the source type or convert deliberately (`Number(x)`), never with
   `as` first.
2. **TS2345 — `Argument of type 'X' is not assignable to parameter of type 'Y'`.** Same
   as TS2322 but at a function call. Check the argument order — a very common variant is
   passing `(projectId, title)` where the function takes `(title, projectId)` because
   both are `string` and the compiler can't tell.
3. **TS2339 — `Property 'x' does not exist on type 'Y'`.** You're reading a field the
   type doesn't declare. Three causes: typo, the type is stale (Go added a field — update
   your zod schema and derived type, Lesson 25), or you're accessing a field that only
   exists on one *member* of a union and haven't narrowed.
4. **TS2551 — `Property 'x' does not exist... Did you mean 'y'?`.** TS2339 with a hint —
   the compiler found a near-miss. Almost always a typo (`task.projectId` vs `task.projId`).
   Note *which* casing the type uses: Go's JSON tags decide (Lesson 24).
5. **TS2739 / TS2741 — `Type 'X' is missing the following properties from type 'Y': a, b`.**
   You're constructing an object (a task, a project) without required fields. Either the
   target type is wrong (should the field be optional with `?`) or your object is
   incomplete. For "create" payloads, `Omit<Task, "id" | "createdAt">` is usually the
   right shape (Lesson 08).
6. **TS2322 on an object literal — `Object literal may only specify known properties, and 'prioirty' does not exist`.** The
   excess property check. It's telling you a property name is wrong or extra. Fix the
   name; don't restructure to dodge the check.
7. **TS2554 — `Expected 2 arguments, but got 1`.** Miscounted arguments. The frequent
   cause: a function signature gained a parameter (someone added `signal` for timeouts,
   Lesson 28) and this call site wasn't updated.
8. **TS7006 — `Parameter 'x' implicitly has an 'any' type`.** `noImplicitAny` (on by
   default in strict mode) refuses to guess. Usually a callback parameter:
   `tasks.filter(t => ...)` with `tasks` itself untyped. Fix the *source* type, then
   inference flows down and the error disappears.
9. **TS18048 / TS2533 — `'x' is possibly 'undefined'` / `'undefined' not assignable`.**
   strict null checks (Lesson 07) caught a field that can be missing. Fix with a real
   decision: `x?.field ?? fallback`, or an early `if (!x) return`, or make the zod schema
   `.optional().default(...)` if Go truly always sends it.
10. **TS18046 — `'x' is of type 'unknown'`.** The good-news error: you got `unknown` from
    `catch (e)` or `JSON.parse`. You must validate before using — which is exactly what
    zod's `safeParse` does. This error is the compiler *protecting your BFF boundary*
    (Lesson 25).
11. **TS2367 — `This comparison appears to be unintentional because the types have no overlap`.**
    You're comparing a narrowed union to a value it can't be: after checking
    `status === "todo"`, a later `status === "todo"` is dead code, or you're comparing
    `task.status === "Todo"` and the union only has lowercase `"todo"`. Casing mismatches
    with the Go API show up here.
12. **TS2307 — `Cannot find module '...' or its corresponding type declarations`.**
    Wrong path (App Router uses `@/` alias — check `tsconfig.json` `paths`), missing
    install (`pnpm add <pkg>`), or a case-sensitive filename mismatch that works on
    macOS but breaks CI on Linux.

### TypeScript errors vs ESLint errors

TypeScript checks *types* (`pnpm tsc --noEmit` or the editor's red squiggles). ESLint
checks *style and dangerous patterns* (`pnpm lint`), and its errors reference a rule name
in parentheses instead of a TS code:

```text
6:3  error  Error: Calling setState synchronously within an effect can trigger cascading renders  react-hooks/set-state-in-effect
```

The three ESLint rules that will fire most on your code:

- **`react-hooks/exhaustive-deps`** — your `useEffect`/`useMemo` dependency array is
  missing a value the callback reads. Don't silence it; add the dep, and if adding the
  dep causes a loop, the loop is the actual bug (Lesson 15).
- **`@typescript-eslint/no-floating-promises`** — you called an `async` function without
  `await`ing or `.catch()`ing it (classic in route handlers, Lesson 20). Unhandled
  rejections crash request handling silently. Fix: `await` it.
- **`@typescript-eslint/no-unused-vars`** — dead import or variable. Delete it. (If it's
  a required-but-unused function parameter, prefix with `_`.)

When the rule is genuinely wrong (rare), your team's convention is a one-line disable with
a reason — the comment is code review bait if the reason is empty:

```ts
// eslint-disable-next-line react-hooks/exhaustive-deps -- ref is stable by design
```

## Follow along

Work in `~/teampulse` (your P3 app from Lesson [30](30-p3-full-stack-teampulse.md)). Make
sure the mock Go API is up: `cd backend-mock && go run main.go`.¹

1. Create a deliberately broken file to collect real errors to dissect:

   ```ts
   // src/lib/bug-lab.ts
   import { z } from "zod";

   const TaskSchema = z.object({
     id: z.string(),
     title: z.string(),
     priority: z.number(),
     status: z.enum(["todo", "doing", "done"]),
   });
   type Task = z.infer<typeof TaskSchema>;

   const raw = { id: "t1", title: "Write tests", prioirty: 2, status: "todo" };
   const task: Task = raw;

   function bump(task: Task, amount: number): Task {
     return { ...task, priority: task.priority + amount };
   }

   const next = bump(task, 1);
   const label: string = next.dueDate;
   export { next, label, raw };
   ```

2. Run the type checker and count errors before fixing anything:

   ```bash
   pnpm tsc --noEmit
   ```

   Expected: four errors — TS2322 on `const task: Task = raw` (with the excess-property
   hint about `prioirty`), TS2739-ish complaints, TS2339 on `next.dueDate`, and possibly
   TS7006-style notes depending on your tsconfig. **Read the first one only.** Its hint
   (`Did you mean 'priority'?`) is the whole fix.

3. Fix the typo `prioirty` → `priority`, recompile. Expected: the excess-property and
   assignability errors vanish; `next.dueDate` (TS2339) remains because `Task` has no
   `dueDate`. Add `dueDate: z.string()` to the schema (Go sends it as an RFC3339 string,
   Lesson 24) and to `raw`, recompile. Expected: **zero TypeScript errors.**

4. Now run ESLint:

   ```bash
   pnpm lint
   ```

   Expected: `raw` is exported so nothing is unused, but if you delete the `export`,
   ESLint flags unused locals. Try deleting `export { next, label, raw };` plus the
   `label` line's usage and watch the rule name appear after the message.

5. Reproduce the "cascade" effect: change `priority: z.number()` to `priority: z.string()`
   and recompile. Expected: one schema change produces errors at the `raw` literal, inside
   `bump`, and at `next.priority` — three errors, one fix. Revert and confirm clean.

## Your turn

1. **Union narrowing drill.** Add to `bug-lab.ts`:

   ```ts
   function describe(status: "todo" | "doing" | "done"): string {
     if (status === "todo") return "not started";
     if (status === "doing") return "in progress";
     return "unrecognized status: " + status.toUpperCase();
   }
   ```

   Why does the compiler error on the last line? Fix it without deleting the branch.

2. **Decode a real one.** In one of your TeamPulse route handlers (Lesson 20), remove the
   `await` before `fetch` and run `pnpm tsc --noEmit`. Read the first error carefully:
   what two types does it say are not assignable, and why? Restore and confirm clean.

3. **Cascade hunt.** In `bump`, change the signature to `(task: Task, amount: string)`
   *without touching the body*. How many errors appear? Which one did you fix?

<details>
<summary>Solution</summary>

1. After the two `if` returns, TypeScript has narrowed `status` to the `never` type —
   the union is exhausted, so `status.toUpperCase()` is invalid and the comparison logic
   is dead. Fix: make the fallthrough honest, e.g.

   ```ts
   function describe(status: "todo" | "doing" | "done"): string {
     switch (status) {
       case "todo": return "not started";
       case "doing": return "in progress";
       case "done": return "shipped";
       default: { const _exhaustive: never = status; return _exhaustive; }
     }
   }
   ```

   The `never` assignment is the idiom: if someone adds a fourth status, the *compiler*
   breaks here instead of silently falling through.

2. `fetch` without `await` returns `Promise<Response>`, so the error is TS2339/TS2345
   shapes like `Property 'json' does not exist on type 'Promise<Response>'` or
   `Type 'Promise<unknown>' is not assignable to type 'Response'`. The two types named
   are the Promise and the awaited value — the fix is `await`.
3. Three or more errors (comparison `+` on string, the `bump(task, 1)` call site,
   possibly downstream `next.priority`). Fixing the signature back to `number` clears
   all of them — one root cause again.

</details>

## Checkpoint

1. A file shows 9 TypeScript errors. What's your first move, and why?

<details>
<summary>Answer</summary>

Fix the *first* error in the file and recompile. Errors cascade from a single bad type;
the first is usually the root and later ones are follow-on noise.

</details>

2. One sentence each: assignability, excess property check, union narrowing.

<details>
<summary>Answer</summary>

Assignability: type A is legal wherever B is expected if A has at least everything B
requires (structural shapes, not names). Excess property check: object literals written
inline are rejected if they have properties the target type doesn't declare — usually a
typo. Union narrowing: `if`/`switch` checks shrink a union type inside each branch, and
the compiler tracks this.

</details>

3. Your BFF's zod schema for `Task` omits `dueDate`, and a component now throws TS2339
   `Property 'dueDate' does not exist on type 'Task'`. Where is the bug and what's the fix?

<details>
<summary>Answer</summary>

The bug is at the contract, not the component. `GET /api/tasks` from the mock Go API
(`http://localhost:8080`, see `backend-mock/main.go`) returns `dueDate` on every task
(Lesson 24), so the zod schema in the route handler is stale — add
`dueDate: z.string()` to the schema so the inferred type regains the field.

</details>

4. What's the difference between a TS2322 and an error tagged `react-hooks/exhaustive-deps`,
   and which one blocks `pnpm build`?

<details>
<summary>Answer</summary>

TS2322 is a type error from `tsc` and blocks the build. `react-hooks/exhaustive-deps` is
an ESLint rule about missing `useEffect` dependencies; it fails `pnpm lint` (and Next's
build lints by default), but it's a different tool — same terminal habit, different fix
category.

</details>

5. ESLint's `no-floating-promises` flags a line in your route handler:
   `validateAndCache(task);`. Why is this dangerous in a route handler specifically?

<details>
<summary>Answer</summary>

The handler may return a response before the async work finishes, and a rejection inside
it becomes an unhandled promise rejection (crashing or poisoning the Node process). In a
route handler every async step must be `await`ed (Lesson 20, Lesson 28).

</details>

## Primary source

[TypeScript Handbook — Type Compatibility](https://www.typescriptlang.org/docs/handbook/type-compatibility.html).
Every "is not assignable" error is a verdict from these rules, and reading the actual
rules once demystifies TS2322 forever.

## Before next session

Run `pnpm tsc --noEmit && pnpm lint` on your TeamPulse repo and paste the first error of
each kind into a scratch note — in Lesson 32 you'll debug a runtime bug, and a clean
compile/lint baseline is how you prove a bug is *not* a type problem.

---

¹ `pnpm lint` is your Next.js script (created by `create-next-app`). If you installed
dependencies with npm instead, substitute `npm run lint` throughout.
