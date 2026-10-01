import type { PageLoad } from './$types';
import { createApi, type Community } from '$lib/api';

export const load: PageLoad = async ({ fetch }) => {
	const api = createApi(fetch);
	const communities: Community[] = await api.communities();
	return { communities };
};