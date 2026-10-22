import type { PageLoad } from './$types';
import { createApi, type Community } from '$lib/api';

export const load: PageLoad = async ({ params, fetch }) => {
	const api = createApi(fetch);
	const [community, membership] = await Promise.all([
		api.findCommunity(params.id),
		api.membership(params.id)
	]);
	return {
		community: community as Community,
		canManage: membership.canManage,
		claims: membership.canManage ? await api.communityPlayerClaims(params.id) : []
	};
};
