# Lesson 20 — Route Handlers: your first BFF

*Phase 3 · Session 20 of 38 · ~2 hours · Prereqs: Lesson(s) 05, 17, 18, 19*

## Why this matters

This is the lesson the course is named for. A **Backend-for-Frontend** is a thin server layer, owned by the frontend team, that sits between the browser and the real backend — reshaping data and, crucially, holding the credentials. In Next.js it's a file: `app/api/.../route.ts`. Get this lesson and you own mission item #2 — "add a route handler that calls the Go backend" — and you understand *why* the password `demo` in Lesson 19's page file was safe while the same string in a client component would be a leak.

## What you'll be able to do

- You can write a `route.ts` that exports `GET` (and `POST`), reads request headers/body, and returns JSON with a chosen status code.
- You can call the Go API from a route handler with a Bearer token that exists only in the Node process.
- You can explain what `NEXT_PUBLIC_` does and why the Go base URL must *not* use it.
- You can read `GO_BASE_URL` from `.env.local` via `process.env` and explain when its value is baked in.
- You can be asked "is `/api/projects` served by Next.js or Go?" and answer from the file tree in seconds.

## Principles

### `route.ts` is a server function wearing a URL

A **route handler** is a file named `route.ts` inside an `app/api/**` folder that *exports functions named after HTTP methods*:

```ts
// app/api/health/route.ts
export async function GET() {
  return Response.json({ ok: true, where: "next.js" });
}
```

That's `http://localhost:3000/api/health`. You saw the pattern from the pre-session: no JSX, no component, no `"use client"` — folders are still URLs ([Lesson 17](17-file-system-routing.md)), but instead of rendering UI the file *answers HTTP*. Rules:

- Export one `async` function per method you support (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, `OPTIONS`); unexported methods get an automatic `405 Method Not Allowed`, other exports are ignored.
- A folder can have `page.tsx` **or** `route.ts` — never both (build error:
"you cannot have a page and a route handler at the same URL").
- Route handlers always run on the server. `"use client"` is meaningless
here; there is no render and no bundle.
- **Never fetch a route handler from another server component with
`fetch("http://localhost:3000/...")`** — that's Next.js calling itself over the network. Call the shared function directly instead. Route handlers are for *external* callers: the browser, mobile apps, curl, webhooks.

### The two types you actually use

```ts
import { NextRequest, NextResponse } from "next/server";

export async function POST(req: NextRequest) {
  const body = (await req.json()) as { title: string };
  // req.headers.get("authorization"), req.nextUrl.searchParams.get("q")
  return NextResponse.json({ created: body.title }, { status: 201 });
}
```

`NextRequest` is the standard `Request` (the fetch API request object from [Lesson 05](05-async-promises-await-fetch.md)) plus Next conveniences; `NextResponse.json(data, { status })` is a shortcut for a JSON response. Plain `Response.json(...)` works identically; team codebases use one of the two interchangeably.

### The BFF move: Go lives behind `/api` on `:3000`

Here is the architecture this course keeps returning to:

```text
before (P1):  browser ──fetch http://localhost:8080──▶ Go          (secrets in the browser)
after  (P2):  browser/client/server ──/api/*──▶ route handlers ──▶ Go  (secrets in Node)
```

The route handler below does the two-hand dance from [Lesson 19](19-fetching-in-server-components.md) — log in, then list — and the credentials are *untouchable by the browser*:

```ts
// app/api/projects/route.ts
const GO = process.env.GO_BASE_URL;

async function goToken(): Promise<string> {
  const res = await fetch(`${GO}/api/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username: "bff", password: "demo" }),
    cache: "no-store",
  });
  if (!res.ok) throw new Error(`login failed: ${res.status}`);
  const data = (await res.json()) as { token: string };
  return data.token;
}

