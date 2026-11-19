import { describe, expect, it, vi } from 'vitest';
import { load as loadReport } from './report/+page';
import { load as loadLeaderboard } from './leaderboard/+page';
import { load as loadHistory } from './history/+page';
import { load as loadHistoryFilters } from './history/+layout';

vi.mock('$lib/auth', () => ({ authHeaders: async () => ({}) }));

function pendingRequests() {
	const pending = new Map<string, (response: Response) => void>();
	const fetcher = vi.fn<typeof fetch>((input) => new Promise((resolve) => {
		pending.set(String(input), resolve);
	}));
	return {
		fetcher,
		respond(path: string, body: unknown, status = 200) {
			const resolve = pending.get(path);
			if (!resolve) throw new Error(`Request has not started: ${path}`);
			resolve(new Response(JSON.stringify(body), { status }));
		}
	};
}

function loadEvent<Route extends '/report' | '/leaderboard' | '/history'>(route: Route, fetcher: typeof fetch, query: string) {
	return {
		url: new URL(`http://localhost:5173${route}?${query}`),
		fetch: fetcher,
		params: {},
		route: { id: route },
		data: null,
		parent: async (): Promise<never> => { throw new Error('Unexpected parent load access'); },
		depends: vi.fn(),
		setHeaders: vi.fn(),
		untrack: <T>(fn: () => T) => fn(),
		get tracing(): never { throw new Error('Unexpected tracing access in a client load test'); }
	};
}

describe('match-report page loading', () => {
	it('starts community, roster and edit requests without waiting for one another', async () => {
		const { fetcher, respond } = pendingRequests();
		const result = loadReport(loadEvent('/report', fetcher, 'communityId=c_one&matchId=m_one'));
		await vi.waitFor(() => expect(fetcher).toHaveBeenCalledTimes(3));
		respond('/api/communities/c_one', { id: 'c_one', name: 'Community' });
		respond('/api/leaderboard?mode=1v1&communityId=c_one&full=1', []);
		respond('/api/matches/m_one', { matchId: 'm_one' });
		await expect(result).resolves.toMatchObject({
			communityId: 'c_one', community: { id: 'c_one' }, players: [], match: { matchId: 'm_one' }
		});
	});

	it('preserves the single-community redirect before loading the report', async () => {
		const { fetcher, respond } = pendingRequests();
		const result = loadReport(loadEvent('/report', fetcher, ''));
		const redirected = expect(result).rejects.toMatchObject({
			status: 307, location: '/report?communityId=c_one'
		});
		await vi.waitFor(() => expect(fetcher).toHaveBeenCalledTimes(1));
		respond('/api/communities', [{ id: 'c_one' }]);
		await redirected;
	});
});

