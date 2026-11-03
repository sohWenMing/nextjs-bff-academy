# Lesson 13 — State: `useState` & Re-renders

*Phase 2 · Session 13 of 38 · ~2 hours · Prereqs: Lessons 03, 11, 12*

## Why this matters

This is the lesson the phase gate is about: *why does clicking a button change
the screen?* State lives in React, setting it schedules a re-render, and
re-render means **calling your function again** — the model that makes every
weird bug report predictable. It's where Lesson 03's immutability rule bites:
mutate state in place and the UI silently freezes.

## What you'll be able to do

- You can explain the render loop out loud: event → setState → React re-calls
  the component → new JSX diffed → DOM updated.
- You can write typed `useState` calls and update state immutably (spread, `map`).
- You can predict the traps — `setCount(count + 1)` twice vs
  `setCount(c => c + 1)` twice — and explain why a module-level `let` fails as
  component "memory" (re-rendering and multi-user).
- You can mark an interactive component `'use client'` and say why, deferring
  the full story to Lesson 18.

## Principles

### Render is a function call. State is React's memory of it.

Your component function is pure (Lesson 11): props in, JSX out. But a counter
"remembers" clicks — and a function can't remember anything between calls.
Resolution: **the function stays pure; the memory lives in React**, *outside*,
keyed by component position. `useState` is the interface to that memory.

```tsx
"use client";

import { useState } from "react";

export function Counter() {
  const [count, setCount] = useState(0);
  return <button onClick={() => setCount(count + 1)}>{count}</button>;
}
```

Line by line:

- `useState(0)` returns a **pair**: the current value for *this render*, and
  a setter. You destructure it (Lesson 02). `0` is the *initial* value —
  React reads it only the very first time this component instance mounts.
- The first render allocates a memory cell holding `0`.
- Click → `setCount(1)` → React marks the component dirty and **calls
  `Counter()` again**. This time `useState` hands back `1` from the cell. New
  JSX out, diff, DOM updated.
- The function was called twice; each call is a **render**. Local `let`
  variables are brand-new every call — which is why the Lesson 12 warm-up's
  `let clicked = 0` never moved, even if clicking had triggered a render.

Say it back until it's boring: **nothing you write calls the component
function — only React does, when it decides your state changed.** The moment
your render has side effects, that's when the model breaks.

### Three rules that fall out of "render = function call"

**1. Hooks are per-position.** React matches memory cells to `useState` calls
*by order in the function body*. This is why hooks must run unconditionally —
no `if (x) useState(...)` — or your second render reads the wrong cell.

**2. Reading state gives you the value *for this render*.** `count` inside
the function is a snapshot. Fire `setCount(count + 1)` twice in one handler
and both calls use the *same* snapshot:

```tsx
function wrong() {
  setN(n + 1); // n is the snapshot from THIS render — say 0
  setN(n + 1); // still 0 + 1 → final: 1, not 2
}

function right() {
  setN((prev) => prev + 1); // updater sees the queued result
  setN((prev) => prev + 1); // → final: 2
}
```

Whenever a new state value depends on the old, use the **updater function**.
Also: multiple `setState` calls in one handler are **batched** into a single
re-render — one render for all of them, not one each.

**3. Updates must be immutable — Lesson 03, now load-bearing.** React decides
"did anything change" partly by comparing references:

```tsx
// BAD: mutate in place, pass the same reference back
const found = tasks.find((t) => t.id === id);
if (found) {
  found.status = "done";
  setTasks(tasks); // SAME array reference → React may skip the update
}

// GOOD: new array, new object for the changed item
setTasks((prev) => prev.map((t) => (t.id === id ? { ...t, status: "done" } : t)));
```

Spread (`{ ...t, status: "done" }`) makes a *new* object with one field
changed; `map` builds a *new* array. That's Lesson 03's immutability, now
with a UI-freezing reason attached. The full set: add → `[...prev, item]`,
remove → `prev.filter(...)`, change one → `prev.map(...)` with
`{ ...item, changes }`.

### `'use client'` — the two-minute version