export async function GET() {
  try {
    const token = await goToken();
    const res = await fetch(`${GO}/api/projects`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (res.status === 404) {
      return NextResponse.json({ error: "not found" }, { status: 404 });
    }
    if (!res.ok) {
      return NextResponse.json({ error: "upstream error" }, { status: 502 });
    }
    const projects = await res.json();
    return NextResponse.json(projects);
  } catch (err) {
    // Go is down/unreachable — the browser gets a clean 502, never a stack trace
    return NextResponse.json(
      { error: "go backend unreachable" },
      { status: 502 },
    );
  }
}
```

Read the `catch` again — the BFF's job in miniature: **the browser sees your status code and your error shape, never Go's raw failure or your stack trace.** The `{"error": "..."}` shape matches Go's own `writeErr` (`backend-mock/main.go`), so the UI handles both uniformly.

Dynamic segments work the same as pages ([Lesson 17](17-file-system-routing.md)): `app/api/projects/[id]/route.ts`, with `params: Promise<{ id: string }>` awaited inside `GET(req, { params })`.

### Environment variables: `process.env`, and the `NEXT_PUBLIC_` landmine

The URL `http://localhost:8080` is configuration, not code — on the team's staging server it's `https://go.internal`. Next.js reads env vars from an `.env.local` file at the project root (git-ignored by create-next-app's `.gitignore`):

```bash
# ~/teampulse/.env.local
GO_BASE_URL=http://localhost:8080
```

Rules that make or break deployments:

- `process.env.GO_BASE_URL` in a **server** file (route handler, server
component) reads the real, live environment. This is where the Go URL and every secret live.
- Code that runs **in the browser** cannot read real env vars at all —
there is no environment there. Next.js *inlines* any variable whose name starts with `NEXT_PUBLIC_` into the JS bundle at **build time**, replacing the `process.env.NEXT_PUBLIC_X` text with the literal string.
- Which means: **`NEXT_PUBLIC_` = "make this public."** It's not a
convenience prefix; it's a publishing prefix. A `NEXT_PUBLIC_GO_PASSWORD=demo` would appear, greppable, in `.next/static/chunks/*.js` — exactly where an attacker greps. The Go base URL is deliberately *not* `NEXT_PUBLIC_`: the browser should never know `:8080` exists. That's the whole point of the BFF.
- Inlined-at-build-time also explains the classic "worked on staging, broke
in prod" bug: the prod bundle was built with staging's values baked in. Server-side `process.env`, by contrast, is read at request time.
- Changing `.env.local` requires **restarting `pnpm dev`** — the file is read
at process start.

### Why a BFF when the browser could call Go directly?

Because in the P1 world the browser needed to know Go's URL, hold the token (Lesson 18's discomfort), survive CORS, and parse Go's raw shapes. Behind the BFF, the browser knows only same-origin `/api/*` URLs, holds no credentials, and receives responses the frontend team designed. When Go renames a field, you edit one route handler — not forty components. That file *is* your job description.

## Follow along

`pnpm dev` in `~/teampulse`, Go mock up on `:8080`.

1. **Health check from Next.js itself.** Create `app/api/health/route.ts`:

   ```ts
   export async function GET() {
     return Response.json({ ok: true, where: "next.js", at: new Date().toISOString() });
   }
   ```

`curl -s http://localhost:3000/api/health`. Expected: your JSON. Now `curl -s http://localhost:3000/api/health -X DELETE`. Expected: `405` (no `DELETE` exported). Note the `at` timestamp changes every call — route handlers run per request; nothing is cached for you.

2. **Externalize the Go URL.** Create `~/teampulse/.env.local` containing
`GO_BASE_URL=http://localhost:8080`. Restart `pnpm dev` (Ctrl-C, rerun) — expected: the startup line may print your env; either way, proceed.

3. **Write the projects handler.** Create `app/api/projects/route.ts` with the
full code from Principles (type it — every line; the `goToken` helper is where the muscle memory forms). Then:

   ```bash
   curl -s http://localhost:3000/api/projects | head -c 200
   ```

Expected: `[{"id":"p1","name":"Billing Service",...` — three projects, from Go, through your BFF, **with no `Authorization` header and no token in the curl command.** The credentials are behind the counter.

