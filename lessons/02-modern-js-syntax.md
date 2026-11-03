# Lesson 02 — Modern JS syntax

*Phase 1 · Session 02 of 38 · ~2 hours · Prereqs: [Lesson 01](01-toolchain-node-pnpm-your-first-scripts.md)*

## Why this matters

Code written in the last decade of JavaScript looks nothing like the JS of
2009. Every file in your team's Next.js repo uses `const`, arrow functions,
destructuring, and spread — if these are unfamiliar, the team's code reads like
noise. This lesson gives you the ~15 syntactic constructs that make up most of
what you will ever type in the BFF layer.

## What you'll be able to do

- You can declare variables with `let`/`const` and explain why `var` is banned.
- You can write arrow functions, including the concise one-expression form.
- You can build strings with template literals without looking anything up.
- You can destructure objects and arrays, with renamed and default fields.
- You can use spread/rest to copy and merge objects and call variadic functions.

## Principles

**Variables and binding.** In JS, a *binding* associates a name with a value.
`const` creates a binding that cannot be reassigned; `let` allows
reassignment; `var` is the pre-2015 mechanism with confusing scoping rules —
treat it as a lint error you never write. Rule of thumb: `const` always,
`let` only when you genuinely reassign (loop counters, accumulators).

```javascript
const MAX = 3;
let count = 0;
count += 1;      // fine
// MAX = 4;      // TypeError: Assignment to constant variable
```

One subtlety: `const` freezes the *binding*, not the value. An object bound
with `const` can still have its properties changed — the immutability story is
[Lesson 03](03-arrays-objects-immutability.md).

**Functions** come in two syntaxes. The classic `function` keyword still
exists; the modern default is the **arrow function** — an expression whose
parameters come before `=>`:

```javascript
const square = (x) => x * x;              // one expression: returned implicitly
const add = (a, b) => {                   // block form: explicit return needed
  const sum = a + b;
  return sum;
};
```

Arrow functions also inherit `this` from where they were defined (the
"lexical" value), which is why team code reaches for them inside callbacks.
You can ignore `this` for Phase 1 — just always use arrows for callbacks.

**Template literals** use backticks and `${...}` for interpolation — the
expression inside `${}` is evaluated and stringified. They span multiple lines
naturally. Never build strings with `+` anymore.

**Destructuring** extracts values from objects and arrays into variables in
one statement. It is everywhere in React code (`const { data } = props`).

```javascript
const user = { id: "u1", name: "Wen", role: "admin" };
const { name, role: title } = user;       // name === "Wen", title === "admin"
const { email = "n/a" } = user;           // default when property missing

const pair = [10, 20];
const [first, second] = pair;             // positional
```

**Spread** (`...`) *expands* an iterable; **rest** (`...`) *collects* the
remainder. Same three dots, opposite directions, judged by position:

```javascript
// Spread — in a value position: expand
const base = { status: "todo", priority: 2 };
const task = { ...base, title: "Write tests" };   // copy + override, new object

// Rest — in a binding/parameter position: collect
const { status, ...withoutStatus } = task;        // drop one field
const sum = (...nums) => nums.reduce((a, b) => a + b, 0);
sum(1, 2, 3);                                     // 6
```

`{ ...base, title }` creates a **new object** — that is the seed of the
"never mutate" discipline that powers React.

**Optional chaining and nullish coalescing** handle "the value might be
missing", which JSON APIs do constantly:

```javascript
const city = response.user?.address?.city;  // undefined if any link is missing
const port = config.port ?? 8080;            // right side only for null/undefined
const fallback = value || "default";         // ⚠ fires for 0 and "" too
```

Prefer `??` over `||` for defaults: `||` treats `0` and `""` as missing,
`??` only treats `null`/`undefined` that way.

## Follow along

Work in `~/learn-js/lesson-02` (create it; make a `package.json` with
`pnpm init` if you like — scripts work the same as [Lesson 01](01-toolchain-node-pnpm-your-first-scripts.md)).

1. Create `syntax.mjs` and type — do not paste — each block below, running
   `node syntax.mjs` after adding each section. Typing is where syntax
   becomes memory.

```javascript
// --- bindings
const PROJECT = "TeamPulse";
let hits = 0;
hits += 1;
console.log(`${PROJECT}: ${hits} hit(s)`);
// expected: TeamPulse: 1 hit(s)

// --- arrow functions
const tasks = ["invoice export", "retry webhook", "keyboard overlap"];
const done = tasks.map((t) => `${t} ✔`);
console.log(done);
// expected: [ 'invoice export ✔', 'retry webhook ✔', 'keyboard overlap ✔' ]

// --- destructuring
const project = {
  id: "p1",
  name: "Billing Service",
  status: "active",
  owner: { name: "Wen", role: "admin" },
};
const { name, owner: { role } } = project;
console.log(`${name} owned by ${role}`);
// expected: Billing Service owned by admin

// --- spread / rest
const updated = { ...project, status: "archived" };
const { owner, ...summary } = updated;
console.log(summary);
// expected: { id: 'p1', name: 'Billing Service', status: 'archived' }
console.log(project.status, updated.status);
// expected: active archived  <- original untouched, new object was made

// --- optional chaining / nullish
const maybe = { user: null };
const city = maybe.user?.address?.city ?? "Singapore";
console.log(city);
// expected: Singapore
```