describe('contextual history loading', () => {
	it('loads combined match filters alongside complete scope options and retains empty intersections', async () => {
		const { fetcher, respond } = pendingRequests();
		const event = loadEvent('/history', fetcher, 'playerId=p_one&communityId=c_one&mode=team');
		const matches = loadHistory(event);
		const filters = loadHistoryFilters(event);
		await vi.waitFor(() => expect(fetcher).toHaveBeenCalledTimes(5));
		respond('/api/history?playerId=p_one&communityId=c_one&mode=team&limit=50', []);
		respond('/api/profile/p_one', {
			player: { id: 'p_one', handle: 'Alice', unclaimed: false },
			ratings: [
				{ communityId: 'c_two', communityName: 'Second', mode: '1v1' },
				{ communityId: 'c_two', communityName: 'Second', mode: 'team' }
			]
		});
		respond('/api/communities/c_one', { id: 'c_one', name: 'First' });
		respond('/api/communities/c_one/membership', { canManage: false });
		respond('/api/history/players?communityId=c_one', [
			{ playerId: 'p_former', handle: 'Former player', unclaimed: true }
		]);
		await expect(matches).resolves.toMatchObject({ history: [], mode: 'team' });
		await expect(filters).resolves.toMatchObject({
			playerId: 'p_one', communityId: 'c_one', canManage: false,
			playerOptions: [
				{ value: 'p_one', label: 'Alice' },
				{ value: 'p_former', label: 'Former player', detail: 'Unclaimed' }
			],
			communityOptions: [{ value: 'c_one', label: 'First' }, { value: 'c_two', label: 'Second' }]
		});
	});

	it('makes an unscoped history URL explicit without losing the mode', async () => {
		const { fetcher, respond } = pendingRequests();
		const result = loadHistoryFilters(loadEvent('/history', fetcher, 'mode=ffa'));
		const redirected = expect(result).rejects.toMatchObject({
			status: 307, location: '/history?mode=ffa&playerId=p_one'
		});
		await vi.waitFor(() => expect(fetcher).toHaveBeenCalledTimes(1));
		respond('/api/me', { playerId: 'p_one' });
		await redirected;
	});

	it('uses the selected community plus memberships when no player is selected', async () => {
		const { fetcher, respond } = pendingRequests();
		const result = loadHistoryFilters(loadEvent('/history', fetcher, 'communityId=c_one'));
		await vi.waitFor(() => expect(fetcher).toHaveBeenCalledTimes(4));
		respond('/api/communities/c_one', { id: 'c_one', name: 'Visiting' });
		respond('/api/communities/c_one/membership', { canManage: true });
		respond('/api/history/players?communityId=c_one', []);
		respond('/api/communities', [{ id: 'c_two', name: 'Home' }]);
		await expect(result).resolves.toMatchObject({
			playerId: '', canManage: true, playerOptions: [],
			communityOptions: [{ value: 'c_two', label: 'Home' }, { value: 'c_one', label: 'Visiting' }]
		});
	});

	it('includes an unplayed unclaimed player home community', async () => {
		const { fetcher, respond } = pendingRequests();
		const result = loadHistoryFilters(loadEvent('/history', fetcher, 'playerId=p_guest'));
		await vi.waitFor(() => expect(fetcher).toHaveBeenCalledTimes(1));
		respond('/api/profile/p_guest', {
			player: { id: 'p_guest', handle: 'Guest', communityId: 'c_home', unclaimed: true }, ratings: []
		});
		await vi.waitFor(() => expect(fetcher).toHaveBeenCalledTimes(2));
		respond('/api/communities/c_home', { id: 'c_home', name: 'Home' });
		await expect(result).resolves.toMatchObject({
			playerOptions: [{ value: 'p_guest', label: 'Guest', detail: 'Unclaimed' }],
			communityOptions: [{ value: 'c_home', label: 'Home' }]
		});
	});

	it('rejects invalid URL modes without making an API request', async () => {
		const fetcher = vi.fn<typeof fetch>();
		await expect(loadHistory(loadEvent('/history', fetcher, 'communityId=c_one&mode=invalid')))
			.rejects.toMatchObject({ status: 400 });
		expect(fetcher).not.toHaveBeenCalled();
	});

	it('surfaces option-loading errors instead of claiming there are no players', async () => {
		const { fetcher, respond } = pendingRequests();
		const result = loadHistoryFilters(loadEvent('/history', fetcher, 'communityId=c_one'));
		const rejected = expect(result).rejects.toThrow('Players unavailable');
		await vi.waitFor(() => expect(fetcher).toHaveBeenCalledTimes(4));
		respond('/api/communities/c_one', { id: 'c_one', name: 'First' });
		respond('/api/communities/c_one/membership', { canManage: false });
		respond('/api/history/players?communityId=c_one', { error: 'Players unavailable' }, 503);
		respond('/api/communities', []);
		await rejected;
	});
});

describe('leaderboard loading after reporting', () => {
	it('starts all four reads together and preserves the selected mode and filters', async () => {
		const { fetcher, respond } = pendingRequests();
		const result = loadLeaderboard(loadEvent('/leaderboard', fetcher, 'communityId=c_one&mode=team&q=Alice&sort=games&full=1'));
		await vi.waitFor(() => expect(fetcher).toHaveBeenCalledTimes(4));
		respond('/api/leaderboard?mode=team&communityId=c_one&q=Alice&sort=games&full=1', [{ playerId: 'p_one' }]);
		respond('/api/leaderboard/modes?communityId=c_one', ['team']);
		respond('/api/communities/c_one', { id: 'c_one' });
		respond('/api/communities/c_one/membership', { status: 'member' });
		await expect(result).resolves.toMatchObject({
			mode: 'team', full: true, rows: [{ playerId: 'p_one' }],
			availableModes: ['team'], community: { id: 'c_one', membership: 'member' }
		});
	});

	it('still surfaces a failed membership read instead of showing an empty leaderboard', async () => {
		const { fetcher, respond } = pendingRequests();
		const result = loadLeaderboard(loadEvent('/leaderboard', fetcher, 'communityId=c_one'));
		const rejected = expect(result).rejects.toThrow('Membership unavailable');
		await vi.waitFor(() => expect(fetcher).toHaveBeenCalledTimes(4));
		respond('/api/leaderboard?mode=1v1&communityId=c_one', []);
		respond('/api/leaderboard/modes?communityId=c_one', []);
		respond('/api/communities/c_one', { id: 'c_one' });
		respond('/api/communities/c_one/membership', { error: 'Membership unavailable' }, 503);
		await rejected;
	});
});
