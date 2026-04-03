# Frontend

## Requirements

- Node.js 20+
- pnpm

## Setup

```bash
pnpm install
cp .env.example .env
```

## Run

```bash
pnpm dev
```

App URL: `http://localhost:5173`

## Environment

`.env`

```env
VITE_API_BASE_URL=http://localhost:8080
```

Use `http://localhost:8080` when the backend is running locally.

## Other Commands

```bash
pnpm build
pnpm preview
pnpm lint
pnpm typecheck
pnpm test
```
