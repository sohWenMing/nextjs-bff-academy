# Lesson 01 — Toolchain: Node, pnpm, your first scripts

*Phase 1 · Session 01 of 38 · ~2 hours · Prereqs: none*

## Why this matters

Your team's frontend and BFF are TypeScript files that must be *run* and *built*
by tools on your machine before any of it becomes a website. Before you can read
a single line of the team's Next.js code, you need to be fluent in the layer
underneath it: the Node.js runtime, the `pnpm` package manager, and the
`package.json` file that every JS project starts with. Every command your team
runs — `pnpm dev`, `pnpm build`, `pnpm test` — is exactly the mechanism you build
today.

## What you'll be able to do

- You can verify which version of Node is installed and explain, in one
  sentence, what Node.js actually is.
- You can enable `pnpm` via Corepack and explain what a package manager does.
- You can write a `.mjs` script and run it with `node`, without looking anything up.
- You can read a `package.json` and explain every field in it.
- You can add a `scripts` entry and run it with `pnpm run <name>`.

## Principles

**JavaScript (JS)** is the programming language that runs in web browsers. It
also runs outside browsers, on servers and laptops, thanks to **Node.js** — a
program that embeds the same JS engine that Chrome uses (V8) and adds
operating-system abilities the browser deliberately withholds: reading files,
opening network ports, spawning processes. Node is a *runtime*: a program that
executes JS code. It is not a language and not a framework.

```bash
node hello.mjs   # "run hello.mjs with the Node runtime"
```

**ES modules** are JavaScript's official module system (module details are
[Lesson 04](04-modules-and-bundler-mental-model.md)). A file ending in `.mjs`
is explicitly an ES module — modern import/export syntax, no configuration
needed. That is the extension we use for throwaway scripts in this lesson.

**A package manager** installs other people's published code (libraries, also
called *packages* or *dependencies*) into your project and tracks which versions
you use. **pnpm** is the package manager this course uses; it is faster and
uses less disk than the historically default npm (see footnote[^npm] once,
then we never use npm again). **Corepack** is a small manager-of-package-managers
that ships *inside* Node itself: `corepack enable` creates the `pnpm` command
on your PATH and keeps its version pinned per project.

**`package.json`** is a project's identity card and instruction manual, in JSON
(JavaScript Object Notation — data syntax: key/value pairs, objects in `{}`,
arrays in `[]`). It lists the project name, its dependencies, and its
**scripts** — named shell commands that anyone on the team can run without
memorizing them. When you later type `pnpm dev` in a Next.js repo, you are
running the script literally named `"dev"` from that repo's `package.json`.

[^npm]: npm is Node's original package manager and still perfectly functional;
`npm install`, `npm init -y`, `npm run dev` are the equivalent commands. This
course standardizes on pnpm because your team does.

## Follow along

Open Terminal. Work in a fresh folder so nothing collides with other projects.

### 1. Verify Node (install via Homebrew if missing)

```bash
node --version
```

