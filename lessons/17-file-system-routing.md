# Lesson 17 — File-system routing

*Phase 3 · Session 17 of 38 · ~2 hours · Prereqs: Lesson(s) 10, 11, 12, 16*

## Why this matters

Your team's Next.js app — the frontend and the BFF — is organized entirely by **where files live**. There is no router config file, no route table, no `routes.jsx`. `app/projects/[id]/page.tsx` *is* the URL `/projects/42`. When you own the BFF layer, every endpoint you add is "create a folder, add a file" — and half your code-review value is knowing which file a bug lives in, which you can only do if folders-are-URLs is reflexive.

## What you'll be able to do

- You can predict the URL of any `page.tsx` in an `app/` folder, and create the folder for any URL, without looking anything up.
- You can write a `layout.tsx` that wraps every page in a folder, and explain what a "root layout" is.
- You can navigate between pages with `<Link>` and explain why you don't use `<a href>` for internal links.
- You can create a dynamic segment `[id]` and read `id` inside the page.
- You can add two new routes to `~/teampulse` and see each render in the browser.

## Principles

### The one rule: folders are URLs

Next.js's **App Router** (the only router this course teaches — never the old `pages/` router) maps the file tree under `app/` directly to URLs:

```text
app/
  page.tsx              →  /
  projects/
    page.tsx            →  /projects
    [id]/
      page.tsx          →  /projects/:id   (e.g. /projects/p1)
  settings/
    page.tsx            →  /settings
```

Three vocabulary terms, each in one clause:

- **Segment**: one path piece of a URL, matched by one folder (`projects` is the segment in `/projects`).
- **Route segment**: the folder *plus* the file inside it that Next.js looks for.
- **Page**: a file named exactly `page.tsx` that returns JSX — its presence makes the URL exist. No `page.tsx`, no URL (404).

That's it. You never register a route. Creating `app/reports/page.tsx` *is* registering `/reports`.

### `page.tsx` is just a component

A page is a normal React function component like the ones you wrote in [Lesson 11](11-components-and-props.md) — except Next.js calls it, and you export it as the file's default export:

```tsx
export default function ProjectsPage() {
  return <h1>Projects</h1>;
}
```

Name the function whatever you like; the **filename** is what matters. `export default` is the "this file's main thing" export from [Lesson 04](04-modules-and-bundler-mental-model.md) — Next.js imports the default and renders it.

### `layout.tsx` wraps every page below it

A `layout.tsx` in a folder renders *around* every page in that folder (and its subfolders). The layout receives your page as a prop named `children` — "the thing I wrap":

```tsx
// app/projects/layout.tsx
export default function ProjectsLayout({
  children,
}: {
  children: React.ReactNode; // ReactNode: "anything JSX-able"
}) {
  return (
    <div>
      <nav>
        <a href="/projects">All projects</a>
      </nav>
      {children}
    </div>
  );
}
```

Your app already has exactly one layout you never wrote yourself: `app/layout.tsx`, the **root layout**. It's required, it must render `<html>` and `<body>`, and everything in the app appears inside it.

### `<Link>` instead of `<a>` for internal navigation

