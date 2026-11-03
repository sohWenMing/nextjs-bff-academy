# Next.js (App Router) cheatsheet

Next.js 15+, App Router only. TeamPulse BFF calls the Go mock at `http://localhost:8080`.

## 1. File conventions

| File | Where | What it is |
|---|---|---|
| `app/layout.tsx` | any segment | Wraps children; root layout must render `<html>`/`<body>`; one per route tree |
| `app/page.tsx` | segment | The UI for that URL. `app/projects/page.tsx` → `/projects` |
| `app/loading.tsx` | segment | Auto Suspense fallback while the segment's data loads |
| `app/error.tsx` | segment | Client component error boundary (`"use client"`, receives `error`, `reset()`) |
| `app/not-found.tsx` | segment | UI rendered when `notFound()` is called in this segment |
| `app/projects/[id]/page.tsx` | segment | Dynamic segment: `/projects/p1` → `params.id === "p1"` (async in Next 15: `const { id } = await params`) |
| `app/api/projects/route.ts` | `app/api/*` | Route handler: export `GET`/`POST`/… functions. Server-only. |
| `middleware.ts` | project root | Runs before requests complete; redirect/rewrite/set headers (edge runtime) |
| `next.config.ts` | project root | Build/server config |

## 2. Server vs Client Components — decision rules

**Default: Server Component.** It runs only on the server, can `await` and hold secrets, and ships ~zero JS.

Switch to `"use client"` (top line of the file) **only** when you need:

- state (`useState`) or effects (`useEffect`)
- browser APIs (`window`, `localStorage`, `navigator`)
- event handlers (`onClick`, `onChange`) or refs
- a hook that requires client context

Rules of the boundary:

1. `"use client"` marks a **boundary** — everything it imports becomes client code too. Push it to the leaves: server page → small client island, not `"use client"` at the top of every page.
2. Server Components **cannot** import client hooks; client components can render server components passed as `children`/props.
3. Props crossing the boundary must be serializable (no functions to a server component, no Date/Map/class instances from server to client — pass ISO strings and arrays).
4. Secrets (`GO_API_KEY`) can appear **only** in server files. Importing a secret module from a client component is a build error — good, that's the guardrail.

```tsx
// app/projects/page.tsx — server component: fetches directly
const res = await fetch(`${process.env.GO_API_URL}/api/projects`, {
  headers: { Authorization: `Bearer ${token}` },
  next: { revalidate: 30 },          // re-cached at most every 30s
});
const projects = ProjectListSchema.parse(await res.json());
return <TaskFilter projects={projects} />;   // client island below
```

## 3. Route handler skeleton (BFF → Go with Bearer token)

```ts
// app/api/projects/[id]/route.ts
import { NextResponse } from "next/server";
import { z } from "zod";

const ProjectSchema = z.object({
  id: z.string(), name: z.string(), description: z.string(),
  status: z.enum(["active", "archived"]), createdAt: z.string(),
});

export async function GET(
  req: Request,
  ctx: { params: Promise<{ id: string }> },      // params is async in Next 15
) {
  const { id } = await ctx.params;
  const token = req.cookies.get("tp_token")?.value;   // read from httpOnly cookie, not JS
  if (!token) return NextResponse.json({ error: "unauthorized" }, { status: 401 });

  const upstream = await fetch(`${process.env.GO_API_URL}/api/projects/${id}`, {
    headers: { Authorization: `Bearer ${token}` },
    signal: AbortSignal.timeout(5000),                // never hang the UI
    cache: "no-store",                                // user-specific data: don't cache
  });

  if (upstream.status === 404) return NextResponse.json({ error: "project not found" }, { status: 404 });
  if (!upstream.ok) return NextResponse.json({ error: "upstream failure" }, { status: 502 });

  const project = ProjectSchema.parse(await upstream.json());  // trust nothing from Go
  return NextResponse.json(project);
}
```

Route handlers are plain `Request → Response` functions. Exported name = HTTP method. Put secrets, validation, and reshaping here — never in the browser.

## 4. Env vars

| Kind | Where defined | When read | Visible to browser? | Example |
|---|---|---|---|---|
| Server var (`GO_API_URL`, `GO_API_KEY`) | `.env.local` / deploy env | at runtime, `process.env.X` in server files | **No** | BFF secret |
| `NEXT_PUBLIC_*` | same | **inlined at build time** | **Yes — treat as public** | analytics id |

- `NEXT_PUBLIC_` values are frozen into the JS bundle at build → changing them needs a **rebuild**, not just a redeploy of env. Classic "works in dev, broken in prod" cause.
- Never name a secret `NEXT_PUBLIC_GO_API_KEY`. If the browser can see it, it's not a secret.
- `.env.local` is gitignored and loaded by `next dev`/`next build`; plain `.env` is committed only for non-secrets.

## 5. fetch with AbortSignal.timeout

```ts
try {
  const res = await fetch(`${process.env.GO_API_URL}/api/slow`, {
    signal: AbortSignal.timeout(2000),   // the mock's /api/slow takes 2s — good test
  });
} catch (err) {
  if (err instanceof Error && err.name === "TimeoutError") {
    return NextResponse.json({ error: "upstream timeout" }, { status: 504 });
  }
  throw err;                             // AbortError for manual AbortController.abort()
}
```

Every upstream call from the BFF gets a timeout. `fetch` only rejects on network/abort — still check `res.ok` for 4xx/5xx.

## 6. Typical BFF repo structure

```
app/
  layout.tsx            # root layout (server)
  page.tsx              # dashboard (server component)
  login/page.tsx        # login form → POST /api/login
  projects/
    page.tsx            # server: awaits BFF or Go
    [id]/page.tsx       # project detail, notFound() on 404
    task-filter.tsx     # "use client" island (useState filter)
  api/
    login/route.ts      # POST: Go login → set httpOnly cookie, strip token from body
    projects/route.ts   # GET: forward Authorization, zod-parse, reshape
    tasks/route.ts      # GET ?projectId= / POST
components/             # shared UI (mostly presentational, props-only)
lib/
  go-api.ts             # typed fetch helpers + zod schemas for Go responses
  auth.ts               # cookie read/write helpers
types/                  # shared TS types (import type everywhere)
middleware.ts           # no tp_token cookie + non-public path → redirect /login
next.config.ts
.env.local              # GO_API_URL, GO_API_KEY (never NEXT_PUBLIC_ for the key)
```

## 7. Data-fetching decision (one line each)

- Public, cacheable page → `fetch` in a Server Component + `next: { revalidate: n }`.
- Per-user page → `fetch` in a Server Component with `cache: "no-store"` (or via your own route handler).
- Interactive client data (filters, mutations) → route handler + TanStack Query.
- Secret-handling upstream call → route handler, always.

---
*Related lessons: [17 file-system routing](../lessons/17-file-system-routing.md) · [18 server vs client components](../lessons/18-server-vs-client-components.md) · [19 fetching in server components](../lessons/19-fetching-in-server-components.md) · [20 route handlers](../lessons/20-route-handlers-your-first-bff.md) · [28 resilience](../lessons/28-resilience-timeouts-retries-error-shapes.md) · [29 middleware & config](../lessons/29-middleware-and-configuration.md)*
