<script lang="ts">
	import type { Snippet } from 'svelte';
	let { children }: { children: Snippet } = $props();
</script>

{@render children()}

<style>
	:global(*) {
		box-sizing: border-box;
		margin: 0;
		padding: 0;
	}

	:global(:root) {
		--bg: #0d0d14;
		--bg-card: #16161f;
		--bg-card-hover: #1e1e2d;
		--bg-input: #1a1a26;
		--border: rgba(255, 255, 255, 0.07);
		--border-focus: rgba(139, 92, 246, 0.5);
		--accent: #7c3aed;
		--accent-light: #a78bfa;
		--accent-dim: rgba(124, 58, 237, 0.15);
		--accent-glow: rgba(124, 58, 237, 0.35);
		--danger: #ef4444;
		--danger-dim: rgba(239, 68, 68, 0.15);
		--success: #10b981;
		--text: #f1f5f9;
		--text-muted: #94a3b8;
		--text-dim: #475569;
		--radius: 12px;
		--radius-sm: 8px;
		--font: 'Inter', system-ui, -apple-system, sans-serif;
	}

	:global(body) {
		background: var(--bg);
		color: var(--text);
		font-family: var(--font);
		min-height: 100vh;
		-webkit-font-smoothing: antialiased;
	}

	:global(button) {
		cursor: pointer;
		font-family: inherit;
	}

	:global(input) {
		font-family: inherit;
	}

	:global(::-webkit-scrollbar) {
		width: 6px;
	}
	:global(::-webkit-scrollbar-track) {
		background: transparent;
	}
	:global(::-webkit-scrollbar-thumb) {
		background: var(--border);
		border-radius: 3px;
	}
</style>
