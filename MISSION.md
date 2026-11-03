# Mission

**Why Wen is learning this:** His team is building a product with a **Next.js (TypeScript) frontend** and a **Backend-for-Frontend (BFF) layer** in Next.js on top of an existing **Go backend**. Wen currently knows **neither TypeScript nor React** and needs to reach "can operate" level: read the team's code, take small tickets, and own the BFF layer confidently.

**Time budget:** 2 hours per day, ~6 days per week. Target: ~7 weeks to operational.

**Definition of done (the mission is achieved when Wen can):**
1. Read any standard Next.js App Router file and know what runs on the server vs the browser.
2. Add a route handler (`app/api/*/route.ts`) that calls the Go backend, validates the response with zod, and reshapes it for the UI.
3. Build a React page with state, forms, and data fetching using the team's patterns.
4. Trace a bug from a browser Network tab → BFF route handler → Go API response.
5. Open a small, reviewable PR into a Next.js/TS codebase without hand-holding.

**Explicitly out of scope:** becoming a React/TS expert, SSR/ISR deep dives, CSS frameworks, Redux, class components, performance micro-optimization.

**The running project ("TeamPulse"):** every phase builds one app — a small team dashboard — against the included mock Go backend (`backend-mock/`), which mimics a real Go JSON API (auth token, projects, tasks, a flaky endpoint, a slow endpoint). Projects build strictly on each other: P0 (TS CLI) → P1 (React UI) → P2 (BFF layer) → P3 (typed contracts + caching + auth) → P4 (capstone, deployed).
