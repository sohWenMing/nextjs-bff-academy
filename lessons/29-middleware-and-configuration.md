# Lesson 29 — Middleware and configuration

*Phase 4 · Session 29 of 38 · ~2 hours · Prereqs: Lessons 20, 27*

## Why this matters

Three questions recur in every BFF PR review: *where does the Go URL come from* (never
hard-coded — dev, staging, and prod point at different backends), *who checks the session
before the page even renders* (middleware), and *why does this work locally but not in
prod* (env vars and CORS — two of the three classic answers live in this lesson). This is
the plumbing your team expects without being asked.

## What you'll be able to do

- Write a `middleware.ts` that redirects unauthenticated users away from private routes
  before any route handler or component runs
- Name the execution order: middleware → route handler → component, and what each layer
  is *for*
- Keep secrets out of the browser using the `NEXT_PUBLIC_` naming rule, and explain how
  you know which side of the wall a variable is on
- Set up per-environment env files and state what loads when
- Explain CORS at working level: why your BFF makes the problem vanish, and read the
  three headers involved

## Principles

### middleware.ts: the bouncer at the door

`middleware.ts` at the project root runs **before** a request reaches any page, layout,
or route handler, inside Next.js's edge runtime (a lightweight V8 isolate, not Node —
no npm packages with native deps in here). It sees the request, can redirect, rewrite,
or set headers, and returns. The classic BFF use: session gating.

```ts
// middleware.ts
import { NextRequest, NextResponse } from "next/server";

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
  const isPrivate = pathname.startsWith("/dashboard");
  const isLogin = pathname === "/login";
  const hasSession = req.cookies.has("tp_token");

  if (isPrivate && !hasSession) {
    return NextResponse.redirect(new URL("/login", req.url));
  }
  if (isLogin && hasSession) {
    return NextResponse.redirect(new URL("/dashboard", req.url));
  }
}

export const config = {
  matcher: ["/dashboard/:path*", "/login"],
};
```

`matcher` (an array of path patterns) is load-bearing: middleware runs on **every**
matched request — without it you'd run this code on images, fonts, and every asset. Keep
the list as small as your auth rule requires.

Middleware is a *convenience gate*, not the security boundary: it reads cookie presence,
not validity — a garbage `tp_token` passes middleware and then dies at `authHeaders`
(Lesson 27) with a 401 when Go rejects it. Defense in depth: middleware bounces honest
users politely, route handlers enforce honestly. Never let middleware be the only lock.

### Env vars: two sides of one wall

Next.js splits environment variables by *name prefix*, and the rule is the whole story:

| Name | Available in | Use for |
|---|---|---|
| `GO_API_BASE_URL` | server only (route handlers, server components, middleware) | upstream URLs, API keys, anything embarrassing in a stack trace |
| `NEXT_PUBLIC_API_PATH` | server **and** browser | things genuinely public: analytics ID, feature flags |

A variable without the `NEXT_PUBLIC_` prefix that you try to read in client code is
`undefined` — not an error, just `undefined`, which is why the mistake survives until
runtime. The inverse mistake is lethal: prefixing a secret makes it a **string literal
baked into the JS bundle** — open DevTools → Sources, search, find it. Your `tp_token`
session and Go's URL must never carry that prefix.

Files load per environment, most-specific wins:

```
.env                # all environments (defaults, committed)
.env.local          # all environments, git-ignored (your dev values)
.env.production     # only `next start` after `next build` (set real values in your host)
```

Dev uses `pnpm dev` → `.env.local`; the deploy host (Lesson 35, Vercel) sets production
values in its dashboard — `.env.production` files mostly exist to document what must be
set. Changing an env file requires a dev-server restart: env is read at boot.

### next.config.ts, at working level

You need almost nothing, and the little you need is typed:

```ts
// next.config.ts
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // headers(): set response headers route-handler middleware can't reach cleanly,
  // e.g. hardening headers on every page:
  async headers() {
    return [
      {
        source: "/:path*",
        headers: [{ key: "X-Content-Type-Options", value: "nosniff" }],
      },
    ];
  },
};

export default nextConfig;
```

Don't collect config options from blog posts. Every line in this file should trace to a
concrete need you can state in one sentence.

### CORS at working level

