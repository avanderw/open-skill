export const DEMO_USER = {
	id: 'u_demo',
	handle: 'Owner'
} as const;

const STORAGE_KEY = 'open-skill.demo-auth';

export function isDemoAuthenticated(): boolean {
	return typeof localStorage !== 'undefined' && localStorage.getItem(STORAGE_KEY) === 'true';
}

export function loginDemo(): void {
	if (typeof localStorage !== 'undefined') {
		localStorage.setItem(STORAGE_KEY, 'true');
		window.dispatchEvent(new Event('demo-auth-change'));
	}
}

export function logoutDemo(): void {
	if (typeof localStorage !== 'undefined') {
		localStorage.removeItem(STORAGE_KEY);
		window.dispatchEvent(new Event('demo-auth-change'));
	}
}

export function isProtectedPath(pathname: string): boolean {
	return (
		pathname.startsWith('/communities') ||
		pathname.startsWith('/team') ||
		pathname.startsWith('/report') ||
		pathname.startsWith('/profile') ||
		pathname.startsWith('/settings') ||
		pathname.startsWith('/history')
	);
}

export function loginPath(returnTo: string): string {
	return `/login?returnTo=${encodeURIComponent(returnTo || '/')}`;
}

export { STORAGE_KEY };
