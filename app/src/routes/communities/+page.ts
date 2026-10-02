import type { PageLoad } from './$types';
import { createApi, type Community } from '$lib/api';

export const load: PageLoad = async ({ fetch }) => {
	const api = createApi(fetch);
	// A reachability failure (e.g. CORS misconfiguration) shouldn't crash the whole page.
	const communities: Community[] = await api.communities().catch(() => []);
	return { communities };
};