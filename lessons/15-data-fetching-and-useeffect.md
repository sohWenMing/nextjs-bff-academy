# Lesson 15 — Data Fetching & the Truth About `useEffect`

*Phase 2 · Session 15 of 38 · ~2 hours · Prereqs: Lessons 05, 13, 14*

## Why this matters

This is the lesson that keeps you from the most common bug a new React
developer ships: fetching inside `useEffect` "because that's where side
effects go." You'll write the anti-pattern, watch it double-fire, then swap
in what the React team endorses — fetch **on events**, with explicit
loading/error state. That's what P1 needs tomorrow. Lesson 22 moves the
fetching server-side; the loading/error discipline survives that surgery.

## What you'll be able to do

- You can fetch in an event handler (`async/await` + `try/catch`, Lesson 05)
  and render every state an async UI has — loading, error, success,
  empty-vs-error.
- You can explain two failure modes of fetch-in-`useEffect` (StrictMode
  double-invocation, races), and state the one legitimate shape —
  fetch-on-mount with `AbortController` — and why it's a last resort.
- You can send a JSON POST and an `Authorization: Bearer <token>` GET to the
  mock Go API from the browser.

## Principles

### A Promise is not a string — await before setState

From Lesson 05: `fetch` returns a `Promise<Response>`; the data doesn't exist
yet. Step 5 builds the canonical handler (`handleLoadTasks`); these
are the lines that bite:

Non-obvious lines:

