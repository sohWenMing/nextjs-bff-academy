# Lesson 03 — Arrays, objects, immutability

*Phase 1 · Session 03 of 38 · ~2 hours · Prereqs: [Lesson 02](02-modern-js-syntax.md)*

## Why this matters

A BFF exists to take arrays of JSON objects from the Go backend and reshape
them for the UI: filter the done tasks, map projects to slimmer rows, reduce
tasks into counts. That is `filter`/`map`/`reduce`, three array methods that
make up a large fraction of real BFF code. The *how* is easy; the part that
separates junior from mid-level is the **immutability discipline** — never
change data in place, always produce new copies. That rule is the root of
every React state rule you meet in Phase 2, so we install it now, deeply.

## What you'll be able to do

- You can use `map`, `filter`, `find`, `some`, and `reduce` without looking anything up.
- You can sort a copy of an array without corrupting the original.
- You can shallow-copy and deep-copy objects and state when each is needed.
- You can explain, in one sentence, why mutating state breaks React.
- You can chain array methods to reshape API data in a single pipeline.

## Principles

**Arrays** are ordered lists, written with `[]`. JS arrays are dynamic
(resizable, mixed types allowed) and zero-indexed. In practice you get them
from `JSON.parse` or `fetch` as lists of **objects** — key/value bundles,
which mirror the Go structs your backend defines.

**The big three.** `map`, `filter`, and `reduce` all take a *callback* — a
function the array calls for you, once per element. They are **higher-order
functions**; the element-handling logic is the argument you pass in.

```javascript
const tasks = [
  { id: "t1", title: "Invoice PDF export", status: "doing", priority: 1 },
  { id: "t2", title: "Retry payment webhook", status: "todo", priority: 2 },
  { id: "t3", title: "Fix keyboard overlap", status: "done", priority: 1 },
];

// map: transform each element, same length out
const titles = tasks.map((t) => t.title);
// ['Invoice PDF export', 'Retry payment webhook', 'Fix keyboard overlap']

// filter: keep elements where the callback returns true
const open = tasks.filter((t) => t.status !== "done");
// t1 and t2 only

// reduce: fold the array into ONE value, carrying an accumulator
const byPriority = tasks.reduce((acc, t) => {
  acc[t.priority] = (acc[t.priority] ?? 0) + 1;
  return acc;
}, {}); // {} is the accumulator's starting value
// { '1': 2, '2': 1 }
```

`reduce` is the one that needs a moment: read it as "start with `{}`, then for
each task update this running object, and hand me the final object." Anything
you'd write with a `for` loop and an accumulator variable is a `reduce`.

Supporting cast you will also use daily: `find` (first match or `undefined`),
`some`/`every` (any/all pass the test), `at(-1)` (last element),
`includes` (contains).

**Mutation is the enemy.** JS objects and arrays are *reference types*: a
variable holds a pointer to the value, not the value itself. Copying a
variable copies the pointer — both names point at the **same** object.

```javascript
const a = { port: 8080 };
const b = a;          // b points at the SAME object a does
b.port = 9090;
console.log(a.port);  // 9090 — a "changed" too. It didn't. It's the same object.
```

**Shallow vs deep copy.** Spread (`{ ...obj }`, `[...arr]`) copies one level
deep. Nested objects are still shared pointers:

```javascript
const project = { name: "Billing", owner: { name: "Wen" } };
const copy = { ...project };
copy.owner.name = "Sarah";
console.log(project.owner.name); // "Sarah" — shared nested object!
```

For the occasional true deep copy, `structuredClone(obj)` (built into Node 17+)
clones every level. For nested updates, spread at each level you change:

```javascript
const updated = { ...project, owner: { ...project.owner, name: "Sarah" } };
```

