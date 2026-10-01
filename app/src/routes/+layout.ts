// SPA mode: no server-side rendering or prerendering. The static adapter
// emits a single fallback page; all routes are handled client-side.
import { browser } from '$app/environment';
import { redirect } from '@sveltejs/kit';
import { isDemoAuthenticated, isProtectedPath, loginPath } from '$lib/auth';
import type { LayoutLoad } from './$types';

export const ssr = false;
export const prerender = false;

export const load: LayoutLoad = ({ url }) => {
	if (browser && !isDemoAuthenticated() && isProtectedPath(url.pathname) && url.pathname !== '/login') {
		redirect(307, loginPath(url.pathname + url.search));
	}
};