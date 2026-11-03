# Lesson 27 — Auth: tokens and cookies through the BFF

*Phase 4 · Session 27 of 38 · ~2 hours · Prereqs: Lessons 14, 20, 25, 26*

## Why this matters

Right now your TeamPulse login (from P1, Lesson 16) hands the Go token to browser
JavaScript — the exact thing your team's security reviewers will reject in a PR. The BFF
pattern fixes it structurally: the route handler performs the login, stores the token
where JS cannot read it, and forwards it to Go on every request. This is MISSION item 2's
"secrets live here, never in the browser" (Lesson 20), applied to the credential that
matters most.

## What you'll be able to do

- Draw the login round-trip: browser → route handler → Go → httpOnly cookie → back
- Write a `POST /api/login` route handler that exchanges credentials for a session
  cookie without ever sending the token to the browser
- Forward `Authorization: Bearer <token>` from the cookie on every protected route
- Explain what `httpOnly`, `sameSite`, `secure`, `path`, and `maxAge` each prevent
- Implement logout as cookie deletion, and reason about what it does *not* do upstream

## Principles

### The token never crosses to the browser

The flow, once, in full:

```
1. browser  POST /api/login  {"username":"wen","password":"demo"}   → Next.js
2. Next.js  POST http://localhost:8080/api/login  (same body)      → Go
3. Go       {"token":"tok-...","user":{...}}              → Next.js
4. Next.js  validates with LoginResponseSchema (Lesson 25)
5. Next.js  Set-Cookie: tp_token=tok-...; HttpOnly; Path=/; SameSite=Lax  → browser
6. browser  receives {"id":"u1","name":"Wen","role":"admin"}   — no token field
```

The browser now holds a *session cookie*, not a token. On every later request the
browser attaches the cookie automatically; the route handler reads it server-side and
sets Go's `Authorization` header. The token's entire life happens between your Next.js
server process and Go.

Why this beats `localStorage` (the P1 way): any XSS bug in any dependency can run
`localStorage.getItem("token")` and exfiltrate it. `HttpOnly` means the browser refuses
to show the cookie to JavaScript at all — `document.cookie` won't list it. The token is
still on the machine, but the blast radius of an XSS shrinks dramatically.

### Every cookie flag is an answer to an attack

```ts
response.cookies.set("tp_token", token, {
  httpOnly: true,                        // JS on the page can't read it (XSS)
  path: "/",                             // send to all /api/* routes, not just /api/login
  sameSite: "lax",                       // not sent on cross-site POSTs (CSRF)
  secure: process.env.NODE_ENV === "production", // HTTPS-only once deployed (Lesson 35)
  maxAge: 60 * 60 * 8,                   // expires in 8h; mock tokens live forever, so *you* set the TTL
});
```

`sameSite: "lax"` is the CSRF defence: a malicious site's form POSTing to
`localhost:3000/api/tasks` arrives *without* your cookie, so the route handler sees no
token and 401s. `secure` must be `true` in production (cookie only travels over HTTPS)
but `false` on `http://localhost` in dev — browsers reject `Secure` cookies over plain
HTTP, which is the classic "works locally, dies in prod" trap (Lesson 35 has more).

### The cookie is the token's transport, not your data model

You are not "storing session state in a cookie" in any clever sense — you are
parking an opaque upstream credential in the one browser slot JS can't read. Your app
code should treat `tp_token` as a black box: set it once, read it in exactly one helper,
delete it on logout. If you find yourself parsing it or putting it in component props,
stop — it must never reach a component.

### One helper forwards the header

Every protected route needs the same two lines, so write them once:

```ts
// lib/upstream.ts (added to Lesson 25's file)
import { NextRequest } from "next/server";

export function authHeaders(req: NextRequest): HeadersInit {
  const token = req.cookies.get("tp_token")?.value;
  if (!token) throw new UpstreamError(401, "not logged in");
  return { Authorization: `Bearer ${token}` };
}
```

