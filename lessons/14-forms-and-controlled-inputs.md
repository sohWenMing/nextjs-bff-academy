# Lesson 14 — Forms & Controlled Inputs

*Phase 2 · Session 14 of 38 · ~2 hours · Prereqs: Lessons 11, 13*

## Why this matters

P1 opens with a login form — `<form>`, two inputs, submit, client-side
validation before it ever calls `POST /api/login`. In React an input's value is
not browser-owned; it is *your state, echoed back through `value` on every
render*. That one idea explains every form bug class: text that won't type,
stale submit values, validation on the wrong text.

## What you'll be able to do

- You can write a controlled input from memory: state + `value` + `onChange`.
- You can explain why `<input value={text} />` with no `onChange` is frozen,
  in render-loop terms.
- You can handle submit correctly: `<form onSubmit>` + `e.preventDefault()`
  (and say what breaks without it).
- You can type form events and convert numeric inputs without `any`.
- You can build a live text filter over a list using one state value.

## Principles

### The input's value IS state

An uncontrolled input (plain HTML) keeps its text in the DOM; you read it out
when needed. React's posture — all UI is a function of data — wants the text
in *your* data:

```tsx
"use client";

import { useState } from "react";

export function NameInput() {
  const [name, setName] = useState("");
  return (
    <>
      <input value={name} onChange={(e) => setName(e.target.value)} />
      <p>Hello, {name || "nobody"}</p>
    </>
  );
}
```

Trace one keystroke, in render-loop language (Lesson 13):

1. DOM fires `onChange`; React calls your handler with an event object whose
   `target` is the input. `e.target.value` = "Al" — what the user *typed*.
2. `setName("Al")` → React re-calls `NameInput()` with state `"Al"`.
3. The new JSX has `value="Al"` → React pushes that back into the DOM.