Expected: something like `v22.14.0`. You need **v20 or newer** (the course
assumes v20+ features). If `node` is not found, install it with Homebrew
(macOS's package installer for command-line tools):

```bash
brew install node
node --version
```

`--version` is a *flag* (an option that changes how a command behaves). Node
prints its version and exits.

### 2. Enable pnpm via Corepack

```bash
corepack enable
pnpm --version
```

Expected: a version like `9.x.x`. If Corepack asks "Do you want to continue?
[Y/n]" type `y`. If `pnpm` is still not found, nudge Corepack to fetch it:

```bash
corepack prepare pnpm@latest --activate
pnpm --version
```

### 3. Create a project folder and your first script

```bash
mkdir -p ~/learn-js/lesson-01
cd ~/learn-js/lesson-01
```

Create `hello.mjs` with your editor (VS Code, or `nano hello.mjs`):

```javascript
const name = "Wen";
console.log(`Hello, ${name}! The runtime is Node.`);
```

Run it:

```bash
node hello.mjs
```

Expected output:

```
Hello, Wen! The runtime is Node.
```

`console.log` prints to the terminal (the "console"). The backtick string is a
*template literal* — `${name}` is substituted with the variable's value
(properly dissected in [Lesson 02](02-modern-js-syntax.md)).

### 4. Read command-line arguments

Create `greet.mjs`:

```javascript
const who = process.argv[2] ?? "stranger";
console.log(`Hello, ${who}!`);
```

`process` is a global object Node provides about the running program;
`process.argv` is its array of command-line arguments (`argv[0]` and `argv[1]`
are Node and the script path, so *your* arguments start at index `2`). `??` is
the nullish-coalescing operator: "use the right side if the left side is
missing". Run both variants:

```bash
node greet.mjs
node greet.mjs Wen
```

Expected: `Hello, stranger!` then `Hello, Wen!`.

### 5. Create a `package.json`

```bash
pnpm init
```

This asks a few questions (name, version...) and writes `package.json`. Open
it. You will see:

```json
{
  "name": "lesson-01",
  "version": "1.0.0",
  "description": "",
  "main": "index.js",
  "scripts": {
    "test": "echo 'Error: no test specified' && exit 1"
  }
}
```

`name` and `version` identify the package. `main` names the default entry file
(we will not use it). `scripts` is where the action is.

### 6. Add scripts and run them

Edit `package.json` so the `scripts` block reads:

```json
"scripts": {
  "start": "node hello.mjs",
  "greet": "node greet.mjs"
}
```

Run:

```bash
pnpm start
pnpm run greet
```

Expected: your two greetings appear. `pnpm start` works for the two magic names
`start` and `test`; everything else needs the `run` keyword. You just executed
the same machinery that powers `pnpm dev` in every Next.js repo you will ever open.

### 7. (Optional peek) What Corepack did

```bash
which pnpm
```

Expected: a path inside your Node installation (e.g. under
`/opt/homebrew/bin`) — Corepack dropped a shim there that launches the pinned
pnpm. No global `pnpm install` was needed.

## Your turn

1. **Warm-up.** Write `sum.mjs` that takes two numbers from `process.argv`,
   adds them, and prints `Sum: 42`. Run it as
   `node sum.mjs 20 22`. (Gotcha you will discover: argv values are *strings*.)

2. **Script muscle.** In `package.json`, add a script `"sum": "node sum.mjs"`
   and run `pnpm run sum 3 4`. It fails — find out why and fix it.

3. **Transfer.** Write `env.mjs` that prints `process.platform` and
   `process.cwd()`. Predict the output before running it, then check. What do
   you think `cwd` stands for?

<details><summary>Solution 1</summary>

```javascript
const a = Number(process.argv[2]);
const b = Number(process.argv[3]);
console.log(`Sum: ${a + b}`);
```

Without `Number(...)`, `"20" + "22"` is string concatenation → `Sum: 2022`.
JS `+` adds numbers but *joins* strings — you will meet this again.
</details>

<details><summary>Solution 2</summary>

`pnpm run sum 3 4` runs `node sum.mjs` — the extra `3 4` are **not** passed on
unless you add `--`:

```json
"sum": "node sum.mjs"
```

```bash
pnpm run sum -- 3 4
```

Expected: `Sum: 7`. `--` tells pnpm "everything after this is for the script".
</details>

<details><summary>Solution 3</summary>

```javascript
console.log(process.platform); // "darwin" on Apple Silicon/Intel Macs
console.log(process.cwd());    // "/Users/<you>/learn-js/lesson-01"
```

`cwd` = current working directory — the folder you were "in" when you ran
Node. Same values a shell's `uname`/`pwd` would show.
</details>

## Checkpoint

1. What is Node.js, in one sentence, and how is it different from a browser?
<details><summary>Answer</summary>
Node.js is a runtime that executes JavaScript outside the browser, embedding
the same engine Chrome uses and adding OS access (files, network, processes)
that browsers withhold for safety.
</details>

2. What does `corepack enable` do, and why does this course prefer pnpm over
   npm?
<details><summary>Answer</summary>
Corepack is bundled with Node; `corepack enable` puts a `pnpm` shim on your
PATH that runs a per-project-pinned pnpm without a separate global install.
pnpm is faster and more disk-efficient, and the team standardizes on it —
consistency matters more than speed.
</details>

3. Name every top-level field you saw in your `package.json` and what it is for.
<details><summary>Answer</summary>
`name` (package name), `version` (semver version string), `description` (free
text), `main` (default entry file for consumers), `scripts` (named commands run
via `pnpm run <name>`).
</details>

4. Why does `pnpm run greet Wen` not pass `Wen` to the script, and what is the
   fix?
<details><summary>Answer</summary>
pnpm swallows arguments after the script name. Add `--` before the extra
arguments: `pnpm run greet -- Wen`.
</details>

5. You will run the TeamPulse mock Go backend from `backend-mock/` for most of
   this course. What command starts it, and what URL does it listen on?
<details><summary>Answer</summary>
`cd backend-mock && go run main.go` — it listens on
`http://localhost:8080` ("localhost" means this machine; 8080 is the TCP port).
</details>

## Primary source

[Node.js — Getting Started guides](https://nodejs.org/en/learn/getting-started/introduction-to-nodejs)
This is the official Node documentation's front door: it defines the runtime,
the module system, and `npm`/`package.json` terminology in the exact words the
ecosystem uses, so you can quote it back at teammates.

## Before next session

Run `node --version` and `pnpm --version` and write both numbers in a file
`~/learn-js/versions.txt`. Then `cd backend-mock && go run main.go` in one
terminal tab, open a second tab, run
`curl -s http://localhost:8080/api/health`, and note the JSON that comes back.
Leave the server running for Lesson 02's warm-up if you like — you will not
need it again until [Lesson 05](05-async-promises-await-fetch.md).
