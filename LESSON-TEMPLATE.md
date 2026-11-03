# Lesson Template (writers MUST follow)

Every lesson file is a single self-contained Markdown document in `lessons/`, named
`NN-short-dash-title.md` where NN is the two-digit session number (e.g. `07-ts-shapes.md`).

Structure (keep these exact headings, in this order):

```markdown
# Lesson NN — Title

*Phase X · Session NN of 38 · ~2 hours · Prereqs: Lesson(s) a, b*

## Why this matters
2–4 sentences tying the lesson directly to the mission: team Next.js+TS frontend/BFF over a Go backend.

## What you'll be able to do
3–5 bullet checklist, phrased as observable outcomes ("You can write a … without looking anything up").

## Principles
The core explanations. Short paragraphs, every claim backed by a code block or a concrete example.
This is where the conceptual "aha" lives. Keep it tight — depth over breadth.

## Follow along
Numbered, copy-pasteable steps. Every command and every code block must be real and runnable.
State expected output after each meaningful step so the learner gets immediate feedback.

## Your turn
1–3 exercises the learner does alone, increasing difficulty. Put solutions in
`<details><summary>Solution</summary> ... </details>` blocks so they must attempt first.

## Checkpoint
5 recall questions (answers in `<details>` blocks). Questions test *recall and transfer*, not
recognition. At least one question involves the TeamPulse mock Go API
(`http://localhost:8080`, see `backend-mock/main.go`).

## Primary source
ONE best external link for this topic (official docs preferred: nextjs.org, react.dev,
typescriptlang.org, developer.mozilla.org). One sentence on why it's the best.

## Before next session
A 5-minute prep task (e.g. "run `npm create next-app@latest` and note any errors").
```

Rules:
- Audience: competent programmer (knows general programming, some Go exposure) but ZERO
  JavaScript/TypeScript/React/Next.js knowledge. Never assume they know JS idioms.
- Explain every term the first time it appears, in one clause.
- Code style: TypeScript, 2-space indent, modern ESM imports. React function components only.
- Next.js 15+, App Router ONLY. Never teach `pages/` router, class components, or `create-react-app`.
- Package manager: `pnpm` (mention npm alternative in a footnote once per lesson if used).
- The mock Go backend runs at `http://localhost:8080` with endpoints documented in
  `backend-mock/README.md`. Use it for all backend-touching exercises. For pure
  frontend-fetch practice where auth is a nuisance, `https://jsonplaceholder.typicode.com`
  is acceptable.
- Target length: 150–350 lines of Markdown. Project sessions (09, 16, 22, 30, 36) may be
  200–450 lines.
- Tone: direct, no flattery, no emoji. Talk to the reader as "you".
- Cross-link lessons you build on: `[Lesson 05](05-async-promises-await-fetch.md)`.
