# Lesson 35 — Deploying to Vercel

*Phase 5 · Session 35 of 38 · ~2 hours · Prereqs: Lesson(s) [20](20-route-handlers-your-first-bff.md), [29](29-middleware-and-configuration.md), [33](33-performance-at-working-level.md)*

## Why this matters

A BFF you've never deployed is a BFF you don't understand yet. Deployment is where the
env-var model from Lesson 29 stops being theory: **build-time vs runtime** is the single
concept behind the most common production incident a junior causes — "it works in dev but
breaks in prod." On the team, your route handlers will run on servers you don't control
(Vercel or similar), talking to the real Go backend at a URL that only exists in prod
config. This lesson rehearses that exact situation with the mock Go API.

## What you'll be able to do

- You can deploy a Next.js App Router repo to Vercel from a Git push, and explain what Vercel actually runs.
- You can classify every env var in your repo as build-time or runtime, and predict what happens if you change it without redeploying.
- You can explain why `NEXT_PUBLIC_*` values are frozen into the JS bundle at build time — and why that makes them public.
- You can use preview deploys: what URL a PR gets, and why teams review on preview URLs instead of local.
- You can diagnose the classic "worked in dev, broke in prod" outage and name the two usual culprits: missing env var and wrong backend URL.

## Principles

### What Vercel runs, in three sentences

You push your repo; Vercel detects Next.js, runs `pnpm install && pnpm build`, stores the
output, and runs `pnpm start` — the *same* `next start` Node server that runs your route
handlers, on their infrastructure. Route handlers are server functions in prod just as in
dev (Lesson 20); what changes is everything *around* them: your terminal becomes the
**Deployments → Runtime Logs** page, your `.env.local` becomes the project's Environment
Variables settings, and `localhost:8080` becomes... nothing. It's gone. That last part is
the plot of this lesson.

### Build-time vs runtime: the fold in the timeline

There are two moments your code is touched:

1. **Build time** (`next build`, once per deploy): the bundler walks your client code and
   **literally pastes** every `process.env.NEXT_PUBLIC_*` value it finds into the JS files
   it emits. `console.log(process.env.NEXT_PUBLIC_API)` in client code becomes
   `console.log("https://api.example.com")` in the shipped bundle — a string constant.
   Changing the env var later changes nothing until you *build again*.
2. **Runtime** (`next start`, every server request): plain `process.env.FOO` reads in
   server code (route handlers, Server Components, `middleware.ts`) happen *per request*
   and reflect the current environment. Change the var, and the next request sees it
   (Vercel redeploys on env change by default, which hides this subtlety — but know it).

The classification table to memorize:

| Code location | Env var form | Read when | Visible to users? |
|---|---|---|---|
| Client component / any bundled JS | `NEXT_PUBLIC_*` | build | **yes — it's in the bundle** |
| Client wanting config | `NEXT_PUBLIC_*` only | build | yes |
| Route handler / Server Component | anything | runtime | no |
| Secrets (API keys, passwords) | **never `NEXT_PUBLIC_`** | runtime, server-only | no |

