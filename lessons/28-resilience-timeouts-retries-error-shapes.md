# Lesson 28 — Resilience: timeouts, retries, error shapes

*Phase 4 · Session 28 of 38 · ~2 hours · Prereqs: Lessons 05, 25, 26, 27*

## Why this matters

The mock backend contains two deliberately hostile endpoints — `/api/flaky` (500s ~30%
of the time) and `/api/slow` (always sleeps 2s) — because real backends are hostile
*all the time*: pods restart, networks stutter, deploys hang. Your team expects the BFF
to absorb that: never wait forever, retry only what's safe, and hand the UI exactly one
error shape so it renders one error component instead of six. MISSION item 4's bug
traces start with an error the UI can actually describe.

## What you'll be able to do

- Put a hard timeout on every upstream `fetch` with `AbortSignal.timeout`
- Catch the timeout distinctly from other failures and label it
- Choose a retry policy and justify it (which methods, which statuses, how many)
- Configure TanStack Query's retry behaviour per query, and stop it retrying 4xx
- Normalise every failure — Go's `{"error":...}`, timeouts, validation crashes,
  network deaths — into a single `UiError` shape the UI renders once

## Principles

### Never await upstream without a deadline

`fetch` has no default timeout — if Go hangs, your route handler hangs, the browser
tab hangs, and users refresh (multiplying the hang). The fix is one option:

```ts
const res = await fetch(url, { signal: AbortSignal.timeout(5_000) });
```

