# Lesson 37 — The Team-Codebase Playbook

*Phase 5 · Session 37 of 38 · ~2 hours · Prereqs: Lesson(s) [20](20-route-handlers-your-first-bff.md), [31](31-decoding-typescript-and-eslint-errors.md), [36](36-p4-capstone-ship-teampulse.md)*

## Why this matters

Everything so far was *your* repo. On Monday it's theirs: a 200-file Next.js codebase
someone else grew, with conventions nobody documented and opinions in every `.eslintrc`.
MISSION.md's fifth item — open a small, reviewable PR into a Next.js/TS codebase without
hand-holding — is a *playbook* skill, not a knowledge skill: a repeatable reading order,
a PR shape, a git minimum, and a list of questions whose answers you're allowed to ask on
day one (everyone asks them; nobody remembers who asked first). This lesson rehearses all
four on foreign code.

## What you'll be able to do

- You can map an unfamiliar Next.js repo in 30 minutes: `package.json` → routes → one vertical slice.
- You can trace and describe one vertical slice (URL → page → component → route handler → backend call) well enough to draw it for someone else.
- You can cut the *smallest reviewable PR*: one behavior, one commit stack, test included, PR description a reviewer can skim in 90 seconds.
- You can run the git loop you'll actually use — branch, commit, push, PR, address review — without fear.
- You can ask the ten day-one questions whose answers determine how fast you become useful.

## Principles

### Onboarding a foreign repo: package.json → routes → one slice

Resist the urge to read files alphabetically (it's how people drown). The order:

**1. `package.json` — five questions, two minutes.**

- Scripts: what does `dev` / `build` / `test` / `lint` actually run? (`test: "vitest"`
  tells you their test runner before you find a single test file.)
- Dependencies: React/Next versions; data layer (`@tanstack/react-query`? SWR? Apollo?);
  validation (`zod`?); UI kit (MUI? shadcn? Tailwind?); anything exotic is *their*
  convention you must learn.
- `packageManager` field: pnpm vs npm vs yarn — use theirs, exactly, or your lockfile
  becomes a PR incident.
- ESLint/Prettier config presence: their rules are law (Lesson 31).
- `engines`/`.nvmrc`: the Node version they run. Match it or debug ghosts.

**2. The route map — ten minutes.** `src/app/` *is* the sitemap. `ls` the folders: every
directory with a `page.tsx` is a URL; every `route.ts` is a BFF endpoint (Lesson 20).
Write the two lists down — pages and endpoints — before reading any code. Now you know the
app's *shape* (a dashboard with 6 pages and 11 endpoints is tractable; that sentence is
worth more than 40 file reads).

**3. One vertical slice — 20 minutes.** Pick the most trivial page. Read it top to bottom,
then *only* the things it imports, one layer deep: the page → the component it renders →
the hook or fetch it uses → the route handler → the backend call. Follow it with the
Network tab open while clicking the real page (Lesson 32). Now you understand **one
complete path end-to-end** — and every future feature you add will walk a path shaped
exactly like this. Depth on one slice beats breadth on thirty files; the team's conventions
live inside slices, not in READMEs.

While tracing, keep a "conventions notebook": where do they put shared types? Do all fetch
goes through one client wrapper? What do they name route folders? How do they do errors?
That notebook is your BFF-ownership foundation.

### The smallest reviewable PR

Reviewers are generous with small PRs and hostile to big ones — it's cognitive capacity,
not politeness. The shape:

- **One behavior.** "Add priority filter to task list." Not the filter, a refactor, and
  an icon swap. If the description needs an "and", split it.
- **A commit stack that tells a story**: `feat: add zod schema for filter params` →
  `feat: wire filter through /api/tasks` → `test: filter cases` → `docs: update README`.
  Each commit compiles; the reviewer can read it like paragraphs.