Route handlers then read uniformly: `callGo("/api/projects", z.array(ProjectSchema), {
headers: authHeaders(req) })`. Note Go's exact demand from `main.go`'s `auth` function:
the prefix is `Bearer ` — capital B, one trailing space. A lowercase `bearer` actually
works against this mock because Go's `strings.HasPrefix` check is what it is, but your
team's real backend may not be as forgiving; send the canonical form.

### Logout deletes your copy

`POST /api/logout` → `response.cookies.delete("tp_token")`. That's honest scope: the
mock Go backend has no logout endpoint and its `tokens` map keeps entries forever, so
you log the user out *of your app*, not out of Go. In a real backend you'd also call its
revoke endpoint if it has one. Say the limitation out loud in the PR description —
silent half-implementations are how "logged out but still accessible" bugs ship.

## Follow along

`~/teampulse`, Go mock running, Lessons 25–26 in place. Start from a clean slate:
DevTools → Application → Cookies → `http://localhost:3000` → delete `tp_token` if
Lesson 25's experiments left one.

1. Replace the login page from P1 with one that posts to the **BFF**, never to Go.
   `app/login/page.tsx` (a Client Component — forms need interactivity, Lesson 14):

   ```tsx
   "use client";

   import { useRouter } from "next/navigation";
   import { useState } from "react";

   export default function LoginPage() {
     const router = useRouter();
     const [username, setUsername] = useState("wen");
     const [error, setError] = useState<string | null>(null);
     const [pending, setPending] = useState(false);

     async function onSubmit(e: React.FormEvent) {
       e.preventDefault();
       setPending(true);
       setError(null);
       const res = await fetch("/api/login", {
         method: "POST",
         headers: { "Content-Type": "application/json" },
         body: JSON.stringify({ username, password: "demo" }),
       });
       setPending(false);
       if (!res.ok) {
         const body = (await res.json().catch(() => null)) as { error?: string } | null;
         setError(body?.error ?? "login failed");
         return;
       }
       router.push("/dashboard");
     }

     return (
       <form onSubmit={onSubmit}>
         <input
           value={username}
           onChange={(e) => setUsername(e.target.value)}
           aria-label="username"
         />
         <p>Password hint: the literal word demo.</p>
         <button disabled={pending}>{pending ? "Signing in…" : "Sign in"}</button>
         {error && <p role="alert">{error}</p>}
       </form>
     );
   }
   ```
   Expected: the page renders; submitting hits `POST /api/login` on **localhost:3000**
   (check Network tab), which 404s until step 2 exists.

2. Create `app/api/login/route.ts` — the heart of this lesson:

   ```ts
   import { NextRequest, NextResponse } from "next/server";
   import { z } from "zod";
   import { callGo, UpstreamError } from "@/lib/upstream";
   import { LoginResponseSchema } from "@/lib/schemas";

   const LoginBodySchema = z.object({
     username: z.string().min(1),
     password: z.string().min(1),
   });

   const GO_API = process.env.GO_API_BASE_URL ?? "http://localhost:8080";

   export async function POST(req: NextRequest) {
     const parsed = LoginBodySchema.safeParse(await req.json().catch(() => null));
     if (!parsed.success) {
       return NextResponse.json({ error: "username and password required" }, { status: 400 });
     }

     try {
       const data = await callGo(
         "/api/login",
         LoginResponseSchema,
         {
           method: "POST",
           headers: { "Content-Type": "application/json" },
           body: JSON.stringify(parsed.data),
         },
       );

       // success: token goes to the cookie jar, ONLY the user goes to the browser
       const response = NextResponse.json(data.user);
       response.cookies.set("tp_token", data.token, {
         httpOnly: true,
         path: "/",
         sameSite: "lax",
         secure: process.env.NODE_ENV === "production",
         maxAge: 60 * 60 * 8,
       });
       return response;
     } catch (err) {
       if (err instanceof UpstreamError && err.status === 401) {
         return NextResponse.json({ error: "invalid credentials" }, { status: 401 });
       }
       throw err;
     }
   }
   ```
   Expected: sign in on `/login` → redirected to `/dashboard`, and DevTools → Cookies
   shows `tp_token` with the HttpOnly column ticked.

