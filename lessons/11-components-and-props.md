# Lesson 11 — Components & Props

*Phase 2 · Session 11 of 38 · ~2 hours · Prereqs: Lessons 02, 06, 10*

## Why this matters

Every file in your team's `components/` folder is one thing: a function that
takes an object in and returns UI out. Once that's boringly obvious, the props
type tells you what a component can display — and props are how the BFF's
reshaped data (Lesson 20) will flow into your screens.

## What you'll be able to do

- You can write a typed function component from scratch: `interface` for
  props, destructured parameter, `return` JSX.
- You can explain "a component is a pure function of its props" and name two
  ways to violate it.
- You can pass objects as props and render their fields (`Project`, `Task`,
  [Lesson 07](07-typescript-ii-shapes-and-narrowing.md)); use `children` for wrappers.
- You can use `children` to build a wrapper component.

## Principles

### A component is a function. The capital letter is the whole trick.

You discovered the puzzle at the end of Lesson 10: you rendered
`<StatusBadge />` without calling `StatusBadge()`. What React does with a JSX
tag:

- **lowercase tag** (`<div>`, `<p>`) → compiles to `jsx("div", ...)` — a
  built-in DOM element.
- **capitalized tag** (`<StatusBadge />`) → compiles to
  `jsx(StatusBadge, { ...props })` — React *stores a reference* to your
  function and calls it later, in its own render pass.

So `<StatusBadge />` isn't "calling a function now" — it hands React the
function *reference*; React decides when to call it. The capital letter is
purely that compile-time switch.

### Props are the function's one parameter — and they're read-only

```tsx
interface TaskRowProps {
  title: string;
  status: "todo" | "doing" | "done";
  priority: number;
}

function TaskRow({ title, status, priority }: TaskRowProps) {
  return (
    <li>
      <strong>{title}</strong> — {status} (P{priority})
    </li>
  );
}
```

Three facts to hold:

1. `TaskRowProps` is a normal TS interface (Lesson 07), named
   `ComponentNameProps` by convention. The compiler checks every prop at
   *build time*, before the browser ever opens.
2. Props arrive as **one object**, so you destructure in the parameter list —
   plain [Lesson 02](02-modern-js-syntax.md) destructuring, nothing React-specific.
3. **You never mutate props.** `status = "done"` in the body is forbidden —
   not by linting taste but by the model: render must be a pure function of
   inputs. Same immutability rule as
   [Lesson 03](03-arrays-objects-immutability.md), now with teeth: React
   decides re-renders partly by comparing references, and mutation leaves
   references untouched — silent staleness follows.

"Pure function of its props" means: same props in → same JSX out, no side
effects. The escape hatch — components that *also* remember things — is
state, Lesson 13.

### Composition and `children`

The React idiom is many tiny components composed together: `page.tsx`
doesn't render a task board; it renders `<TaskList />`, which renders many
`<TaskRow />`. The tool for "component that wraps other components" is the
special prop `children` — whatever sits *between* the open and close tags:

```tsx
interface CardProps {
  title: string;
  children: React.ReactNode; // ReactNode = "anything renderable"
}

function Card({ title, children }: CardProps) {
  return (
    <section>
      <h2>{title}</h2>
      {children}
    </section>
  );
}

// usage
<Card title="Active projects">
  <TaskRow title="Fix login overlap" status="done" priority={1} />
</Card>
```

### Props all the way down