2. Prove `const` is about the binding, not the value — create `mutate.mjs`:

```javascript
const cfg = { port: 8080 };
cfg.port = 9090;        // allowed! we changed a property, not the binding
console.log(cfg.port);  // expected: 9090
```

Run `node mutate.mjs`, watch it succeed, then add `cfg = {};` after it and
confirm Node now throws `TypeError: Assignment to constant variable`.

3. Explore `||` vs `??` in a one-liner:

```bash
node -e 'console.log(0 || 10, 0 ?? 10, "" || "d", "" ?? "d")'
```

Expected: `10 0 d ` (the last is blank — `??` kept the empty string).

## Your turn

1. **Rewrite.** Convert to arrow + template literal + destructuring:

```javascript
function greet(user) {
  return "Hello " + user.name + " (" + user.role + ")";
}
```

2. **Reshape.** Given `const t = { id: "t1", title: "x", status: "todo", priority: 2, dueDate: "2026-12-01" }`,
   produce a *new* object that (a) drops `dueDate` and (b) sets `status` to
   `"doing"`, in a single expression. The original must not change.

3. **BFF-flavored.** Write `flatten.mjs`: given the `project` object from
   Follow-along step 1, produce
   `{ id, name, ownerName, ownerRole }` using destructuring with renames and
   spread. Print the result.

<details><summary>Solution 1</summary>

```javascript
const greet = ({ name, role }) => `Hello ${name} (${role})`;
```

Destructuring in the parameter list is the idiomatic React/BFF style.
</details>

<details><summary>Solution 2</summary>

```javascript
const { dueDate, ...rest } = t;
const fixed = { ...rest, status: "doing" };
```

Or one expression: `const fixed = (({ dueDate, ...r }) => ({ ...r, status: "doing" }))(t);`
— clever, but the two-liner is what reviewers prefer.
</details>

<details><summary>Solution 3</summary>

```javascript
const project = {
  id: "p1",
  name: "Billing Service",
  status: "active",
  owner: { name: "Wen", role: "admin" },
};

const { id, name, owner: { name: ownerName, role: ownerRole } } = project;
const flat = { id, name, ownerName, ownerRole };
console.log(flat);
// { id: 'p1', name: 'Billing Service', ownerName: 'Wen', ownerRole: 'admin' }
```

`{ id, name }` uses *shorthand*: when the variable and the key share a name,
`{ id }` means `{ id: id }`. Team code uses this constantly.
</details>

## Checkpoint

1. When is `let` justified over `const`? Give one legitimate use.
<details><summary>Answer</summary>
Only when you reassign the binding: a loop counter, an accumulator summing in
a `for` loop, a retry attempt count. Everything else is `const`.
</details>

2. What does `{ ...base, title: "x" }` produce, and does `base` change?
<details><summary>Answer</summary>
A brand-new object containing base's properties plus/overriding `title`.
`base` is untouched — spread copies values into a new object.
</details>

3. Difference between `value ?? "default"` and `value || "default"` when
   `value` is `0`?
<details><summary>Answer</summary>
`0 ?? "default"` → `0` (zero is not nullish). `0 || "default"` → `"default"`
(zero is falsy). Use `??` for defaults from API data.
</details>

4. What does `const { data: rows = [] } = response;` do in one sentence?
<details><summary>Answer</summary>
It creates a variable named `rows` from `response.data`, defaulting to an
empty array if `response.data` is `undefined`.
</details>

5. You will later fetch tasks from the TeamPulse mock API
(`http://localhost:8080`). Using destructuring and spread, write an expression
that takes a task `{ id, title, status: "todo" }` and returns it with status
`"done"`. (Server not required — just write the expression.)
<details><summary>Answer</summary>

```javascript
const task = { id: "t1", title: "x", status: "todo" };
const completed = { ...task, status: "done" };
```
</details>

## Primary source

[MDN — JavaScript guide](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide)
MDN is the de-facto JS dictionary: every construct in this lesson has a page
there with live examples, and it is the reference senior JS developers
actually link each other.

## Before next session

Take the `project` object from Follow-along step 1 and, in a scratch file,
predict what each of these prints, then check:
`Object.keys(project)`, `Object.values(project.owner)`, and
`"status" in project`. Look up `Object.keys` on MDN if the answers surprised
you — you will need all three heavily in [Lesson 03](03-arrays-objects-immutability.md).