3. Prove the browser can't see it. In DevTools Console:
   ```js
   document.cookie
   ```
   Expected: `""` — or at least no `tp_token`. The cookie exists (Application tab
   confirms) but JS is blind to it. Then try the fake token path: sign in with password
   `wrong` → your form shows "invalid credentials" (Go's 401, translated by the route
   handler).

4. Upgrade every protected route to the uniform pattern. `app/api/projects/route.ts`:

   ```ts
   import { NextRequest, NextResponse } from "next/server";
   import { z } from "zod";
   import { authHeaders, callGo, UpstreamError } from "@/lib/upstream";
   import { ProjectSchema } from "@/lib/schemas";

   export async function GET(req: NextRequest) {
     try {
       const projects = await callGo("/api/projects", z.array(ProjectSchema), {
         headers: authHeaders(req),
       });
       return NextResponse.json(projects);
     } catch (err) {
       if (err instanceof UpstreamError) {
         return NextResponse.json({ error: err.message }, { status: err.status });
       }
       throw err;
     }
   }
   ```
   Do the same for `app/api/tasks/route.ts` (GET and POST) and add
   `app/api/me/route.ts` proxying Go's `GET /api/me` with `UserSchema`.
   Expected: with the cookie set, `curl -s localhost:3000/api/me` returns
   `{"id":"u1","name":"Wen","role":"admin"}`; without it (new incognito window or
   `curl` with no cookie), 401 `{"error":"not logged in"}` — note the 401 came from
   `authHeaders`, Go was never even called.

5. Add `app/api/logout/route.ts` and wire a Sign out button on the dashboard:

   ```ts
   import { NextResponse } from "next/server";

   export async function POST() {
     const response = NextResponse.json({ ok: true });
     response.cookies.delete("tp_token");
     return response;
   }
   ```
   Expected: after clicking Sign out, `curl -s localhost:3000/api/me` (in the same
   browser session) returns 401, and the Network tab's response headers for the logout
   call show `set-cookie: tp_token=; ... Max-Age=0` — deletion is just a
   set-with-expiry-zero.

6. End-to-end trace, the Lesson 23 drill with auth added: sign in, open Network tab,
   click through the dashboard. Find the one request whose *response* sets the cookie,
   and three whose *requests* carry it. Confirm none of them contain the string `tok-`
   anywhere in the browser-visible payload — except in the Application → Cookies view,
   which is the browser telling you what it will attach, not your JS.

## Your turn

1. Your route handlers currently 401 with `{"error":"not logged in"}` when the cookie is
   missing. TanStack Query surfaces that in `error.message`. Extend `getJson` (Lesson 26)
   so a 401 from any BFF route redirects to `/login`.

   <details><summary>Solution</summary>

   ```ts
   async function getJson<T>(url: string, init?: RequestInit): Promise<T> {
     const res = await fetch(url, init);
     if (res.status === 401) {
       window.location.href = "/login";
       throw new Error("session expired"); // queryFn must still throw
     }
     if (!res.ok) {
       const body = (await res.json().catch(() => null)) as { error?: string } | null;
       throw new Error(body?.error ?? `request failed: ${res.status}`);
     }
     return res.json() as Promise<T>;
   }
   ```
   Redirect is a side effect, so the throw still happens — TanStack never renders data
   that was never coming.
   </details>

