<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { base } from '$app/paths';
	import { api, type Community, type CommunityJoinRequest, type CommunityPlayerClaim } from '$lib/api';
	let { data } = $props();
	const community = data.community as Community;
	let name = $state(community.name);
	let mode = $state<'open' | 'approval'>(community.mode === 'approval' ? 'approval' : 'open');
	let draws = $state(community.draws);
	let message = $state(''); let saving = $state(false); let copied = $state(false);
	let tab = $state<'general' | 'members' | 'join' | 'integrations'>('general');
	let members = $state<CommunityJoinRequest[]>([]); let discord = $state(''); let webhook = $state('');
	const claims = $derived(data.claims as CommunityPlayerClaim[]);
	let pendingMerge = $state<string | null>(null);
	let claimBusy = $state(false);
	let claimError = $state('');
	let claimMessage = $state('');
	function claimKey(claim: CommunityPlayerClaim) { return `${claim.playerId}:${claim.userId}`; }
	async function reviewClaim(claim: CommunityPlayerClaim, approve: boolean) {
		if (claimBusy) return;
		claimBusy = true;
		claimError = '';
		claimMessage = '';
		try {
			if (approve) await api.approvePlayerClaim(claim.playerId, claim.userId);
			else await api.rejectPlayerClaim(claim.playerId, claim.userId);
			pendingMerge = null;
			await invalidateAll();
			claimMessage = approve ? `${claim.handle} merged into ${claim.targetHandle}.` : `Claim for ${claim.handle} rejected.`;
		} catch (err) {
			claimError = (err as Error).message;
		} finally {
			claimBusy = false;
		}
	}
	async function saveGeneral() { saving=true; message=''; try { const {api}=await import('$lib/api'); await api.updateCommunity(community.id,{name:name.trim(),mode,draws}); message='Community settings saved.';} catch(e){message=(e as Error).message;} finally{saving=false;} }
	let confirmName = $state(''); let deleting = $state(false);
	async function deleteCommunity() { if (confirmName !== community.name) return; deleting=true; message=''; try { const {api}=await import('$lib/api'); await api.deleteCommunity(community.id); await goto(`${base}/communities`); } catch(e){message=(e as Error).message; deleting=false;} }
	async function saveIntegration() { saving=true; message=''; try { const {api}=await import('$lib/api'); const saved=await api.updateIntegration(community.id,{discord,webhook}); discord=saved.discord; webhook=saved.webhook; message='Integration settings saved.';} catch(e){message=(e as Error).message;} finally{saving=false;} }
	async function loadMembers(){try{const {api}=await import('$lib/api'); members=await api.communityMembers(community.id);}catch(e){message=(e as Error).message}}
	async function loadIntegration(){try{const {api}=await import('$lib/api'); const i=await api.integration(community.id); discord=i.discord; webhook=i.webhook;}catch(e){message=(e as Error).message}}
	function selectTab(next:typeof tab){tab=next;message='';if(next==='members')loadMembers();if(next==='integrations')loadIntegration()}
	async function copyCode(){await navigator.clipboard?.writeText(community.code);copied=true}
	let shared = $state(false);
	async function shareInvite(){
		const url = `${window.location.origin}${base}/communities?code=${encodeURIComponent(community.code)}`;
		const text = `Join ${community.name} on Open-Skill — use code ${community.code}`;
		if (navigator.share) { try { await navigator.share({ title: community.name, text, url }); } catch { /* user cancelled */ } return; }
		try { await navigator.clipboard.writeText(url); shared = true; setTimeout(() => (shared = false), 2000); } catch { /* clipboard unavailable */ }
	}
