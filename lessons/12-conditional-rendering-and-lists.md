# Lesson 12 — Conditional Rendering & Lists

*Phase 2 · Session 12 of 38 · ~2 hours · Prereqs: Lessons 02, 03, 11*

## Why this matters

Two patterns account for most of the JSX you'll read in the team's codebase:
"show this only when X" and "render one of these per item". Both are just
JavaScript — ternaries, `&&`, and `.map()` (Lesson 03) — used inside JSX.
Both also have one trap that produces the most infamous React bug class:
unstable list `key`s, which is why the console warning you collected last
night gets its own section here. This lesson's filter UI is the exact skeleton
of P1's task filter on Friday.

## What you'll be able to do

- You can render conditionally with ternary, `&&`, and early `return null`,
  and say which to use when.
- You can explain why `{count && <Banner />}` can print a naked `0`.
- You can render a list with `.map()` and a stable `key`, and state the two
  rules for keys.
- You can explain, with a concrete scenario, what breaks when you key by
  array index.
- You can build a filtered list — `.filter()` in render, no state — for
  TeamPulse tasks by project.

## Principles

### Conditionals are just expressions in `{}`

Lesson 10's rule — only expressions in `{}` — *is* the conditional-rendering
API. Three forms, in order of preference:

**1. Ternary — the default.** Two branches, always explicit:

```tsx
<p>{task.status === "done" ? "Shipped" : "In progress"}</p>
```

**2. `&&` — when the "else" is nothing.** `a && b` evaluates to `b` only if
`a` is truthy (Lesson 02), and rendering `false`/`null`/`undefined` prints
nothing:

```tsx
{task.priority === 1 && <span className="hot">🔥 top</span>}
```

**The `0` trap.** `0 && <x/>` evaluates to `0` — and unlike `false`, the
number `0` *is* rendered: `{unread && <Badge/>}` with `unread = 0` prints a
bare `0`. Fix patterns: `{unread > 0 && ...}`, `{!!unread && ...}`, or a
ternary. You'll hit it in week one of real code.

**3. Early return — whole-component conditions.** When the *entire* component
renders nothing (or a placeholder) in some case, return early like any
function:

```tsx
function TaskList({ tasks }: { tasks: Task[] }) {
  if (tasks.length === 0) return <p>No tasks. Enjoy the quiet.</p>;
  return <ul>{tasks.map((t) => <TaskRow key={t.id} task={t} />)}</ul>;
}
```

A component returning `null` renders nothing at all — that's how "render this
banner only for admins" is often expressed at the call site:
`{isAdmin && <AdminBanner />}`.

### Lists: `.map()` produces children, `key` identifies them

```tsx
<ul>
  {tasks.map((task) => (
    <TaskRow key={task.id} task={task} />
  ))}
</ul>
```

`tasks.map(...)` returns an *array of React elements* — and React renders
arrays as multiple children. Nothing React-specific about the map; it's
Lesson 03.

`key` is the one prop with a special meaning: React diffs old children
against new ones every render and uses `key` to decide which elements are
"the same item moved/updated" vs "new" vs "deleted". Without stable keys,
editing a list means React rebuilds — or worse, *reuses the wrong* — DOM
nodes: inputs keep the wrong text, checkboxes point at the wrong row.

The two rules:

1. **Unique among siblings** — two `<TaskRow key="t1">` in the same list is a
   bug and a console warning.
2. **Stable across renders** — same item, same key, forever. Which rules out:

```tsx
// NEVER: position, not identity
{tasks.map((task, i) => <TaskRow key={i} task={task} />)}

// NEVER: generated per render
{tasks.map((task) => <TaskRow key={Math.random()} task={task} />)}
```

The concrete failure with index keys: your list is `[A, B, C]` keyed `0, 1,
2`. The user deletes A. Now `[B, C]` are keyed `0, 1` — React looks at key
`0`'s DOM node, which still holds A's rendered output, and *patches it into
B*. Any DOM-level state inside that row (a typed-into input, focus, an open
`<details>`) now belongs to the wrong task. With id keys, React sees "key t1
gone", removes that node outright, and t2/t3 keep their exact DOM. (And
note: `key` is consumed by React — it never reaches `TaskRow` as a prop.)