**Why never mutate — the React connection (Lesson 13 makes it concrete).** React decides whether to re-render by
comparing the *identity* (the pointer) of state before and after an update.
`state.push(x)` mutates in place — same pointer — React concludes "nothing
changed" and your UI silently freezes. `setState([...state, x])` produces a
new pointer, React sees a new value, and re-renders. Every "my state updated
but the screen didn't" bug you will ever have in Phase 2 is this rule violated
early. Learning immutability now is learning React early.

The same identity logic makes immutability *cheap and correct* for caches and
change-detection in TanStack Query (Phase 4). Immutability is not pedantry; it
is the mechanism the whole UI layer runs on.

## Follow along

Create `~/learn-js/lesson-03/arrays.mjs` and type each section, running
`node arrays.mjs` as you go. The data mirrors the TeamPulse mock API's real
task shape (see `backend-mock/main.go`).

1. **The big three on real-shaped data:**

```javascript
const tasks = [
  { id: "t1", projectId: "p1", title: "Invoice PDF export", status: "doing", priority: 1 },
  { id: "t2", projectId: "p1", title: "Retry payment webhook", status: "todo", priority: 2 },
  { id: "t3", projectId: "p2", title: "Fix keyboard overlap", status: "done", priority: 1 },
  { id: "t4", projectId: "p2", title: "Push notification opt-in", status: "todo", priority: 3 },
  { id: "t5", projectId: "p1", title: "Load test checkout", status: "todo", priority: 2 },
];

const p1Open = tasks
  .filter((t) => t.projectId === "p1" && t.status !== "done")
  .map((t) => `${t.title} (P${t.priority})`);
console.log(p1Open);
// [ 'Invoice PDF export (P1)', 'Retry payment webhook (P2)', 'Load test checkout (P2)' ]

const countByStatus = tasks.reduce((acc, t) => {
  acc[t.status] = (acc[t.status] ?? 0) + 1;
  return acc;
}, {});
console.log(countByStatus); // { doing: 1, todo: 3, done: 1 }

console.log(tasks.find((t) => t.id === "t3")?.title); // Fix keyboard overlap
console.log(tasks.some((t) => t.priority === 3));     // true
```

2. **Prove aliasing bites.** Append to `arrays.mjs`:

```javascript
const sorted = tasks.sort((a, b) => a.priority - b.priority);
console.log(sorted[0].id, tasks[0].id);
// expected: t1 t1 — looks innocent... but sort mutated `tasks` IN PLACE!
```

`Array.prototype.sort` reorders the *original* array and returns it — the
single most common mutation bug in JS. Now fix it:

```javascript
const tasks2 = [/* re-type the array, or restart the file */];
const safe = [...tasks2].sort((a, b) => a.priority - b.priority);
console.log(tasks2[0].id); // t1 — original preserved
```

3. **Prove shallow copy leaks.** Continue with:

```javascript
const project = { name: "Billing", tags: ["go", "payments"] };
const copy = { ...project };
copy.tags.push("urgent");
console.log(project.tags); // ['go', 'payments', 'urgent'] — leaked!

const deep = structuredClone(project);
deep.tags.push("nope");
console.log(project.tags.length); // still 3 — structuredClone is truly deep
```

4. **Immutably update one nested field.** Write it, predict, run:

```javascript
const project = { name: "Billing", owner: { name: "Wen", role: "admin" } };
const promoted = { ...project, owner: { ...project.owner, role: "lead" } };
console.log(project.owner.role, promoted.owner.role); // admin lead
```

## Your turn

1. **Sort safely.** Produce tasks ordered by priority (1 first) without
   mutating `tasks`, then print the first title.

2. **Reshape for a UI.** From `tasks`, build an array of objects shaped
   `{ projectId, count, urgent }` — one entry per project, where `count` is
   the number of that project's tasks and `urgent` is whether any has
   `priority === 1`. (Hint: `filter` per project id, or one `reduce`.)

3. **Bug hunt.** This code intends to mark a task done without touching the
   array. Find the two bugs:

```javascript
const done = tasks.map((t) => {
  if (t.id === "t2") t.status = "done";
  return t;
});
```

