# Client

## Running the Dev Server

This project uses [Moon](https://moonrepo.dev/) as a monorepo task runner. From the repository root:

```bash
moon run client:dev
```

This starts the Vite dev server with hot-module replacement.

## Architecture

The app has two views and a central Pinia store that owns the WebSocket connection.

- **HomeView** (`/`) — Register with a username, pick 3 Pokemon, and join the matchmaking queue. Redirects to battle once a match is found.
- **BattleView** (`/battle`) — Real-time battle UI. Attack, switch Pokemon, and watch the battle unfold through a live log panel.
- **wsStore** (`src/stores/wsStore.ts`) — Single Pinia store that manages the WebSocket lifecycle, sends/receives all messages, and exposes reactive battle state to both views.

## Environment Variables

Environment variables are managed through `.env` files and must be prefixed with `VITE_` to be exposed to the client.

| Variable       | Description                          | Example                             |
| -------------- | ------------------------------------ | ----------------------------------- |
| `VITE_WS_URL`  | WebSocket URL of the backend server  | `ws://localhost:3003/battle`        |

- `.env` — local development defaults (`ws://localhost:3003/battle`)
- `.env.production` — production overrides (used during `pnpm build`)
