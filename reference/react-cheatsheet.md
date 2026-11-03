# React cheatsheet — the mental model + daily patterns

React is only in `"use client"` files (client components). Domain data: projects like `p1 "Billing Service"`, tasks with `status: todo|doing|done`.

## 1. A component is a function

```tsx
function ProjectCard({ project }: { project: Project }) {
  return <h2>{project.name}</h2>;    // JSX: returns UI, not side effects
}
```

React calls your function to *produce* UI. **Render = pure function of (props, state).** Same inputs → same output, no fetching/mutating/log side effects during render.

## 2. State + render loop

```tsx
const [filter, setFilter] = useState<TaskStatus | "all">("all");
```

- `useState(initial)` returns `[currentValue, setter]` — one slot of state per call.
- Calling the setter asks React to **re-render** the component: your function runs again from the top with new state.
- The old `filter` variable in an already-running callback is **stale** — each render closes over its own values.
- Updates are batched/async: `setCount(c => c + 1)` twice in a row adds 2 (functional update), `setCount(count + 1)` twice adds 1.

## 3. Lists + keys

```tsx
{tasks.map(task => <TaskRow key={task.id} task={task} />)}
```

`key` must be **stable and unique among siblings** — use `task.id`, never the array index. Index keys make React reuse the wrong row when you filter/reorder (TeamPulse's task filter looks broken: check your keys first).

## 4. Conditional rendering

```tsx
{tasks.length === 0 && <p>No tasks yet.</p>}
{user ? <span>{user.name}</span> : <span>Guest</span>}
```

Caution: `count && <X/>` renders `0` when count is 0 — write `count > 0 && <X/>`.

## 5. Controlled input pattern (forms)

The input's value IS state; `onChange` writes it back; render feeds it forward.

```tsx
function LoginForm({ onLogin }: { onLogin: (u: string, p: string) => void }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  return (
    <form onSubmit={e => { e.preventDefault(); onLogin(username, password); }}>
      <input value={username} onChange={e => setUsername(e.target.value)} />
      <input value={password} onChange={e => setPassword(e.target.value)} />
      <button type="submit">Log in</button>
    </form>
  );
}
```

Login against the mock: `POST /api/login` with `{ username: "wen", password: "demo" }`.

## 6. Lifting state up

Two sibling components need the same data? The data isn't theirs — it belongs to their nearest common parent.

```tsx
// Parent owns it:
const [selectedProjectId, setSelectedProjectId] = useState("p1");
// Child receives value + callback props:
<ProjectPicker value={selectedProjectId} onSelect={setSelectedProjectId} />
<TaskList projectId={selectedProjectId} />
```

Props flow down; events (callbacks) flow up. There is no "write to a parent" — you pass the parent's setter down as a prop.

## 7. Immutability — the root of every React rule

State must be **replaced**, never mutated, or React sees the same object reference and skips the update.

```tsx
setTasks(prev => [...prev, newTask]);                     // add
setTasks(prev => prev.map(t => t.id === id ? { ...t, status: "done" } : t)); // update
setTasks(prev => prev.filter(t => t.projectId !== id));   // remove
```

`prev.push(newTask)` + `setTasks(prev)` renders nothing new. Same for objects: `{ ...task, status: "done" }`.

## 8. `useEffect` — the escape hatch, not the default

`useEffect(fn, deps)` runs `fn` after render, re-runs when deps change. Legitimate uses: subscribing to an external system, syncing something outside React (browser APIs, timers), one-time browser-only setup.

**Fetching data in `useEffect` is usually an anti-pattern** — do it in a Server Component (`await fetch` directly) or with TanStack Query on the client. If you must:

```tsx
useEffect(() => {
  const ac = new AbortController();
  fetch("/api/tasks?projectId=p1", { signal: ac.signal })
    .then(r => r.json()).then(setTasks).catch(() => setError("failed"));
  return () => ac.abort();          // cleanup: runs on unmount / before next run
}, [projectId]);                    // every value read inside must be here
```

Rules of hooks: top level only, never inside `if`/`for`/callbacks; same order every render.

## 9. "Why does my component render twice?"

Dev-only `StrictMode` (in `app/layout.tsx`) mounts → unmounts → remounts components to expose bad effects. Double render in **dev is normal**; one extra render means something else:

| Symptom | Cause | Fix |
|---|---|---|
| Twice in dev only | StrictMode | ignore; it's a feature |
| Render loop (max depth) | `setX` called during render or in effect with bad deps | move the set into an event handler; fix the dep array |
| Parent re-renders everything | parent state changed | only the stateful subtree re-renders; split components if needed |

## 10. "Why is my state stale?"

| Symptom | Cause | Fix |
|---|---|---|
| `setCount(count + 1)` twice adds 1 | closure captured old `count` | functional update `setCount(c => c + 1)` |
| `useEffect` sees old `filter` | missing dep in array | add every value read inside to `[deps]` |
| Event handler uses old value | handler from an old render (e.g. stored in a ref/timer) | include state in deps of the effect that registers it |
| `setTasks(prev => { prev.push(x); return prev; })` doesn't update | returned same reference | `return [...prev, x]` |
| Read state right after setting it, get old value | updates are async; this render's variables are frozen | derive from the new value directly, or read it next render |

## 11. Component-splitting heuristic

When a file grows, split where **state lives**: the smallest component that owns a piece of state wraps the UI that needs it. Props are the API between components — keep them few, typed, read-only.

---
*Related lessons: [11 components & props](../lessons/11-components-and-props.md) · [12 conditional rendering & lists](../lessons/12-conditional-rendering-and-lists.md) · [13 state & re-renders](../lessons/13-state-usestate-and-rerenders.md) · [14 forms](../lessons/14-forms-and-controlled-inputs.md) · [15 useEffect](../lessons/15-data-fetching-and-useeffect.md) · [26 TanStack Query](../lessons/26-tanstack-query-on-the-client.md)*
