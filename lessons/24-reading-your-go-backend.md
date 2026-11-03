# Lesson 24 — Reading your Go backend

*Phase 4 · Session 24 of 38 · ~2 hours · Prereqs: Lessons 05, 07, 20, 22*

## Why this matters

Your team's Go backend is something you will never maintain — but you must be able to
*read* it. Half of the "fix the BFF" tickets you'll take start with the question "what
does Go actually return here?" The answer is sitting in `backend-mock/main.go`, and this
lesson teaches you to extract it without knowing a line of Go. This is MISSION item 4:
trace a bug from the browser Network tab → BFF route handler → Go API response.

## What you'll be able to do

- Look at a Go struct with `json` tags and write the matching TypeScript interface cold
- Explain why a Go `time.Time` field becomes a `string` in your TS types
- List every status code the TeamPulse API can return and what each one means for your
  route handlers
- Predict the exact JSON body of any endpoint before you run the `curl`
- Maintain `~/teampulse/lib/api-types.ts` as the single mirror of the Go shapes

## Principles

### A contract, not code

You read `main.go` the way you'd read an API doc someone handed you. Three things matter;
everything else is plumbing you can skip:

1. **The routes** — the `mux.HandleFunc(...)` lines near the bottom of `main.go` list
   every URL the server answers.
2. **The structs** — `User`, `Project`, `Task` at the top define every JSON shape.
3. **The errors** — every `writeErr(...)` call is one possible failure, with its status
   code and message baked in.

### Go struct → JSON → TS interface

A Go *struct* is a typed bag of fields, roughly a TS `interface`. The part in backticks
is a **struct tag** — metadata attached to the field. The `json:"..."` tag tells Go what
key to use when it serialises the struct to JSON:

```go
type Task struct {
  ID        string    `json:"id"`
  ProjectID string    `json:"projectId"`
  Title     string    `json:"title"`
  Status    string    `json:"status"` // todo | doing | done
  Priority  int       `json:"priority"`
  DueDate   time.Time `json:"dueDate"`
  Assignee  string    `json:"assignee"`
}
```

The tag value wins — always. Go field names are capitalised (`ProjectID`) because Go
requires capitalisation to export a field, but the JSON key is whatever the tag says
(`projectId`). Your TS interface mirrors the **JSON keys**, never the Go names:

```ts
// lib/api-types.ts — mirrors Task in backend-mock/main.go
interface Task {
  id: string;
  projectId: string;
  title: string;
  status: "todo" | "doing" | "done";
  priority: number;
  dueDate: string;
  assignee: string;
}
```

Two conversions to memorise: Go `string` → TS `string`, Go `int` → TS `number`.

### Dates are strings

JSON has no date type. Go's `time.Time` serialises as an RFC 3339 string — e.g.
`mustDate("2026-05-01")` in `main.go` becomes `"2026-05-01T00:00:00Z"`. Type date fields
as `string`; parse into a `Date` only at display time (see
[Lesson 07](07-typescript-ii-shapes-and-narrowing.md) on modelling precisely).

### Status-code semantics

The Go code decides the status in each `writeJSON(w, status, ...)` / `writeErr(w, status,
msg)` call. For TeamPulse the full set is:

| Status | Where in main.go | Meaning for your BFF |
|---|---|---|
| 200 OK | most GETs, login | body is the success shape |
| 201 Created | `POST /api/tasks` | task created — **pass 201 through**, don't downgrade to 200 |
| 400 Bad Request | malformed JSON body; `projectId and title are required` | your request was wrong — a bug in *your* code |
| 401 Unauthorized | missing/invalid token; wrong password | session problem — clear the cookie, don't retry (Lesson 28) |
| 404 Not Found | `project not found` | resource doesn't exist — render "not found", don't retry |
| 405 Method Not Allowed | wrong HTTP verb | your route handler called the wrong method |
| 500 Internal Server Error | flaky endpoint, "user vanished" | upstream problem — the only class worth retrying |