CORS (Cross-Origin Resource Sharing) is the browser rule that says: a page at
`http://localhost:3000` may not read responses from `http://localhost:8080` unless
8080 sends `Access-Control-Allow-Origin: http://localhost:3000` (plus
`Access-Control-Allow-Methods`/`-Headers` for non-simple requests). For "pre-flight"ed
requests (custom headers like `Authorization`, JSON POSTs), the browser first fires an
`OPTIONS` probe and aborts if the reply lacks the right headers. The server opts origins
in; everything else is denied.

Now the punchline: **your BFF has no CORS problem**, because browser →
`localhost:3000` is same-origin, and route handler → Go is server-to-server — `fetch`
from Node ignores CORS entirely (it's a browser-only policy). You sidestepped the whole
subject by architecture. What you *must* be able to do at working level:

- recognise the browser console error (`No 'Access-Control-Allow-Origin' header...`)
  and know it means *someone is calling Go directly from the browser* — usually legacy
  P1 code; fix by moving the call behind a route handler
- read the three response headers with curl to see what Go actually sends
- explain to a teammate why "just disable CORS checks in the browser" means "delete a
  security feature while debugging"

## Follow along

`~/teampulse`, Go mock running, Lesson 27's cookie flow working.

1. Create `.env.local` in `~/teampulse` (git-ignored by the default Next template):
   ```
   GO_API_BASE_URL=http://localhost:8080
   NEXT_PUBLIC_APP_NAME=TeamPulse
   ```
   Restart `pnpm dev` (env is read at boot). Expected: everything still works —
   `callGo`'s `process.env.GO_API_BASE_URL` now reads your file instead of the fallback.

2. Prove the wall. Add to any Client Component: `console.log(process.env.GO_API_BASE_URL)`.
   Expected in the browser console: `undefined`, while `console.log(process.env.NEXT_PUBLIC_APP_NAME)`
   prints `TeamPulse`. Now search the sources: in DevTools → Cmd+F (global search),
   search `TeamPulse` — found in a bundle chunk; search `8080` — absent. The prefix
   *is* the wall between server world and browser world. Remove the log lines.

3. Prove the fallback was load-bearing: temporarily delete `.env.local`'s
   `GO_API_BASE_URL` line *and* the `?? "http://localhost:8080"` fallback in
   `lib/upstream.ts`, restart, hit the dashboard. Expected: `{"code":"network",
   "message":"Cannot reach the server.","retryable":true}` — Lesson 28's normalised
   error catches you, because `fetch("undefined/api/projects")` rejects as a network
   failure. Restore both.

4. Add `middleware.ts` from the Principles section (root of `~/teampulse`). Test the
   bouncer: log out (`POST /api/logout` or clear the cookie), then navigate the browser
   to `http://localhost:3000/dashboard`. Expected: instant redirect to `/login` — check
   the Network tab and you'll see the dashboard request never rendered a page. Log in;
   visiting `/login` now bounces you to `/dashboard`.

5. Prove middleware is presence-only. With DevTools you can't edit an httpOnly cookie —
   so from the terminal:
   ```bash
   curl -si localhost:3000/dashboard --cookie "tp_token=fake-token"
   ```
   Expected: **no redirect** (middleware saw *a* cookie) but the page's data fetches
   401 — `authHeaders` forwarded `fake-token`, Go's `requireAuth` rejected it, and your
   `UiError` renders `unauthorized`. That pair of behaviours is the layering, demonstrated.

6. CORS fieldwork. Ask Go directly what it says about cross-origin access:
   ```bash
   curl -si -X OPTIONS http://localhost:8080/api/projects \
     -H "Origin: http://localhost:3000" \
     -H "Access-Control-Request-Method: GET"
   ```
   Expected: a response with **no** `Access-Control-Allow-Origin` header at all (the
   mock never sets CORS headers — check `main.go`: no mention of `Access-Control`).
   Translation: any direct browser→Go call would die at the pre-flight. Since your
   architecture forbids those, this is evidence for the architecture, not a to-do list.

7. Add the `headers()` block from `next.config.ts` above, restart, and verify:
   ```bash
   curl -sI localhost:3000/login | grep -i x-content-type
   ```
   Expected: `x-content-type-options: nosniff`.

## Your turn

1. Extend the matcher and rule: `/dashboard` *and* `/tasks` are private, but
   `/healthz`-style public pages must not pay the middleware cost. Write the `config`
   object and the private-path test that keeps both properties.

   <details><summary>Solution</summary>

   ```ts
   export const config = {
     matcher: ["/dashboard/:path*", "/tasks/:path*", "/login"],
   };
   ```
   and inside: `const isPrivate = pathname.startsWith("/dashboard") || pathname.startsWith("/tasks");`
   Properties: every private subtree runs the check (defense in depth behind route
   handlers); unmatched paths (assets, `/`, `/api/*`) skip middleware entirely. Adding a
   new private page means touching *both* the matcher and the predicate — teams often
   extract a `const PRIVATE_PREFIXES = [...]` both read.
   </details>

2. Your deploy host sets `GO_API_BASE_URL` in its dashboard. A teammate adds a
   `.env.production` file to the repo containing the *staging* Go URL. What happens at
   deploy time and why?

   <details><summary>Solution</summary>
   It depends on load order — but either way it's a bug. Next loads `.env.production`
   only for values not already in the real process environment (actual env vars win), so
   on most hosts the dashboard value wins and the file is dead weight that *lies* to
   every reader. On a host that doesn't set the var, you silently ship pointing at
   staging. Commit `.env.example` with documented *placeholders*; keep real values in
   the host.
   </details>

3. A PR lands with browser code: `fetch("http://localhost:8080/api/projects", { headers:
   { Authorization: \`Bearer ${token}\` } })` where `token` was read from a global. File
   two review comments, each citing a lesson.

   <details><summary>Solution</summary>
   1. **CORS (this lesson):** no `Access-Control-Allow-Origin` comes back from Go —
      the pre-flight fails and the request never completes; move the call behind a
      route handler as Lesson 22/25 established.
   2. **Token exposure (Lesson 27):** a token in browser code defeats the httpOnly
      cookie design — XSS can exfiltrate it. Route handlers read the cookie and attach
      the `Authorization` header server-side; the browser must never hold it.
   </details>

## Checkpoint

1. In one sentence each: what runs in middleware, and what can it *not* do that a route
   handler can?

   <details><summary>Answer</summary>
   Middleware runs on matched requests before routing resolves — redirect/rewrite/
   set-header, cheaply, on every hit. It can't do route-handler work: it runs in the
   edge runtime (no Node APIs, no zod-and-database business logic conveniently, no
   response bodies that depend on validated upstream data) — and it must not be the
   authority on session *validity*.
   </details>

2. Why does `process.env.SOMETHING` read as `undefined` in a Client Component while
   `process.env.NEXT_PUBLIC_SOMETHING` works, mechanically?

   <details><summary>Answer</summary>
   At build time Next inlines only `NEXT_PUBLIC_`-prefixed variables into the client
   bundle as string literals; everything else is replaced with `undefined` so secrets
   can't leak even accidentally. Server code reads the real `process.env` at runtime.
   </details>

3. You set `GO_API_BASE_URL` in `.env.local`, fix a typo an hour later, and nothing
   changes. Why, and what's the habit?

   <details><summary>Answer</summary>
   Env files are read once when the dev server boots. Restart `pnpm dev` after any env
   change — before debugging anything else.
   </details>

4. (Go API question) `curl -X OPTIONS` to Go's `/api/projects` with an Origin header
   returns no `Access-Control-Allow-*` headers. Your teammate says TeamPulse therefore
   needs a CORS library in `main.go`. What's your answer, and what would justify a
   *yes*?

   <details><summary>Answer</summary>
   No — the BFF makes it same-origin for the browser, and server-to-server `fetch`
   ignores CORS. A yes requires a genuine browser-to-Go consumer: a mobile WebView or
   third-party page calling Go directly, in which case Go opts *specific origins* in
   (never `*` with credentials).
   </details>

5. Your middleware lets a request through to `/dashboard`, and the route handler still
   401s. Is the middleware broken? Explain the design.

   <details><summary>Answer</summary>
   No — middleware checks cookie *presence* only (cheap, edge, before rendering).
   Validity is checked when `authHeaders` forwards to Go's `requireAuth`, which knows
   which tokens it minted. Presence-gating gets real users to the right page without a
   round-trip; the enforcement that matters lives behind it.
   </details>

## Primary source

[nextjs.org — middleware](https://nextjs.org/docs/app/api-reference/file-conventions/middleware) —
the official page for `middleware.ts`: execution order, the `config.matcher` syntax,
`NextResponse.redirect`/`rewrite`, and the edge-runtime constraints that decide what
code belongs there.

## Before next session

P3 (Lesson 30) is the full consolidation. Tonight, just get the workshop clean: delete
`node_modules` and the `.next` folder in `~/teampulse`, run `pnpm install && pnpm build`,
and write down the build output verbatim (route table and all) — Lesson 30 opens with
reading that output together.
