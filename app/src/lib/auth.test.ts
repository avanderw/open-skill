import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const firebase = vi.hoisted(() => ({
	auth: { currentUser: null as { uid: string } | null, authStateReady: vi.fn() },
	initializeApp: vi.fn(),
	onAuthStateChanged: vi.fn(),
	signInWithPopup: vi.fn()
}));

vi.mock('firebase/app', () => ({ initializeApp: firebase.initializeApp }));
vi.mock('firebase/auth', () => ({
	getAuth: () => firebase.auth,
	GoogleAuthProvider: class {},
	GithubAuthProvider: class {},
	onAuthStateChanged: firebase.onAuthStateChanged,
	signInWithPopup: firebase.signInWithPopup,
	signOut: vi.fn()
}));

let storage: Map<string, string>;

beforeEach(() => {
	vi.resetModules();
	vi.clearAllMocks();
	vi.stubEnv('VITE_DEMO_AUTH', '0');
	vi.stubGlobal('window', new EventTarget());
	storage = new Map();
	vi.stubGlobal('localStorage', {
		getItem: (key: string) => storage.get(key) ?? null,
		setItem: (key: string, value: string) => storage.set(key, value)
	});
	firebase.auth.currentUser = null;
	firebase.auth.authStateReady.mockResolvedValue(undefined);
	firebase.signInWithPopup.mockResolvedValue(undefined);
});

afterEach(() => {
	vi.unstubAllGlobals();
	vi.unstubAllEnvs();
	vi.restoreAllMocks();
});

describe('saved session restoration', () => {
	it('shares the pending check and does not remember an anonymous visit', async () => {
		let resolve!: () => void;
		firebase.auth.authStateReady.mockReturnValue(new Promise<void>((done) => { resolve = done; }));
		const { authReady, hasSignedInBefore } = await import('./auth');
		const first = authReady();
		const second = authReady();
		expect(first).toBe(second);
		await Promise.resolve();
		expect(firebase.auth.authStateReady).toHaveBeenCalledTimes(1);
		expect(firebase.onAuthStateChanged).not.toHaveBeenCalled();
		resolve();
		await first;
		expect(hasSignedInBefore()).toBe(false);
	});

	it('allows retrying a failed check, including initialization failures', async () => {
		firebase.initializeApp.mockImplementationOnce(() => { throw new Error('initialization failed'); });
		const { authReady } = await import('./auth');
		await expect(authReady()).rejects.toThrow('initialization failed');
		firebase.auth.authStateReady.mockRejectedValueOnce(new Error('session failed'));
		await expect(authReady()).rejects.toThrow('session failed');
		await expect(authReady()).resolves.toBeUndefined();
		expect(firebase.onAuthStateChanged).toHaveBeenCalledTimes(1);
	});

	it('remembers a restored session and subsequent auth changes', async () => {
		firebase.auth.currentUser = { uid: 'existing-user' };
		const { authReady, hasSignedInBefore } = await import('./auth');
		await authReady();
		expect(hasSignedInBefore()).toBe(true);
		storage.clear();
		const callback = firebase.onAuthStateChanged.mock.calls[0]![1];
		callback({ uid: 'another-user' });
		expect(hasSignedInBefore()).toBe(true);
	});
});

describe('returning-user preference', () => {
	it('is saved only after successful sign-in', async () => {
		const { signIn, hasSignedInBefore } = await import('./auth');
		firebase.signInWithPopup.mockRejectedValueOnce(new Error('popup closed'));
		await expect(signIn('google')).rejects.toThrow('popup closed');
		expect(hasSignedInBefore()).toBe(false);
		await signIn('google');
		expect(hasSignedInBefore()).toBe(true);
	});

	it('reports preference storage failures without failing sign-in', async () => {
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});
		vi.stubGlobal('localStorage', {
			getItem: () => { throw new Error('storage blocked'); },
			setItem: () => { throw new Error('storage blocked'); }
		});
		const { signIn, hasSignedInBefore } = await import('./auth');
		expect(hasSignedInBefore()).toBe(false);
		await expect(signIn('google')).resolves.toBeUndefined();
		expect(warn).toHaveBeenCalledTimes(2);
	});
});
