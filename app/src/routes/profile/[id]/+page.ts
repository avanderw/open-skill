import type { PageLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { base } from '$app/paths';
import { createApi, type Profile } from '$lib/api';

export const load: PageLoad = async ({ params, fetch }) => {
	const api = createApi(fetch);
	const profile: Profile = await api.profile(params.id);
	if (profile.player.id !== params.id) redirect(307, `${base}/profile/${encodeURIComponent(profile.player.id)}`);
	return { profile };
};