4. **Prove the secret didn't cross.** Check the two places a leak would show:

   `curl -s http://localhost:3000/api/projects | grep -o 'tok-[^"]*'` (browser-visible body). Expected: no output. And DevTools → Network on any page: no request to `:8080` anywhere; the only API call is to `:3000`.

5. **Watch the fail path.** Stop the Go mock (Ctrl-C in its tab). Reload the
curl from step 3. Expected: `{"error":"go backend unreachable"}` with status `502` (verify: `curl -s -o /dev/null -w '%{http_code}\n' http://localhost:3000/api/projects`). Restart the Go mock. This is mission item #4 in embryo: browser Network tab shows *your* 502 shape; the Go-vs-BFF distinction is now visible in tooling.

6. **A POST handler with a request body.** Create `app/api/tasks/route.ts`:

   ```ts
   import { NextResponse } from "next/server";

   const GO = process.env.GO_BASE_URL;

   async function goToken(): Promise<string> {
     const res = await fetch(`${GO}/api/login`, {
       method: "POST",
       headers: { "Content-Type": "application/json" },
       body: JSON.stringify({ username: "bff", password: "demo" }),
       cache: "no-store",
     });
     const data = (await res.json()) as { token: string };
     return data.token;
   }

   export async function POST(req: Request) {
     const body = (await req.json()) as {
       projectId: string;
       title: string;
       priority: number;
     };
     if (!body.projectId || !body.title) {
       return NextResponse.json(
         { error: "projectId and title are required" },
         { status: 400 },
       );
     }
     const token = await goToken();
     const res = await fetch(`${GO}/api/tasks`, {
       method: "POST",
       headers: {
         Authorization: `Bearer ${token}`,
         "Content-Type": "application/json",
       },
       body: JSON.stringify(body),
       cache: "no-store",
     });
     const task = await res.json();
     return NextResponse.json(task, { status: res.status });
   }
   ```

   ```bash
   curl -s -X POST http://localhost:3000/api/tasks \
     -H 'Content-Type: application/json' \
     -d '{"projectId":"p2","title":"Via the BFF","priority":2}'
   ```

Expected: the created task JSON with a `t<number>` id and `"status":"todo"`. Also try omitting `title`: your handler returns the 400 *before* Go is even called — validation at the border.

## Your turn

1. Write `app/api/tasks/route.ts`'s sibling: add a `GET` to the same file that
forwards to Go's `/api/tasks`, passing through the `projectId` query param (`/api/tasks?projectId=p1` on `:3000` → Go). Verify with curl; expect the three p1 tasks.

   <details><summary>Solution</summary>

   ```ts
   export async function GET(req: Request) {
     const url = new URL(req.url);
     const projectId = url.searchParams.get("projectId");
     const token = await goToken();
     const qs = projectId ? `?projectId=${projectId}` : "";
     const res = await fetch(`${GO}/api/tasks${qs}`, {
       headers: { Authorization: `Bearer ${token}` },
       cache: "no-store",
     });
     const tasks = await res.json();
     return NextResponse.json(tasks, { status: res.status });
   }
   ```

`curl -s 'http://localhost:3000/api/tasks?projectId=p1'` → tasks `t1`, `t2`, `t5`.
   </details>

2. Now create `app/api/projects/[id]/route.ts` serving `/api/projects/p1`
(remember: `params` is a Promise, per Lesson 17). Return Go's 404 as a 404. Verify `/api/projects/p3` works and `/api/projects/p9` gives `{"error":"project not found"}` with status 404.

   <details><summary>Solution</summary>

   ```ts
   // app/api/projects/[id]/route.ts
   import { NextResponse } from "next/server";

   const GO = process.env.GO_BASE_URL;

   async function goToken(): Promise<string> {
     const res = await fetch(`${GO}/api/login`, {
       method: "POST",
       headers: { "Content-Type": "application/json" },
       body: JSON.stringify({ username: "bff", password: "demo" }),
       cache: "no-store",
     });
     const data = (await res.json()) as { token: string };
     return data.token;
   }

   export async function GET(
     _req: Request,
     ctx: { params: Promise<{ id: string }> },
   ) {
     const { id } = await ctx.params;
     const token = await goToken();
     const res = await fetch(`${GO}/api/projects/${id}`, {
       headers: { Authorization: `Bearer ${token}` },
       cache: "no-store",
     });
     const data = await res.json();
     return NextResponse.json(data, { status: res.status });
   }
   ```