Next.js runs component code on the **server first** (the App Router's core
move — Lesson 18 gives the full story). Server-rendered code can't have
button handlers or memory: the server doesn't see your clicks. Marking a file
`"use client"` (a string literal at the very top, a *directive*) tells
Next.js: this component and everything it imports runs **in the browser**.
Anything with `useState`/`onClick`/`onChange` lives in a `'use client'`
file. (Lesson 22 later moves some *data fetching* server-side into route
handlers; state stays client-side forever — the browser is where the user is.)

### Where state goes

State lives in the **lowest component that needs it, and in the closest
common ancestor when several need it**. If `TaskList` and a badge both need
`tasks`, `tasks` lives above both and flows down as props (Lesson 11). And
unchanged from Lesson 12: anything computable from state/props is *derived
during render*, not stored in a second `useState`. Two copies of one fact is
where UI bugs breed.

## Follow along

In `~/teampulse`, dev server running. Watch the screen on every save — hot
reload is your feedback loop.

### 1. Fix the broken counter

You left `let clicked = 0` in `app/page.tsx` last night. Create
`components/counter.tsx`:

```tsx
"use client";

import { useState } from "react";

export function Counter() {
  const [count, setCount] = useState(0);
  console.log("render: count is", count);
  return <button onClick={() => setCount(count + 1)}>{count}</button>;
}
```

Render `<Counter />` in `page.tsx`. Open devtools console, click twice.
Expected: `render: count is 0` appears **twice**, then `1` twice, then `2`
twice — StrictMode's dev double-call: two log pairs per click = one render
per click. The number updates because the *cell* holds it, not the function.

### 2. The stale-snapshot trap

Add a second button to `Counter`:

```tsx
<button onClick={() => { setCount(count + 1); setCount(count + 1); }}>+2?</button>
```

Click `+2?`. Expected: up by **1**, not 2 — both calls read this render's
snapshot; React batched them into one re-render. Replace the body with
`setCount((c) => c + 1)` twice. Expected: now genuinely +2. You've met
batching and stale closures.

### 3. State with real data: toggleable tasks

Rewrite `components/task-list.tsx` — state lives here (the page doesn't care
about per-task status):

```tsx
"use client";

import { useState } from "react";
import type { Task } from "@/lib/types";

export function TaskList({ initialTasks }: { initialTasks: Task[] }) {
  const [tasks, setTasks] = useState<Task[]>(initialTasks);
  const [showDone, setShowDone] = useState(true);

  function toggleDone(id: string) {
    setTasks((prev) =>
      prev.map((t) =>
        t.id === id
          ? { ...t, status: t.status === "done" ? "todo" : "done" }
          : t,
      ),
    );
  }

  function addSample() {
    setTasks((prev) => [
      ...prev,
      {
        id: `t${Date.now()}`, projectId: "p1", title: `Sample ${prev.length + 1}`,
        status: "todo", assignee: "Wen", priority: 3,
        dueDate: "2026-12-31T00:00:00Z",
      },
    ]);
  }

  const visible = showDone ? tasks : tasks.filter((t) => t.status !== "done");
  if (visible.length === 0) return <p>No tasks. Add one.</p>;

  return (
    <>
      <ul>
        {visible.map((t) => (
          <li key={t.id}>
            <button onClick={() => toggleDone(t.id)}>
              {t.status === "done" ? "[x]" : "[ ]"}
            </button>{" "}
            {t.title} P{t.priority}
          </li>
        ))}
      </ul>
      <p>{tasks.filter((t) => t.status === "done").length} / {tasks.length} done</p>
      <button onClick={() => setShowDone((v) => !v)}>Toggle done</button>{" "}
      <button onClick={addSample}>Add sample</button>{" "}
      <button onClick={() => setTasks([])}>Clear</button>
    </>
  );
}
```

Update `page.tsx` to `<TaskList initialTasks={tasks} />`. Expected:
checkbox-style toggles work (done → strikethrough-style `[x]`, counter
updates), `Toggle done` hides completed rows, `Add sample` appends, `Clear`
shows the empty state. Note the progress line **derives** from `tasks`, not
`visible` — filtering is a view, not a deletion. Also: add
`console.log("TaskList render")` at the top of the function — every click
re-runs the *entire* function; React's diff, not your code, makes that cheap.

### 4. Prove immutability matters

