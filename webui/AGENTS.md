# Agent Guideline

## Technique Choice

- Language: TypeScript
- UI Framework: React
- Build Tool: Vite
- Styling: Tailwind CSS
- End-to-end testing: Playwright

## Project Description

- This repository is for a Web UI project.
- Prefer patterns that keep the UI easy to iterate on, test, and restyle.
- Favor simple composition over heavy abstractions until repeated patterns are proven.

## Project Structure And Module Organization

- Keep application code under `src`.
- Define and maintain the canonical folder layout in `specs/structure.md`.
- Organize the app around route features so screens, loaders, actions, and feature-local UI stay close together.
- Use `createBrowserRouter` from `react-router-dom` with explicit route definitions rather than ad hoc route declarations spread across the app.
- Use `easy-peasy` for client state management. Keep store models focused on shared client state, not ephemeral component-local UI state.
- Keep shared UI primitives, utilities, hooks, and types in clearly named shared modules.
- Avoid deeply nested folders unless they reflect a stable domain boundary.
- Keep files focused. Split large components or helpers when readability drops.

## Documentation

- Keep `README.md` and `AGENTS.md` in the repository root.
- Keep project structure and architecture notes under `specs` when they become stable enough to guide future work.
- Always add new environment variables to `.env.example`.
- Document intent when behavior is not obvious. Avoid comments that only restate the code.
- Add brief inline comments for complex UI state, data flow, or interaction logic when needed.

## Development Process

- Reuse existing components, hooks, utilities, and styles before adding new patterns.
- Prefer incremental changes that keep the UI working at each step.
- When adding a new reusable pattern, refactor the older code that overlaps with it.
- Keep commit messages clear and specific. Add a body when extra context is useful.

## Coding Style And Naming Conventions

- Use standard TypeScript and React conventions.
- Prefer function components and hooks.
- Keep components presentational when possible. Move data fetching, transformation, and side effects into hooks or higher-level modules when that improves clarity.
- Put route configuration in a dedicated router module and keep route-specific logic close to the route feature that owns it.
- Prefer easy-peasy actions, thunks, and selectors for cross-route client state. Do not introduce a second global client-state library unless the project explicitly adopts one later.
- Use descriptive names based on domain meaning, not temporary implementation details.
- Prefer explicit prop types and avoid overly broad types such as `any`.
- Keep styling consistent with the chosen Tailwind patterns. Extract repeated UI patterns into reusable components instead of copying long class lists everywhere.
- Colocate small feature-specific helpers with the feature. Move only broadly reusable code into shared modules.

## Testing Guidelines

- Use Playwright for end-to-end coverage of important user flows.
- Prioritize tests for critical paths, regressions, and cross-page interactions.
- Keep test names behavior-focused and readable.
- Do not add brittle end-to-end coverage for trivial rendering details.
- Pure test additions or fixes generally do not require broader documentation updates unless user-facing behavior changes.

## Planning New Features

- Start from the user flow, then define the component, state, and API boundaries needed to support it.
- Prefer extending an existing pattern when it already fits the new feature.
- If a new feature exposes duplication in older code, refactor toward a reusable component or hook as part of the change.
- Balance flexibility with speed: do not over-engineer for hypothetical future screens.
