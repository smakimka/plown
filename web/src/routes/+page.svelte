<script lang="ts">
	import { goto } from '$app/navigation';

	let joining = $state('');
	let creating = $state(false);
	let error = $state('');

	async function createLobby() {
		creating = true;
		error = '';
		try {
			const res = await fetch('/api/lobby', { method: 'POST' });
			if (!res.ok) throw new Error('Failed to create lobby');
			const { lobby_id } = await res.json();
			goto(`/lobby/${lobby_id}`);
		} catch {
			error = 'Could not create a lobby. Is the server running?';
			creating = false;
		}
	}

	function joinLobby() {
		const id = joining.trim().toLowerCase();
		if (id) goto(`/lobby/${id}`);
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') joinLobby();
	}
</script>

<main>
	<div class="hero">
		<div class="logo">🎲</div>
		<h1>Plown</h1>
		<p class="tagline">Pick a game with your friends — no arguments, just vibes.</p>

		<div class="card">
			<button class="btn-primary" onclick={createLobby} disabled={creating}>
				{creating ? 'Creating…' : 'Create Lobby'}
			</button>

			<div class="divider"><span>or join one</span></div>

			<div class="join-row">
				<input
					type="text"
					placeholder="Lobby code"
					bind:value={joining}
					onkeydown={onKeydown}
					maxlength={6}
					spellcheck={false}
				/>
				<button class="btn-secondary" onclick={joinLobby} disabled={!joining.trim()}>
					Join
				</button>
			</div>

			{#if error}
				<p class="error">{error}</p>
			{/if}
		</div>
	</div>
</main>

<style>
	main {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
		background: radial-gradient(ellipse at 50% 0%, rgba(124, 58, 237, 0.12) 0%, transparent 60%);
	}

	.hero {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		text-align: center;
		width: 100%;
		max-width: 400px;
	}

	.logo {
		font-size: 4rem;
		line-height: 1;
		filter: drop-shadow(0 0 24px rgba(124, 58, 237, 0.6));
	}

	h1 {
		font-size: 3rem;
		font-weight: 800;
		letter-spacing: -0.03em;
		background: linear-gradient(135deg, #fff 30%, var(--accent-light));
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.tagline {
		color: var(--text-muted);
		font-size: 1rem;
		margin-bottom: 0.5rem;
	}

	.card {
		background: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 1.75rem;
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.btn-primary {
		width: 100%;
		padding: 0.85rem 1.5rem;
		background: var(--accent);
		color: #fff;
		border: none;
		border-radius: var(--radius-sm);
		font-size: 1rem;
		font-weight: 600;
		transition: opacity 0.15s, box-shadow 0.15s;
	}

	.btn-primary:hover:not(:disabled) {
		opacity: 0.88;
		box-shadow: 0 0 20px var(--accent-glow);
	}

	.btn-primary:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.divider {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		color: var(--text-dim);
		font-size: 0.8rem;
	}

	.divider::before,
	.divider::after {
		content: '';
		flex: 1;
		height: 1px;
		background: var(--border);
	}

	.join-row {
		display: flex;
		gap: 0.5rem;
	}

	input {
		flex: 1;
		padding: 0.75rem 1rem;
		background: var(--bg-input);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		color: var(--text);
		font-size: 1rem;
		outline: none;
		text-transform: lowercase;
		letter-spacing: 0.1em;
		transition: border-color 0.15s;
	}

	input:focus {
		border-color: var(--border-focus);
	}

	.btn-secondary {
		padding: 0.75rem 1.25rem;
		background: var(--bg-card-hover);
		color: var(--text);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		font-size: 0.9rem;
		font-weight: 500;
		transition: background 0.15s;
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--border);
	}

	.btn-secondary:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.error {
		color: var(--danger);
		font-size: 0.85rem;
		text-align: center;
	}
</style>
