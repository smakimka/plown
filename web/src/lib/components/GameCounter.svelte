<script lang="ts">
	let { count, total }: { count: number; total: number } = $props();

	// R=88, viewBox 200x200, circumference ≈ 553
	const R = 88;
	const C = +(2 * Math.PI * R).toFixed(2);

	let dashOffset = $derived(total > 0 ? C * (1 - count / total) : 0);

	const DIGITS = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9];

	// Split a number into per-character parts, keyed by position from right
	// so that DOM elements are stable as digit count changes.
	function digitParts(n: number) {
		const s = n.toLocaleString('en-US');
		let digitIdx = 0;
		const totalDigits = s.replace(/,/g, '').length;
		return s.split('').map((c) => {
			if (c >= '0' && c <= '9') {
				const key = totalDigits - 1 - digitIdx;
				digitIdx++;
				return { char: c, isDigit: true as const, key };
			}
			return { char: c, isDigit: false as const, key: `s${digitIdx}` };
		});
	}

	let parts = $derived(digitParts(count));
</script>

<div class="wrap">
	<svg viewBox="0 0 200 200" class="ring" aria-hidden="true">
		<defs>
			<linearGradient id="cg-grad" x1="0" y1="0" x2="0" y2="1">
				<stop offset="0%" style="stop-color: var(--accent-light)" />
				<stop offset="100%" style="stop-color: var(--accent)" />
			</linearGradient>
		</defs>
		<!-- Track -->
		<circle cx="100" cy="100" r={R} fill="none" stroke-width="8" style="stroke: var(--border)" />
		<!-- Drain ring -->
		<circle
			cx="100"
			cy="100"
			r={R}
			fill="none"
			stroke="url(#cg-grad)"
			stroke-width="8"
			stroke-linecap="round"
			stroke-dasharray={C}
			stroke-dashoffset={dashOffset}
			transform="rotate(-90 100 100)"
			style="transition: stroke-dashoffset 0.8s cubic-bezier(0.4, 0, 0.2, 1)"
		/>
	</svg>

	<div class="center">
		<div class="odometer">
			{#each parts as part (part.key)}
				{#if part.isDigit}
					<div class="slot">
						<div class="drum" style="transform: translateY(-{+part.char * 10}%)">
							{#each DIGITS as d}
								<div class="drum-item">{d}</div>
							{/each}
						</div>
					</div>
				{:else}
					<span class="sep">{part.char}</span>
				{/if}
			{/each}
		</div>
		<div class="label">games remaining</div>
	</div>
</div>

<style>
	.wrap {
		position: relative;
		width: 220px;
		height: 220px;
		flex-shrink: 0;
	}

	.ring {
		width: 100%;
		height: 100%;
		filter: drop-shadow(0 0 10px rgba(124, 58, 237, 0.35));
	}

	.center {
		position: absolute;
		inset: 0;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.35rem;
	}

	/* Odometer */
	.odometer {
		display: flex;
		align-items: center;
		font-size: 2.4rem;
		font-weight: 800;
		letter-spacing: -0.04em;
		line-height: 1;
		font-variant-numeric: tabular-nums;
	}

	.slot {
		height: 1em;
		width: 1ch;
		overflow: hidden;
	}

	.drum {
		display: flex;
		flex-direction: column;
		transition: transform 0.45s cubic-bezier(0.34, 1.56, 0.64, 1);
	}

	.drum-item {
		height: 1em;
		display: flex;
		align-items: center;
		justify-content: center;
		background: linear-gradient(135deg, #fff 40%, var(--accent-light));
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.sep {
		color: var(--text-dim);
		font-size: 2.4rem;
		font-weight: 800;
		line-height: 1;
		margin: 0 0.02em;
	}

	.label {
		font-size: 0.72rem;
		color: var(--text-dim);
		letter-spacing: 0.1em;
		text-transform: uppercase;
	}
</style>
