import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) => filename.split(/[/\\\\]/).includes('node_modules') ? undefined : true
			},
			adapter: adapter({
				// SPA mode: dynamic routes (/profile/[id]) and query-driven leaderboards
				// are handled client-side; serve a single fallback page.
				fallback: 'index.html',
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