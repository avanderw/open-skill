import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

// GitHub Pages serves this project under a subpath (https://avanderw.co.za/open-skill);
// leave empty for local dev and for hosts that serve the app at the origin root.
const rawBase = process.env.VITE_BASE_PATH ?? '';
if (rawBase !== '' && !rawBase.startsWith('/')) {
	throw new Error(`VITE_BASE_PATH must start with "/" or be empty, got ${JSON.stringify(rawBase)}`);
}
const base = rawBase as '' | `/${string}`;

export default defineConfig({
	plugins: [
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) => filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			paths: { base },
			adapter: adapter({
				// SPA mode: dynamic routes (/profile/[id]) and query-driven leaderboards
				// are handled client-side. GitHub Pages serves 404.html for unknown paths,
				// so using it as the fallback makes deep-link refreshes work.
				fallback: '404.html',
				pages: 'build',
				assets: 'build'
			})
		})
	],
	server: {
		proxy: {
			// Dev: forward /api to the Go service so the app and API share an origin.
			'/api': 'http://localhost:8080'
		}
	}
});