Footgun from Lesson 33, formalized: a value in `NEXT_PUBLIC_*` is a value you have
published. And the converse trap — a *server* env var read from client code compiles to
`undefined` (it wasn't inlined), which is the "works locally, `undefined` in prod" special
case of everything below.

### Why prod broke but dev worked

Memorize this incident shape; you will cause it or review it within a month:

> Dev: `GO_URL=http://localhost:8080` in `.env.local`, everything green.
> Prod: every `/api/*` returns 500. Runtime logs show
> `fetch failed: getaddrinfo ENOTFOUND` or `GO_URL is not defined`.

Cause: the var exists in your laptop's `.env.local` but was never added to Vercel's
Environment Variables. Dev never proves your prod env is complete — **only prod does.**
The professional habit: after the first deploy, run an "env audit" — grep every
`process.env.X` in server code, and check each X exists in the Vercel dashboard:

```bash
grep -rhoE "process\.env\.[A-Z_]+" src/ | sort -u
```

Second culprit: the URL itself. Locally the Go backend is `http://localhost:8080`. Vercel's
servers have *your* localhost nowhere in them — `GO_URL=http://localhost:8080` in prod
means "call Vercel's own container, where nothing listens on 8080." In the real job,
prod `GO_URL` points at the team's real Go service (e.g. `https://api.teampulse.internal`
— provided by the team; this is question #9 on Lesson 37's list). For this course's demo,
see the honest workaround below.

### The mock-stays-local demo pattern

Your mock Go API is stdlib Go running in a terminal on your machine. It does **not**
deploy to Vercel (Vercel runs Node, not arbitrary Go binaries, and a stateful in-memory
demo backend doesn't belong there anyway). So the deployed app can't reach it directly.
Two accepted demo configurations:

1. **Honest prod config** (what you document): `GO_URL` in Vercel points at a *reachable*
   Go backend. In the real job that's the team's deployment. For the course, either run
   the mock somewhere reachable (any server you have, `GO_URL=http://<that-host>:8080`) or
   accept option 2.
2. **Local-only API demo**: deploy the *UI and marketing paths* publicly with
   `GO_URL` pointed at a stub, and demo the data-driven pages by running `pnpm dev`
   locally with the mock. Write this into your README like a professional: *"The mock Go
   backend runs locally on :8080; in production GO_URL targets the real Go service — this
   Vercel deployment is a UI demo."* Reviewers read that sentence and trust you; they
   panic when they have to guess it.

The transferable lesson — **where does each dependency run in prod?** — is worth more than
the deployment itself.

### Preview deploys: every branch gets a URL

Push a branch or open a PR and Vercel builds a **preview deployment** at a URL like
`teampulse-git-feature-task-form-wen.vercel.app`, alongside the **production** deployment
at the project URL. Each PR's preview is a frozen, shareable, clickable version of that
branch — which is why teams review UI changes on preview URLs ("LGTM on preview") instead
of asking reviewers to check out your branch. Preview deploys usually use preview env vars
(a third env scope in Vercel's UI); point `GO_URL` wherever your demo backend lives.

## Follow along

Work in `~/teampulse` (Lessons [30](30-p3-full-stack-teampulse.md) and
[34](34-minimal-testing-teams-actually-use.md)).¹

1. **Prep the repo for the world.**
   - Make sure `pnpm build && pnpm test` are both green locally — Vercel runs the build,
     and a red build is a red deploy.
   - Add the deploy README section now (the mock-stays-local honesty statement above).
   - Run the env audit: `grep -rhoE "process\.env\.[A-Z_]+" src/ | sort -u`. Write the
     list down; this is your Vercel checklist. Expected for a typical P3: `GO_URL` and
     maybe `SESSION_SECRET`, `NEXT_PUBLIC_APP_NAME`.

2. **Push to GitHub and connect Vercel.** Create a private GitHub repo, push, sign up at
   vercel.com (GitHub login), **Add New → Project → Import** the repo. Vercel detects
   Next.js; before deploying, open **Settings → Environment Variables** and add every var
   from step 1 for Production *and* Preview (use `GO_URL=http://localhost:8080` for now —
   we *want* to watch it fail). Click Deploy.
   Expected: build succeeds in ~1 minute; you get a `*.vercel.app` URL; the shell page
   renders.

3. **Reproduce the outage on purpose.** Open the deployed dashboard and log in. Expected:
   the login fails; the UI shows its error state (Lesson 21). Open
   **Deployments → your deploy → View function logs** (or `vercel logs <url>` via the
   CLI). Expected: your `[bff /api/auth/login]` log line (Lesson 32) followed by a fetch
   failure — `connect ECONNREFUSED 127.0.0.1:8080` or DNS error. You have just seen the
   BFF↔world seam fail in prod, and the logs localized it in under a minute.

4. **Fix the env, not the code.** Get the mock reachable: run it on any host you control
   (`go run main.go` bound to a LAN/public interface, or a cheap VM), and set the Vercel
   `GO_URL` to that reachable URL — no code change, no redeploy-with-different-commits.
   Redeploy (Vercel redeploys on env change; otherwise Deployments → ⋯ → Redeploy).
   Log in with any username and password `demo`.
   Expected: green. If it fails with a CORS-looking error, remember: server-to-server
   fetches have no CORS — that error would mean a client-side direct call exists; grep
   for `8080` in client files (Lesson 32's check).

5. **Prove the NEXT_PUBLIC freeze.** Temporarily put
   `console.log("app name:", process.env.NEXT_PUBLIC_APP_NAME)` in a client component,
   set the var to `TeamPulse-A` in Vercel, deploy, open DevTools console, and *find the
   literal string* in the bundle (Sources tab, or View Source). Then change the var to
   `TeamPulse-B` in the dashboard and — without redeploying — hard-refresh: the console
   still says `TeamPulse-A`. Redeploy; now it's B. Expected observation: **NEXT_PUBLIC
   values are build artifacts, not configuration.** Remove the log.

6. **Preview deploy ritual.** `git checkout -b feat/deploy-notes`, change some visible
   copy, push, open a PR. Expected: Vercel posts a comment (or check the Deployments tab)
   with a preview URL. Open it — the *old* copy (prod) and *new* copy (preview) side by
   side. This is the review workflow of Lesson 37, pre-installed.

## Your turn

1. Your team's reviewer asks: "You set `SESSION_SECRET` as a `NEXT_PUBLIC_` var. What just leaked, and what's the fix?" Answer in writing, precisely.
2. A teammate deploys and gets a **build failure** on Vercel while `pnpm build` passes locally. List the two most likely causes on a fresh clone, and the one command that reproduces theirs most faithfully.
3. Write the README paragraph that tells a stranger where every piece of TeamPulse runs in "production": UI, BFF, Go backend, session storage, and where the demo's `GO_URL` points vs where a real deployment's would.

<details>
<summary>Solution</summary>

1. The session secret was inlined into the public JS bundle at build time — anyone can
   read it and forge session tokens. Fix: rename it to non-prefixed `SESSION_SECRET`, set
   it as a runtime env var in Vercel, read it only in server code (route handlers /
   cookie signing, Lesson 27), redeploy, and *rotate the secret's value* since the old one
   is already public.
2. (a) A var required at build time (e.g. `NEXT_PUBLIC_*` or a build-time check) that
   Vercel doesn't have set — your local `.env.local` hides it. (b) Lockfile/registry drift
   — a different resolved dependency version (usually `pnpm-lock.yaml` out of sync, or the
   project built with npm locally). Most faithful repro: clone the repo *fresh into a new
   directory*, `pnpm install --frozen-lockfile && pnpm build` with no `.env.local`.
3. Must state: UI + BFF both run as one Next.js app on Vercel (route handlers are the BFF,
   server-side); the Go backend is a separate service — the mock runs locally on `:8080`
   for demos, and a real deployment sets `GO_URL` to the team's Go service; sessions live
   in an httpOnly cookie (Lesson 27), not in localStorage; the demo deployment's `GO_URL`
   points at a locally-reachable mock and is explicitly a UI demo, not production wiring.

</details>

## Checkpoint

1. You change a runtime env var in the Vercel dashboard but skip the redeploy. Name the one kind of var that changes behavior immediately and the kind that does not.

<details>
<summary>Answer</summary>

Server-side (non-`NEXT_PUBLIC_`) reads are per-request, so they take effect (and Vercel
redeploys by default anyway); `NEXT_PUBLIC_*` values were baked into the bundle at build
time and don't change until a new build.

</details>

2. Why is `GO_URL=http://localhost:8080` set correctly in Vercel still broken?

<details>
<summary>Answer</summary>

`localhost` resolves to the machine running the code — Vercel's container, where nothing
listens on 8080. The var must point at a network-reachable Go backend; in the real job,
the team's deployed Go service.

</details>

3. Where is your server-side `console.log` output in prod, and what's the dev-tool
   equivalent you used in Lesson 32?

<details>
<summary>Answer</summary>

Vercel's deployment/function Runtime Logs (dashboard or `vercel logs`). The dev equivalent
is the terminal running `pnpm dev`.

</details>

4. What URL does a PR get on a Vercel-connected repo, and why do teams prefer reviewing there to "check out my branch"?

<details>
<summary>Answer</summary>

A unique preview deployment URL (branch-derived, e.g. `project-git-<branch>-<user>.vercel.app`).
Reviewers click one link, see the exact build with the right env, no local setup, and
comments attach to the PR.

</details>

5. You log into the deployed app with password `demo` and every `/api/*` route 500s while
   the public homepage renders. Name the first two checks, in order, and which layer each
   eliminates.

<details>
<summary>Answer</summary>

(1) Runtime logs for the failing route — proves whether the handler executed and what the
Go fetch returned (eliminates BFF-logic vs BFF↔Go connectivity). (2) The Vercel
Environment Variables list vs your `grep process.env` audit (eliminates missing-env-var).
A rendering homepage eliminates the UI/deploy itself.

</details>

## Primary source

[Next.js docs — Environment Variables](https://nextjs.org/docs/app/guides/environment-variables).
It states the `NEXT_PUBLIC_` inlining rule and the build-vs-runtime loading order
directly, which is the exact mechanism behind every "worked in dev, broke in prod" story
in this lesson.

## Before next session

Keep the Vercel project alive — Lesson 36's capstone deploys your finished feature there.
Spend 5 minutes: open the Vercel dashboard's Environment Variables page and, beside each
var, write the one-sentence reason it's prefixed (or not). Any var you can't justify gets
fixed in Lesson 36.

---

¹ npm alternative: `npm run build` / `npm test` throughout; Vercel auto-detects npm if there's no `pnpm-lock.yaml`.
