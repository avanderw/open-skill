import { describe, expect, it, vi } from 'vitest';
import { load as loadReport } from './report/+page';
import { load as loadLeaderboard } from './leaderboard/+page';

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

function loadEvent<Route extends '/report' | '/leaderboard'>(route: Route, fetcher: typeof fetch, query: string) {
	return {
		url: new URL(`http://localhost:5173${route}?${query}`),
		fetch: fetcher,
		params: {},
		route: { id: route },
		data: null,
		parent: async () => ({}),
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