</script>
<svelte:head><title>{community.name} settings — Open-Skill</title></svelte:head>
<div class="page-wrap"><div class="page-inner"><div class="nav"><button class="btn btn-ghost" onclick={()=>goto(`${base}/leaderboard?communityId=${encodeURIComponent(community.id)}`)}>← Leaderboard</button></div><section class="card hero"><div><p class="label-caps">Community settings</p><h1>{community.name}</h1><p>{community.members} members · Invite code <strong>{community.code}</strong></p></div><div style="display:flex; gap:10px; flex-wrap:wrap;"><button class="btn btn-secondary" onclick={copyCode}>{copied?'Copied':'Copy invite code'}</button><button class="btn btn-primary" onclick={shareInvite}>{shared?'Shared!':'Share invite'}</button></div></section><nav class="tabs" aria-label="Community settings"><button class:active={tab==='general'} onclick={()=>selectTab('general')}>General</button><button class:active={tab==='members'} onclick={()=>selectTab('members')}>Members</button><button class:active={tab==='join'} onclick={()=>selectTab('join')}>Join mode</button><button class:active={tab==='integrations'} onclick={()=>selectTab('integrations')}>Integrations</button></nav><section class="card panel" aria-live="polite">{#if tab==='general'}<p class="label-caps">General</p><h2>Community details</h2><label for="name">Name</label><input id="name" class="text-input" bind:value={name}/><label class="check"><input type="checkbox" bind:checked={draws}/> Draws allowed</label><button class="btn btn-primary" disabled={saving} onclick={saveGeneral}>{saving?'Saving…':'Save settings'}</button>{#if community.membership==='owner'}<div class="danger-zone"><h3>Delete community</h3><p>This permanently deletes the community, its members, matches and ratings. This cannot be undone.</p><label for="confirm-name">Type <strong>{community.name}</strong> to confirm</label><input id="confirm-name" class="text-input" autocomplete="off" bind:value={confirmName}/><button class="btn btn-danger" disabled={deleting||confirmName!==community.name} onclick={deleteCommunity}>{deleting?'Deleting…':'Delete this community'}</button></div>{/if}{:else if tab==='members'}<p class="label-caps">Members</p><h2>Pending requests</h2>{#if members.length===0}<p class="muted">No pending requests.</p>{:else}{#each members as member}<div class="list-row">{JSON.stringify(member)}<span>Pending</span>{#if community.membership === 'owner'}<button class="btn btn-primary" onclick={async () => { const parsed = member as { userId?: string }; if (parsed.userId) { const { api } = await import('$lib/api'); await api.approveMember(community.id, parsed.userId); await loadMembers(); } }}>Approve</button>{/if}</div>{/each}{/if}{:else if tab==='join'}<p class="label-caps">Join mode</p><h2>Who can join?</h2><div class="join-options"><button class:active={mode==='open'} onclick={()=>mode='open'}><strong>Open joining</strong><span>Players join immediately.</span></button><button class:active={mode==='approval'} onclick={()=>mode='approval'}><strong>Approval required</strong><span>Players wait for owner approval.</span></button></div><button class="btn btn-primary" disabled={saving} onclick={saveGeneral}>{saving?'Saving…':'Save join mode'}</button>{:else}<p class="label-caps">Integrations</p><h2>Preview connections</h2><label for="discord">Discord URL</label><input id="discord" class="text-input" bind:value={discord}/><label for="webhook">Webhook URL</label><input id="webhook" class="text-input" bind:value={webhook}/><button class="btn btn-primary" disabled={saving} onclick={saveIntegration}>{saving?'Saving…':'Save integrations'}</button>{/if}{#if message}<p class="status" role="status">{message}</p>{/if}</section></div></div>
{#if data.canManage}
	<section class="claim-queue" aria-label="Player claims">
		<div class="claim-heading"><h2>Player claims</h2><span class="muted">{claims.length} pending</span></div>
		{#if claims.length === 0}
			<p class="muted">No pending player claims.</p>
		{:else}
			<ul class="claim-list">
				{#each claims as claim (claimKey(claim))}
					<li>
						<div class="claim-details"><a href={`${base}/profile/${encodeURIComponent(claim.playerId)}`}>{claim.handle}</a> <span class="unclaimed-badge">Unclaimed</span><p>Claimed by {claim.display}</p><p class="muted">Merge into <a href={`${base}/profile/${encodeURIComponent(claim.targetPlayerId)}`}>{claim.targetHandle}</a></p></div>
						<div class="claim-buttons"><button class="btn btn-primary" disabled={claimBusy} onclick={() => { pendingMerge = claimKey(claim); claimError = ''; }} aria-label={`Approve and merge ${claim.handle}`}>Approve and merge</button><button class="btn btn-ghost" disabled={claimBusy} onclick={() => reviewClaim(claim, false)} aria-label={`Reject claim for ${claim.handle}`}>Reject</button></div>
						{#if pendingMerge === claimKey(claim)}
							<div class="merge-confirm"><p>Merge {claim.handle} into {claim.targetHandle}? Match history transfers, ratings are recalculated, and the separate player is removed.</p><div class="claim-buttons"><button class="btn btn-primary" disabled={claimBusy} onclick={() => reviewClaim(claim, true)}>{claimBusy ? 'Merging...' : 'Confirm merge'}</button><button class="btn btn-ghost" disabled={claimBusy} onclick={() => (pendingMerge = null)}>Cancel</button></div></div>
						{/if}
					</li>
				{/each}
			</ul>
		{/if}
		{#if claimError}<p class="claim-error" role="alert">{claimError}</p>{/if}
		{#if claimMessage}<p role="status">{claimMessage}</p>{/if}
	</section>
{/if}
<style>
.claim-queue{width:min(980px,calc(100% - 40px));margin:0 auto 48px;padding-top:24px;border-top:1px solid var(--theme-border);display:grid;gap:16px}.claim-heading{display:flex;justify-content:space-between;align-items:baseline;gap:12px;flex-wrap:wrap}.claim-heading h2{font-size:22px}.claim-list{list-style:none;display:grid;gap:16px}.claim-list li{display:flex;justify-content:space-between;align-items:center;gap:16px;flex-wrap:wrap;padding-bottom:16px;border-bottom:1px solid var(--theme-border)}.claim-details{min-width:0;overflow-wrap:anywhere;flex:1 1 240px}.claim-details>a{font-weight:700}.claim-details p{font-size:13px;margin-top:4px}.claim-buttons{display:flex;gap:8px;flex-wrap:wrap}.claim-buttons .btn{white-space:normal}.merge-confirm{flex-basis:100%;display:grid;gap:12px;padding:16px 0;font-size:14px}.claim-error{color:var(--danger)}
.page-inner{display:flex;flex-direction:column;gap:18px;max-width:980px}.nav{display:flex}.hero,.panel{padding:28px}.hero{display:flex;justify-content:space-between;gap:18px;align-items:flex-start;flex-wrap:wrap}.hero h1{font-size:clamp(30px,5vw,48px);line-height:1.05;margin-bottom:10px}.hero p{color:var(--theme-text-muted)}.label-caps{font-size:12px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:var(--theme-accent);margin-bottom:8px}.tabs{display:flex;gap:8px;flex-wrap:wrap}.tabs button,.join-options button{padding:10px 14px;border:1px solid var(--theme-border);border-radius:999px;background:rgba(255,255,255,.04);color:var(--theme-text-muted);cursor:pointer}.tabs button.active,.join-options button.active{border-color:var(--theme-accent);background:rgba(240,90,26,.14);color:#ffd9c9}.panel h2{font-size:24px;margin-bottom:18px}.panel label{display:block;margin:12px 0 8px;font-size:13px;font-weight:700}.text-input{width:100%;margin-bottom:12px}.check{display:flex!important;gap:8px;align-items:center}.check input{accent-color:var(--theme-accent)}.join-options{display:grid;gap:10px;margin-bottom:18px}.join-options button{display:flex;flex-direction:column;align-items:flex-start;gap:5px;text-align:left;border-radius:14px}.join-options span{font-size:12px;color:var(--theme-text-muted)}.list-row{padding:12px;border-bottom:1px solid var(--theme-border);display:flex;justify-content:space-between}.danger-zone{margin-top:28px;padding:18px;border:1px solid var(--theme-danger,#c0392b);border-radius:12px}.danger-zone h3{margin-bottom:8px}.danger-zone p{color:var(--theme-text-muted);font-size:13px}.status{margin-top:16px;padding:10px;border-radius:10px;background:var(--ok-bg);color:var(--ok)}
</style>
        