The rule that saves you hours: **4xx means fix your call, 5xx means the upstream is
sick.** Your route handlers should map each class differently (Lesson 28 does this
properly).

### One error shape

Every failure goes through `writeErr`, which always sends:

```json
{ "error": "project not found" }
```

That's the entire error contract. Model it once:

```ts
// lib/api-types.ts
interface ApiError {
  error: string;
}
```

Your BFF can trust exactly one error shape from this backend — and must still *verify*
that trust, which is Lesson 25's job.

### Auth, from the consumer's seat

`handleLogin` checks the password (the literal word `demo`; any username accepted),
mints a token shaped `tok-` + 32 hex characters, and stores it in a `tokens` map.
Every protected handler runs `requireAuth`, which demands an
`Authorization: Bearer <token>` header. Tokens never expire in the mock. You will store
the token **in the BFF, in an httpOnly cookie** — the browser never sees it (Lesson 27).

## Follow along

1. Start the backend from the repo root:
   ```bash
   cd ~/workspace/nextjs-bff-academy/backend-mock && go run main.go
   ```
   Expected: a log line ending `TeamPulse mock backend listening on http://localhost:8080`.
   Leave this terminal tab open.

2. Health check (no auth needed):
   ```bash
   curl -s http://localhost:8080/api/health
   ```
   Expected: `{"status":"ok","time":"2026-..."}`. Note `time` is a string — RFC 3339,
   UTC, exactly as predicted above.

3. Log in. The password is the literal word `demo`:
   ```bash
   curl -s -X POST http://localhost:8080/api/login \
     -H 'Content-Type: application/json' \
     -d '{"username":"wen","password":"demo"}'
   ```
   Expected: `{"token":"tok-...","user":{"id":"u1","name":"Wen","role":"admin"}}`.
   Read the `handleLogin` source (around line 104) and point at the line that produced
   each key.

4. Stash the token and fetch tasks. Predict the JSON *before* pressing enter — you
   should expect tasks `t1`, `t2`, `t5` only, since you filter on `p1`:
   ```bash
   TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
     -H 'Content-Type: application/json' \
     -d '{"username":"wen","password":"demo"}' | sed 's/.*"token":"\([^"]*\)".*/\1/')
   curl -s 'http://localhost:8080/api/tasks?projectId=p1' \
     -H "Authorization: Bearer $TOKEN"
   ```
   Expected: an array of three objects whose keys are exactly
   `id, projectId, title, status, priority, dueDate, assignee` — the struct order.

5. Hunt the 404. `-i` makes curl print the status line and headers:
   ```bash
   curl -si http://localhost:8080/api/projects/p999 -H "Authorization: Bearer $TOKEN"
   ```
   Expected: `HTTP/1.1 404 Not Found` then `{"error":"project not found"}`.

6. Hunt the 401 — same URL, no header:
   ```bash
   curl -si http://localhost:8080/api/projects/p999
   ```
   Expected: `401` + `{"error":"missing or invalid token"}`.

7. Enumerate every possible failure in one grep, from the repo root:
   ```bash
   grep -n 'writeErr(' backend-mock/main.go
   ```
   Expected: about a dozen lines. Each one is `(w, <status>, "<message>")` — build your
   own status-code table from it and diff against the one above.

8. Create `~/teampulse/lib/api-types.ts` with `User`, `Project`, `Task`, `ApiError`,
   and a `LoginResponse` type (`{ token: string; user: User }`). Then:
   ```bash
   cd ~/teampulse && pnpm tsc --noEmit
   ```
   Expected: no output — clean compile.

## Your turn