`_req` is required positionally but unused — the underscore says so honestly. Note passing `res.status` through: Go's 404 becomes your 404.
   </details>

3. Harder (the trap question): a teammate, trying to reuse your work, writes
this in a `"use client"` component:

   ```tsx
   const GO = process.env.NEXT_PUBLIC_GO_BASE_URL; // set in .env.local
   const res = await fetch(`${GO}/api/flaky`, {
     headers: { Authorization: `Bearer ${token}` },
   });
   ```

Their local build works. Name three distinct problems with those five lines.

   <details><summary>Answer</summary>
(1) `NEXT_PUBLIC_GO_BASE_URL` is inlined into the shipped bundle at build time — the browser now knows Go's internal URL, which the BFF exists to hide; it will also be stale forever (baked at build). (2) The browser must hold a token again — Lesson 18's leak. (3) Browser → `:8080` is cross-origin and the mock Go sends no CORS headers, so it only appears to work in environments without enforcement — it will fail with a CORS error in the browser console. The fix is all three at once: call `/api/flaky` (same origin, no token, no env var) and add a route handler for it.
   </details>

## Checkpoint

1. What three things must be true for `http://localhost:3000/api/orders` to be
served by your code?

   <details><summary>Answer</summary>
The file `app/api/orders/route.ts` exists; it exports `async function GET` (for a GET request); no `page.tsx` occupies the same URL. Any other method returns 405 unless exported.
   </details>

2. Where is the value of `GO_BASE_URL` (no prefix) readable — and where is
`NEXT_PUBLIC_GO_BASE_URL` readable?

   <details><summary>Answer</summary>
`GO_BASE_URL`: server files only (route handlers, server components), read at request time. `NEXT_PUBLIC_GO_BASE_URL`: browser code — but the mechanism is build-time string inlining, so it's also "readable" by anyone who greps the JS bundle. Public prefix = public value.
   </details>

3. Your `/api/projects` handler must call Go, but Go is temporarily down.
What does the browser receive, and why is that preferable to the exception propagating?

   <details><summary>Answer</summary>
Your `catch` returns `NextResponse.json({ error: ... }, { status: 502 })` — a clean JSON 502. The alternative is a crashed render or a raw stack trace leaking internals; the UI can only branch on a status code it controls.
   </details>

4. TeamPulse: you changed `.env.local` to point at a mock running on `:8090`
but `curl localhost:3000/api/projects` still hits `:8080` (per Go's logs). Two possible reasons, one relating to env loading and one to bundle magic.

   <details><summary>Answer</summary>
(1) `pnpm dev` wasn't restarted — `.env.local` is read at process start. (2) If the value were read in a `NEXT_PUBLIC_` style (inlined at build), a stale `.next` build would keep the old literal — clear `.next` / rebuild. Server-side `process.env` without restart is the far more common cause.
   </details>

5. Why must the browser never learn the string `http://localhost:8080`
anywhere in this course's architecture?

   <details><summary>Answer</summary>
Because then it could (and would) call Go directly — bypassing your validation, error-shaping, and (later) auth layer, and requiring tokens in browser JS. The BFF's guarantees only hold if `:3000/api/*` is the only door the browser knows about.
   </details>

## Primary source

[Next.js Docs — Route Handlers](https://nextjs.org/docs/app/building-your-application/routing/route-handlers) — defines `route.ts`, the HTTP-method exports, `NextRequest`/`NextResponse`, the page/route conflict rule, and exactly which features (middleware, headers) apply to them.

## Before next session

Kill the Go mock and leave it dead. Reload `http://localhost:3000/pulse` (from Lesson 19) several times and note *what the user actually sees* when the server-side fetch throws — a white screen? an error? nothing? Write your observation in one sentence; Lesson 21 turns it into `error.tsx`.
