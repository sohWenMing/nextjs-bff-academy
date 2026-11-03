# Lesson 06 — TypeScript I: types & inference

*Phase 1 · Session 06 of 38 · ~2 hours · Prereqs: [Lesson 02](02-modern-js-syntax.md), [Lesson 05](05-async-promises-await-fetch.md)*

## Why this matters

Your team's entire frontend and BFF are TypeScript. TS is JavaScript plus a
compiler that checks code *before* it runs: wrong field names, `undefined`
access, a string where a number belongs — caught at edit time, not in
production. Two ideas make TS click: **types are erased at build time** (the
browser never sees them; TS is a checker, not a runtime) and **inference does
~90% of your typing**. Get these and team code stops being guesswork.

## What you'll be able to do

- You can set up a strict TS project (`pnpm add -D typescript tsx`) from memory.
- You can annotate variables, function parameters, and return types with the
  primitive and array types.
- You can predict what the compiler infers for `const`/`let` initializers.
- You can read and fix the first line of a typical compiler error.
- You can explain why `const x = 5` infers `5`, not `number`.

## Principles

**What TypeScript is.** TypeScript = JavaScript + a static type system + a
compiler (`tsc`). You write `.ts` files; the compiler *checks* them and
*transpiles* (translates to equivalent code) plain JavaScript. Nothing about
runtime behavior changes: `console.log(1 + "2")` still prints `12` if you ask
for it — TS just refuses to compile code that's *provably* nonsense. And the
key corollary: **type erasure**. Type annotations are deleted during
compilation. `const port: number = 8080` compiles to exactly
`const port = 8080`. Therefore TS types cannot validate an API response at
runtime — they are a promise you make to the compiler. (Remember this for
Lesson 25, where zod re-adds real runtime checking.)

**Annotations vs inference.** You *annotate* by writing `: Type` after a name:

```typescript
let retries: number = 3;
const projectName: string = "TeamPulse";
const isDone: boolean = false;
const tags: string[] = ["go", "bff"];
```

But inference means the compiler derives types from initializers and you
rarely annotate what it can see:

```typescript
const port = 8080;            // inferred number — no annotation needed
const task = { id: "t1", priority: 1 }; // inferred { id: string; priority: number }
```

Rule: annotate function boundaries (parameters, returns) and nothing else
until the compiler asks. Over-annotating is the newcomer's tell.

**Functions.** Annotate parameters and (optionally) the return:

```typescript
function totalPriority(tasks: Task[]): number { ... }   // shape types in Lesson 07
const double = (n: number): number => n * 2;
```

**Literal types.** `const x = 5` infers the type `5` — not `number` — because
a `const` can never hold anything else. This is a feature, not a quirk:
literal types power discriminated unions and status fields in
[Lesson 07](07-typescript-ii-shapes-and-narrowing.md). `let x = 5` infers
`number`, because `let` may later hold other numbers.

**The toolchain you use in this lesson.**
- `typescript` — the compiler package (gives you `tsc` and the language
  service your editor talks to). Dev-only dependency: `-D`.
- `tsx` — a runner that executes `.ts` files directly (transpile-on-the-fly),
  so you don't need a build step for scripts. It is what `node hello.mjs`
  was for JS.
- `tsconfig.json` — the compiler's settings file. `"strict": true` enables
  the strict family of checks (no implicit `any`, null-safety); every serious
  team turns it on, so every lesson here runs strict.

**Reading compiler errors.** Format: `file.ts:12:3 - error TS2322: Type
'string' is not assignable to type 'number'.` Read the **first line** —
file, position, and the two types in conflict. The numbered code (TS2322) is
searchable. Ignore elaboration paragraphs until you're fluent.

## Follow along

### 1. Scaffold

```bash
mkdir -p ~/learn-js/lesson-06 && cd ~/learn-js/lesson-06
pnpm init            # accept defaults
pnpm add -D typescript tsx
```

Expected: `dependencies` empty, `devDependencies` gains `typescript` and
`tsx`; `node_modules/` populated.

### 2. Write a strict tsconfig

Create `tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  }
}
```