1. `handleLogin` and `handleTasks` use *anonymous* inline structs for their request
   bodies (search for `var body struct`). Write TS types for the login request body and
   the create-task request body from them.

   <details><summary>Solution</summary>

   ```ts
   // POST /api/login request body
   interface LoginRequest {
     username: string;
     password: string;
   }

   // POST /api/tasks request body
   interface CreateTaskRequest {
     projectId: string;
     title: string;
     priority: number;
   }
   ```
   Note what the Go struct does *not* include: `status`, `dueDate`, `assignee` — the
   server fills those in. Mirroring the request struct prevents you from "helpfully"
   sending fields Go will silently ignore.
   </details>

2. POST a task with `title` missing. Predict the exact status line and body, then verify
   with curl.

   <details><summary>Solution</summary>

   ```bash
   curl -si -X POST http://localhost:8080/api/tasks \
     -H "Authorization: Bearer $TOKEN" \
     -H 'Content-Type: application/json' \
     -d '{"projectId":"p1","priority":1}'
   ```
   Expected: `HTTP/1.1 400 Bad Request` and `{"error":"projectId and title are required"}`
   — from the guard in `handleTasks`. A 400 here is *your* bug, not the backend's.
   </details>

3. Read `handleFlaky` and `handleSlow` at the bottom of `main.go`. Write one comment
   above each in your own words, stating what kind of BFF failure each is designed to
   rehearse.

   <details><summary>Solution</summary>

   `handleFlaky` fails ~30% of requests with a 500 `{"error":"upstream hiccup..."}` — it
   rehearses **retry policy**: your BFF must decide when re-calling is safe. `handleSlow`
   always sleeps 2 seconds — it rehearses **timeouts**: your BFF must stop waiting and
   degrade gracefully instead of hanging. Lesson 28 uses both.
   </details>

## Checkpoint

1. A Go field reads `DueDate time.Time `json:"dueDate"``. What is the JSON key, and what
   TS type do you write?

   <details><summary>Answer</summary>
   Key `dueDate` (the tag wins over the Go name). TS type `string` — JSON has no date
   type; `time.Time` serialises as an RFC 3339 string like `"2026-11-20T00:00:00Z"`.
   </details>

2. Your route handler proxies `POST /api/tasks`. Your teammate's handler returns 200 on
   success. What's wrong?

   <details><summary>Answer</summary>
   Go returns **201 Created** for that endpoint. The BFF should pass 201 through —
   clients (and TanStack Query invalidation logic, Lesson 26) may branch on it, and
   silently downgrading the status lies about what happened.
   </details>

3. Go replies `{"error":"project not found"}`. What status code accompanies it, and what
   should your route handler do — retry, re-auth, or render?

   <details><summary>Answer</summary>
   404. It's a 4xx — the resource genuinely doesn't exist. Don't retry, don't re-auth;
   surface a not-found state to the UI (Lesson 21's `not-found.tsx`).
   </details>

4. A teammate adds a Go field `Assignee string` with **no** json tag. What will the JSON
   key be, and how do you find out without asking them?

   <details><summary>Answer</summary>
   Without a tag, Go uses the Go field name as-is — key `Assignee`, capital A. Find out
   yourself: `go run main.go`, curl the endpoint, read the key. Never type from
   assumption; the wire format is observable.
   </details>

5. You have zero Go fluency. Name the three things in `main.go` you actually read, and
   the thing you skip.

   <details><summary>Answer</summary>
   Read: the `mux.HandleFunc` route registrations, the structs with their `json` tags,
   and every `writeErr` call (status + message). Skip: the handler logic in between —
   slicing, map lookups, rand calls. The contract lives in the three you read.
   </details>

## Primary source

[encoding/json — Struct tagged fields](https://pkg.go.dev/encoding/json#hdr-Structs) —
the canonical explanation of the one Go feature you must be able to read: `json:"..."`
tags, including what happens when the tag is missing.

## Before next session

In `~/teampulse`, run `pnpm add zod`, then open <https://zod.dev> and read only the
"Basic usage" section. Note one thing zod gives you that the plain `interface` types in
`lib/api-types.ts` cannot — you'll use it all of Lesson 25.
