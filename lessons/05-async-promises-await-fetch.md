# Lesson 05 — Async: Promises, async/await, fetch

*Phase 1 · Session 05 of 38 · ~2 hours · Prereqs: [Lesson 03](03-arrays-objects-immutability.md), [Lesson 04](04-modules-and-bundler-mental-model.md)*

## Why this matters

The BFF layer is, at heart, asynchronous glue: receive a request, call the Go
backend over HTTP, transform the JSON, respond. Every step is an **async**
operation — code that starts now and finishes later — and async is the biggest
source of confusion and of bugs (unhandled rejections, missing awaits,
forgotten error paths) for programmers from synchronous languages. You will
use the TeamPulse mock Go API today, for real: health check and login.

## What you'll be able to do

- You can explain what the event loop does, at working level, in two sentences.
- You can read and write `async`/`await` code and know what a `Promise<T>` is.
- You can call `fetch`, check `response.ok`, and parse JSON cold.
- You can wrap async calls in `try/catch` and say what `fetch` does *not* throw on.
- You can run independent fetches concurrently with `Promise.all`.

## Principles

**The problem async solves.** JavaScript runs your code on a single thread.
If waiting for a network response blocked that thread, the browser could not
paint while a request was in flight; a Node server could not serve anyone
else. So JS never waits in place: slow operations (network, timers, disk) are
*handed off*, and your "when it's done" code runs later.

**The event loop, at working level.** Node runs your synchronous code first;
when it finishes, Node runs the callbacks of completed async operations from a
queue; repeat. Two rules fall out: synchronous code always wins — nothing
interleaves into the middle of a function — and a callback only runs after the
current tick completes. That is all Phase 1 needs.

**A Promise** is an object representing a value that will exist later, in one
of three states: *pending* (still working), *fulfilled* (has a value),
*rejected* (failed with an error). Notation `Promise<T>` means "a promise that
will deliver a `T`". You mostly *consume* them via `await`:

```javascript
const response = await fetch(url);  // pause here (async only) until fulfilled
```

`await` only appears inside `async` functions, which always return a Promise —
returning `5` gives callers `Promise<number>`. "Pause" is shorthand: the
function *suspends*, control returns to the event loop, and execution resumes
when the promise settles. Nothing blocks.

**`fetch`** is the built-in HTTP client (global in Node 18+ and all browsers).
It returns `Promise<Response>`. The critical gotcha, the most-misunderstood
fact in JS networking:

> `fetch` only rejects on **network failure** (DNS dead, connection refused).
> An HTTP 401 or 500 is a *perfectly fulfilled* Response. You must check
> `response.ok` (true for status 200–299) yourself.

Your team's API returns errors as `{"error": "message"}` JSON with status
codes — exactly why `response.ok` checking is non-negotiable in the BFF.

**Parsing.** `response.json()` also returns a Promise (the body may still be
streaming) and yields the parsed JS object — untyped for now; typed in
[Lesson 07](07-typescript-ii-shapes-and-narrowing.md).

**Concurrency.** Independent requests should be launched together and awaited
with `Promise.all([...])` — an array of promises in, a promise of an array
out.

**`try/catch` around `await`** reads like sync error handling: rejecting
awaits inside `try` jump to `catch` with the error object.

## Follow along

Start the mock backend in its own terminal tab if it is not already up:

```bash
cd ~/workspace/nextjs-bff-academy/backend-mock && go run main.go
# 2026/08/13 TeamPulse mock backend listening on http://localhost:8080
```

Work in `~/learn-js/lesson-05`. Type every file, run with `node <file>`.

### 1. Feel the ordering, without any API

`order.mjs` — run with `node order.mjs`:

```javascript
console.log("1 start");
setTimeout(() => console.log("2 timer fired"), 0);
console.log("3 end");
```

Expected: `1 start`, `3 end`, `2 timer fired` — even at 0 ms, the callback
waited for your synchronous code to finish. The event loop in one experiment.

