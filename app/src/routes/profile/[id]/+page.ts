import type { PageLoad } from './$types';
import { createApi, type Profile } from '$lib/api';

export const load: PageLoad = async ({ params, fetch }) => {
	const api = createApi(fetch);
	const profile: Profile = await api.profile(params.id);
	return { profile };
};