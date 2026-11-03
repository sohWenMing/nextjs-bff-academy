# Lesson 34 — Minimal Testing That Teams Actually Use

*Phase 5 · Session 34 of 38 · ~2 hours · Prereqs: Lesson(s) [05](05-async-promises-await-fetch.md), [20](20-route-handlers-your-first-bff.md), [25](25-typed-contracts-with-zod.md)*

## Why this matters

Nobody on a real team asks you to "write tests for everything." They ask you to not break
things: a PR that touches the BFF boundary without a test is a PR the reviewer must test by
hand. Your bar is much smaller than a testing textbook's — **one unit test on a pure
function, one route-handler test against the live mock Go API, one component test** — and
that trio is exactly the shape of the test files you'll find (and extend) in your team's
repo. You are not learning testing theory; you are learning to add the fourth file to
their existing three.

## What you'll be able to do

- You can set up vitest in a Next.js repo (config, `pnpm test` script) and run it.
- You can write a unit test for a pure function using `describe`/`it`/`expect`.
- You can test a route handler by importing its exported `GET`/`POST` and calling it with a `Request` — no mock server needed, because your mock Go API *is* the integration target at `http://localhost:8080`.
- You can render a component with `@testing-library/react` and assert on what the user sees, querying by role/text — not by CSS class.
- You can look at a foreign team's test directory and recognize which of these three kinds each file is, so extending theirs is a fill-in-the-blanks exercise.

## Principles

### The testing pyramid, working-level edition

Three kinds of test, three costs:

| Kind | Tests | Speed | Your frequency |
|---|---|---|---|
| **Unit** | one pure function, no I/O | milliseconds | many |
| **Route-handler (integration)** | one BFF endpoint, real dependencies | ~100 ms | a few — one per important endpoint |
| **Component** | one component's user-visible behavior | ~1 s | one per critical interaction |

The instinct to test *everything* is the instinct that kills testing cultures. Teams keep
the bar: pure logic that had a bug gets a unit test; every route handler you *own* gets at
least one happy-path test; components get tests only for interactions users actually rely
on. MISSION.md wants "small, reviewable PRs" — a PR with a behavior change plus one test
adjustment is the reviewable kind.

### vitest, in one page

vitest is the standard test runner for modern TS/React projects (Jest's successor: same
APIs, Vite-powered, TS out of the box). Three concepts, no more:

```ts
import { describe, it, expect } from "vitest";

describe("summarizeTasks", () => {          // a group of related tests
  it("counts done tasks", () => {           // one behavior, one sentence name
    const tasks = [
      { id: "t1", title: "a", status: "done" as const, priority: 1 },
      { id: "t2", title: "b", status: "todo" as const, priority: 1 },
      { id: "t3", title: "c", status: "done" as const, priority: 1 },
    ];
    expect(summarizeTasks(tasks).done).toBe(2);   // the assertion
  });
});
```

- `describe` groups, `it` is one test, `expect(x).toBe/.toEqual/.toThrow` asserts.
- `.toBe` is identity (primitives); `.toEqual` is deep structural equality (objects/arrays).
- A test file is any `*.test.ts` / `*.test.tsx` — colocated with the code it tests, which
  is where most teams put them.

### Why route-handler tests don't need mock servers here

The usual painful part of testing a BFF is the "call the backend" part — teams that can't
hit a real backend reach for MSW or hand-rolled `fetch` mocks. You skip all of it: your
mock Go API *is* a real backend that runs locally on `http://localhost:8080`
(Lesson 24). Testing your route handler against it is an **integration test** — the same
kind your team runs against staging. If your team's CI can't start the Go backend, *that*
is the day you learn `vi.mock` — but for now: real Go in, real assertions out.

A route handler is just an exported async function taking a `Request` (Lesson 20), so a
test is embarrassingly direct:

```ts
const res = await GET(new Request("http://localhost:3000/api/projects", {
  headers: { Cookie: `tp_token=${token}` },
}));
```

### Testing components like a user

`@testing-library/react` renders a component in a fake DOM and gives you *user-facing*
queries: `getByRole("button", { name: "Save" })`, `getByText("No tasks yet")`,
`getByRole("textbox")`. The rule that separates good component tests from useless ones:
**query what the user perceives (role, label, text), never implementation** (no
`.find(".task-row")` — CSS classes are refactor bait; roles and labels are the contract
with the user and with accessibility). Async content (your fetches) uses `await
screen.findByText(...)` or `waitFor`, because React updates after the promise resolves.

## Follow along

All work is in `~/teampulse`. Start the mock Go API (`cd backend-mock && go run main.go`)
and leave it running — your tests will hit it for real.¹