In `toggleDone`, swap in the broken version (found → mutate → `setTasks(tasks)`
from Principles). Click a checkbox. Expected: **nothing visible happens** —
same array reference, React sees no change; the data is corrupted while the
UI is frozen. Restore the `map`/spread version and confirm toggling works.

## Your turn

1. **Warm-up:** in `TaskList`, label the filter button `{showDone ?
   "showing all" : "hiding done"}`, and make `Add sample` set `showDone` to
   `true` (a new task must be visible). Expected: with done-tasks hidden,
   adding a sample reveals the list again.

2. **Transfer:** build `components/like-button.tsx` (`'use client'`): state
   `likes: number`, a number, `Like`/`Unlike` buttons (`Unlike` never below
   0), and a `Burst` button that fires five queued updater calls in one
   click. Predict before saving: +5 or +1?

3. **Debug:** a teammate's `ProjectCard` has
   `const [taskCount, setTaskCount] = useState(tasks.filter(t => t.projectId === p.id).length)`
   and the badge goes stale when tasks are added above it. One sentence: what
   did they do wrong, and what's the fix?

<details><summary>Solution</summary>

1. 

   ```tsx
   <button onClick={() => setShowDone((v) => !v)}>
     {showDone ? "showing all" : "hiding done"}
   </button>
   // and inside addSample's handler: setShowDone(true); setTasks((prev) => ...)
   ```

2. 

   ```tsx
   "use client";

   import { useState } from "react";

   export function LikeButton() {
     const [likes, setLikes] = useState(0);
     return (
       <>
         <span>{likes}</span>
         <button onClick={() => setLikes((n) => n + 1)}>Like</button>
         <button onClick={() => setLikes((n) => Math.max(0, n - 1))}>Unlike</button>
         <button
           onClick={() => {
             for (let i = 0; i < 5; i++) setLikes((n) => n + 1);
           }}
         >
           Burst
         </button>
       </>
     );
   }
   ```

   Prediction: **+5**. Five queued updater functions run in order against the
   queued value, in a single re-render (batching).

3. They stored a **derived** value in state — `useState` captures the count
   once at mount; nothing re-runs the filter when `tasks` changes. Fix:
   delete the state, compute
   `tasks.filter(t => t.projectId === p.id).length` during render.

</details>

## Checkpoint

1. Walk the full chain: user clicks → screen updates. Name every step,
   including who calls your function.
2. Why does a module-level `let` fail as state? (Two independent reasons:
   re-rendering and multi-user.)
3. `setN(n + 1)` twice in one handler renders what, and why? What's the fix
   and what does it receive as its argument?
4. Why does `task.status = "done"; setTasks(tasks)` often leave the UI
   unchanged, and which Lesson 03 rule is violated?
5. TeamPulse: your P1 page will keep the login token in state. Why does
   `const [token, setToken] = useState<string | null>(null)` use `null`
   instead of `""`, and which Lesson 12 pattern is the future
   `if (!token) return <LoginForm />` line?

<details><summary>Answers</summary>

1. Click → `onClick` handler → setter call → React marks dirty + batches →
   React **re-calls the component** → new JSX → diff → only changed DOM nodes.
2. Nothing about assigning to it tells React to re-render, and a module is
   shared across all requests/users — it isn't even per-user memory.
3. +1 total: both calls read the same snapshot `n`, and the batched
   re-render applies the final value once. Fix: updater form
   `setN(prev => prev + 1)`; the argument is the queued latest value.
4. The array reference is unchanged, so React's reference check skips the
   re-render — mutation violates immutability; build a new array via `map` +
   `{ ...t, ... }`.
5. `null` means "no token", unlike the string-shaped lie `""`. The early
   return is Lesson 12's whole-component conditional rendering.

</details>

## Primary source

[react.dev — State: A Component's Memory](https://react.dev/learn/state-a-components-memory)
— it argues *why* locals and module variables can't be state before
introducing `useState`, which is exactly the order your brain needs things in.

## Before next session

In `TaskList`, add a bare `<input />` (no value, no onChange) above the
list. Type into it, then click a checkbox — your typed text survives.
Tomorrow (Lesson 14) the input's value comes *from* state, and you'll watch
what breaks and what finally gets powerful.
