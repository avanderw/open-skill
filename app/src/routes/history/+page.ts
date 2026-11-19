import type { PageLoad } from './$types';
import { error } from '@sveltejs/kit';
import { createApi } from '$lib/api';

export const load: PageLoad = async ({ fetch, url, parent }) => {
	const api = createApi(fetch);
	const playerId = url.searchParams.get('playerId') || '';
	const communityId = url.searchParams.get('communityId') || '';
	const mode = url.searchParams.get('mode') || '';
	if (mode !== '' && mode !== '1v1' && mode !== 'team' && mode !== 'ffa') {
		error(400, 'Invalid match mode');
	}
	if (!playerId && !communityId) {
		await parent();
		error(400, 'Choose a player or community');
	}
	const history = await api.history(playerId, {
		communityId: communityId || undefined, mode: mode || undefined, limit: 50
	});
	return { history, mode };
};
