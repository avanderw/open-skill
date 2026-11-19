# sv

Everything you need to build a Svelte project, powered by [`sv`](https://github.com/sveltejs/cli).

## Creating a project

If you're seeing this, you've probably already done this step. Congrats!

```sh
# create a new project
npx sv create my-app
```

To recreate this project with the same configuration:

```sh
# recreate this project
npx sv@0.16.6 create --template minimal --types ts --add sveltekit-adapter="adapter:static" --install npm app
```

## Developing

Once you've created a project and installed dependencies with `npm install` (or `pnpm install` or `yarn`), start a development server:

```sh
npm run dev

# or start the server and open the app in a new browser tab
npm run dev -- --open
```

### Sign-in experience

Sign-in appears first on mobile, with the introductory hero underneath; desktop
shows them side-by-side. After a successful sign-in on this browser, the hero
defaults to a collapsed "Learn about Open-Skill" disclosure. This preference
stores only a boolean, survives sign-out, and does not grant authentication.

Saved-session checks show a visible status panel. After four seconds, the panel
offers connection guidance and a retry button that reloads the page. A slow check
never treats the user as signed out; it can still complete normally. Failed checks
show an error and retry action. Google sign-in shows progress until the popup
completes or is cancelled.

To investigate delays, reload the login page and inspect browser Performance
entries (durations are milliseconds):

```js
performance.getEntriesByType('mark')
  .filter(entry => entry.name.startsWith('open-skill:'))
performance.getEntriesByType('measure')
  .filter(entry => entry.name.startsWith('open-skill:'))
  .map(({ name, startTime, duration }) => ({ name, startTime, duration }))
```

`login-mounted` marks when the login UI mounts after JavaScript startup.
`firebase-init` measures SDK initialization; `session-restore` measures the shared
Firebase readiness check including initialization; `login-session-wait` measures
how much of that check the login page waits for. `login-redirect` measures navigation
after authentication. These overlapping measurements are not additive and contain
no identity or token data. Compare them with navigation/resource timings to
distinguish app download/startup, Firebase restoration, and destination loading.
Development server compilation and dependency optimization can inflate startup;
use a production build when measuring deployed performance.

## Building

To create a production version of your app:

```sh
npm run build
```

You can preview the production build with `npm run preview`.

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.