### 2. Your first fetch: the health check

`health.mjs` (again `node health.mjs`):

```javascript
const API = "http://localhost:8080";
const res = await fetch(`${API}/api/health`);
console.log("status:", res.status, "ok:", res.ok);
const body = await res.json();
console.log("body:", body);
```

Expected: `status: 200 ok: true`, then a body like
`{ status: 'ok', time: '2026-...' }`. Top-level `await` works in `.mjs` files —
an ES module feature, not magic.

**Prove the gotcha.** Temporarily add:

```javascript
const bad = await fetch(`${API}/api/projects`); // needs auth, we sent no token
console.log(bad.status, bad.ok); // 401 false — but NO exception was thrown
```

`fetch` happily fulfilled: if you assumed "no throw = success", you just
ingested an error body as if it were data.

### 3. Login — POST with a JSON body

`login.mjs`:

```javascript
const API = "http://localhost:8080";
async function login(username) {
  const res = await fetch(`${API}/api/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password: "demo" }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(`login failed: ${res.status} ${err.error ?? ""}`.trim());
  }
  return res.json();
}
const { token, user } = await login("wen");
console.log(`token=${token.slice(0, 12)}… user=${user.name} (${user.role})`);
```

Expected: something like `token=tok-3f9a1c2b… user=Wen (admin)`. Any username
works; the password is literally `demo` (`backend-mock/main.go`'s demo rule).
`res.json().catch(() => ({}))` covers non-JSON error responses.

### 4. Use the token: authenticated request with real error handling

`me.mjs` — build on the previous file (type it fully):

```javascript
const API = "http://localhost:8080";
async function api(path, options = {}, token) {
  const res = await fetch(`${API}${path}`, {
    ...options,
    headers: { ...options.headers, Authorization: Bearer ${token}` },
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(`${res.status} ${err.error ?? res.statusText}`);
  }
  return res.json();
}

try {
  const loginRes = await fetch(`${API}/api/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username: "wen", password: "demo" }),
  });
  if (!loginRes.ok) throw new Error("login failed");
  const { token } = await loginRes.json();
  console.log("I am:", await api("/api/me", {}, token));
  // { id: 'u1', name: 'Wen', role: 'admin' }
  const projects = await api("/api/projects", {}, token);
  console.log("projects:", projects.map((p) => p.name));
  // [ 'Billing Service', 'Mobile App', 'Legacy Portal' ]
} catch (err) {
  console.error("FAILED:", err.message);
  process.exitCode = 1;
}
```

### 5. Kill the server, see the *other* failure

Stop the Go server (Ctrl-C in its tab) and run `node me.mjs`. Expected:
`FAILED: fetch failed` — the ONLY category `fetch` rejects for. Restart the
server before continuing.

### 6. Concurrent fetches

`parallel.mjs`:

```javascript
const API = "http://localhost:8080";
const t0 = Date.now();
// /api/slow sleeps 2s on purpose; two SEQUENTIAL awaits = ~4s
const [a, b] = await Promise.all([
  fetch(`${API}/api/slow`).then((r) => r.json()),
  fetch(`${API}/api/slow`).then((r) => r.json()),
]);
console.log(a.ok, b.ok, `took ${Date.now() - t0}ms`);
// expected: true true took ~20xx ms — concurrent, not additive
```

## Your turn

1. **Warm-up.** Write `flaky.mjs` that logs in, calls `/api/flaky` (auth
   required) ten times, and prints how many succeeded vs failed with 500.
   Expect roughly 7/3 — that endpoint fails 30% of the time deliberately.

2. **Error path.** Make a login send password `nope` and print
   `rejected: 401 invalid credentials` — using the `{"error": ...}` body the
   Go server returns, not the generic statusText.

3. **Transfer (BFF preview).** Write `board.mjs`: after login, fetch
   `/api/projects` and `/api/tasks` **concurrently** (one `Promise.all`, same
   Authorization header), then print each project with its task count, e.g.
   `Billing Service — 3 task(s)`.

<details><summary>Solution 1</summary>

```javascript
const API = "http://localhost:8080";
const loginRes = await fetch(`${API}/api/login`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ username: "wen", password: "demo" }),
});
const { token } = await loginRes.json();
let ok = 0; let failed = 0;
for (let i = 0; i < 10; i++) {
  const res = await fetch(`${API}/api/flaky`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (res.ok) ok++; else failed++;
}
console.log(`ok=${ok} failed=${failed}`);
```
</details>

<details><summary>Solution 2</summary>

```javascript
const res = await fetch(`${API}/api/login`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ username: "wen", password: "nope" }),
});
if (!res.ok) {
  const err = await res.json();
  console.log(`rejected: ${res.status} ${err.error}`);
}
```

Prints `rejected: 401 invalid credentials` — note `fetch` did not reject; the
`!res.ok` branch caught it.
</details>

<details><summary>Solution 3</summary>

```javascript
const API = "http://localhost:8080";
const loginRes = await fetch(`${API}/api/login`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ username: "wen", password: "demo" }),
});
const { token } = await loginRes.json();
const auth = { headers: { Authorization: `Bearer ${token}` } };

