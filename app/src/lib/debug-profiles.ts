export const DEBUG_PROFILES = [
	{ id: 'u_demo', handle: 'Owner', label: 'Owner / admin preview' },
	{ id: 'u_member', handle: 'CommunityMember', label: 'Joined member preview' },
	{ id: 'u_guest', handle: 'GuestPlayer', label: 'Non-member preview' }
] as const;

export const DEBUG_PROFILE_KEY = 'open-skill.debug-profile';

export interface DebugProfile {
	id: string;
	handle: string;
	label: string;
}

export function selectedDebugProfile(): DebugProfile {
	if (typeof localStorage === 'undefined') return DEBUG_PROFILES[0];
	const selected = localStorage.getItem(DEBUG_PROFILE_KEY)?.trim();
	const fixed = DEBUG_PROFILES.find((profile) => profile.id === selected);
	if (fixed) return fixed;
	if (selected) return { id: selected, handle: selected, label: 'Email preview' };
	return DEBUG_PROFILES[0];
}

export function setDebugProfile(id: string) {
	if (typeof localStorage !== 'undefined') {
		localStorage.setItem(DEBUG_PROFILE_KEY, id.trim().toLowerCase());
		window.dispatchEvent(new Event('demo-auth-change'));
	}
}

export function clearDebugProfile() {
	if (typeof localStorage !== 'undefined') localStorage.removeItem(DEBUG_PROFILE_KEY);
}