### Filtering is derived data — compute it, don't store it

"How do I show only p1's tasks?" You already know how, from Lesson 03 —
`.filter()`. The React addition: **derive during render, don't copy into
separate storage.**

```tsx
const visible = tasks.filter((t) => t.projectId === projectId);
if (visible.length === 0) return <p>No tasks for this project.</p>;
return <ul>{visible.map((t) => <TaskRow key={t.id} task={t} />)}</ul>;
```

`visible` is *derived* from props — recompute it every render and it can
never go stale. Keeping a second list that you manually sync to the first is
how "the badge says 3, the list has 2" bugs are born. Rule to keep: **one
source of truth; everything else is computed from it.** (Lesson 13 adds state;
this rule becomes "derive from props + state, never mirror".)

## Follow along

Continue in `~/teampulse` (Lesson 11's `lib/data.ts`, `components/task-row.tsx`
still in place). Dev server: `pnpm dev`. No backend needed today.

### 1. Empty state via early return

In `app/page.tsx`, extract the task `<ul>` into a component in
`components/task-list.tsx`:

```tsx
import { TaskRow } from "@/components/task-row";
import type { Task } from "@/lib/types";

export function TaskList({ tasks }: { tasks: Task[] }) {
  if (tasks.length === 0) return <p>No tasks match. Try a different filter.</p>;
  return <ul>{tasks.map((t) => <TaskRow key={t.id} task={t} />)}</ul>;
}
```

Use it in `page.tsx`: `<TaskList tasks={tasks} />`. Expected: all three
tasks, as yesterday.

### 2. Filter by project, derived in render

Add a component `components/project-tasks.tsx`:

```tsx
import { TaskList } from "@/components/task-list";
import type { Task } from "@/lib/types";

export function ProjectTasks({
  tasks,
  projectId,
}: {
  tasks: Task[];
  projectId: string;
}) {
  const visible = tasks.filter((t) => t.projectId === projectId);
  return <TaskList tasks={visible} />;
}
```

In `page.tsx`, add below the cards:

```tsx
<h2>Checkout v2 tasks</h2>
<ProjectTasks tasks={tasks} projectId="p1" />
```

Expected: exactly the two p1 tasks. Change `projectId="p9"` → the empty
state earns its keep.

### 3. Conditional badges with && and ternary

Extend `components/task-row.tsx`:

```tsx
interface TaskRowProps {
  task: {
    id: string;
    title: string;
    status: "todo" | "doing" | "done";
    priority: number;
  };
}

export function TaskRow({ task }: TaskRowProps) {
  return (
    <li>
      <StatusBadge status={task.status} />{" "}
      <span style={{ opacity: task.status === "done" ? 0.5 : 1 }}>
        {task.title}
      </span>{" "}
      {task.priority === 1 && <strong>(TOP)</strong>}
    </li>
  );
}

interface StatusBadgeProps {
  status: "todo" | "doing" | "done";
}

function StatusBadge({ status }: StatusBadgeProps) {
  return <span>[{status.toUpperCase()}]</span>;
}
```

Expected: the done task is 50% faded; `t3` (priority 1) shows `(TOP)`; the
other rows don't.

### 4. Trigger the `0` trap, then fix it

Above the list in `page.tsx` add:

```tsx
{tasks.filter((t) => t.projectId === "p9").length && (
  <p>Some p9 tasks exist</p>
)}
```

Expected on screen: a lone **`0`**. Why: `0 && x` → `0`, and React renders
numbers. Fix: change the condition to
`tasks.filter((t) => t.projectId === "p9").length > 0 &&`. Save — the `0`
is gone. Delete the block.

### 5. Trigger the index-key warning properly

In `components/task-list.tsx`, temporarily switch to `key={tasks.indexOf(t)}`
(an index key). Save, open devtools console, then reorder `t1`/`t2` in
`lib/data.ts`. Expected: no crash — but note how fast the warning fires, and
memorize the Principles scenario: positional keys + changing lists = DOM
state attached to the wrong row. Restore `key={t.id}`; the warning stops.

## Your turn

