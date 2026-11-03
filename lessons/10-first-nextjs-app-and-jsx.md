# Lesson 10 — Your First Next.js App & JSX

*Phase 2 · Session 10 of 38 · ~2 hours · Prereqs: Lessons 01, 02, 04*

## Why this matters

Your team's entire frontend is one Next.js App Router codebase. Before you can
read a single business-logic file, you need two things cold: what the
`create-next-app` scaffold actually put on your disk, and what JSX really is
(a function call, not HTML). Every `.tsx` file you'll ever open in the BFF
repo — route handlers, pages, components — is JSX plus TypeScript. This lesson
also creates the app you'll grow for the next six sessions: **TeamPulse** in
`~/teampulse`.

## What you'll be able to do

- You can generate a fresh Next.js app with pnpm and name every top-level file
  the scaffold created and what it does.
- You can write JSX by hand — elements, `className`, interpolation with `{}`,
  attributes — without reaching for HTML habits (`class`, multiple roots).
- You can explain, in one sentence, what a `.tsx` file compiles down to.
- You can run the dev server, edit a file, and explain why the browser updated
  without a full page reload.
- You can state which of your `create-next-app` answers you chose and why.

## Principles

### Next.js in one paragraph

Next.js is a framework on top of React (the UI library) that adds the things
every real app needs: file-based routing, a dev server with hot reload, a
build system, and a server runtime. Your team runs the **App Router** — the
modern routing system where folders map to URLs. There is an older `pages/`
router; you will never see it taught here. Next.js 15+ means App Router only.

