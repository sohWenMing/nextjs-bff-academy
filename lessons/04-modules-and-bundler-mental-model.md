# Lesson 04 — Modules & the bundler mental model

*Phase 1 · Session 04 of 38 · ~2 hours · Prereqs: [Lesson 02](02-modern-js-syntax.md), [Lesson 03](03-arrays-objects-immutability.md)*

## Why this matters

A Next.js app is hundreds of `.ts`/`.tsx` files that import each other, plus
thousands of files from packages in `node_modules`. Nothing "just runs" — a
**bundler** (Next.js includes one) reads your import graph and produces
browser-ready JS. When the build fails with a module-resolution error, or you
wonder why `"use client"` and import direction matter in Phase 3, the mental
model you build today is what saves you. This lesson also makes you fluent in
`import`/`export` — the syntax on line 1 of every team file.

## What you'll be able to do

- You can write named and default exports and the matching imports, cold.
- You can explain what `node_modules` and a lockfile are and where they came from.
- You can trace a module's import graph by hand across three files.
- You can explain, in three sentences, what a bundler does to your files.
- You can predict whether an `import` path resolves (extension, folder, package).

## Principles

**Modules** are files that explicitly share names with each other. Without
modules, every top-level variable in your program would be one global
namespace — chaos. JS's official module system is **ES modules (ESM)**: a file
*exports* names, another file *imports* them. Nothing is shared unless
exported; every file has its own scope.

There are two export flavors. **Named exports** — the team default — let a
module export many things, imported by exact name inside `{}`:

```javascript
// format.mjs
export const currency = (n) => `$${n.toFixed(2)}`;
export const slugify = (s) => s.toLowerCase().replace(/\s+/g, "-");

// app.mjs
import { currency, slugify } from "./format.mjs";
```

**Default export** — one per file, arbitrary import name. You will see it in
React component files and Next.js special files (`page.tsx` must default-export
a component):

```javascript
// header.mjs
export default function Header() { return "TeamPulse"; }

// app.mjs
import Header from "./header.mjs";   // no braces — it's *the* default
```

**Resolution rules** (how a string after `from` becomes a file): paths
starting with `./` or `../` are *your* files, relative to the importing file.
A bare name like `react` or `zod` is looked up in `node_modules`. For local
files in Node ESM (`.mjs` scripts), the **extension is required** —
`import { x } from "./format.mjs"`, not `"./format"`. In bundler-land
(Next.js), extensions and folder-index resolution are added by the tool —
that difference trips up almost everyone once, so know which world you are in.

**`node_modules`** is the folder where `pnpm` physically places installed
packages (and *their* dependencies, recursively — it gets huge). A
**lockfile** (`pnpm-lock.yaml`) records the exact resolved version tree so
every machine gets identical installs; commit it. `package.json` says "we want
`zod@^3.23`", the lockfile says "we actually got 3.23.8" — that pairing is
why builds are reproducible.

**Installing a dependency** is `pnpm add <pkg>` (runtime code) or
`pnpm add -D <pkg>` (dev-only: build tools, type checkers, test runners —
tools not needed in production). Both write to `package.json`
(`dependencies` vs `devDependencies`) and update the lockfile.

**What a bundler actually does.** Next.js's bundler (Turbopack/Webpack under
the hood) takes your entry points, walks the import graph, and emits a few
optimised `.js` files for the browser. Concretely it: (1) resolves every
import string to a real file, (2) concatenates those files into *chunks*
(bundle output files) in dependency order, (3) rewrites the import/export
plumbing between them into direct references, (4) tree-shakes — drops exported
code nothing imports — and minifies. Your dozens of files become a handful of
chunks; your dozens of `node_modules` packages likewise. Two consequences to
internalize: **unused exports cost nothing in the browser** (tree-shaking),
and **the import graph is the app** — file X importing file Y means "Y is
part of X's world," which is exactly the mechanism Phase 3 uses to decide what
runs on the server vs what ships to the browser.

**Circular imports** (A imports B imports A) usually "work" until a value is
suddenly `undefined` at module-evaluation time. If you ever see that, look for
a cycle before looking anywhere else.

## Follow along

Build a tiny module graph in `~/learn-js/lesson-04`. Create the files by
typing them.

### 1. A leaf module

`data.mjs`:

```javascript
export const projects = [
  { id: "p1", name: "Billing Service", status: "active" },
  { id: "p2", name: "Mobile App", status: "active" },
  { id: "p3", name: "Legacy Portal", status: "archived" },
];
```

### 2. A module that imports and re-exports

`format.mjs`:

```javascript
import { projects } from "./data.mjs";

export const activeProjects = () =>
  projects.filter((p) => p.status === "active");

export const summaryLine = (p) => `${p.name} [${p.status}]`;

export default activeProjects; // both flavors, one file (legal, pick one in real code)
```

### 3. The entry point

`index.mjs`:

```javascript
import { activeProjects, summaryLine } from "./format.mjs";
import { projects } from "./data.mjs";

for (const p of activeProjects()) {
  console.log(summaryLine(p));
}
console.log(`${projects.length} total, ${activeProjects().length} active`);
```

### 4. Run it

```bash
node index.mjs
```

Expected:

```
Billing Service [active]
Mobile App [active]
3 total, 3 active
```

Trace the graph on paper: `index.mjs` → `format.mjs` → `data.mjs`, and
`index.mjs` → `data.mjs` directly. `data.mjs` is **evaluated once** —
`projects` is the same array object whichever path reaches it (modules are
singletons, cached after first execution). Verify with one appended line in
`index.mjs`:

```javascript
import { projects as p } from "./data.mjs";
console.log(p === projects); // true — same cached module, same reference
```

### 5. Install a real package

