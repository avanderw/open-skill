<script lang="ts">
	import { onMount } from 'svelte';
	import type { Community } from '$lib/api';
	import { AUTH_CHANGE_EVENT, authReady, isAuthenticated } from '$lib/auth';
	import { base } from '$app/paths';
	let { data } = $props();
	const initial = data.communities as Community[];
	let query = $state('');
	let authenticated = $state(false);
	const communities = $derived(initial.filter((community) => community.name.toLowerCase().includes(query.trim().toLowerCase())));
	onMount(() => {
		const sync = () => (authenticated = isAuthenticated());
		authReady().then(sync);
		window.addEventListener(AUTH_CHANGE_EVENT, sync);
		return () => window.removeEventListener(AUTH_CHANGE_EVENT, sync);
	});
</script>
<svelte:head><title>Browse communities — Open-Skill</title></svelte:head>
<div class="page-wrap"><div class="page-inner"><div class="top"><a class="btn btn-ghost" href={`${base}/communities`}>← My Communities</a></div><section class="card hero"><div class="hero-head"><div><p class="label-caps">Find your next competition</p><h1>Browse Communities</h1><p>Explore communities you can join and open their leaderboard before deciding.</p></div>{#if authenticated}<a class="btn btn-primary" href={`${base}/communities/start`}>Create community</a>{/if}</div><input class="text-input" aria-label="Search communities" placeholder="Search communities…" bind:value={query}/></section><section class="grid">{#if communities.length===0}<div class="card empty"><h2>No communities found</h2><p class="muted">Try another search{#if authenticated}, or start your own{/if}.</p>{#if authenticated}<a class="btn btn-primary" href={`${base}/communities/start`}>Create community</a>{/if}</div>{:else}{#each communities as community}<a class="card community" href={`${base}/leaderboard?communityId=${encodeURIComponent(community.id)}`}><div class="community-head"><div><p class="label-caps">{community.membership === 'non_member' ? 'Available to join' : community.membership}</p><h2>{community.name}</h2></div><span class="mode">{community.mode}</span></div><p class="muted">{community.members} members · {community.draws ? 'Draws allowed' : 'No draws'}</p>{#if community.membership === 'pending'}<span class="pending">Request pending</span>{:else if community.membership === 'member' || community.membership === 'owner'}<span class="joined">Already joined</span>{:else}<span class="join">View leaderboard and join →</span>{/if}</a>{/each}{/if}</section></div></div>
<style>
.page-inner{display:flex;flex-direction:column;gap:18px;max-width:1120px}.top{display:flex}.hero{padding:28px}.hero-head{display:flex;align-items:flex-start;justify-content:space-between;gap:18px;flex-wrap:wrap}.hero h1{font-size:clamp(32px,5vw,48px);line-height:1.05;margin-bottom:10px}.hero p:not(.label-caps){color:var(--theme-text-muted);margin-bottom:18px}.label-caps{font-size:12px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:var(--theme-accent);margin-bottom:8px}.hero .text-input{max-width:520px}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:16px}.community{padding:22px;color:var(--theme-text);transition:transform .15s,border-color .15s}.community:hover{transform:translateY(-2px);border-color:var(--theme-accent)}.community-head{display:flex;justify-content:space-between;gap:12px}.community h2{font-size:20px;margin-bottom:8px}.mode{font-size:11px;color:var(--theme-accent);text-transform:uppercase}.join,.joined,.pending{display:inline-block;margin-top:18px;font-size:13px;font-weight:800}.join{color:var(--theme-accent)}.joined{color:var(--ok)}.pending{color:#ffd166}.empty{padding:48px;text-align:center;grid-column:1/-1}.empty .btn{margin-top:18px}
</style>