1. **Warm-up:** in `ProjectCard` (Lesson 11), show `[ARCHIVED]` next to the
   project name only when `project.status === "archived"`, using `&&`. Set
   one project in `lib/data.ts` to `"archived"` to see it. Expected: the tag
   appears for exactly one card.

2. **Transfer:** write `components/project-filter.tsx` — a component taking
   `{ tasks: Task[]; show: "all" | "open" | "done" }` that renders
   `<TaskList>` of, respectively, all tasks, tasks where `status !== "done"`,
   and tasks where `status === "done"` — using a plain helper function
   *outside* the component. Render all three variants stacked in the page
   with `<h3>` labels. Expected: three lists, 3 / 2 / 1 tasks.

3. **Debug:** a teammate writes
   `{tasks.length ? tasks.map((t) => <TaskRow key={Math.floor(Math.random() * 1000)} task={t} />) : <p>Empty</p>}`.
   It renders correctly on first load. Name two distinct problems and the
   one-line fix for each.

<details><summary>Solution</summary>

1. 

   ```tsx
   <h3>
     {project.name}{" "}
     {project.status === "archived" && <span>[ARCHIVED]</span>}
   </h3>
   ```

2. 

   ```tsx
   function pick(tasks: Task[], show: "all" | "open" | "done"): Task[] {
     if (show === "open") return tasks.filter((t) => t.status !== "done");
     if (show === "done") return tasks.filter((t) => t.status === "done");
     return tasks;
   }

   export function ProjectFilter({
     tasks,
     show,
   }: {
     tasks: Task[];
     show: "all" | "open" | "done";
   }) {
     return <TaskList tasks={pick(tasks, show)} />;
   }
   ```

   The helper lives outside the component; TS is satisfied because `pick`
   returns on every path of the union.

3. (a) **Unstable keys:** a new key per render makes React rebuild every row
   each render — in-DOM state (typed text, focus) vanishes. (b) **No
   uniqueness guarantee:** random ids collide — `Math.floor(random()*1000)`
   yields duplicates with real probability. Fix both with `key={t.id}` — the
   API's ids are unique and stable.

</details>

## Checkpoint

1. `{unread && <Badge />}` — what renders when `unread` is `0`, and why?
   Give two fixes.
2. When do you use an early `return null` / early return instead of `&&` in
   the JSX?
3. State the two rules for `key`. Which one does `key={i}` violate, and which
   does `key={crypto.randomUUID()}` (called every render) violate?
4. In your own words: the concrete "user deletes the first row" failure with
   index keys — what does React do with the DOM nodes, and what user-visible
   glitch results?
5. TeamPulse question: `GET /api/tasks?projectId=p1` on
   `http://localhost:8080` returns `[]` for an unknown `projectId` — no 404.
   Why must your `TaskList`'s empty state be built to handle "server returned
   success with zero items" separately from "request failed"?

<details><summary>Answers</summary>

1. A literal `0` — `0 && x` evaluates to `0`, and React renders numbers (but
   skips `false`/`null`/`undefined`). Fixes: `unread > 0 && ...`,
   `!!unread && ...`, or a ternary.
2. When the whole component (or a large subtree) disappears in one case;
   `&&` suits inline show/hide of a fragment.
3. Unique among siblings; stable across renders. `key={i}` violates
   stability (an item's key changes when the list reorders);
   `randomUUID()` per render violates stability maximally (every item is
   "new" each render).
4. After deleting row A, B moves into key `0`'s slot; React patches A's old
   DOM node to look like B instead of removing it — so anything stored in the
   DOM itself (typed text, focus) shows up on B.
5. `[]` is a success: render "No tasks" for a 200. A failure (network down,
   401) is not an empty list — claiming "No tasks" hides real outages. Two
   states; Lesson 15 builds both.

</details>

## Primary source

[react.dev — Rendering Lists](https://react.dev/learn/rendering-lists) —
the definitive explanation of what `key` is *for* (item identity across
renders), including the "don't use indexes as keys" reasoning with examples.

## Before next session

In `page.tsx`, add above everything: `let clicked = 0;` (module-level) and a
button `<button onClick={() => { clicked = clicked + 1; }}>{clicked}</button>`.
Click it. Note what you see — Lesson 13 starts from why this is broken.