1. **Install the toolchain.**

   ```bash
   pnpm add -D vitest @vitejs/plugin-react @testing-library/react @testing-library/jest-dom @testing-library/user-event jsdom
   ```

   Expected: five packages added, no peer-dependency errors on a Next 15 repo.

2. **Configure.** Create `vitest.config.ts` at the repo root:

   ```ts
   import { defineConfig } from "vitest/config";
   import react from "@vitejs/plugin-react";
   import path from "node:path";

   export default defineConfig({
     plugins: [react()],
     test: {
       environment: "jsdom",
       include: ["src/**/*.test.{ts,tsx}"],
       env: { GO_URL: "http://localhost:8080" },
     },
     resolve: { alias: { "@": path.resolve(__dirname, "src") } },
   });
   ```

   Add to `package.json` scripts: `"test": "vitest run"`. Expected: `pnpm test` runs and
   reports `No test files found, exiting with code 1` — that's success for now.

3. **Unit test a pure function.** Create `src/lib/summarize.ts` (or reuse a real pure
   helper from your P3, such as the task filter):

   ```ts
   export function summarizeTasks(tasks: { status: "todo" | "doing" | "done" }[]) {
     return {
       total: tasks.length,
       done: tasks.filter((t) => t.status === "done").length,
       open: tasks.filter((t) => t.status !== "done").length,
     };
   }
   ```

   Then `src/lib/summarize.test.ts`:

   ```ts
   import { describe, expect, it } from "vitest";
   import { summarizeTasks } from "./summarize";

   const t = (status: "todo" | "doing" | "done") => ({ status });

   describe("summarizeTasks", () => {
     it("splits done from open", () => {
       expect(summarizeTasks([t("todo"), t("doing"), t("done")])).toEqual({
         total: 3, done: 1, open: 2,
       });
     });
     it("handles the empty list", () => {
       expect(summarizeTasks([])).toEqual({ total: 0, done: 0, open: 0 });
     });
   });
   ```

   Run `pnpm test`. Expected: `2 passed`. Now break it on purpose: change `open` to count
   `=== "done"` and re-run. Read the failure diff — vitest prints expected vs received
   side by side. Restore.

