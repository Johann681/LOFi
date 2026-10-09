# Sideby Web

Next.js App Router frontend for the Fiber student matching and chat API.

## Run locally

1. Start the Go API on `http://localhost:3000` with MongoDB available.
2. From this folder run `npm install` and `npm run dev`.
3. Open `http://localhost:3001` and create a student profile.

The API defaults to `http://localhost:3000` and the WebSocket defaults to `ws://localhost:3000`. Override these with `NEXT_PUBLIC_API_URL` and `NEXT_PUBLIC_WS_URL` when needed.

## Deploy

For a Vercel deployment using the Render API at `https://lofi-sk54.onrender.com`, set these Vercel environment variables for the deployment environments you use:

- `NEXT_PUBLIC_API_URL=https://lofi-sk54.onrender.com`
- `NEXT_PUBLIC_WS_URL=wss://lofi-sk54.onrender.com`

On Render, set `FRONTEND_ORIGINS` to the exact Vercel origin, for example `https://your-app.vercel.app` (no trailing slash). If the app uses multiple domains, provide a comma-separated list with no spaces. The API uses this setting for both CORS and WebSocket origin checks. The default only allows local development origins on port 3001.

The browser stores only the student's ID in `localStorage`. Authentication is not currently enabled by the backend, so this is a development flow and must not be treated as secure identity verification.
