# Glossary — every term this course uses

Alphabetical, one paragraph each, grounded in TeamPulse. When you forget a word mid-ticket, look it up here first.

**AbortSignal** — an object you pass to `fetch` to cancel a request: `fetch(url, { signal: AbortSignal.timeout(5000) })` kills the call to `/api/slow` after 5s instead of hanging the UI. The route handler catches an `AbortError` and returns a clean 504. (Lesson 28)

**any** — the TypeScript "shut up" type: disables all checking on a value. Never reach for it; use `unknown` when you don't know a shape yet. A stray `any` on the Go response defeats the whole point of the zod boundary. (Lessons 07, 25)

**App Router** — Next.js's routing system where folders in `app/` are URLs and files like `page.tsx` and `route.ts` define what each URL is. The only router this course teaches. (Lessons 17–23)

**assignability** — TS's rule for "can a value of type A be used where type B is expected?": a `Task` is assignable to `Pick<Task, "id">`, but `{ title: string }` is not assignable to `Task` (missing `priority: number`). The verb behind TS2322 errors. (Lessons 06–07, 31)

**async/await** — syntax that lets you write Promise-consuming code top-to-bottom: `const res = await fetch(url)`. `await` pauses the function (not the thread) until the Promise settles. (Lesson 05)

**BFF (Backend-for-Frontend)** — the Next.js layer between the browser UI and the Go backend: route handlers in `app/api/*/route.ts` hold secrets, forward `Authorization` headers, validate Go's JSON with zod, and reshape it for the UI. This is the layer Wen owns. (Lessons 20, 22, 24–30)

**bundler** — the tool (Turbopack/webpack via Next.js) that at build time merges your hundreds of ESM modules into a few optimized JS files the browser downloads. "Bundle size is the cost of `use client`" refers to how much code the bundler must ship for the browser. (Lessons 04, 33)

**client component** — a component that runs in the browser (and also on the server for initial HTML). Marked `"use client"`; only these can use `useState`, `onClick`, `useEffect`. Your TeamPulse task-filter lives in one. (Lesson 18)

**controlled input** — a form field whose `value` comes from state and whose `onChange` sets that state, so React owns the text. The TeamPulse login form's username/password fields are controlled inputs. (Lesson 14)

**cookies()** — Next.js server-side API to read request cookies (`const token = (await cookies()).get("tp_token")?.value`). In Next 15 it's `async` — you must `await` it. Only usable in server components/route handlers. (Lessons 27, 29)

**discriminated union** — a union of object types sharing one literal "tag" field, letting TS narrow: `{ status: "ok"; data: Task } | { status: "error"; message: string }`. Check `result.status === "ok"` and TS knows `result.data` exists. The idiomatic way to model a Go API result. (Lesson 07)

**env var** — configuration injected by the environment: `GO_API_URL`, `GO_API_KEY`. Read on the server via `process.env.X` at runtime; anything the browser needs must be baked at build time as `NEXT_PUBLIC_X`. (Lessons 29, 35)

**ESM (ECMAScript Modules)** — the modern JS module system: `import`/`export` with statically-analyzable bindings. All course code is ESM; static imports are what let the bundler tree-shake. (Lesson 04)

**event loop** — the single-threaded JS scheduler: run code to completion, then drain the microtask queue (Promise callbacks), then macrotasks (timers, I/O). Explains why two `console.log`s around an `await` print in a surprising order. (Lesson 05)

**fetch** — the browser/Node built-in for HTTP: `await fetch("http://localhost:8080/api/projects", { headers })`. Returns a Promise resolving to a `Response`; you still must `await res.json()` and check `res.ok` yourself. (Lessons 05, 19, 20)

**hydration** — the browser-side step where React runs your component functions against the server's HTML and attaches event handlers/state to it. Server renders markup → browser "hydrates" it into a live app. (Lessons 18, 21)

**hydration mismatch** — the error when server HTML and the first client render disagree (e.g. `new Date()` differs between server and browser). React warns or throws. Fix: render deterministically, or gate browser-only code behind `useEffect`/`mounted` state. (Lessons 18, 21, 31)

**hook** — a function React gives components to use its machinery: `useState`, `useEffect`. Rules: call only at the top level of a component/hook, never conditionally; only in components or other hooks. Enforced by the `react-hooks` ESLint plugin. (Lessons 13, 15, 31)

