import type { PageLoad } from './$types';
import { createApi, type Community } from '$lib/api';

export const load: PageLoad = async ({ params, fetch }) => {
	const api = createApi(fetch);
	try {
		return { community: await api.findCommunity(params.id) as Community };
	} catch {
		throw new Error('Community not found');
	}
};
