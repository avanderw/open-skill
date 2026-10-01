import type { PageLoad } from './$types';
import type { Community } from '$lib/api';
import { createApi } from '$lib/api';

export const load: PageLoad = async ({ fetch }) => {
	const api = createApi(fetch);
	const communities: Community[] = await api.browseCommunities();
	return { communities };
};