- `if (!res.ok)` — `fetch` **only rejects on network failure**. A 401 or 500
  is a *successful* fetch (Lesson 05's gotcha); without your own check
  you'll render `{error: "missing or invalid token"}` as if it were data.
- `finally { setLoading(false) }` — loading must end on *every* exit path;
  `finally` is the only place that guarantees it.
- `null` vs `[]`: "never loaded" vs "loaded, zero items" — "No tasks" and
  "Backend unreachable" must not lie to each other (Lesson 12's checkpoint 5).

The Go backend's errors are always `{"error": "message"}`
(`backend-mock/README.md`) — reading it into the thrown message is why the
error UI shows `missing or invalid token`, not `HTTP 401`.

### Fetch in event handlers. That's the pattern.

When does data load? When the user *does something*: submits the login form,
clicks Retry — an **event**. `onClick`/`onSubmit` → `async` handler → fetch →
three `setState`s. No hook required.

### The anti-pattern you're required to try: fetch-in-`useEffect`

`useEffect(fn)` runs `fn` after the component paints — meant for
*synchronizing with external systems* (subscriptions, timers), not for
"place I do stuff". The old habit — `useEffect(() => fetch(url).then(r =>
r.json()).then(setTasks), [url])` — ships with no cancel and no error branch.

**StrictMode double-fire:** in dev, React mounts → unmounts → remounts, so
the effect runs twice and the Network tab shows two identical requests —
tolerate that habit until a double `POST` charges someone twice. **Races:**
two renders fire two fetches; if the earlier response lands *last*, the
screen shows the previous query's results; and loading/error state gets
trapped in a closure, unreachable from a Retry button. The React team's
verdict is the doc linked below: **You Might Not Need an Effect** — most
effects are events mislabeled as lifecycle.

### The legitimate residue: fetch-on-mount, done right

Sometimes data must load when the component appears, with no user event. The
sanctioned shape adds cancellation:

```tsx
useEffect(() => {
  const controller = new AbortController();
  (async () => {
    try {
      const res = await fetch(url, { signal: controller.signal });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      setData(await res.json());
    } catch (err) {
      if (controller.signal.aborted) return; // expected during unmount
      setError(String(err));
    }
  })();
  return () => controller.abort(); // React calls this on unmount (and StrictMode's fake unmount)
}, []);
```

The cleanup function is the effect's promise to undo itself;
`AbortController` cancels the in-flight request, and the `aborted` check
keeps the expected unmount abort out of your error UI. Even this is a
stopgap — Lessons 19/22 move page-initial fetching server-side.

### Where the token goes (temporary)

The token lives in `useState<string | null>` — per tab, gone on reload. Fine
for P1 (the mock's token is worthless); the rule: **whatever can't leak to
page JS must never be in it** — why real secrets live in route handlers
(Lesson 20+) and the session in an httpOnly cookie (Lesson 27).

## Follow along

**Terminal 1:** `cd backend-mock && go run main.go`. **Terminal 2:** `pnpm dev` in `~/teampulse`.

### 1. Trust the backend from the terminal

`curl -s http://localhost:8080/api/health` → `{"status":"ok",...}`;
`curl -si http://localhost:8080/api/projects | head -1` → `HTTP/1.1 401
Unauthorized` — you met `res.ok === false` before the browser did.

### 2. One-time CORS patch (browser-only problem)

On `http://localhost:3000`, devtools console:
`fetch("http://localhost:8080/api/health").then(r => r.json()).then(console.log)`.

Expected: a red CORS error — `:8080` didn't send
`Access-Control-Allow-Origin` (curl doesn't care; browsers do). CORS lives
on **the server being called**, so patch `backend-mock/main.go`: add above
`main()`:

```go
// withCORS lets the Next.js dev server (port 3000) call this API from a browser.
// A real backend would list exact origins; "*" is demo-grade.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == http.MethodOptions { // preflight
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

Change `main()`'s last line to
`log.Fatal(http.ListenAndServe(":8080", withCORS(mux)))`, restart, re-run:
`{status: 'ok', time: '...'}`. Lesson 29 gives CORS a proper page.

### 3. Event-driven fetch: a health widget

Create `components/health-widget.tsx`:

```tsx
"use client";

import { useState } from "react";

type State =
  | { kind: "idle" } | { kind: "loading" }
  | { kind: "ok"; time: string } | { kind: "error"; message: string };

export function HealthWidget() {
  const [state, setState] = useState<State>({ kind: "idle" });

  async function check() {
    setState({ kind: "loading" });
    try {
      const res = await fetch("http://localhost:8080/api/health");
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const body = (await res.json()) as { time: string };
      setState({ kind: "ok", time: body.time });
    } catch (err) {
      setState({ kind: "error", message: err instanceof Error ? err.message : "fetch failed" });
    }
  }

  return (
    <>
      <button onClick={check} disabled={state.kind === "loading"}>
        {state.kind === "loading" ? "Checking…" : "Check backend"}
      </button>
      {state.kind === "ok" && <p>Backend up — server clock {state.time}</p>}
      {state.kind === "error" && <p style={{ color: "crimson" }}>{state.message}</p>}
    </>
  );
}
```

Render `<HealthWidget />` in `page.tsx`. Click: `Checking…` → `Backend up`.
Stop the Go server, click: `Failed to fetch`. Restart, click, recover — no
reload anywhere. One `useState` holding a **discriminated union** (Lesson 07)
makes impossible states (`ok` *and* `error`) unrepresentable.

### 4. The anti-pattern, witnessed

Open `components/health-widget.tsx` and do damage: `console.log("BAD effect
fired")` at the top of the body, then move the `check()` **call** into
`useEffect(() => { check(); }, [])` and delete the button's `onClick` (add
`useEffect` to the react import). Refresh with devtools open. Expected: the
log **twice**, **two** `health` requests, and nothing can Retry — with a
`POST` where that `GET` is, it's the double-charge demo. Restore the
event-handler version.

### 5. Real flow: login POST, then authorized GET

Wire Lesson 14's `LoginForm` to the API in `page.tsx` (now `"use client"`,
Lesson 18 formalizes):

```tsx
"use client";

import { useState } from "react";
import { LoginForm } from "@/components/login-form";
import type { Task } from "@/lib/types";

export default function Home() {
  const [token, setToken] = useState<string | null>(null);
  const [tasks, setTasks] = useState<Task[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleLogin(username: string, password: string) {
    try {
      const res = await fetch("http://localhost:8080/api/login", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ username, password }) });
      const body = (await res.json()) as { token?: string; error?: string };
      if (!res.ok || !body.token)
        return setError(body.error ?? `HTTP ${res.status}`); // own error slot: Lesson 16
      setToken(body.token);
    } catch {
      setError("backend unreachable");
    }
  }

  async function handleLoadTasks() {
    if (!token) return;
    setLoading(true); setError(null);
    try {
      const res = await fetch("http://localhost:8080/api/tasks", { headers: { Authorization: `Bearer ${token}` } });
      if (!res.ok) {
        const b = (await res.json()) as { error?: string };
        throw new Error(b.error ?? `HTTP ${res.status}`);
      }
      setTasks((await res.json()) as Task[]);
    } catch (err) {
      setError(err instanceof Error ? err.message : "failed");
    } finally {
      setLoading(false);
    }
  }

  if (!token) {
    return (
      <main>
        <h1>TeamPulse</h1> <LoginForm onLogin={handleLogin} />
        {error && <p style={{ color: "crimson" }}>{error}</p>}
      </main>
    );
  }

  return (
    <main>
      <h1>TeamPulse — signed in</h1>
      <button onClick={handleLoadTasks} disabled={loading}>{loading ? "Loading…" : "Load tasks"}</button>
      <button onClick={() => setToken(null)}>Logout</button>
      {error && <p style={{ color: "crimson" }}>{error}</p>}
      {tasks && <p>{tasks.length} tasks loaded</p>}
    </main>
  );
}
```

Test every path. Any username works; the password is always the word
`demo` (d-e-hyphen-m-o):

1. Log in `wen` / `demo` → signed-in view; reload → login again (the token
   lived in React memory, as designed).
2. Wrong password → red `invalid credentials`, the backend's own
   `{"error":...}` message.
3. `Load tasks` → `5 tasks loaded`; Network shows the auth header.
   Click twice fast → two requests, second wins. If this were an
   effect-fetch, what's your cancellation plan?

## Your turn

1. **Warm-up:** in `HealthWidget`, walk the whole state machine in React
   DevTools: `idle → loading → ok` → kill backend → `loading → error` →
   restart → `loading → ok`. The disabled-state never lied.

2. **Transfer:** add a `Load /api/me` button next to `Load tasks`: fetch
   with the token header, render `name` and `role`. Give it its *own*
   loading/error states — one button's loading must never cover the other's
   request. Expected: `Wen · admin`.

3. **Race it:** make `handleLoadTasks` `await` a fetch of
   `http://localhost:8080/api/slow` first (ignore the result), then load
   tasks. Click `Load tasks`, then `Logout` mid-load. Predict, then test:
   the response lands after logout — harmless *here*? Harmful if the screen
   rendered off `tasks` instead of `token`?

<details><summary>Solution</summary>

1. `idle → loading → ok → loading → error → loading → ok`, with `disabled`
   following `kind === "loading"` throughout.

2. Copy `handleLoadTasks` verbatim and rename: states `me`/`meError`/
   `loadingMe`, URL `…/api/me`, render `{me && <p>{me.name} · {me.role}</p>}`.
   The point is the *separate* state triple — one `loading` flag must never
   cover two different requests.

3. The late response calls `setTasks` while `token === null` — harmless
   today only because `if (!token)` early-returns the login view. Key the
   screen off `tasks` instead and a logged-out user sees a stranger's list.
   Robust fix: an `AbortController` per load, aborted on logout.

</details>

## Checkpoint

1. Why must `res.ok` be checked manually when `fetch` "didn't throw"? What
   are the *only* things `fetch` rejects on?
2. Which user-facing message maps to `null` data vs `[]` data, and why does
   confusing them hide outages?
3. Name the two failure modes of fetch-in-`useEffect` you demonstrated, and
   what devtools showed for each.
4. Sketch fetch-on-mount-with-cleanup: what cancels, which function returns
   the cleanup, what the `aborted` check is for.
5. From memory, write the authorized tasks-fetch with its `!res.ok` branch —
   then name which *future* lesson takes the token out of browser state,
   and how.

<details><summary>Answers</summary>

1. `fetch` treats any HTTP response as success — 401/500 included; it only
   rejects on network failure (DNS, refused connection, CORS, abort).
2. `null` = never loaded (skeleton); `[]` = zero items ("No tasks").
   Rendering the empty state for a failure tells users the board is empty
   while the backend is down.
3. StrictMode double-fire (log twice, two Network requests per mount) and
   races (earlier request resolves last, stale data wins).
4. `new AbortController()`, `signal` into fetch; the effect **returns**
   `() => controller.abort()`; `if (controller.signal.aborted) return` keeps
   the unmount abort out of the error UI.
5. Step 5's `handleLoadTasks`: auth header, `if (!res.ok)` → read `{error}`
   → throw, then `setTasks`. Lesson 27 moves the session to an httpOnly
   cookie via the BFF (Lessons 20/22 first move the calling server-side).

</details>

## Primary source

[react.dev — You Might Not Need an Effect](https://react.dev/learn/you-might-not-need-an-effect)
— the official case that "fetch on event, not on mount" is the default.

## Before next session

Tomorrow is [P1](16-p1-teampulse-explorer-page.md) — full build, no new
concepts. Tonight: decide on paper *which component owns which piece of
state* (token? selected project? filter query?); you'll diff your sketch in
step 1.
