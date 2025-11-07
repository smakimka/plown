<script lang="ts">
	let {
		players,
		activePlayer,
		myName
	}: {
		players: string[];
		activePlayer: string;
		myName: string;
	} = $props();
</script>

<aside class="player-list">
	<h3>Players</h3>
	<ul>
		{#each players as name}
			<li class:active={name === activePlayer} class:you={name === myName}>
				<span class="dot"></span>
				<span class="name">{name}</span>
				{#if name === myName}
					<span class="badge you-badge">you</span>
				{/if}
				{#if name === activePlayer}
					<span class="badge turn-badge">turn</span>
				{/if}
			</li>
		{/each}
	</ul>
</aside>

<style>
	.player-list {
		padding: 1rem;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	h3 {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	ul {
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	li {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.6rem;
		border-radius: var(--radius-sm);
		font-size: 0.9rem;
		color: var(--text-muted);
		transition: background 0.15s;
	}

	li.active {
		background: var(--accent-dim);
		color: var(--text);
	}

	li.you {
		font-weight: 500;
	}

	.dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: var(--text-dim);
		flex-shrink: 0;
	}

	li.active .dot {
		background: var(--accent-light);
		box-shadow: 0 0 6px var(--accent-glow);
	}

	.name {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.badge {
		font-size: 0.65rem;
		font-weight: 700;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		padding: 0.15rem 0.4rem;
		border-radius: 4px;
		flex-shrink: 0;
	}

	.you-badge {
		background: rgba(255, 255, 255, 0.06);
		color: var(--text-dim);
	}

	.turn-badge {
		background: var(--accent-dim);
		color: var(--accent-light);
	}
</style>