`AbortSignal.timeout(ms)` (MDN's page below) returns a signal that fires after the
deadline and rejects the fetch with a `DOMException` whose `name` is `"TimeoutError"`.
Two rules: the budget covers connect + body read, so 5s means "5s total, not 5s of
silence"; and pick a number *under* your users' patience — 1–5s for interactive calls.
Your `/api/slow` endpoint exists so you can watch this fire on purpose.

### Retry policy: not everything deserves a second chance

Retries turn one hiccup into a storm if you're careless. The policy that survives code
review:

- **Retry only idempotent requests.** `GET /api/tasks` twice is harmless; `POST
  /api/tasks` twice is two tasks — unless you add idempotency keys (beyond scope;
  mention it in PRs when asked).
- **Retry only transient failures.** 5xx (Go is sick right now) and network/timeout
  errors: yes. 4xx: never — 401 won't fix itself, 404 never grows a project.
- **Bound it.** 2–3 attempts, exponential backoff (1s, 2s, 4s), with jitter. TanStack
  Query gives you backoff for free.
- **Only on the read path in TanStack** — mutations don't auto-retry, which is correct.

For the flaky endpoint (30% failure), one retry cuts the observed failure rate to
~9% (0.3 × 0.3); two retries, ~2.7%. That math is why "retry: 2" is the common default.

### One error shape, produced at the boundary

Go already gives you one error shape (`{"error": "..."}`, Lesson 24). Your job is to
extend it to *every* failure mode and normalize upstream variations — including your
own crashes — into a single shape:

```ts
// lib/errors.ts
export type UiErrorCode =
  | "unauthorized"   // 401 — go log in again
  | "not_found"      // 404 — resource gone
  | "bad_request"    // 4xx from Go, or your zod rejected the body
  | "upstream"       // 5xx — Go is sick; retryable
  | "timeout"        // AbortSignal fired
  | "network"        // couldn't reach Go at all
  | "unknown";       // anything else (bugs)

export interface UiError {
  code: UiErrorCode;
  message: string;   // safe to show a user
  retryable: boolean;
}

export function toUiError(err: unknown): UiError {
  if (err instanceof UpstreamError) {
    // err.status came from Go — map by class, keep Go's message
    const code = err.status === 401 ? "unauthorized"
      : err.status === 404 ? "not_found"
      : err.status >= 500 ? "upstream"
      : "bad_request";
    return { code, message: err.message, retryable: code === "upstream" };
  }
  if (err instanceof Error && err.name === "TimeoutError") {
    return { code: "timeout", message: "The server took too long to respond.", retryable: true };
  }
  if (err instanceof z.ZodError) {
    // contract violation: log the detail, tell the user something calm
    console.error("upstream contract broken", err.issues);
    return { code: "unknown", message: "Unexpected data from the server.", retryable: false };
  }
  if (err instanceof TypeError) {
    // fetch rejects with TypeError on DNS/connection failures
    return { code: "network", message: "Cannot reach the server.", retryable: true };
  }
  return { code: "unknown", message: "Something went wrong.", retryable: false };
}
```

Every route handler's `catch` becomes one line: `return NextResponse.json(toUiError(err), { status: statusFor(err.code) })`.
Every `queryFn` throws things the UI can *classify*, not just display — `retryable`
drives whether you show a Retry button.

### Timeouts live in callGo; retries live in TanStack

Put `AbortSignal.timeout` inside `callGo` (Lesson 25) — one knob for every upstream
call, server-side. Put retry policy in TanStack Query — because retries belong to the
*caller*, which knows if the user is still waiting and whether the query is visible.
A retried route handler that succeeds on attempt 2 returns one success to the browser
while hiding that Go was limping; caller-side retries keep that visible in
`isFetching`/`fetchStatus`, which is honest UI.

## Follow along

`~/teampulse`, Go mock running. You logged in in Lesson 27, so you have the `tp_token`
cookie. Add `GO_API_BASE_URL=http://localhost:8080` to `.env.local` if Lesson 29
hasn't formalized it yet; the fallback in `callGo` covers it meanwhile.

1. Add the timeout to `lib/upstream.ts`'s `callGo`:

   ```ts
   const res = await fetch(`${GO_API}${path}`, {
     ...init,
     cache: "no-store",
     signal: AbortSignal.timeout(5_000),
   });
   ```

2. Build the error module. Add `statusFor` next to `toUiError` in `lib/errors.ts`:

   ```ts
   const STATUS_BY_CODE: Record<UiErrorCode, number> = {
     unauthorized: 401, not_found: 404, bad_request: 400,
     upstream: 502, timeout: 504, network: 502, unknown: 500,
   };
   export const statusFor = (code: UiErrorCode) => STATUS_BY_CODE[code];
   ```
   Note `upstream` maps to **502 Bad Gateway** and `timeout` to **504 Gateway
   Timeout** — as the proxy, you report upstream sickness with gateway codes, keeping
   5xx reserved for actual bugs in your own handlers.
   Expected: `pnpm tsc --noEmit` clean after wiring the imports.

3. Retrofit the catch blocks on two routes (`app/api/projects/route.ts` and
   `app/api/tasks/route.ts`) to the one-line form:

   ```ts
   } catch (err) {
     const ui = toUiError(err);
     return NextResponse.json(ui, { status: statusFor(ui.code) });
   }
   ```
   Expected: `curl -si localhost:3000/api/projects` (log out first, or use plain curl
   with no cookie) → `401` and `{"code":"unauthorized","message":"not logged in","retryable":false}`.
   One shape, everywhere.

4. **Timeout demo with /api/slow.** Add `app/api/slow/route.ts`:

   ```ts
   import { NextResponse } from "next/server";
   import { z } from "zod";
   import { callGo, UpstreamError } from "@/lib/upstream";
   import { toUiError, statusFor } from "@/lib/errors";

   const SlowSchema = z.object({ ok: z.boolean(), note: z.string() });

   export async function GET() {
     try {
       return NextResponse.json(await callGo("/api/slow", SlowSchema));
     } catch (err) {
       const ui = toUiError(err);
       return NextResponse.json(ui, { status: statusFor(ui.code) });
     }
   }
   ```
   Now drop `callGo`'s timeout to `1_000` temporarily and run
   `time curl -s localhost:3000/api/slow`.
   Expected: it returns *at ~1 second* (not 2):
   `{"code":"timeout","message":"The server took too long to respond.","retryable":true}`
   with a 504. Go's 2s sleep was abandoned server-side. Restore 5_000, re-run: ~2s,
   `{"ok":true,"note":"I slept 2 seconds on purpose"}`.

5. **Retry demo with /api/flaky.** Add `app/api/flaky/route.ts` exactly like step 4's
   but with `FlakySchema = z.object({ ok: z.boolean(), note: z.string() })`
   (`/api/flaky` requires auth in `main.go`, so pass `{ headers: authHeaders(req) }`
   and call it from a page, or test via `curl --cookie "tp_token=..."`). Then on the
   client page:

   ```tsx
   const { data, isError, error, failureCount } = useQuery({
     queryKey: ["flaky"],
     queryFn: () => getJson("/api/flaky"),
     retry: 2,
     retryDelay: (attempt) => Math.min(1000 * 2 ** attempt, 4000),
   });
   ```
   Open the Network tab and click Reload a handful of times. Expected: some page loads
   show one red `flaky` request that then succeeds on a second or third — `retry: 2`
   re-firing after ~1–2s. When all three attempts fail, `isError` flips true and your
   error UI shows `error.message` with "unexpected data"/upstream wording plus a Retry
   button gated on `error.retryable` (parse it in `getJson`: throw an `Error` with the
   `code` attached, or reject with the `UiError` itself).
   Count: of ~10 reloads, you should see near-zero visible failures — 30% per-attempt
   became ~3% after two retries.

6. Prove 4xx is never retried. Set `retry` to a logger temporarily:
   ```ts
   retry: (failureCount, err) => {
     console.log("retry decision", failureCount, err);
     return false; // observe, then decide
   },
   ```
   Clear cookies, reload a data page, watch the console. Expected: exactly **one**
   fetch for the 401 — your `getJson` threw an error marked non-retryable; once you
   restore `retry: 2`, add the guard
   `retry: (count, err) => (err as UiError).code === "unauthorized" ? false : count < 2`
   and confirm the 401 still yields one request, while flaky retries again.

## Your turn

1. Mutations and retries: your `AddTaskForm` (Lesson 26) posts to `/api/tasks`. The
   user double-clicks "Add task" while `isPending` is false for a frame. What protects
   Go from duplicate tasks, and what doesn't?

   <details><summary>Solution</summary>
   `disabled={addTask.isPending}` stops the second click while the request is in
   flight. TanStack won't auto-retry mutations, so no accidental POST storm. What
   *doesn't* protect: a user who clicks, sees an error (e.g. the 30% flaky failure),
   and hits the button again — POST is not idempotent, so each click is a real task.
   Real teams fix this with an idempotency key: the client generates one UUID per form
   session and Go dedupes on it.
   </details>

2. Write the client-side counterpart of `toUiError`: make `getJson` parse the BFF's
   `UiError` JSON on non-ok responses and `throw` it, so components can render
   `retryable`-gated Retry buttons.

   <details><summary>Solution</summary>

   ```ts
   import { UiError } from "@/lib/errors"; // shared shape; safe to import client-side

   export class ApiRequestError extends Error implements UiError {
     constructor(
       public code: UiError["code"],
       message: string,
       public retryable: boolean,
     ) {
       super(message);
     }
   }

   async function getJson<T>(url: string, init?: RequestInit): Promise<T> {
     const res = await fetch(url, init);
     if (!res.ok) {
       const body = (await res.json().catch(() => null)) as UiError | null;
       throw new ApiRequestError(
         body?.code ?? "unknown",
         body?.message ?? `request failed: ${res.status}`,
         body?.retryable ?? false,
       );
     }
     return res.json() as Promise<T>;
   }
   ```
   In the component: `{isError && error.retryable && <button onClick={() => refetch()}>Retry</button>}`.
   </details>

3. Your team copies the BFF pattern; a junior wraps *every* call in
   `retry: 5, retryDelay: 10_000`. Using flaky's 30% number and slow's 2s, calculate
   roughly what one impatient user hitting Reload five times does to Go during an
   outage, and name the two policy lines from this lesson that prevent it.

   <details><summary>Solution</summary>
   Each page load = up to 6 attempts (1 + 5 retries) spaced 10s. Five reloads ≈ 30
   attempts per user per minute against an already-sick Go — the outage extends
   because Go is burning cycles on retries instead of serving whoever's healthy. The
   two violated lines: "bound it — 2–3 attempts with exponential backoff" (capped
   total attempts) and "retry only transient failures" (with jitter so reload-storms
   don't synchronize). Retries are a budget, not a reflex.
   </details>

## Checkpoint

1. What exception does an expired `AbortSignal.timeout` produce, and how do you catch
   that case specifically?

   <details><summary>Answer</summary>
   A `DOMException` (seen as `Error` in Node/undici) with `name === "TimeoutError"`.
   Check `err instanceof Error && err.name === "TimeoutError"` in the catch, before the
   generic branches — `toUiError` maps it to `code: "timeout"`.
   </details>

2. Why is `retryable` false for 401 and true for 500? Give the reasoning, not the
   answer key.

   <details><summary>Answer</summary>
   401 is a *deterministic* consequence of your request state (no/expired credential) —
   an identical retry produces an identical 401; only a credential change can fix it.
   500 is *state-of-the-world* — Go was sick at that instant and might not be in 500ms.
   Retry only what can change between two identical attempts.
   </details>

3. (Go API question) `/api/slow` sleeps exactly 2 seconds. With
   `AbortSignal.timeout(1_000)` in `callGo`, what does the user see and what does Go do?

   <details><summary>Answer</summary>
   The user sees your 504 `{"code":"timeout",...}` at ~1s. Go is oblivious — it still
   sleeps its full 2s and then writes into a connection nobody reads. Timing out is
   *your* protection, not Go's; which is also why the timeout must be shorter than
   whatever the user waits.
   </details>

4. Where do timeouts go and where do retries go in this architecture, and why that
   split?

   <details><summary>Answer</summary>
   Timeout in `callGo` (server-side, every upstream fetch gets a deadline regardless of
   caller). Retries in TanStack Query (caller-side, because the caller knows visibility
   and user presence, and can back off per query). Route handlers that silently retry
   hide upstream sickness from the UI.
   </details>

5. Name four distinct failure modes and the `UiError.code` each produces.

   <details><summary>Answer</summary>
   Go 401 → `unauthorized`; Go 500 (flaky) → `upstream`; deadline expired → `timeout`;
   Go process not running (fetch TypeError) → `network`; zod contract violation →
   `unknown` (with detail in server logs); malformed user input → `bad_request`. The UI
   renders one component that reads `code`/`retryable`.
   </details>

## Primary source

[MDN — AbortSignal.timeout()](https://developer.mozilla.org/en-US/docs/Web/API/AbortSignal/timeout_static) —
the definitive description of the static method, the `TimeoutError` DOMException it
raises, and how it composes with `fetch`'s `signal` option.

## Before next session

In `~/teampulse`, create `.env.local` containing exactly:
```
GO_API_BASE_URL=http://localhost:8080
```
Restart `pnpm dev` and confirm nothing changed in behaviour — you're about to find out
*why* it still worked when you delete the fallback, in Lesson 29.