2. The mock's tokens never expire, so a stolen cookie copy works forever. Add a
   server-side expiry check without calling Go: what do you change in
   `route.ts`'s `cookies.set` call and in `authHeaders`, and what does each side catch?

   <details><summary>Solution</summary>
   `maxAge: 60 * 60 * 8` already bounds *browser* storage — the browser drops the cookie
   after 8h. But a copied cookie value replayed by another client ignores maxAge, because
   the server never checks age. Fix: also set a second cookie `tp_login_at` (not
   httpOnly-relevant; it's just a timestamp) and in `authHeaders` compare
   `Date.now() - Number(req.cookies.get("tp_login_at")?.value)` against 8h, throwing
   `UpstreamError(401, "session expired")` when older. The browser enforces politely;
   the server enforces honestly. Real backends issue tokens with embedded expiry — this
   is the poor-man's version of why Lesson 25's contract should include expiry if it
   exists.
   </details>

3. A teammate suggests: "Skip the cookie — put the token in a normal React state
   variable and send it in a header from the client." What is the concrete failure mode
   of that design versus the httpOnly cookie?

   <details><summary>Solution</summary>
   Token-in-JS-state means any XSS (a malicious npm package, an injected script tag) can
   read the state or simply intercept `fetch` and exfiltrate the token; the token is
   deliberately visible to page code. The httpOnly cookie is never readable by page code,
   so the same XSS cannot steal it — it can only make requests *as* you while you sit
   there, which SameSite=Lax further limits. Same-origin headers from the client are
   also per-tab: refresh loses the token, multiple tabs need resync. The cookie rides
   with every request for free.
   </details>

## Checkpoint

1. Sketch the six messages of the login round-trip (browser, BFF, Go) and say which two
   never contain the raw token.

   <details><summary>Answer</summary>
   Browser→BFF (credentials), BFF→Go (same body), Go→BFF (token), BFF→browser
   (Set-Cookie carries token), BFF→browser body (user only), later requests
   (cookie→BFF→Authorization header). The two without the raw token in the *body JS can
   see*: the login response body (user JSON only) and — after `HttpOnly` — even the
   cookie isn't token-readable via JS. Browser code never handles the token.
   </details>

2. What does `httpOnly` block exactly, and what class of attack is it the defence for?

   <details><summary>Answer</summary>
   It makes the browser omit the cookie from anything JavaScript can reach
   (`document.cookie`, browser extensions' page-context scripts). Defence for XSS token
   theft.
   </details>

3. Why is `secure: true` set conditionally on `NODE_ENV`, and what breaks if you set it
   unconditionally during development?

   <details><summary>Answer</summary>
   `Secure` means "only send over HTTPS". Dev is `http://localhost`, so setting it
   unconditionally makes the browser refuse to store the cookie — login appears to
   succeed then every request 401s, a miserable debugging session. Production (Lesson
   35, HTTPS) needs it on.
   </details>

4. (Go API question) You forgot the cookie and hit your BFF's `/api/projects`. Two
   401s are possible in this architecture — where can each originate, and how do you
   tell them apart from the response alone?

   <details><summary>Answer</summary>
   The BFF's `authHeaders` 401s with `{"error":"not logged in"}` before Go is called
   (cookie missing). If the cookie exists but Go's `requireAuth` rejects the token
   (unknown token in its map), the BFF forwards Go's `{"error":"missing or invalid
   token"}`. The message tells you which layer answered — Lesson 28 will formalize this
   into error `code`s.
   </details>

5. Logout on the mock only deletes a cookie. Why is that both sufficient *and* a
   limitation worth writing in the PR description?

   <details><summary>Answer</summary>
   Sufficient because the app's only access path is the cookie — without it, every route
   401s at `authHeaders`. A limitation because Go's `tokens` map still holds the token
   (no revoke endpoint), so a previously exfiltrated copy would keep working against Go
   directly. Honest PRs name what logout does and doesn't cover.
   </details>

## Primary source

[nextjs.org — cookies() (Route Handler API)](https://nextjs.org/docs/app/api-reference/functions/cookies) —
the official reference for reading and setting cookies in route handlers, including the
exact option names (`httpOnly`, `sameSite`, `secure`, `maxAge`) and the
`NextResponse.cookies` methods you set and delete with.

## Before next session

With the Go mock running, run each of these ten times and count how many fail:
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"wen","password":"demo"}' | sed 's/.*"token":"\([^"]*\)".*/\1/')
curl -s http://localhost:8080/api/flaky -H "Authorization: Bearer $TOKEN"; echo
```
Write the failure count next to the terminal tab — Lesson 28 starts from your number.