const getJSON = async (path) => {
  const res = await fetch(`${API}${path}`, auth);
  if (!res.ok) throw new Error(`${path}: ${res.status}`);
  return res.json();
};
const [projects, tasks] = await Promise.all([
  getJSON("/api/projects"),
  getJSON("/api/tasks"),
]);
for (const p of projects) {
  const count = tasks.filter((t) => t.projectId === p.id).length;
  console.log(`${p.name} — ${count} task(s)`);
}
```

This — login, shared header, concurrent fetches, `ok` checks, reshape — is
80% of a Phase 3 route handler.
</details>

## Checkpoint

1. Does `await` block the thread? What actually happens while it waits?
<details><summary>Answer</summary>
No. The async function suspends, control returns to the event loop to run
other work, and the function resumes when the promise settles.
</details>

2. When does `fetch` reject, and when does it fulfill? What must you check for
   the latter?
<details><summary>Answer</summary>
It rejects only on network-level failure (connection refused, DNS). HTTP
error statuses (401, 500) fulfill normally — check `response.ok` yourself.
</details>

3. What is the return type of `async function f() { return 5; }`, and what
   must callers do with it?
<details><summary>Answer</summary>
`Promise<number>`. Callers must `await f()` (or `.then`) to get the 5.
</details>

4. Why does `await Promise.all([a, b])` beat `await a; await b` when
   a and b are independent?
<details><summary>Answer</summary>
`Promise.all` starts both immediately and waits once — total ≈ the slower one.
Sequential awaits pay both latencies (the /api/slow demo: ~2s vs ~4s).
</details>

5. Against the TeamPulse mock API, write the lines that log in and print the
   token (any username, correct demo password), no error handling.
<details><summary>Answer</summary>

```javascript
const res = await fetch("http://localhost:8080/api/login", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ username: "wen", password: "demo" }),
});
const { token } = await res.json();
console.log(token);
```
</details>

## Primary source

[MDN — `fetch()` API](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API/Using_Fetch)
The fetch spec lives with browsers, and MDN is their canonical manual:
request/response anatomy, every option you will pass, and the explicit
statement that HTTP errors do not reject — the semantics your BFF's error
handling stands on.

## Before next session

Kill the Go server and run your `login.mjs` — read the `fetch failed` error
closely, including its `cause` (print `err.cause` in your catch). Saying
"that's connection-refused, not an HTTP error" when a teammate reports 'the
API is broken' is a superpower already. Then come to
[Lesson 06](06-typescript-i-types-and-inference.md), where we stop guessing
what shape `res.json()` returns.
