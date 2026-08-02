<script lang="ts">
	import { onMount } from 'svelte';

	let {
		tagName,
		included,
		delta,
		onDismiss
	}: {
		tagName: string;
		included: boolean;
		delta: number;
		onDismiss: () => void;
	} = $props();

	const DURATION = 4000;
	// delta > 0 = pool shrank, delta < 0 = pool grew
	const shrank = $derived(delta > 0);
	const absDelta = $derived(Math.abs(delta));

	onMount(() => {
		const t = setTimeout(onDismiss, DURATION);
		return () => clearTimeout(t);
	});
</script>

<div class="toast" class:shrank class:grew={!shrank && delta !== 0} class:neutral={delta === 0}>
	<div class="row">
		<span class="dot" class:inc={included} class:exc={!included}></span>
		<span class="name">{tagName}</span>
	</div>
	{#if delta !== 0}
		<div class="delta" class:shrank class:grew={!shrank}>
			<span class="sign">{shrank ? '−' : '+'}{absDelta.toLocaleString()}</span>
			<span class="sub">games</span>
		</div>
	{/if}
	<div class="bar-track">
		<div
			class="bar-fill"
			class:inc={included}
			class:exc={!included}
			style="animation-duration: {DURATION}ms"
		></div>
	</div>
</div>

<style>
	.toast {
		width: 240px;
		background: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		overflow: hidden;
		pointer-events: auto;
	}

	.toast.shrank {
		border-color: rgba(239, 68, 68, 0.2);
	}

	.toast.grew {
		border-color: rgba(16, 185, 129, 0.2);
	}

	.row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.6rem 0.75rem 0.25rem;
	}

	.dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.dot.exc { background: var(--danger); }
	.dot.inc { background: #10b981; }

	.name {
		flex: 1;
		font-size: 0.84rem;
		font-weight: 600;
		color: var(--text);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.delta {
		display: flex;
		align-items: baseline;
		gap: 0.3rem;
		padding: 0 0.75rem 0.5rem;
	}

	.sign {
		font-size: 0.95rem;
		font-weight: 700;
		font-variant-numeric: tabular-nums;
	}

	.delta.shrank .sign {
		color: #f87171;
	}

	.delta.grew .sign {
		color: #34d399;
	}

	.sub {
		font-size: 0.72rem;
		color: var(--text-dim);
	}

	/* Timer bar */
	.bar-track {
		height: 2px;
		background: var(--border);
	}

	.bar-fill {
		height: 100%;
		animation: shrink linear forwards;
	}

	.bar-fill.exc { background: var(--danger); }
	.bar-fill.inc { background: #10b981; }

	@keyframes shrink {
		from { width: 100%; }
		to   { width: 0%; }
	}
</style>
