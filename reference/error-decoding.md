# Error decoding — "what does this error mean?"

Rule one: **read the first line, ignore the rest.** The compiler names the file, line, both types, and what it wanted. The follow-on "did you mean…" noise is usually derivative.

## TypeScript compiler errors

| Code | Raw shape | Plain English | Fix |
|---|---|---|---|
| TS2345 | `Argument of type 'X' is not assignable to parameter of type 'Y'` | You handed a function the wrong shape | Match the param's type at the call site |
| TS2322 | `Type 'X' is not assignable to type 'Y'` | Assignment/return shape mismatch (e.g. `string` into `priority: number`) | Convert (`Number(x)`) or fix the type |
| TS2339 | `Property 'foo' does not exist on type 'Bar'` | Typo, wrong type, or field missing from your interface | Check spelling against Go's JSON tags; add the field |
| TS7006 | `Parameter 'x' implicitly has an 'any' type` | Strict mode demands an annotation on a param | Annotate it |
| TS18046 | `'x' is of type 'unknown'` | You're using an unvalidated/`catch` value | Narrow or zod-parse first |
| TS2532 | `Object is possibly 'undefined'` | Optional/`.find()` result used without a guard | `?.`, `??`, or `if (!x) return` |
| TS2345-ish: `is missing properties from type` | shape has too few fields | Provide the missing fields or `Pick`/`Omit` the type |
| TS2353 | `Object literal may only specify known properties, and 'foo' does not exist in type 'Y'` | Extra/misspelled key in an object literal | Rename/remove the key |
| TS2739 | `Type 'X' is missing the following properties from type 'Y': a, b` | Object literal missing required props (classic untyped-props bug) | Pass all required props |
| TS2551 | `Property 'createdat' does not exist... Did you mean 'createdAt'?` | Casing typo — Go's `CreatedAt` is `createdAt` in JSON | Fix casing exactly |
| TS2307 | `Cannot find module '@/lib/go-api' or its corresponding type declarations` | Wrong path/alias, or file doesn't exist | Check `tsconfig.json` `paths`, filename case |
| TS7016 | `Could not find a declaration file for module 'x'. implicitly has an 'any' type` | Untyped npm package | `pnpm add -D @types/x` or write a `.d.ts` |

**TeamPulse examples:**

```ts
// TS2345 — string where number expected (search params are always strings)
openTask(task.priority);            // ❌ if priority came from searchParams.get() unconverted
openTask(Number(searchParams.get("priority") ?? "1"));   // ✅

// TS2339 — field named after Go's struct, not its JSON tag
const n = project.Name;             // ❌ Property 'Name' does not exist on type 'Project'
const n = project.name;             // ✅ json tag is lowercase

// TS2532 — optional assignee used directly
const c = task.assignee.charAt(0);  // ❌ possibly undefined
const c = task.assignee?.charAt(0) ?? "U";  // ✅

// TS18046 — JSON.parse / catch result
try { await callGo(); } catch (e) {
  showError(e.message);              // ❌ 'e' is of type 'unknown'
  showError(e instanceof Error ? e.message : String(e)); // ✅
}

// TS2739 — forgot required props on a component
<TaskRow task={t} />                 // ❌ missing 'onSelect' declared in TaskRowProps
<TaskRow task={t} onSelect={pick} /> // ✅
```

## ESLint — react-hooks rules

**`react-hooks/exhaustive-deps` (warning):** "React Hook useEffect has a missing dependency: 'projectId'." Every value your effect/callback *reads* from render scope must be in its dep array — otherwise the effect closes over a stale render's value. Fix: add the dep. If adding it causes a loop, the real problem is that the value is recreated each render (a new object/function) — move it out, `useMemo`/`useCallback` it, or restructure the effect. Never paste `// eslint-disable` to silence it on data-fetching effects.

**`react-hooks/rules-of-hooks` (error):** "React Hook 'useState' is called conditionally." Hooks must run in the **same order every render**: top level of a component or custom hook only — never inside `if`, loops, `try`, or nested functions. Fix: hoist the call, put the condition *around* the code that uses the result, or extract a child component that's mounted/unmounted instead of conditionally calling hooks.

## Next.js runtime errors

**1. Hydration mismatch** — "Hydration failed because the initial UI does not match what was rendered on the server" / "Text content did not match."
Server HTML ≠ first client render. Usual causes: `Date.now()`/`Math.random()`/locale formatting during render, browser-only values (`window`) rendered server-side as `undefined`, invalid nesting (`<div>` inside `<p>`, block inside inline). Fix: make render deterministic — pass server-computed strings as props, or render browser-only UI after mount:

```tsx
const [mounted, setMounted] = useState(false);
useEffect(() => setMounted(true), []);
if (!mounted) return <span suppressHydrationWarning>—</span>;  // prefer: deterministic skeleton
return <span>{new Date(task.dueDate).toLocaleDateString()}</span>;
```

**2. `useState only works in a Client Component`** — "useState, useEffect, useLayoutEffect... but it was rendered as a Server Component."
You used a hook (or an event handler) in a file without `"use client"`, or a client hook leaked through an import. Fix: add `"use client"` to the top of the file that owns the state — or if the state is trivial, move it into a small leaf client component and keep the page a server component.

**3. `cookies()` in Next 15** — "You're calling cookies() during render but the return type has changed / Missing 'await'."
In Next 15 `cookies()`, `headers()`, and `params` are **async**. `const t = cookies().get("tp_token")` fails or returns a Promise. Fix:

```ts
import { cookies } from "next/headers";
const token = (await cookies()).get("tp_token")?.value;   // ✅ in an async server component
```

**4. Unsupported (PNA) / CORS failure** — "Failed to fetch" in the browser console; Network tab shows the request blocked.
Two flavors: (a) **CORS** — the browser blocked a *browser→Go* call because Go didn't send `Access-Control-Allow-Origin`. Fix: the browser should only ever talk to your BFF (`/api/*`, same origin — no CORS); keep Go calls server-side. (b) **Private Network Access** — Chrome blocks a public page fetching a private (`localhost`) address. Fix: same-origin BFF, or dev on `http://localhost:3000`, not a tunnel URL.

**5. Unhandled fetch rejection** — "Unhandled Runtime Error: TypeError: Failed to fetch" / "This Suspense boundary received an error before it finished executing."
`fetch` rejects on network failure/timeout and nobody caught it — either an `await fetch` in a server component with no `try/catch` (crashes the page into `error.tsx` or a 500) or an un-awaited Promise in a client handler. Fix: wrap upstream calls in `try/catch`, distinguish `TimeoutError`/`AbortError`, return a typed error response from route handlers, and give the segment an `error.tsx`. Note `fetch` does **not** reject on 404/500 — always check `res.ok` too.

## Triage order for any error

1. First line: code + both types + file:line.
2. Is it TS (build/type time) or runtime (browser console / server terminal)?
3. Network tab: request to `/api/*` failing → BFF bug; request to Go failing → upstream; request absent → client bug.
4. Only then read the stack/suggestions.

---
*Related lessons: [31 decoding TS & ESLint errors](../lessons/31-decoding-typescript-and-eslint-errors.md) · [32 debugging](../lessons/32-debugging-devtools-network-server-logs.md) · [28 resilience](../lessons/28-resilience-timeouts-retries-error-shapes.md) · [27 auth & cookies](../lessons/27-auth-tokens-and-cookies-through-the-bff.md)*
