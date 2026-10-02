import { initializeApp } from 'firebase/app';
import {
	getAuth,
	GithubAuthProvider,
	GoogleAuthProvider,
	onAuthStateChanged,
	signInWithPopup,
	signOut as firebaseSignOut,
	type Auth
} from 'firebase/auth';
import { base } from '$app/paths';
import { selectedDebugProfile } from '$lib/debug-profiles';

// Demo auth (trusted X-Demo-User header) is only ever available in dev builds.
export const DEMO_AUTH = import.meta.env.DEV && import.meta.env.VITE_DEMO_AUTH === '1';

export const AUTH_CHANGE_EVENT = 'auth-change';
const DEMO_STORAGE_KEY = 'open-skill.demo-auth';

export type AuthProvider = 'google' | 'github';

let auth: Auth | undefined;
let ready: Promise<void> | undefined;

function firebaseAuth(): Auth {
	if (!auth) {
		const app = initializeApp({
			apiKey: import.meta.env.VITE_FIREBASE_API_KEY,
			authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
			projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID,
			appId: import.meta.env.VITE_FIREBASE_APP_ID
		});
		auth = getAuth(app);
	}
	return auth;
}

function notify() {
	window.dispatchEvent(new Event(AUTH_CHANGE_EVENT));
}

/** Resolves once the persisted sign-in state is known. */
export function authReady(): Promise<void> {
	if (DEMO_AUTH || typeof window === 'undefined') return Promise.resolve();
	if (!ready) {
		const a = firebaseAuth();
		ready = a.authStateReady().then(() => {
			onAuthStateChanged(a, notify);
		});
	}
	return ready;
}

export function isAuthenticated(): boolean {
	if (DEMO_AUTH) return typeof localStorage !== 'undefined' && localStorage.getItem(DEMO_STORAGE_KEY) === 'true';
	return typeof window !== 'undefined' && firebaseAuth().currentUser !== null;
}

export async function signIn(provider: AuthProvider): Promise<void> {
	await signInWithPopup(firebaseAuth(), provider === 'google' ? new GoogleAuthProvider() : new GithubAuthProvider());
}

export function loginDemo(): void {
	if (DEMO_AUTH && typeof localStorage !== 'undefined') {
		localStorage.setItem(DEMO_STORAGE_KEY, 'true');
		notify();
	}
}

export async function signOut(): Promise<void> {
	if (DEMO_AUTH) {
		localStorage.removeItem(DEMO_STORAGE_KEY);
		notify();
		return;
	}
	await firebaseSignOut(firebaseAuth());
}

/** Headers identifying the caller to the API. */
export async function authHeaders(): Promise<Record<string, string>> {
	if (DEMO_AUTH) return { 'X-Demo-User': selectedDebugProfile().id };
	if (typeof window === 'undefined') return {};
	await authReady();
	const user = firebaseAuth().currentUser;
	// getIdToken refreshes the token when it is close to expiry.
	return user ? { Authorization: `Bearer ${await user.getIdToken()}` } : {};
}

export function isProtectedPath(pathname: string): boolean {
	// GitHub Pages can serve this app under a subpath; match against the
	// in-app path once the base is stripped off.
	const path = base && pathname.startsWith(base) ? pathname.slice(base.length) || '/' : pathname;
	return (
		path.startsWith('/communities') ||
		path.startsWith('/team') ||
		path.startsWith('/report') ||
		path.startsWith('/profile') ||
		path.startsWith('/settings') ||
		path.startsWith('/history')
	);
}

export function loginPath(returnTo: string): string {
	return `${base}/login?returnTo=${encodeURIComponent(returnTo || `${base}/`)}`;
}