You *can* use `<a href="/projects">`, but it forces a full page reload — the browser throws away all your JS and re-downloads everything (the same distinction as [Lesson 15](15-data-fetching-and-useeffect.md)'s "fetch on click vs fetch on render"). `next/link`'s `<Link>` swaps pages client-side: instant, and it prefetches the destination when it scrolls into view:

```tsx
import Link from "next/link";

<Link href="/projects/p1">Billing Service</Link>
```

Rule to memorize: **internal links use `<Link>`, external links (another domain entirely) use `<a>`.**

### Dynamic segments: `[id]`

A folder named with square brackets captures its URL piece as a parameter. `app/projects/[id]/page.tsx` matches `/projects/anything`, and Next.js hands you the captured value via the `params` prop. In Next.js 15, `params` is a **Promise** (an async value, as in [Lesson 05](05-async-promises-await-fetch.md)), so the page is `async` and you `await` it:

```tsx
// app/projects/[id]/page.tsx
export default async function ProjectPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <h1>Project {id}</h1>;
}
```

Why a Promise? Next.js may still be resolving routing (or revalidating) while the component starts running; awaiting is the price of the app router being async all the way down. The type `Promise<{ id: string }>` is exactly the generic-reading skill from [Lesson 08](08-typescript-iii-generics-and-utility-types.md).

### Two useful conventions you'll meet on the team

- **Private folders** start with an underscore — `_components/`, `_utils/`.
Next.js ignores them for routing, so you can keep components near the page that uses them without inventing URLs.
- **Route groups** wrap folders in parentheses — `(dashboard)/` — to share a
layout across routes *without* adding a URL segment.

## Follow along

Work in the app you built in [Lesson 16](16-p1-teampulse-explorer-page.md): `~/teampulse`. Keep `go run main.go` running in a separate tab so `http://localhost:8080` is up, and run the dev server in another:

```bash
cd ~/teampulse
pnpm dev
```

Expected: `Ready in ...` and `Local: http://localhost:3000`.

1. **Map what already exists.** Before changing anything, write down the URL
of every `page.tsx` in `app/`:

   ```bash
   find app -name "page.tsx"
   ```

Expected on a stock app: `app/page.tsx` → `/`. Check your guess in the browser before continuing.

2. **Create `/about`.** Create `app/about/page.tsx` (type it, don't paste):

   ```tsx
   export default function AboutPage() {
     return (
       <main>
         <h1>About TeamPulse</h1>
         <p>A team dashboard over a Go backend.</p>
       </main>
     );
   }
   ```

Visit `http://localhost:3000/about`. Expected: your heading, *and* the app's existing chrome (font, background) — because `app/layout.tsx` still wraps it.

3. **Add a layout to the folder.** Create `app/about/layout.tsx`:

   ```tsx
   export default function AboutLayout({
     children,
   }: {
     children: React.ReactNode;
   }) {
     return (
       <section style={{ padding: 24, border: "2px dashed #999" }}>
         {children}
       </section>
     );
   }
   ```

Reload `/about`. Expected: a dashed box around your content. Visit `/` — unchanged, because this layout only covers the `about/` folder.

4. **Link to it.** In `app/page.tsx`, add above the existing content:

   ```tsx
   import Link from "next/link";
   // ...inside your returned JSX:
   <p>
     <Link href="/about">About TeamPulse</Link>
   </p>
   ```

Click it from `http://localhost:3000`. Expected: the page swaps with no full reload — no flash, scroll preserved elsewhere, and hover the link to see the prefetch happen in the Network tab (look for a `?_rsc=` request; that's Next.js pre-fetching the page's rendered payload).

5. **Add a dynamic route.** Create `app/projects/[id]/page.tsx`:

   ```tsx
   export default async function ProjectPage({
     params,
   }: {
     params: Promise<{ id: string }>;
   }) {
     const { id } = await params;
     return (
       <main>
         <h1>Project {id}</h1>
       </main>
     );
   }
   ```

Visit `/projects/p1`, then `/projects/p2`, then `/projects/anything-at-all`. Expected: `Project p1`, `Project p2`, `Project anything-at-all` — the same file answers all of them, with `id` captured from the URL.

6. **Prove there's no route table.** Kill the dev server (`Ctrl-C`) and run:

   ```bash
   pnpm build
   ```

Expected: the build output prints a route list — `Route (app)` with `○ /`, `○ /about`, `ƒ /projects/[id]`. That table is *generated from your folders*, not configured by you. (`○` = static, `ƒ` = dynamic.)

## Your turn

1. Create `app/projects/page.tsx` that lists the three known project ids
(`p1`, `p2`, `p3` — from `backend-mock/main.go`) as `<Link>`s to `/projects/[id]`. Click one and confirm the dynamic page renders it.

   <details><summary>Solution</summary>

   ```tsx
   // app/projects/page.tsx
   import Link from "next/link";

   const IDS = ["p1", "p2", "p3"];

   export default function ProjectsPage() {
     return (
       <main>
         <h1>Projects</h1>
         <ul>
           {IDS.map((id) => (
             <li key={id}>
               <Link href={`/projects/${id}`}>{id}</Link>
             </li>
           ))}
         </ul>
       </main>
     );
   }
   ```

`/projects` shows the list; clicking `p2` renders `/projects/p2` via `app/projects/[id]/page.tsx`.
   </details>

2. Add a `app/projects/layout.tsx` with a heading that says `TeamPulse
Projects`, and verify both `/projects` and `/projects/p1` render inside it.

   <details><summary>Solution</summary>

   ```tsx
   // app/projects/layout.tsx
   export default function ProjectsLayout({
     children,
   }: {
     children: React.ReactNode;
   }) {
     return (
       <main>
         <h1>TeamPulse Projects</h1>
         {children}
       </main>
     );
   }
   ```

Both URLs now show `TeamPulse Projects` above their own `<h1>`. If your page files also have `<main>` tags, that's legal HTML-nesting noise for now — trim one to keep it clean.
   </details>

3. Harder: create `app/team/[name]/page.tsx` that renders `Hello, {name} —
role unknown` and a `<Link>` back to `/`. Then guess: what happens at `/team/` (trailing slash, no name)? Try it and explain.

   <details><summary>Solution</summary>

   ```tsx
   // app/team/[name]/page.tsx
   import Link from "next/link";

   export default async function TeamMemberPage({
     params,
   }: {
     params: Promise<{ name: string }>;
   }) {
     const { name } = await params;
     return (
       <main>
         <h1>Hello, {name} — role unknown</h1>
         <Link href="/">Back home</Link>
       </main>
     );
   }
   ```

`/team/` (no name) matches nothing — `[name]` requires exactly one segment — so Next.js serves the built-in 404 page. The URL space is exhaustive: every URL either matches a `page.tsx` or 404s.
   </details>

## Checkpoint

1. You need a page at URL `/tasks/archive`. What exact file do you create, and
what must it export?

   <details><summary>Answer</summary>
`app/tasks/archive/page.tsx` with `export default` of a React component. No registration anywhere.
   </details>

2. Why does Next.js 15 make `params` a Promise, and what does that force your
page component to become?

   <details><summary>Answer</summary>
Routing/revalidation data may not be ready when the component starts running, so params arrive asynchronously. The page becomes an `async` function and you `const { id } = await params;`.
   </details>

3. What breaks (functionally, not stylistically) if you use `<a href="/projects">`
instead of `<Link href="/projects">` for internal navigation?

   <details><summary>Answer</summary>
A full browser reload: client state is lost, all JS re-downloads and re-executes, and there's no prefetch-on-hover. The page still loads, but like a hard navigation.
   </details>

4. Against the mock Go API: your teammate says `http://localhost:3000/api/login`
must be hitting the Go backend because it returns JSON. Using only what you know about routing, why is that claim not yet provable, and what file would prove it either way?

   <details><summary>Answer</summary>
A URL returning JSON tells you nothing about *which server* answered — :3000 is Next.js, :8080 is Go. Only the file tree tells you what Next.js serves: if `app/api/login/` has no route/page file, Next.js 404s and the request must have gone elsewhere. You'd look for `app/api/login/route.ts` (Lesson 20's subject) to prove Next.js served it.
   </details>

5. You create `app/_data/page.tsx`. What URL serves it, and why?

   <details><summary>Answer</summary>
None — folders starting with `_` are private: Next.js excludes them from routing entirely, so `/_data` 404s. It's a place to keep code near a page without exposing URLs.
   </details>

## Primary source

[Next.js Docs — Layouts and Pages](https://nextjs.org/docs/app/getting-started/layouts-and-pages) — the official page that defines the segments/pages/layouts model in the exact terms (and exact Next.js 15 async-`params` shape) this lesson uses.

## Before next session

Open your `~/teampulse` app and add this one line to the very top of `app/page.tsx`, above the imports: `"use client";`. Save, reload, and note in one sentence what the dev-server output says differently (any error or warning is fine — bring it; Lesson 18 is about exactly what that line means).
