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

### Reporting consecutive matches

After reviewing a new match, choose **Save & log another** to stay on the report
page, or **Save & view leaderboard** to navigate once after saving. Editing an
existing match still saves changes and opens the leaderboard.

The repeat-entry flow shows the saved standings and server-returned rating changes,
keeps the community, played date, selected players, and teams, and clears winners
and finishing places. **Change players** clears participant selection without
reloading the roster. Team/FFA results do not overwrite the roster's 1v1 ratings.
Newly created unclaimed players remain available for the next match.

Saving disables the match controls and prevents duplicate in-flight submissions.
A failed save preserves the draft. A saved match is not submitted again when
retrying leaderboard navigation; its confirmation is passed through typed SvelteKit
navigation state rather than session storage. If destination data loading fails,
the leaderboard error page confirms the match was saved and offers a read-only
retry. Report and leaderboard pages load
independent API reads concurrently, and reporting does not invalidate the report
page before navigation. These changes do not alter backend match ordering or
rating recalculation.

### Filtering match history

History combines **Community**, **Player**, and match-mode filters. Choose a value
in either searchable picker to apply it immediately; the picker is a bottom sheet
on mobile. With both scopes selected, **All players** or **All communities** clears
just that filter. Keep at least one scope selected; a bare `/history` opens your
own history rather than implying a global feed.

Filters are stored in `playerId`, `communityId`, and `mode` URL parameters, so
refresh, shared links, and browser Back/Forward retain the view. Mode changes
reload only matches, not the scope metadata. Empty intersections keep the chosen
filters and show an explicit no-results message.

Community player options include the full current roster, unclaimed players, and
former participants, independently of the displayed 50 matches or leaderboard
limits. With a player selected, community options come from their ratings,
their unclaimed-player home community, and the currently selected community.
Choose a community before browsing a different player. Without a player filter,
the community picker lists your communities and the currently selected community.

Profile community sections link directly to that player's history in the
community. Match participant profile links remain unchanged; a separate **History
here** link filters to that participant in the match's community, preserving mode.

## Building

To create a production version of your app:

```sh
npm run build
```

You can preview the production build with `npm run preview`.

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.