Data flows **one direction: parent → child**, via props. A child cannot edit a
prop; it asks the parent to change the source of truth (Lesson 13's story).
Find where a value is born by searching for `<ComponentName`.

## Follow along

All work in `~/teampulse`, dev server from Lesson 10 running (`pnpm dev`).
The mock Go backend is *not* needed today — all data is hardcoded.

### 1. The data model, as TS

Create `lib/types.ts` (`lib/` is the course convention for non-component
code):

```ts
export interface Project {
  id: string;
  name: string;
  description: string;
  status: "active" | "archived";
  createdAt: string;
}

export interface Task {
  id: string;
  projectId: string;
  title: string;
  status: "todo" | "doing" | "done";
  assignee: string;
  priority: number;
  dueDate: string;
}
```

These mirror the Go structs documented in `backend-mock/README.md` — the
exact shapes the API returns. `status` is a union of string literals
(Lesson 07): `status: "wip"` will be a compile error, which is the point.

### 2. Fake data

Create `lib/data.ts`:

```ts
import type { Project, Task } from "@/lib/types";

export const projects: Project[] = [
  { id: "p1", name: "Checkout v2", description: "Rewrite checkout flow",
    status: "active", createdAt: "2026-09-01T00:00:00Z" },
  { id: "p2", name: "Mobile App", description: "iOS + Android shells",
    status: "active", createdAt: "2026-09-15T00:00:00Z" },
];

export const tasks: Task[] = [
  { id: "t1", projectId: "p1", title: "Add retry to payment webhook",
    status: "todo", assignee: "Wen", priority: 2, dueDate: "2026-11-25T00:00:00Z" },
  { id: "t2", projectId: "p1", title: "Load test checkout endpoint",
    status: "todo", assignee: "Sarah", priority: 2, dueDate: "2026-12-01T00:00:00Z" },
  { id: "t3", projectId: "p2", title: "Fix login screen keyboard overlap",
    status: "done", assignee: "Wen", priority: 1, dueDate: "2026-11-01T00:00:00Z" },
];
```

Note `import type` — type-only imports are erased at build time (Lesson 08);
no runtime cost.

### 3. First real components

Create `components/task-row.tsx` (course convention: kebab-case filenames,
capitalized component names):

```tsx
interface TaskRowProps {
  task: { title: string; status: "todo" | "doing" | "done"; priority: number };
}

export function TaskRow({ task }: TaskRowProps) {
  return (
    <li>
      [{task.status}] {task.title} — P{task.priority}
    </li>
  );
}
```

Create `components/project-card.tsx`:

```tsx
import type { Project } from "@/lib/types";

export function ProjectCard({ project, taskCount }: { project: Project; taskCount: number }) {
  return (
    <article>
      <h3>{project.name}</h3>
      <p>{project.description}</p>
      <p>
        {taskCount} task{taskCount === 1 ? "" : "s"} · {project.status}
      </p>
    </article>
  );
}
```

Expected at every save: clean recompile, no red overlay.

### 4. Compose the page

Replace `app/page.tsx`:

```tsx
import { ProjectCard } from "@/components/project-card";
import { TaskRow } from "@/components/task-row";
import { projects, tasks } from "@/lib/data";

export default function Home() {
  return (
    <main>
      <h1>TeamPulse</h1>
      <section>
        <h2>Projects</h2>
        {projects.map((p) => (
          <ProjectCard
            key={p.id}
            project={p}
            taskCount={tasks.filter((t) => t.projectId === p.id).length}
          />
        ))}
      </section>
      <section>
        <h2>All tasks</h2>
        <ul>
          {tasks.map((t) => (
            <TaskRow key={t.id} task={t} />
          ))}
        </ul>
      </section>
    </main>
  );
}
```

You just used `.map()` (Lesson 03) to turn an array into an array of React
elements, and passed `key` — React's identity marker for list items. Its
rules are Lesson 12's headline; for now, use the item's `id` and move on.

Expected browser output: two project cards with correct task counts (2 and
1) and a list of all three tasks with status and priority.

### 5. Prove props are checked

Change `<TaskRow key={t.id} task={t} />` to `<TaskRow key={t.id} />`: red
build error naming the missing prop. Put it back.

### 6. A wrapper with children

Write the `Card` component from Principles into `components/card.tsx` and wrap
the `page.tsx` output in `<Card title="Dashboard">…</Card>`. Expected: one
framed panel, inner content untouched.

## Your turn

1. **Warm-up:** add an `AssigneeTag` component that takes
   `assignee: string` and renders `<span>@{assignee}</span>`; use it inside
   `TaskRow` (add `assignee` to the prop type — `Task` already has it).
   Expected: each task line shows `@Wen` or `@Sarah`.

2. **Transfer:** create `components/status-badge.tsx` with
   `interface StatusBadgeProps { status: "todo" | "doing" | "done" }` that
   renders the status uppercased in brackets. Use it in `TaskRow`. Expected:
   `[TODO]` / `[DOING]` / `[DONE]` prefixes each row.

3. **Read like a reviewer:** in `ProjectCard`, someone wrote
   `project.status = "archived";` as the first line of the function body
   "to normalize display". The app appears to work. Explain in two sentences
   why this is a bug even when the screen looks right — one sentence must
   mention StrictMode's double render (Lesson 10, step 6).

<details><summary>Solution</summary>

1. 

   ```tsx
   function AssigneeTag({ assignee }: { assignee: string }) {
     return <span>@{assignee}</span>;
   }
   ```

   Inline `{ assignee: string }` as the prop type is legal for tiny
   components; render `<AssigneeTag assignee={task.assignee} />`.

2. 

   ```tsx
   export function StatusBadge({ status }: { status: "todo" | "doing" | "done" }) {
     return <span>[{status.toUpperCase()}]</span>;
   }
   ```

3. It mutates shared data from `lib/data.ts` — the change leaks to every other
   reader and render is no longer pure. StrictMode's double call makes the
   drift visible; in prod React may skip re-renders and lie silently.

</details>

## Checkpoint

1. What exactly does React do differently for `<div>` vs `<Div />`, and what
   does the capital letter switch between at compile time?
2. Write from memory a typed `PriorityLabel({ level }: ...)` component that
   renders `"P" + level`. What is the conventional name of its props type?
3. Why is mutating props (or imported module data) inside a component worse
   than "just a bit impure" — what silently breaks?
4. What is `children`, and when do you reach for it over more props?
5. Against the mock Go API: `GET /api/tasks?projectId=p1` (authorized) returns
   fields `id, projectId, title, status, assignee, priority, dueDate`. Which
   file you wrote today mirrors that contract, and what would the compiler
   catch if you renamed `assignee` to `owner` in `lib/types.ts`?

<details><summary>Answers</summary>

1. Lowercase compiles to `jsx("div", ...)` — a built-in descriptor; capitalized
   compiles to `jsx(Div, props)` and React calls it while rendering.
2. `function PriorityLabel({ level }: { level: number }) { return <>P{level}</>; }`
   — conventionally `PriorityLabelProps` as a named interface.
3. Other components read the same mutated object; React decides re-renders by
   reference equality, so in-place mutation can look "unchanged" while the UI
   freezes on stale output.
4. The prop holding whatever sits between a component's open and close tags —
   for composing content you don't control (wrappers, panels).
5. `lib/types.ts`. The compiler errors on `task.assignee` — it guards the
   *declared* contract, not the server's drift; Lesson 25's zod validates responses.

</details>

## Primary source

[react.dev — Your First Component](https://react.dev/learn/your-first-component)
— the "capitalized = function reference, lowercase = built-in" rule and the
read-only-props contract, the exact model this lesson assumes.

## Before next session

Delete the `projects.map(...)` call's `key={p.id}` prop in `page.tsx`, save,
and read the browser-console warning verbatim. Write the warning sentence in
your notes — Lesson 12 begins by decoding it.
