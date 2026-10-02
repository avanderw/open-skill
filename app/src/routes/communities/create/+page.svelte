<script lang="ts">
	import { goto } from '$app/navigation';
	import { base } from '$app/paths';
	import { api } from '$lib/api';

	let name = $state('');
	let mode = $state<'open' | 'approval'>('open');
	let draws = $state(true);
	let error = $state('');
	let submitting = $state(false);
	let createdCode = $state('');

	async function submit() {
		if (!name.trim()) {
			error = 'Give the community a name.';
			return;
		}
		submitting = true;
		error = '';
		try {
			const c = await api.createCommunity({ name: name.trim(), mode, draws });
			createdCode = c.code;
			goto(`${base}/communities/${encodeURIComponent(c.id)}`);
		} catch (e) {
			error = (e as Error).message;
		} finally {
			submitting = false;
		}
	}
</script>

<div class="page-wrap">
	<div class="page-inner">
		<div class="page-header">
			<div>
				<p class="label-caps">Create flow</p>
				<h1>Create Community</h1>
				<p>Set up a new community, choose how members join, and preview the invite code.</p>
			</div>
			<a class="btn btn-secondary" href={`${base}/communities`} style="padding:10px 16px; font-size:12px;">My communities</a>
		</div>

		<div class="grid-create">
			<section class="card form">
				<p class="label-caps">Setup</p>
				<div class="field">
					<label for="community-name">Community name</label>
					<input id="community-name" class="text-input" bind:value={name} placeholder="e.g. Open Cup Johannesburg" />
				</div>

				<div class="inner-surface">
					<div class="row">
						<div>
							<p class="row-title">Draws allowed</p>
							<p class="row-sub">Allow ties when a match ends level.</p>
						</div>
						<input type="checkbox" class="themed-checkbox" bind:checked={draws} />
					</div>
					<div class="row" style="border-top:1px solid var(--theme-border); padding-top:14px; margin-top:14px;">
						<div>
							<p class="row-title">Join mode</p>
							<p class="row-sub">Players join instantly or request approval.</p>
						</div>
						<div class="seg">
							<button
								class:active={mode === 'open'}
								class="seg-btn"
								onclick={() => (mode = 'open')}
							>Open</button>
							<button
								class:active={mode === 'approval'}
								class="seg-btn"
								onclick={() => (mode = 'approval')}
							>Approval</button>
						</div>
					</div>
				</div>

				{#if error}<p class="alert-err">{error}</p>{/if}

				<div style="display:flex; gap:12px; flex-wrap:wrap;">
					<a class="btn btn-secondary" href={`${base}/communities`}>Cancel</a>
					<button class="btn btn-primary" disabled={submitting} onclick={submit}>
						{submitting ? 'Creating…' : 'Create community'}
					</button>
				</div>
			</section>

			<aside class="card preview">
				<p class="label-caps">Outcome</p>
				<h3 style="font-size:20px; font-weight:700; margin-bottom:6px;">Community preview</h3>
				<p class="muted" style="font-size:13px; margin-bottom:18px;">
					A unique invite code is generated when the community is created, and members can join with it.
				</p>
				<div class="preview-box">
					<div class="preview-avatar">{name.trim() ? name.trim()[0].toUpperCase() : '?'}</div>
					<div class="preview-name">{name.trim() || 'Your Community'}</div>
					<div class="preview-mode">{mode === 'open' ? 'Open join mode' : 'Approval required'}</div>
					<div class="preview-code">
						<span class="muted">Code</span>
						<span>{createdCode || 'Generated on create'}</span>
					</div>
				</div>
			</aside>
		</div>
	</div>
</div>

<style>
	.page-inner {
		display: flex;
		flex-direction: column;
	}
	.page-header {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 24px;
		margin-bottom: 28px;
		flex-wrap: wrap;
	}
	.page-header h1 {
		font-size: 36px;
		font-weight: 700;
	}
	.page-header p:last-child {
		color: var(--theme-text-muted);
		max-width: 620px;
		margin-top: 8px;
	}
	.grid-create {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(280px, 360px);
		gap: 24px;
		align-items: start;
	}
	.form,
	.preview {
		padding: 28px;
	}
	.field {
		margin-bottom: 18px;
	}
	.field label {
		display: block;
		font-size: 13px;
		font-weight: 700;
		color: var(--theme-text-muted);
		margin-bottom: 8px;
	}
	.inner-surface {
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid var(--theme-border);
		border-radius: 20px;
		padding: 18px;
		margin-bottom: 18px;
	}
	.row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
	}
	.row-title {
		font-size: 14px;
		font-weight: 700;
		color: var(--theme-text);
		margin-bottom: 4px;
	}
	.row-sub {
		font-size: 12px;
		color: var(--theme-text-muted);
	}
	.themed-checkbox {
		width: 22px;
		height: 22px;
		accent-color: var(--theme-accent);
		cursor: pointer;
	}
	.seg {
		display: flex;
		gap: 6px;
	}
	.seg-btn {
		padding: 8px 14px;
		border-radius: var(--radius-pill);
		border: 1px solid var(--theme-border);
		background: rgba(255, 255, 255, 0.05);
		color: var(--theme-text-muted);
		font-size: 12px;
		font-weight: 700;
		cursor: pointer;
		transition: all 0.15s ease;
	}
	.seg-btn.active {
		border-color: rgba(240, 90, 26, 0.5);
		background: rgba(240, 90, 26, 0.14);
		color: #ffd9c9;
	}
	.alert-err {
		color: var(--danger);
		background: var(--danger-bg);
		border: 1px solid var(--danger-border);
		border-radius: 14px;
		padding: 12px 14px;
		font-size: 13px;
		font-weight: 600;
		margin-bottom: 16px;
	}
	.preview-box {
		padding: 20px;
		background: linear-gradient(135deg, rgba(240, 90, 26, 0.08), rgba(255, 255, 255, 0.03));
		border: 1px solid var(--theme-border);
		border-radius: 20px;
	}
	.preview-avatar {
		width: 64px;
		height: 64px;
		margin-bottom: 14px;
		background: linear-gradient(135deg, #b83c08, var(--brand-primary));
		border-radius: 18px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
		font-size: 24px;
		font-weight: 700;
	}
	.preview-name {
		font-size: 18px;
		font-weight: 700;
		color: var(--theme-text);
		margin-bottom: 4px;
	}
	.preview-mode {
		font-size: 13px;
		color: var(--theme-text-muted);
		margin-bottom: 14px;
	}
	.preview-code {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 14px;
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid var(--theme-border);
		border-radius: 16px;
		font-size: 13px;
		font-weight: 700;
		color: var(--theme-accent);
		letter-spacing: 0.08em;
	}
	@media (max-width: 760px) {
		.grid-create {
			grid-template-columns: 1fr;
		}
	}
</style>