- **A test** — always less work than the feature (Lesson 34's patterns).
- **A PR description a stranger can act on**: what + why, out-of-scope line, how to test
  it ("open preview URL → filter by priority"), and one screenshot/gif if UI moved.
- **Preview deploy link** (Lesson 35) so review is clicking, not checking-out.

The meta-rule: *the reviewer's time is the scarce resource; every choice above spends less
of it.* Small PRs also protect you — a 40-line PR that's wrong costs a day, a 400-line PR
that's wrong costs a week and your calendar of shame.

### Git: the loop you'll actually use

Forget git's 150 commands; you need this loop, and the two escape hatches:

```bash
git checkout -b feat/task-filter      # 1. branch off a fresh main
# ...edit, then, repeatedly:
git status                            #    what changed?
git add src/app/api/tasks/route.ts    #    stage *this* coherent change (not `git add -A` reflexively)
git commit -m "feat: validate filter params with zod"   # 2. small, imperative-message commits
git push -u origin feat/task-filter   # 3. first push records the upstream
# 4. open the PR on GitHub (or `gh pr create`), link it in your team channel
# 5. review comes back → edit → commit → push again; the PR updates itself
```

Escape hatch 1 — you committed to the wrong branch or want to reshape history *on your
un-pushed branch*: `git reset --soft HEAD~1` un-commits but keeps changes staged. Nothing
git says can hurt you until you `push --force` someone else's branch — don't.

Escape hatch 2 — main moved under you: `git checkout main && git pull`, then
`git checkout feat/task-filter && git merge main` (or rebase if that's team law — ask
question #10). Conflicts in files you understand are fine; conflicts in files you don't
are a moment to ask, not to guess-delete lines.

Commit-message grammar that reads well on any team: `type: imperative sentence` —
`feat:`, `fix:`, `test:`, `docs:`, `refactor:`. "Add filter", never "Added filter" or
"filter stuff" (your capstone's `docs: feature spec` commit from Lesson 36 was rehearsal
for exactly this).

### The ten day-one questions

Ask these in your first week — they're the questions a staff engineer would ask; nobody
loses respect for them, and not asking them is how people build the wrong thing for two
weeks:

1. **State management?** Server state (TanStack Query? SWR? none?) vs client/UI state
   (useState? Context? Zustand/Redux?) — and the line between them.
2. **Data-fetching convention?** Do components fetch, or only through a hooks layer? Do
   Server Components fetch directly, or call internal endpoints?
3. **The env story?** Which env vars exist per environment, who holds prod values, and
   what's the "works in dev, broke in prod" history (Lesson 35)?
4. **CI?** What runs on a PR — lint, typecheck, tests, e2e? Is it red-blocking or advisory?
5. **Lint/format config?** Which rules bite, is Prettier enforced, and are there
   home-grown ESLint rules with opinions behind them?
6. **Who owns BFF contracts?** When the Go API changes, who updates the zod schemas —
   you (the BFF owner), the Go team, a shared `contracts/` package? What's the process?
7. **OpenAPI?** Does the Go backend publish a spec? Generate TS types from it, or are the
   hand-written schemas the source of truth — and who reconciles drift?
8. **Error reporting?** Where do prod errors surface (Sentry? logs?) and what error shape
   does the UI *actually* receive from the BFF (Lesson 28's normalization — is theirs honored)?
9. **Design system?** Component library, token source, do you restyle or reuse; a11y
   expectations.
10. **Review norms?** PR size ceiling, who reviews BFF changes, rebase vs squash-merge,
    branch naming, and the sacred one: "what's the first PR a new person should take?"

Questions 6 and 7 are *your* beat specifically — you're the person who will feel contract
drift first, because the zod boundary is where Go's truth lands on your side.

## Follow along

You'll onboard a real foreign repo. The recommended stand-in is the official learning
app: [https://github.com/vercel/next-learn](https://github.com/vercel/next-learn)
(Dashboard example) — public, moderate size, full of real conventions. Any open-source
Next.js repo you admire works equally well; the steps are identical.

1. **The two-minute pass.** Open `package.json` only. Write answers to the five questions
   (scripts, data layer, UI kit, package manager, lint). Expected: e.g. "tests run via
   `jest`, app uses Tailwind, npm as package manager" — *whatever theirs is, theirs wins*;
   you're reading for facts, not opinions.

2. **The ten-minute route map.** List every `page.tsx` path and every `route.ts` path
   (`search: page.tsx`, or `find src/app -name 'page.tsx'`). Expected output: two lists —
   URLs like `/dashboard`, `/dashboard/invoices`, and endpoints like `/api/...`. Write one
   sentence describing what the app does, using only those lists.

3. **The twenty-minute slice.** Pick the simplest page. Trace: page → imports → data call
   → (route handler or direct fetch — note *which*, it's a convention finding) → backend.
   Keep DevTools Network open against their live demo (or your local clone: `pnpm install
   && pnpm dev`¹). Write the slice as a five-line story: "Clicking /dashboard/invoices →
   server component → queries the DB through `lib/data.ts` → renders `<InvoicesTable>`."

4. **Cut the smallest PR (on YOUR capstone, against YOUR repo).** Back in
   `~/teampulse`, pick one 15-line improvement you postponed in Lesson 36 (better empty
   state copy; a missing `aria-label`; a stray `console.log`). Do the full loop:
   `git checkout -b fix/empty-state`, tiny commits with `type:` prefixes, push, open a
   real PR on GitHub with: description (what/why/how-to-test), out-of-scope line, preview
   deploy link, and the Lesson 34 test that covers it (add one if missing).
   Expected: a PR you'd be comfortable showing a reviewer — and a preview URL that renders
   the change.

5. **Interview yourself with the ten.** Answer all ten questions *for your own
   TeamPulse repo*, in writing, one line each. Any question you can't answer about a repo
   you built yourself is the exact question to ask the team about theirs.

## Your turn

1. Repeat the onboarding drill (steps 1–3, 35 minutes total, timer running) on a second foreign repo of your choice. Compare your convention notebook: name two conventions repo #2 has that TeamPulse lacks.
2. Take your PR from step 4 and make it *smaller than feels necessary*: split the description/test commit from the code commit, and rewrite the PR description in ≤ 60 words. Would a reviewer know what to click in the preview deploy? (Fix until yes.)
3. Your team's Go team adds `priority` values 4–5 next week with no notice. Using day-one questions #6/#7, what do you ask, and what one guard would you add to your zod schema today so the breakage is a loud, local 500-with-context instead of a silent `undefined` in the UI?

<details>
<summary>Solution</summary>

1. Personal; a strong answer names things like: they colocate tests in `__tests__/`
   instead of beside files, they centralize fetch in a typed `lib/api.ts` wrapper, they
   use path aliases differently, or they run e2e Playwright against preview deploys.
2. Done when a stranger can state the change, know exactly what to click, and know what's
   *not* included — in under 90 seconds. A reviewer asking "what should I look at?" means
   iterate again.
3. Ask: who owns the contract update, is there an OpenAPI spec we generate from, and how
   will we be notified of the change (changelog channel? PR review request?). Guard: make
   priority a strict union — `priority: z.union([z.literal(1), /* ... */ z.literal(3)])`
   (or `z.number().int().min(1).max(3)`) in the response schema, so an unexpected value
   fails your zod check loudly at the BFF boundary — logged with context, surfaced via
   your normalized error shape — rather than flowing into the UI as a surprise (Lesson 25).

</details>

## Checkpoint

1. You have 30 minutes with a foreign Next.js repo. Name the reading order and why the third step is one slice, not the whole `components/` folder.

<details>
<summary>Answer</summary>

package.json → route map (`page.tsx`/`route.ts` lists) → one vertical slice traced
end-to-end. One slice, because conventions (fetching, errors, types, naming) are fully
visible inside a complete path; a folder of components shows none of how they're wired.

</details>

2. What four things make a PR "reviewable" beyond passing CI?

<details>
<summary>Answer</summary>

One behavior only; a coherent commit stack; a test included; a skimmable description
(what/why, out-of-scope, how to test) with a preview-deploy link.

</details>

3. Translate: "stage this one change, record it with a story commit message, get it to the remote the first time." Which three commands?

<details>
<summary>Answer</summary>

`git add <file>` → `git commit -m "feat: ..."` (imperative, `type:` prefix) →
`git push -u origin <branch>`.

</details>

4. Which two of the ten day-one questions are most specific to the BFF owner, and why?

<details>
<summary>Answer</summary>

"Who owns BFF contracts?" and "Is there an OpenAPI spec / how do we catch drift?" — the
zod boundary between your route handlers and the Go service is *your* failure surface;
whoever ignores contract ownership breaks your layer silently (Lesson 25).

</details>

5. A teammate says "just `git add -A`, what's the difference?" Give the concrete failure
   this habit eventually produces.

<details>
<summary>Answer</summary>

Staging everything sweeps in unrelated edits — half-finished experiments, `.env.local`,
debug logs — which then get committed and reviewed (or leaked). Reviewers lose trust in
every subsequent commit because they can no longer assume a commit contains what its
message claims.

</details>

## Primary source

[GitHub Docs — About pull requests](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests).
The canonical description of the PR object itself — branches vs PRs, what review threads
attach to, and what "mergeable" means — the mental model under the whole smallest-PR
workflow.

## Before next session

Five minutes: re-read your five answers from step 5 (the ten questions, TeamPulse
edition) and mark the ones that felt like guesswork. Lesson 38's final quiz will probe
exactly those soft spots — and after that, the reference cheat sheets become your
permanent kit.

---

¹ npm alternative throughout foreign repos: always install with whatever `packageManager`/lockfile the repo dictates — never `pnpm install` inside an npm-locked repo just because it's your habit.