The browser's suggestion is *overruled by your render output every single
keystroke*. This is a **controlled input**. The cost is one re-render per
keystroke (cheap; the team's code does this everywhere).

### Frozen input: `value` without `onChange`

```tsx
<input value={name} />           // warning; typing does NOTHING
<input value={name} onChange={...} /> // works
<input defaultValue={name} />    // uncontrolled: React sets it once, hands over
```

Why frozen: state never changed (no `onChange`), so every render re-asserts
the same `value`. Uncontrolled inputs exist, but the course standardizes on
controlled — validation, defaults, and clearing all get easy.

### Submit is an event on the form, not the button

```tsx
function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
  e.preventDefault(); // stop the browser's native GET-navigation
  console.log("submitting", { username, password }); // values ARE state
}
```

- **`e.preventDefault()` is mandatory.** A `<form>`'s native submit navigates
   (GET with query params) and reloads, wiping state. Your handler owns
   submission entirely.
- **Read the state, not the DOM.** Every keystroke already flowed into state
  — the values are just variables. No `document.getElementById` moment.

Typing: `React.FormEvent<HTMLFormElement>` — React's synthetic event
generic over the element. `onChange` handlers usually infer
`ChangeEvent<HTMLInputElement>` without annotation.

### Numeric inputs give you strings

`<input type="number">` still yields a **string** from `e.target.value`.
Convert at the edge into a number-typed state:

```tsx
const [priority, setPriority] = useState(1);

<input type="number" min={1} max={3} value={priority}
  onChange={(e) => setPriority(Number(e.target.value) || 1)} />
```

`Number(...)` converts; `|| 1` guards the emptied field (`""` → `NaN`/`0`).
The Go backend won't police `priority` bounds — your side of the bargain.

### One state value, two jobs: the live filter

The P1 task filter is a controlled input plus derived data (Lesson 12) —
nothing new required:

```tsx
const [query, setQuery] = useState("");
const visible = tasks.filter((t) =>
  t.title.toLowerCase().includes(query.toLowerCase()),
);
```

`query` in state (UI memory), `visible` computed every render (never stored).

## Follow along

In `~/teampulse`, dev server up. No backend today — you *simulate* login.
(Real `fetch` tomorrow, Lesson 15.)

### 1. Skeleton login form

Create `components/login-form.tsx`:

```tsx
"use client";

import { useState } from "react";

interface LoginFormProps {
  onLogin: (username: string, password: string) => void;
}

export function LoginForm({ onLogin }: LoginFormProps) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    onLogin(username, password);
  }

  return (
    <form onSubmit={handleSubmit}>
      <input
        value={username}
        onChange={(e) => setUsername(e.target.value)}
        placeholder="username"
        autoComplete="username"
      />
      <input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="password"
        autoComplete="current-password"
      />
      <button type="submit">Log in</button>
    </form>
  );
}
```

Note the design in `LoginFormProps`: the form doesn't decide what login *means*
— the parent does, via `onLogin`. Presentational child, smart parent; you'll
reuse this exact component in Lesson 15. Now wire it into `page.tsx`:

```tsx
<LoginForm onLogin={(u, p) => console.log("would POST", { u, p })} />
```

Expected: typing works in both boxes; empty-username submit shows the red
validation error; Network shows **no** page reload.

### 2. Kill preventDefault, observe the crime

Temporarily delete `e.preventDefault()`. Submit. Expected: full page reload,
state wiped, `?username=...` in the URL — the native GET you just prevented.
Restore it.

### 3. Client-side validation before the callback

Extend `LoginForm` with `const [error, setError] = useState<string | null>(null)`
and guard the handler:

```tsx
function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
  e.preventDefault();
  if (username.trim() === "") {
    setError("Username is required.");
    return;
  }
  if (password === "") {
    setError("Password is required.");
    return;
  }
  setError(null);
  onLogin(username.trim(), password);
}
```

Render inside the form (Lesson 12's `&&`):
`{error && <p style={{ color: "crimson" }}>{error}</p>}`.

Expected: empty submit → red `Username is required.`; filling the field and
submitting clears it. The server adds its own verdict tomorrow with a 401.

### 4. Live filter on TaskList

Open Lesson 13's `components/task-list.tsx`. At the top of the component add:

```tsx
const [query, setQuery] = useState("");
const visible = tasks.filter((t) =>
  t.title.toLowerCase().includes(query.toLowerCase()),
);
```

Render above the list:

```tsx
<input
  value={query}
  onChange={(e) => setQuery(e.target.value)}
  placeholder="Filter tasks…"
/>
```

Make the empty-check and `.map()` use `visible` (message:
`No tasks match "q".`), add a `{visible.length} / {tasks.length} done` count.
A filter is a *view*, not a deletion.

### 5. Numeric input

In `addSample`, priority is hardcoded `3`. Add a number input (`min=1`,
`max=3`, default `2`) and thread its value in — the new sample row shows `P2`.

## Your turn

1. **Warm-up:** make the filter *count* its matches: render
   `{visible.length} of {tasks.length} tasks` under the input. Type to
   confirm. Expected: `2 of 3 tasks` when filtering hits two.

2. **Transfer:** build `components/task-create.tsx` (`'use client'`): title
   input + priority select (1/2/3); `onSubmit` calls
   `onCreate({title, priority})` after a non-empty-title check. In `page.tsx`,
   append to `tasks`. Expected: a new row after picking priority 2.

3. **Debug:** a teammate's search box "loses every other keystroke" with
   `onChange={(e) => setQ(Math.random() > 0.5 ? e.target.value : q)}`. One
   sentence: what breaks, and what's the fix?

<details><summary>Solution</summary>

1. 

   ```tsx
   <p>
     {visible.length} of {tasks.length} tasks
   </p>
   ```

2. Lift: page holds
   `const [tasks, setTasks] = useState<Task[]>(initialTasks)`, renders
   `<TaskList tasks={tasks} />` and:

   ```tsx
   "use client";

   import { useState } from "react";

   export function TaskCreate({
     onAdd,
   }: {
     onAdd: (title: string, priority: number) => void;
   }) {
     const [title, setTitle] = useState("");
     const [priority, setPriority] = useState(1);
     const [error, setError] = useState<string | null>(null);

     function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
       e.preventDefault();
       if (title.trim() === "") {
         setError("Title is required.");
         return;
       }
       setError(null);
       onAdd(title.trim(), priority);
       setTitle("");
     }

     return (
       <form onSubmit={handleSubmit}>
         <input value={title} onChange={(e) => setTitle(e.target.value)}
           placeholder="New task title" />
         <input type="number" min={1} max={3} value={priority}
           onChange={(e) => setPriority(Number(e.target.value) || 1)} />
         <button type="submit">Add</button>
         {error && <p style={{ color: "crimson" }}>{error}</p>}
       </form>
     );
   }
   ```

3. Typing "ab": the `a` keystroke's delayed `setQ("a")` is superseded by
   `setQ("ab")` — the box always shows what was typed; debounced *work* sees
   one "ab". Debouncing defers work, never state.

</details>

## Checkpoint

1. Trace one keystroke in a controlled input end-to-end, naming the re-render.
2. Why does `<input value={q} />` without `onChange` reject all typing?
3. What does `e.preventDefault()` in `onSubmit` stop; what do you see if you skip it?
4. Why is `Number(e.target.value) || 1` written that way?
5. TeamPulse: `POST /api/login` expects body `{"username":"wen","password":"demo"}`
   — the password is the literal word `demo`` — and returns
   `{"token":"tok-...","user":{...}}` or 401 `{"error":"invalid credentials"}`.
   Where in today's `LoginForm` does the fetch go, and why does the `onLogin`
   prop keep the form reusable?

<details><summary>Answers</summary>

1. Keystroke → `onChange(event)` → `setName(e.target.value)` → React re-calls
   the component with new state → new JSX carries `value="Al"` → React writes
   it into the DOM. The text survives only because you echoed it.
2. State never updates, so every render re-asserts the old `value`.
3. Native submit → navigation (GET with query params) + reload + wiped state.
4. `type="number"` still yields a string; `""` → `0` and junk → `NaN`; `|| 1`
   gives both a valid default.
5. In `handleSubmit` after validation (or the parent's `onLogin`) — submit is
   the only moment both values are final. The prop keeps `LoginForm`
   presentational for any parent, tests, or a future error wrapper.

</details>

## Primary source

[react.dev — `input` reference, "Usage in React"](https://react.dev/reference/react-dom/components/input)
— the canonical controlled-vs-uncontrolled explanation with the exact "value
without onChange" warning semantics and state-echo reasoning.

## Before next session

Run the mock backend (`cd backend-mock && go run main.go`) and click through
your login form once more — tomorrow's fetch slots in right under the
validation branch.
