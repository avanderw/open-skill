import { afterEach, describe, expect, it, vi } from 'vitest';
import { createApi } from '$lib/api';

function installLocalStorage() {
	const values = new Map<string, string>();
	vi.stubGlobal('localStorage', {
		getItem: (key: string) => values.get(key) ?? null,
		setItem: (key: string, value: string) => values.set(key, value),
		removeItem: (key: string) => values.delete(key)
	});
	return values;
}

function response(body: unknown, init: ResponseInit = {}) {
	return new Response(typeof body === 'string' ? body : JSON.stringify(body), {
		status: 200,
		headers: { 'Content-Type': 'application/json' },
		...init
	});
}

afterEach(() => vi.unstubAllGlobals());

describe('API client contract', () => {
	it('sends the played date when reporting and editing a match', async () => {
		installLocalStorage();
		const fetcher = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(async () => response({}));
		const client = createApi(fetcher);
		await client.reportMatch({ communityId: 'c_one', participants: [], playedAt: '2026-10-12' });
		await client.updateMatch('match/one', { participants: [], playedAt: '2026-10-10' });
		expect(fetcher.mock.calls[0]![0]).toBe('/api/matches');
		expect(JSON.parse(fetcher.mock.calls[0]![1]!.body as string).playedAt).toBe('2026-10-12');
		expect(fetcher.mock.calls[1]![0]).toBe('/api/matches/match%2Fone');
		expect(JSON.parse(fetcher.mock.calls[1]![1]!.body as string).playedAt).toBe('2026-10-10');
	});

	it('sends the selected demo identity and encodes query parameters', async () => {
		const storage = installLocalStorage();
		storage.set('open-skill.debug-profile', 'u_member');
		const fetcher = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(async () => response([]));

		await createApi(fetcher).leaderboard('1v1', {
			q: 'A&B / test',
			communityId: 'community/one',
			full: 'yes'
		});

		const [url, init] = fetcher.mock.calls[0]!;
		expect(url).toBe('/api/leaderboard?mode=1v1&communityId=community%2Fone&q=A%26B+%2F+test&full=yes');
		expect((init?.headers as Record<string, string>)['X-Demo-User']).toBe('u_member');
	});

	it('sends an email identity selected from the login form', async () => {
		const storage = installLocalStorage();
		storage.set('open-skill.debug-profile', 'tester@example.com');
		const fetcher = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(async () => response({ ok: true }));

		await createApi(fetcher).health();

		const [, init] = fetcher.mock.calls[0]!;
		expect((init?.headers as Record<string, string>)['X-Demo-User']).toBe('tester@example.com');
	});

	it('preserves zero-valued history options', async () => {
		installLocalStorage();
		const fetcher = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(async () => response([]));

		await createApi(fetcher).history('p/one', { limit: 0, offset: 0 });

		expect(fetcher.mock.calls[0]![0]).toBe('/api/history?playerId=p%2Fone&limit=0&offset=0');
	});

	it('uses the API error message from JSON responses', async () => {
		installLocalStorage();
		const fetcher = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(async () => response({ error: 'membership required' }, { status: 403 }));

		await expect(createApi(fetcher).communities()).rejects.toThrow('membership required');
	});

	it('sends JSON bodies for mutations', async () => {
		installLocalStorage();
		const fetcher = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(async () => response({}));

		await createApi(fetcher).createCommunity({ name: 'Test community', mode: 'open', draws: true });

		const [, init] = fetcher.mock.calls[0]!;
		expect(init?.method).toBe('POST');
		expect(init?.body).toBe(JSON.stringify({ name: 'Test community', mode: 'open', draws: true }));
		expect((init?.headers as Record<string, string>)['Content-Type']).toBe('application/json');
	});

	it('keeps member operation routes typed and encoded', async () => {
		installLocalStorage();
		const fetcher = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(async () => response({ ok: true, role: 'admin' }));

		await createApi(fetcher).promoteMember('community/one', 'user/two');

		const [url, init] = fetcher.mock.calls[0]!;
		expect(url).toBe('/api/communities/community%2Fone/members/user%2Ftwo/promote');
		expect(init?.method).toBe('POST');
	});

	it('creates an unclaimed player in the selected community', async () => {
		installLocalStorage();
		const fetcher = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(async () => response({ id: 'p_new', unclaimed: true, userId: '' }));
		const player = await createApi(fetcher).createUnclaimedPlayer('community/one', 'Visiting Player');
		const [url, init] = fetcher.mock.calls[0]!;
		expect(url).toBe('/api/communities/community%2Fone/players');
		expect(init?.method).toBe('POST');
		expect(JSON.parse(init?.body as string)).toEqual({ handle: 'Visiting Player' });
		expect(player.unclaimed).toBe(true);
	});

	it('encodes claim requests and approval routes', async () => {
		installLocalStorage();
		const fetcher = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(async () => response({ ok: true, playerId: 'p_canonical' }));
		const client = createApi(fetcher);
		await client.requestPlayerClaim('p/one');
		const approval = await client.approvePlayerClaim('p/one', 'u/two');
		await client.rejectPlayerClaim('p/one', 'u/two');
		expect(approval.playerId).toBe('p_canonical');
		expect(fetcher.mock.calls.map(([url, init]) => [url, init?.method])).toEqual([
			['/api/profile/p%2Fone/claims', 'POST'],
			['/api/profile/p%2Fone/claims/u%2Ftwo/approve', 'POST'],
			['/api/profile/p%2Fone/claims/u%2Ftwo', 'DELETE']
		]);
	});

	it('loads the community settings claim queue with encoded community IDs', async () => {
		installLocalStorage();
		const pending = [{ playerId: 'p_guest', handle: 'Guest', userId: 'u_member', targetPlayerId: 'p_member', targetHandle: 'Member' }];
		const fetcher = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(async () => response(pending));
		const claims = await createApi(fetcher).communityPlayerClaims('community/one');
		expect(fetcher.mock.calls[0]![0]).toBe('/api/communities/community%2Fone/player-claims');
		expect(claims[0]?.targetPlayerId).toBe('p_member');
	});

	it('prefixes requests with a configured API base URL', async () => {
		installLocalStorage();
		const fetcher = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(async () => response({ ok: true }));

		await createApi(fetcher, 'https://api.example.test/').health();

		expect(fetcher.mock.calls[0]![0]).toBe('https://api.example.test/api/health');
	});

	it('falls back to non-JSON error text', async () => {
		installLocalStorage();
		const fetcher = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(async () => response('upstream unavailable', { status: 502 }));

		await expect(createApi(fetcher).health()).rejects.toThrow('upstream unavailable');
	});

	it('falls back to status when an error body is empty', async () => {
		installLocalStorage();
		const fetcher = vi.fn<(input: RequestInfo | URL, init?: RequestInit) => Promise<Response>>(async () => response('', { status: 503 }));

		await expect(createApi(fetcher).health()).rejects.toThrow('Request failed: 503');
	});
});