```bash
pnpm add date-fns   # a popular date utility library
```

Open `package.json`: `date-fns` now sits under `dependencies`. `pnpm-lock.yaml`
appeared. `node_modules/` exists. Create `dates.mjs`:

```javascript
import { formatDistanceToNow } from "date-fns";

const due = new Date("2026-12-01");
console.log(`Due ${formatDistanceToNow(due)} from now`);
```

```bash
node dates.mjs
```

Expected: `Due 1 month from now` (roughly — depends on today). Note the import
was the *bare* name `date-fns`: pnpm put it in `node_modules`, Node resolved
it there.

### 6. Break it on purpose — read a resolution error

In `index.mjs`, change one import to `from "./format"` (drop the extension) and
run `node index.mjs`. Expected error:

```
Error [ERR_MODULE_NOT_FOUND]: Cannot find module '.../format'
```

This is the classic "can't resolve" family of errors you will meet in every
bundler project. Restore the `.mjs` and confirm it runs again.

## Your turn

1. **Extract.** Move the task data and a new `groupByProject(tasks)` function
   out of a script you wrote in [Lesson 03](03-arrays-objects-immutability.md)
   into `tasks.mjs` (named exports); import both into a new `report.mjs` that
   prints per-project counts. Run it.

2. **Direction matters.** Create `cycle-a.mjs` and `cycle-b.mjs` that import
   each other (each logging a value the other exports). Run `node cycle-a.mjs`
   and write down what surprised you about the first `console.log`.

3. **Predict, then verify.** Without running: what does `import "./data.mjs"`
   (no names at all) do — load the file or skip it? Check by adding
   `console.log("data.mjs loaded")` at the top of `data.mjs` and using that
   side-effect-only import in a new file.

<details><summary>Solution 1</summary>

```javascript
// tasks.mjs
export const tasks = [
  { id: "t1", projectId: "p1", title: "Invoice PDF export", status: "doing" },
  { id: "t2", projectId: "p1", title: "Retry payment webhook", status: "todo" },
  { id: "t3", projectId: "p2", title: "Fix keyboard overlap", status: "done" },
];

export const groupByProject = (tasks) =>
  tasks.reduce((acc, t) => {
    acc[t.projectId] = (acc[t.projectId] ?? []).concat({ ...t });
    return acc;
  }, {});

// report.mjs
import { tasks, groupByProject } from "./tasks.mjs";

for (const [projectId, list] of Object.entries(groupByProject(tasks))) {
  console.log(`${projectId}: ${list.length} task(s)`);
}
// p1: 2 task(s)
// p2: 1 task(s)
```
</details>

<details><summary>Solution 2</summary>

```javascript
// cycle-a.mjs
import { b } from "./cycle-b.mjs";
export const a = "A";
console.log("a sees b =", b);

// cycle-b.mjs
import { a } from "./cycle-a.mjs";
export const b = "B";
console.log("b sees a =", a);
```

Running `node cycle-a.mjs` prints `b sees a = undefined` *before*
`a sees b = B`. Because execution starts at A, which immediately dives into B,
which asks A for `a` — but A hasn't executed its `export const a` line yet.
Modules run top-to-bottom, once; cycles expose ordering to you.
</details>

<details><summary>Solution 3</summary>

It loads and fully evaluates the file — the `{}` import executes it for
**side effects** only (registering something, printing, patching globals),
while importing zero names. You saw `data.mjs loaded` printed. Real use:
Polyfills and `import "dotenv/config"`.
</details>

## Checkpoint

1. Difference between `export const x = ...` + `import { x }` and
   `export default ...` + `import X`?
<details><summary>Answer</summary>
Named exports: many per module, imported by exact name in braces. Default: at
most one per module, imported with any name you choose, no braces. Teams
prefer named exports because editors can rename/refactor them reliably.
</details>

2. Where does Node look for `import "zod"` vs `import "./zod.mjs"`?
<details><summary>Answer</summary>
The bare name goes through package resolution: `node_modules` (starting near
the importing file, walking up). The `./` path is relative to the importing
file itself and must include the extension in Node ESM.
</details>

3. What is in `pnpm-lock.yaml` and why do you commit it?
<details><summary>Answer</summary>
The exact resolved version of every package in the tree (and integrity
hashes). Committing it makes every machine and CI install byte-identical
dependencies, so "works on my machine" stops happening.
</details>

4. Name four things the bundler does between your source files and the JS the
   browser downloads.
<details><summary>Answer</summary>
Resolves import strings to files; concatenates them into chunks; rewrites
module plumbing to direct references; tree-shakes unused exports and minifies.
</details>

5. Your future P0 CLI ([Lesson 09](09-p0-teampulse-cli.md)) will import a
   helper from `./api.mjs`. If you get `ERR_MODULE_NOT_FOUND`, name two
   plausible causes.
<details><summary>Answer</summary>
Missing/incorrect extension in the import path, wrong relative path (e.g.
`./` from the wrong folder), the file genuinely not created, or it lives under
a different name/case (macOS is often case-insensitive, Linux servers are not).
</details>

## Primary source

[Node.js — Modules: ES modules](https://nodejs.org/api/esm.html)
The official reference for exactly what we covered here: how `.mjs` resolution
works, why extensions are mandatory in Node ESM, and how module caching makes
every module a singleton. When import errors confuse you, this page has the
rules the error messages are built from.

## Before next session

In `~/learn-js/lesson-04`, split `index.mjs` so that every file has exactly one
job (data / formatting / entry) and no file has both export flavors. Then make
sure the mock backend is running (`cd backend-mock && go run main.go`) and
memorize the URL you will hit in [Lesson 05](05-async-promises-await-fetch.md)
with `curl -s http://localhost:8080/api/health`.