A *component* (the term you'll meet all week) is a JavaScript function that
returns UI. Next.js gives you the project skeleton; React gives you the
component model; TypeScript gives you the type checking you learned in
Lessons 06–08. Three tools, one repo.

### JSX is a function call wearing a costume

This is the claim of the lesson. When you write this in a `.tsx` file:

```tsx
const name = "Wen";
const el = <h1 className="greeting">Hello {name}</h1>;
```

the compiler rewrites it to something like this before anything runs:

```ts
const el = jsx("h1", { className: "greeting", children: "Hello " + name });
```

`jsx(...)` is a function call that produces a plain JavaScript object — a
*React element*: an object describing *what* should be on screen, not the
screen itself. React walks those objects later and updates the real DOM
(Document Object Model — the browser's live tree of elements). Three
consequences follow directly, and they explain every JSX "rule":

1. **It's JavaScript, so `{}` holds any *expression*.** `{name}`, `{2 + 2}`,
   `{user.isAdmin ? "Admin" : "Member"}` all work. Statements (`if`, `for`,
   `const x = ...`) do **not** go inside `{}` — expressions produce a value,
   statements don't. You already know this distinction from
   [Lesson 02](02-modern-js-syntax.md) if you've internalized it; if not,
   memorize: *inside braces, things that have a value*.
2. **Attributes are object properties, so they're camelCased.** HTML's `class`
   collides with the `class` keyword, so JSX uses `className`. `tabindex`
   becomes `tabIndex`. The mapping is regular: HTML attribute → camelCase
   property.
3. **The function call returns one value, so there is one root.** A component
   must return a single element. When you need "several things with no
   wrapper", use a fragment: `<>...</>` — shorthand for `<React.Fragment>`,
   which renders nothing itself.

### Rules you'll break first (from muscle memory)

| HTML habit | JSX says | Why |
|---|---|---|
| `<div class="a">` | `<div className="a">` | attribute = object property |
| `<input value={x}>` with no onChange | add `onChange`, or use `defaultValue` | you'll meet this in Lesson 14 |
| text next to elements at top level | wrap in one root or `<>...</>` | one return value |
| `<p>if (x) ...</p>` | `<p>{x ? "a" : "b"}</p>` | expressions only |

Comments inside JSX use `{/* like this */}` because, again, you're inside a
function call's arguments, not inside a markup parser.

### Why a file is `.tsx`, not `.html`

A `.tsx` file is TypeScript (Lesson 06) plus JSX syntax. The `x` means
"this file may contain JSX". Next.js compiles it to plain JavaScript. Nothing
JSX-related exists at runtime in the browser — just function calls and
objects, exactly like the `jsx(...)` line above.

### What the dev server does when you save

`pnpm dev` starts a Node process that watches your files. On save, it
recompiles just what changed and pushes the update to the browser over a
websocket — that's *hot reload*: your React tree re-renders with the new code,
no full page refresh, and (usually) your state survives. This is why you
should **keep the dev server running the entire course** and never click
"reload" reflexively: hot reload is your feedback loop.

## Follow along

### 1. Generate the app

Leave the mock Go backend running if it's up (`cd backend-mock && go run
main.go` in another tab — you won't touch it until Lesson 15, but trust it).
Then:

```bash
cd ~
pnpm create next-app@latest teampulse
```

The prompts and the answers to choose (record them in your notes — your team's
scaffold may differ, and knowing the difference is the point):

| Prompt | Choose | Why |
|---|---|---|
| TypeScript? | **Yes** | the team's frontend is TS |
| ESLint? | **Yes** | catches real bugs; Lesson 31 |
| Tailwind CSS? | **Yes (team default)** | keep team default; we won't style much, but dropping it fights the scaffold |
| `src/` directory? | **No** | flatter tree for learning; many teams say Yes |
| Use App Router? | **Yes** | the only router this course teaches |
| Turbopack for next dev? | **Yes** | faster dev builds |
| Customize import alias (`@/*`)? | **No** (accept `@/`) | `@/components/x` resolves to `./components/x` |

Expected: `Creating teampulse in /Users/<you>/teampulse...` then `Success!`
and pnpm's install summary. If you ever need the npm version:
`npm create next-app@latest teampulse`.¹

### 2. Read the output

```bash
cd ~/teampulse && ls
```

Expected: `app/`, `public/`, `eslint.config.mjs`, `next.config.ts`,
`next-env.d.ts`, `package.json`, `pnpm-lock.yaml`, `postcss.config.mjs`,
`tsconfig.json`, plus git files. Know these four by heart:

- `app/` — **the routes**. `app/page.tsx` is the `/` URL. A folder in `app/`
  is a URL segment (Lesson 17 makes this official).
- `app/layout.tsx` — the shell around every page (imports fonts, wraps the app
  in `<html>`/`<body>`). Every page renders *inside* it.
- `package.json` — note the scripts: `dev`, `build`, `start`, `lint`.
- `next.config.ts` — build/runtime config. You'll open it in Lesson 29.

Open `app/page.tsx`. Everything between `return (` and `)` is one big JSX
expression — one root element, nested children, `{}` interpolations.
`app/page.tsx` exports a function named `Page`:

```tsx
export default function Page() {
  return (
    <main className="...">
      <h1 className="...">Get started by editing&nbsp;<code>app/page.tsx</code></h1>
      ...
    </main>
  );
}
```

That function *is* the page. Next.js calls it; its return value — React
elements — becomes the screen.

### 3. Run it

```bash
pnpm dev
```

Expected: `Ready in ...`, `Local: http://localhost:3000`. Open that URL.
Keep this terminal parked.

### 4. Feel the edit loop

Open `app/page.tsx`, replace the entire return with:

```tsx
  const project = { name: "TeamPulse", openTasks: 4 };
  return (
    <>
      <h1>{project.name}</h1>
      <p className="note">Open tasks: {project.openTasks * 2 / 2}</p>
      <p>{project.openTasks > 0 ? "Work to do" : "All clear"}</p>
    </>
  );
```

Save. The browser updates on its own. Expected on screen:

```
TeamPulse
Open tasks: 4
Work to do
```

You just wrote an expression inside `{}` (`project.openTasks * 2 / 2`), a
fragment (`<>`), and a ternary in JSX. That's the whole syntax you need for
today.

### 5. Prove JSX is a function call

Open `node_modules/react/package.json` and note the version — React 19.
Now look at the bottom of `app/page.tsx`'s compiled sibling: run

```bash
grep -r "jsx-runtime" .next/ 2>/dev/null | head -3
```

Expected: several hits — your JSX compiled into imports of `react/jsx-runtime`
and `jsx(...)` calls. The costume, removed.

### 6. Meet StrictMode

Open `app/layout.tsx` and find `<React.StrictMode>`. In development only,
StrictMode deliberately calls your component functions **twice** to expose
impure rendering. Expect double console.logs in dev — it's a feature, and
Lesson 13 uses it to teach what a "render" is.

## Your turn

1. **Warm-up:** add a `<ul>` below your `<p>` tags that renders `<li>Get
   started</li>`, `<li>Ship P1</li>`, `<li>Ship P2</li>` — written out
   literally (no `.map()` yet; that's Lesson 12). Expected: a bullet list
   under the ternary line.

2. **Break it on purpose:** put a raw `<div>` and `<span>` side by side as
   two roots in a `return`. Read the error. Now "fix" it with a fragment, and
   then again with a wrapper `<div>`. In one sentence each: what did the
   compiler complain about, and why do both fixes work?

3. **TypeScript + JSX:** write a small function in `page.tsx`:

   ```tsx
   function label(priority: number): string {
     return priority === 1 ? "high" : priority === 2 ? "medium" : "low";
   }
   ```

   Render `<p>Priority: {label(2)}</p>`. Then call `label("high")` and note
   the exact squiggle the TS checker gives you in your editor (or via `pnpm
   lint`/terminal red output). First line only — Lesson 31's habit starts now.

<details><summary>Solution</summary>

1. 

   ```tsx
   <ul>
     <li>Get started</li>
     <li>Ship P1</li>
     <li>Ship P2</li>
   </ul>
   ```

2. Error: *"JSX expressions must have one parent element."* A return value is
   one function-call result, so it must be one object. `<></>` wraps without
   adding DOM; `<div>` adds a real `<div>` to the page — same fix from the
   compiler's view, different output in the browser.

3. `label("high")` → `Argument of type 'string' is not assignable to parameter
   of type 'number'`. The JSX compiled fine; the TS checker rejected the call
   before React ever ran.

</details>

## Checkpoint

Answer cold. If any wobbles, redo Follow along step 4 before continuing.

1. What is a React element, in one sentence, and what is `className` a
   symptom of?
2. Why can't you put `if (x) { ... }` directly between `<p>` and `</p>`?
   What do you write instead for a yes/no string?
3. You return two sibling `<p>` tags with no wrapper. What error do you get
   and what are the two fixes — and how do they differ in the browser?
4. What does `pnpm dev` do with your file save, and why does the browser
   update without a reload?
5. The mock Go API at `http://localhost:8080` returns
   `{"error":"missing or invalid token"}` when you GET `/api/projects` with
   no `Authorization` header. Nothing about that message is JSX — so where
   does that string eventually get *rendered* in your future app, and what
   file type will the component be?

<details><summary>Answers</summary>

1. A plain JS object describing UI (tag + props + children), produced by the
   `jsx(...)` call your JSX compiled into; `className` exists because JSX
   attributes are object properties, and `class` is a reserved word.
2. Only *expressions* (value-producing things) go in `{}`; `if` is a
   statement. Write `{x ? "yes" : "no"}`.
3. "JSX expressions must have one parent element." Wrap in `<>...</>`
   (renders nothing extra) or a `<div>` (adds a real DOM node).
4. It watches files, recompiles the changed modules, and pushes the update
   over a websocket; React re-renders in place — hot reload.
5. Inside a React component's return value — in a `.tsx` file (specifically
   the error UI you'll build in Lessons 15 and 16).

</details>

## Primary source

[react.dev — Writing Markup with JSX](https://react.dev/learn/writing-markup-with-jsx)
— the official React docs make the "JSX is a function call, with exactly the
consequences we listed" argument in five minutes, and everything else this
week builds on that framing.

¹ npm works (`npm create next-app@latest`), but the course standardizes on
pnpm because your team does — Lesson 01 covers why.

## Before next session

In `app/page.tsx`, create one more tiny component above `Page`:
`function StatusBadge() { return <span>[ok]</span>; }` and render
`<StatusBadge />` in the page. Note in your notes: what did you have to do
differently from calling a normal function? (Bring the answer; Lesson 11
starts there.)