4. **Route-handler test against the live Go mock.** Pick a handler that requires auth
   (Lesson 27). Write `src/app/api/projects/route.test.ts`:

   ```ts
   import { describe, expect, it } from "vitest";
   import { GET } from "./route";

   async function login(): Promise<string> {
     const res = await fetch("http://localhost:8080/api/login", {
       method: "POST",
       headers: { "Content-Type": "application/json" },
       body: JSON.stringify({ username: "wen", password: "demo" }),
     });
     const body = await res.json();
     return body.token as string;
   }

   describe("GET /api/projects (BFF)", () => {
     it("returns the project list when authenticated", async () => {
       const token = await login();
       const req = new Request("http://localhost:3000/api/projects", {
         headers: { Cookie: `tp_token=${token}` }, // match your cookie name from Lesson 27
       });
       const res = await GET(req);
       expect(res.status).toBe(200);
       const projects = await res.json();
       expect(Array.isArray(projects)).toBe(true);
       expect(projects[0]).toHaveProperty("name");
     });

     it("rejects requests without a session", async () => {
       const res = await GET(new Request("http://localhost:3000/api/projects"));
       expect(res.status).toBe(401);
     });
   });
   ```

   Run `pnpm test`. Expected: `4 passed`. If the first test gets `401`, your handler reads
   the cookie/token differently than the test sends it — this mismatch *is* the kind of
   bug the test exists to pin down; align the test with your handler's actual contract
   (or better, fix the handler so no-token reliably 401s — Lesson 27's rule).

5. **Component test.** Write `src/components/task-count.test.tsx` for a tiny component you
   create:

   ```tsx
   // src/components/task-count.tsx
   "use client";
   import { useState } from "react";

   export function TaskCount({ total, done }: { total: number; done: number }) {
     const [showOpen, setShowOpen] = useState(false);
     return (
       <div>
         <button onClick={() => setShowOpen((v) => !v)}>toggle</button>
         <p>{showOpen ? `${total - done} open` : `${done} of ${total} done`}</p>
       </div>
     );
   }
   ```

   ```tsx
   // src/components/task-count.test.tsx
   import { describe, expect, it } from "vitest";
   import { render, screen } from "@testing-library/react";
   import userEvent from "@testing-library/user-event";
   import { TaskCount } from "./task-count";

   describe("TaskCount", () => {
     it("toggles between done and open view", async () => {
       render(<TaskCount total={5} done={2} />);
       expect(screen.getByText("2 of 5 done")).toBeInTheDocument();
       await userEvent.click(screen.getByRole("button", { name: "toggle" }));
       expect(screen.getByText("3 open")).toBeInTheDocument();
     });
   });
   ```

   Add `import "@testing-library/jest-dom/vitest";` at the top so `toBeInTheDocument`
   exists (or a `vitest.setup.ts` referenced via `test.setupFiles`). Expected: `5 passed`.

6. **Recognize theirs.** Browse any open-source Next.js repo's `__tests__` or colocated
   `*.test.tsx` files (e.g. the TanStack Query examples repo). For each of five files,
   label it unit / route-handler / component, and note which queries it uses. That
   vocabulary is your onboarding shortcut for Lesson 37.

## Your turn

1. Write a **POST test** for your task-creation route handler (Lesson 20): log in, `POST /api/tasks` with `{ projectId: "p1", title: "written by a test", priority: 1 }` through your handler (not directly to Go), and assert `201` and that the returned body has an `id`. Remember the mock resets its data on restart.
2. Write a component test for your login form: it renders a password input and a submit button with accessible names, and submitting empty shows the validation message (the one you built in Lesson 14).
3. **Transfer question.** Your team's repo already has 30 tests. A reviewer says "your PR needs a test." What is the *minimum* honest response for a PR that adds one new BFF endpoint, and which of today's three patterns does it follow?

<details>
<summary>Solution</summary>

1. Same shape as step 4's GET test, but construct a `POST` Request with
   `method: "POST"`, a JSON body, and the session cookie; assert
   `res.status === 201` and `typeof body.id === "string"`. Pin `projectId: "p1"` — the
   mock's known project (`backend-mock/main.go`).
2. `render(<LoginForm />)`; `screen.getByLabelText(/password/i)`;
   `screen.getByRole("button", { name: /log in|sign in/i })`; `await userEvent.click(...)`
   with the field empty; `await screen.findByText(/password is required/i)` — match your
   actual message.
3. One route-handler test for the new endpoint: a happy path with the real mock/staging
   backend proving status + response shape, plus (cheap, high value) one negative path
   like missing auth → 401. That's the integration pattern from step 4 — colocated, named
   after the behavior, no new tooling. Reviewers accept "extended your existing pattern"
   far more readily than an invented testing framework.

</details>

## Checkpoint

1. Which vitest assertion would you use for each: a string status code, or a whole reshaped task object? Why the difference?

<details>
<summary>Answer</summary>

`toBe` for the status code (primitive identity) and `toEqual` for the object (deep
structural equality — compares every field, which is the point at a reshape boundary).

</details>

2. Why can this course skip MSW or `fetch` mocking for route-handler tests, and would a team's CI be able to skip it too?

<details>
<summary>Answer</summary>

Because a real Go backend — the mock on `http://localhost:8080` — is running locally and
is deterministic enough to assert against. Team CI can skip mocking only if it can start
the backend in the pipeline (dockerized); otherwise they mock at the boundary.

</details>

3. Name the rule for how `@testing-library/react` says to find elements, and the reason.

<details>
<summary>Answer</summary>

Query by what the user perceives — `getByRole`/`getByText`/`getByLabelText`, not CSS
classes or internals — because roles and labels are the user/a11y contract, so tests
survive styling refactors and fail when UX actually breaks.

</details>

4. Your route-handler test asserts `expect(res.status).toBe(200)` on `/api/tasks` while the
   mock Go API is up, and it passes. You kill the Go process and re-run: what happens and
   why is that behavior a *feature* of this test style?

<details>
<summary>Answer</summary>

The test fails (your handler's error path returns 5xx, or the fetch rejects —
Lesson 28's timeout). Feature: it's an integration test, so it detects upstream being
down — exactly the BFF↔Go seam failure you're paid to own.

</details>

5. Against the mock API at `http://localhost:8080`, what credentials does every test
   fixture log in with, and what must your test know about data persistence between runs?

<details>
<summary>Answer</summary>

Any username plus password `demo` (you always land as user `u1`, "Wen"). Data is
in-memory and resets when Go restarts — so never assert on tasks created by an earlier
test run, and don't assume your `POST`ed task still exists tomorrow.

</details>

## Primary source

[Vitest Guide](https://vitest.dev/guide/) — the official runner docs: install, config,
assertions, and environment setup, so every step above is checkable against the source
of truth rather than blog folklore.

## Before next session

Add a second route-handler test for an endpoint we didn't cover (tasks-by-project with a
`projectId` filter is a good one). Leave the suite green: Lesson 35 deploys to Vercel, and
"my tests pass locally" is the sentence that buys you patience when prod misbehaves.

---

¹ npm alternative: `npm i -D vitest @vitejs/plugin-react @testing-library/react @testing-library/jest-dom @testing-library/user-event jsdom` and `npm test`.