<details><summary>Solution 1</summary>

```javascript
const sorted = [...tasks].sort((a, b) => a.priority - b.priority);
console.log(sorted[0].title); // Invoice PDF export
```
</details>

<details><summary>Solution 2</summary>

```javascript
const projectIds = [...new Set(tasks.map((t) => t.projectId))]; // unique ids
const summary = projectIds.map((projectId) => ({
  projectId,
  count: tasks.filter((t) => t.projectId === projectId).length,
  urgent: tasks.some((t) => t.projectId === projectId && t.priority === 1),
}));
// [{ projectId: 'p1', count: 3, urgent: true },
//  { projectId: 'p2', count: 2, urgent: true }]
```

`new Set(array)` deduplicates; spreading it back gives an array. The
readability of this beats a clever one-liner `reduce`.
</details>

<details><summary>Solution 2 alternative (single reduce)</summary>

```javascript
const summary = tasks.reduce((acc, t) => {
  const cur = acc.find((s) => s.projectId === t.projectId);
  if (!cur) {
    acc.push({ projectId: t.projectId, count: 1, urgent: t.priority === 1 });
  } else {
    acc.splice(acc.indexOf(cur), 1, {
      ...cur,
      count: cur.count + 1,
      urgent: cur.urgent || t.priority === 1,
    });
  }
  return acc;
}, []);
```

Note the `splice` replacing-in-place is fine *here*: `acc` is local to the
reduce and nobody else holds a reference. Immutability protects *shared* data.
</details>

<details><summary>Solution 3</summary>

Bug 1: `t.status = "done"` mutates the original task object (same pointer).
Bug 2: it returns the same object for every element, so `done` aliases
`tasks`. Fix:

```javascript
const done = tasks.map((t) =>
  t.id === "t2" ? { ...t, status: "done" } : t
);
```

Returning the untouched `t` for other elements is fine — they are not
modified, and React only cares that *changed* things get new identities.
</details>

## Checkpoint

1. What does each return: `map`, `filter`, `reduce`?
<details><summary>Answer</summary>
`map`: a new array, same length, each element transformed. `filter`: a new
array with the elements that passed the test. `reduce`: one single value
(number, object, array — whatever you choose) folded from all elements.
</details>

2. Why is `tasks.sort(...)` dangerous, and what is the standard fix?
<details><summary>Answer</summary>
`sort` mutates the array in place *and* returns it, so every other holder of
that reference sees the reorder. Fix: `[...tasks].sort(...)` — copy, then sort.
</details>

3. What does `{ ...project }` *not* copy, and what do you reach for instead
   when that bites?
<details><summary>Answer</summary>
Spread is shallow: nested objects/arrays are still shared references. Use
per-level spread for updates, or `structuredClone` for a full deep copy.
</details>

4. In one sentence: why does mutating React state leave the screen stale?
<details><summary>Answer</summary>
React compares state by reference identity; in-place mutation keeps the same
reference, so React decides nothing changed and skips the re-render.
</details>

5. Against the TeamPulse mock API (`http://localhost:8080`), you fetched all
   tasks into `tasks` and want p2's undone tasks for project "p2". Write the
   expression.
<details><summary>Answer</summary>

```javascript
const p2Open = tasks.filter((t) => t.projectId === "p2" && t.status !== "done");
```
</details>

## Primary source

[MDN — Array objects reference](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array)
Every method in this lesson lives on this page with signatures, edge cases,
and examples — including which methods mutate (`sort`, `reverse`, `splice`,
`push`) and which return new arrays (`map`, `filter`, `slice`, `toSorted`).
Bookmark it.

## Before next session

Run `node -e 'const a=[3,1,2]; const b=a.toSorted(); console.log(a.join(), b.join())'`.
Note which order each prints, and look up `toSorted` on the MDN page above —
it is the 2023 addition that sorts without mutating. Bring the answer to
[Lesson 04](04-modules-and-bundler-mental-model.md), where we finally split
your code across files.