**httpOnly cookie** — a cookie JS cannot read (`document.cookie` won't see it), set via `Set-Cookie` from a route handler. The BFF stores the Go token in one so XSS can't steal it — the core of Lesson 27's auth design.

**inference** — TS figuring out a type for you: `const tasks = [...]` gets a type without annotation. Trust it; annotate at boundaries (function params, parsed API data) and let inference do the other 90%. (Lesson 06)

**invalidation** — telling TanStack Query that cached server state is stale, forcing a refetch: `queryClient.invalidateQueries({ queryKey: ["tasks"] })` after a `POST /api/tasks`. You invalidate; the cache refetches. (Lesson 26)

**JSX** — HTML-looking syntax inside TS that compiles to function calls: `<TaskRow task={t} />` becomes `createElement(TaskRow, { task: t })`. Attributes become the component's props. (Lesson 10)

**key** — the stable `key={task.id}` you attach to each element from `.map()`, so React can match old to new rows. Using array index breaks reordering/selection — TeamPulse's `key` is always `task.id`. (Lesson 12)

**middleware** — `middleware.ts` at the project root: a function running before a request completes, able to redirect/rewrite/set headers. Use for coarse auth gating (no token → `/login`), not business logic. Runs on the edge runtime. (Lesson 29)

**microtask** — a callback queued by Promises; runs after the current synchronous code, before timers. This is why `await` continuations interleave the way they do in the event loop. (Lesson 05)

**NEXT_PUBLIC_** — prefix that inlines an env var's value into the client bundle at build time. Anything with this prefix is public — never put `GO_API_KEY` behind it, only things like a public API base URL. (Lessons 29, 35)

**narrowing** — refining a wide type to a specific one inside a branch: after `if (typeof x === "string")` or `if (r.status === "error")`, TS treats `x`/`r` as the specific type. Your main tool for handling Go's varying responses. (Lesson 07)

**optional property (`?`)** — a field that may be absent: `assignee?: string` means `string | undefined`. Reading `.toUpperCase()` on it without a check is TS2532. Prefer `?? "Unassigned"` or optional chaining `?.`. (Lesson 07)

**props** — the read-only input object a parent passes a component: `<ProjectCard project={p} />` gives `ProjectCard` a prop named `project`. A component is a pure function of its props. (Lesson 11)

**Promise** — an object representing a future value with states pending/fulfilled/rejected; `.then`/`catch`, or `await` inside `async`. Every `fetch` and every route-handler call is a Promise underneath. (Lesson 05)

**route handler** — a `route.ts` file in `app/api/.../` exporting `GET`/`POST`/etc. functions. It's a server-side HTTP endpoint: the BFF's unit of work, e.g. `app/api/projects/route.ts` calling Go and reshaping the JSON. (Lessons 20, 22)

**Server Component** — the default in the App Router: a component that runs only on the server, can `await` directly, hold secrets, and never ships its code to the browser. TeamPulse's project-list page is one. (Lessons 18, 19)

**state** — data a component owns that, when changed via its setter (`setFilter("doing")`), triggers a re-render. `useState` returns `[value, setter]`. Unlike a plain variable, state survives re-renders. (Lesson 13)

**status code** — the HTTP result of a call: 200 OK, 201 created, 401 bad credentials, 404 unknown project id, 500 (the flaky endpoint). `fetch` does NOT reject on 4xx/5xx — check `res.ok` yourself. (Lessons 20, 24)

**tree-shaking** — the bundler dropping exported code nothing imports. Works because ESM imports are static. Dead exports from your BFF helpers cost the browser nothing. (Lessons 04, 33)

**TS2345** — "Argument of type X is not assignable to parameter of type Y": you handed a function the wrong shape, e.g. passing `string` where a `Task` is expected. Read the first line only; it names both types. (Lesson 31)

**`"use client"`** — the directive at the top of a file marking its components client components and turning everything it imports into client code too. It's a boundary marker, not a switch that moves rendering. (Lesson 18)

**utility types** — TS built-in type transformers: `Partial<Task>` (all optional), `Pick`, `Omit`, `Record<K, V>`, `ReturnType<f>`. How you derive the BFF's reshaped DTOs from the Go-mirroring types. (Lesson 08)

**zod** — a runtime validation library. Define a schema, call `schema.parse(json)` on the Go response; it throws if the data lies. Combined with `z.infer` you get a type from the same source of truth. (Lesson 25)

**`z.infer`** — extract the TS type from a zod schema: `type Project = z.infer<typeof ProjectSchema>`. One definition (the schema) yields both runtime validation and a static type. (Lesson 25)

**`unknown`** — the type-safe "I don't know yet": you can't use it until you check/narrow it, unlike `any`. `JSON.parse` should land in `unknown`, then zod `parse` promotes it to a real type. (Lessons 07, 25)

---
*Related lessons: [04 modules & bundler](../lessons/04-modules-and-bundler-mental-model.md) · [05 async](../lessons/05-async-promises-await-fetch.md) · [06–08 TypeScript](../lessons/07-typescript-ii-shapes-and-narrowing.md) · [13 state](../lessons/13-state-usestate-and-rerenders.md) · [18 server vs client](../lessons/18-server-vs-client-components.md) · [20 route handlers](../lessons/20-route-handlers-your-first-bff.md) · [25 zod](../lessons/25-typed-contracts-with-zod.md) · [31 error decoding](../lessons/31-decoding-typescript-and-eslint-errors.md)*