Two settings beyond the minimum, both deliberate: `strict` (the null-safety
and implicit-any checks teams expect) and `noUncheckedIndexedAccess` (array
indexing returns `T | undefined`, forcing you to handle the empty-array case —
Phase 4's resilience habits start here).

### 3. Your first typed file

Create `types.mjs`'s TS twin, `types.ts` — type it by hand:

```typescript
const API = "http://localhost:8080";
function projectLine(name: string, status: string): string {
  return `${name} [${status}]`;
}
const projects = [
  { id: "p1", name: "Billing Service", status: "active" },
  { id: "p2", name: "Mobile App", status: "archived" },
];
for (const p of projects) {
  console.log(projectLine(p.name, p.status));
}
console.log(`API at ${API}`);
```

Run it with tsx (add a `"dev": "tsx types.ts"` script to `package.json` if
you like):

```bash
pnpm tsx types.ts
```

Expected:

```
Billing Service [active]
Mobile App [archived]
API at http://localhost:8080
```

### 4. Make the compiler angry — five errors, five fixes

Append each line, run `pnpm tsx types.ts`, read the **first line** of the
error, then fix or delete it:

```typescript
const count: number = "3";            // TS2322: Type 'string' is not assignable to 'number'
const retry = 5;
retry = 6;                            // TS2588: Cannot assign to 'retry' because it is a constant
console.log(projects[0].titel);       // TS2551: Property 'titel' does not exist... Did you mean 'title'?
projectLine("a", "b", "c");           // TS2554: Expected 2 arguments, but got 3
const missing = projects[5].name;     // with noUncheckedIndexedAccess: TS18048 'projects[5]' is possibly 'undefined'
```

That last one is the paywall of `noUncheckedIndexedAccess`: index access is
honest about possibly missing. Fix pattern:

```typescript
const maybe = projects[5];
if (maybe) console.log(maybe.name); // here TS knows it exists
```

### 5. Watch inference work

Create `infer.ts` with **zero annotations** and inspect what the compiler
thinks by hovering in an editor — or by forcing errors:

```typescript
const port = 8080;
let attempts = 0;
const task = { id: "t1", title: "x", done: false };
attempts += 1;        // fine — attempts is number
// attempts = "one";  // uncomment: TS2322 — inferred number, rejected string
// port = 9090;       // uncomment: readonly-ish — const
// task.done = 1;     // uncomment: TS2322 — inferred boolean, rejected number
```

You annotated nothing and still got checks on every line. That is the 90%.

### 6. Compile for real

```bash
pnpm tsc --noEmit infer.ts types.ts   # type-check without writing files
```

Expected: no output = success (the compiler's praise is silence). Later
projects will emit with `tsc`; for now `--noEmit` + tsx for running is the
loop.

## Your turn

1. **Annotate.** Write `stats.ts`: a function `summarize(tasks)` — give the
   parameter a type using the inline array-of-object annotation
   `{ status: string }[]` — returning a `string` like `"2 task(s), 1 done"`.
   Call it with a two-task array and print.

2. **Strict-mode detective.** This code compiles in sloppy mode but errors
   twice under our strict tsconfig. Find both errors and fix them *without
   deleting functionality*:

```typescript
function greet(who) {
  let message;
  if (who) {
    message = "Hello " + who;
  }
  return message.toUpperCase();
}
console.log(greet("wen"));
```

3. **Transfer.** Write `login-types.ts` that fetches
   `http://localhost:8080/api/login` (POST, JSON body `{ username: "wen",
   password: "demo" }` — the literal demo password from the mock README),
   checks `res.ok` manually, and prints the token. No `any`; annotate only
   function parameters. The response can stay inferred from `res.json()`.

<details><summary>Solution 1</summary>

```typescript
function summarize(tasks: { status: string }[]): string {
  const done = tasks.filter((t) => t.status === "done").length;
  return `${tasks.length} task(s), ${done} done`;
}
console.log(
  summarize([
    { status: "done" },
    { status: "todo" },
  ]),
); // 2 task(s), 1 done
```

`{ status: string }[]` is an inline shape annotation — [Lesson 07](07-typescript-ii-shapes-and-narrowing.md)
names these properly with `interface`/`type`.
</details>

<details><summary>Solution 2</summary>

Error 1: `who` has no type — strict mode forbids implicit `any` (TS7006).
Error 2: `message` is used where it may never have been assigned (TS2454,
"used before assigned").

```typescript
function greet(who: string | undefined): string {
  let message = "Hello stranger"; // give it a definite initial value
  if (who) {
    message = "Hello " + who;
  }
  return message.toUpperCase();
}
```

Both fixes are the strict-mode classics: annotate the parameter, initialize
the `let`.
</details>

<details><summary>Solution 3</summary>

```typescript
async function getToken(username: string): Promise<string> {
  const res = await fetch("http://localhost:8080/api/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password: "demo" }),
  });
  if (!res.ok) {
    throw new Error(`login failed: ${res.status}`);
  }
  const data = await res.json();
  return data.token; // data is any from .json() for now — Lesson 07 fixes this
}
getToken("wen").then((token) => console.log(token));
```

`res.json()` gives you `any` — the compiler blindly trusts you. That hole is
precisely what Lessons 07–08 and (at runtime) Lesson 25's zod exist to close.
</details>

## Checkpoint

1. What happens to type annotations when your code runs in the browser, and
   what does that imply for validating API responses?
<details><summary>Answer</summary>
Erased — invisible at runtime. Types cannot validate untrusted data; a
`: Project` annotation on parsed JSON is a claim, not a check. Runtime
validation needs a library (zod, Lesson 25).
</details>

2. Why does `const x = 5` infer type `5` but `let x = 5` infer `number`?
<details><summary>Answer</summary>
A `const` binding can never hold another value, so the compiler narrows to
the exact literal; `let` may be reassigned to any number, so it widens to
`number`.
</details>

3. Which dependencies go under `devDependencies` and why are `typescript` and
   `tsx` there?
<details><summary>Answer</summary>
Tools used only during development/building, never by the running app. The
compiler and the dev-runner exist only on dev machines and CI — the shipped
JS needs neither.
</details>

4. You see `error TS2345: Argument of type 'string' is not assignable to
   parameter of type 'number'.` What are the two most likely causes?
<details><summary>Answer</summary>
You called a function with the wrong-typed argument (e.g. forgot
`Number(...)` on a CLI/query-string value), or the parameter's declared type
is wrong/stale versus how callers actually use it.
</details>

5. The mock API's `/api/me` returns `{ id, name, role }` — all strings. Write
   a function signature `describeUser(...)` that takes that object with a
   fully inline type annotation and returns `string`.
<details><summary>Answer</summary>

```typescript
function describeUser(user: { id: string; name: string; role: string }): string {
  return `${user.name} (${user.role})`;
}
```
</details>

## Primary source

[TypeScript Handbook — Everyday Types](https://www.typescriptlang.org/docs/handbook/2/everyday-types.html)
The official handbook's core chapter: primitives, arrays, type aliases vs
interfaces, structural typing — written by the TS team in the exact vocabulary
your teammates use in code review.

## Before next session

Run `pnpm tsc --noEmit types.ts` with `"strict": false`, introduce one
deliberate type error, and note the compiler goes quiet. Set strict back.
Then bring your `login-types.ts` to
[Lesson 07](07-typescript-ii-shapes-and-narrowing.md), where we give those
API responses real named shapes.
