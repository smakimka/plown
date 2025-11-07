<script lang="ts">
	import type { Tag, TagCount, LabelCount } from '$lib/types';

	let {
		tags,
		includedIds,
		excludedIds,
		isMyTurn,
		onTag,
		onRemoveTag,
		rawTagCounts,
		labels,
		labelCounts,
		includedLabels,
		excludedLabels,
		onLabel,
		onRemoveLabel,
		count,
		freeCount,
		onlyFree,
		onFree
	}: {
		tags: Tag[];
		includedIds: number[];
		excludedIds: number[];
		isMyTurn: boolean;
		onTag: (id: number, include: boolean) => void;
		onRemoveTag: (id: number) => void;
		rawTagCounts: TagCount[];
		labels: string[];
		labelCounts: LabelCount[];
		includedLabels: string[];
		excludedLabels: string[];
		onLabel: (value: string, is: boolean) => void;
		onRemoveLabel: (value: string) => void;
		count: number;
		freeCount: number;
		onlyFree: boolean | null;
		onFree: (isFree: boolean | null) => void;
	} = $props();

	let search = $state('');

	const includedSet = $derived(new Set(includedIds));
	const excludedSet = $derived(new Set(excludedIds));
	const includedLabelSet = $derived(new Set(includedLabels));
	const excludedLabelSet = $derived(new Set(excludedLabels));

	// id → name lookup from the tags HTTP response
	const nameMap = $derived.by(() => {
		const m = new Map<number, string>();
		for (const t of tags) m.set(t.id, t.name);
		return m;
	});

	// Label count lookup
	const labelCountMap = $derived.by(() => {
		const m = new Map<string, number>();
		for (const lc of labelCounts) m.set(lc.value, lc.count);
		return m;
	});

	// Available tags (not yet applied), sorted by count from backend
	interface TagEntry { id: number; name: string; count: number }
	const availableTags = $derived<TagEntry[]>(
		rawTagCounts
			.filter((tc) => !excludedSet.has(tc.id) && !includedSet.has(tc.id) && nameMap.has(tc.id))
			.map((tc) => ({ id: tc.id, name: nameMap.get(tc.id)!, count: tc.count }))
	);

	const filtered = $derived(
		availableTags.filter(
			(t) => !search.trim() || t.name.toLowerCase().includes(search.trim().toLowerCase())
		)
	);

	// Label chips: all labels with their current state
	type LabelState = 'neutral' | 'included' | 'excluded';
	interface LabelEntry { value: string; count: number | null; state: LabelState }
	const labelEntries = $derived<LabelEntry[]>(
		labels.map((l) => ({
			value: l,
			count: labelCountMap.get(l) ?? null,
			state: includedLabelSet.has(l) ? 'included' : excludedLabelSet.has(l) ? 'excluded' : 'neutral'
		}))
	);

	// Tags already applied (for summary)
	const includedTags = $derived(tags.filter((t) => includedSet.has(t.id)));
	const excludedTags = $derived(tags.filter((t) => excludedSet.has(t.id)));

	const hasAppliedFilters = $derived(
		includedTags.length > 0 || excludedTags.length > 0 || onlyFree !== null
	);

	const paidCount = $derived(count - freeCount);

	// --- Popover state ---
	type PopoverTarget =
		| { kind: 'tag'; id: number; name: string }
		| { kind: 'label'; value: string };
	let popover = $state<{ target: PopoverTarget; x: number; y: number } | null>(null);

	function openPopover(target: PopoverTarget, e: MouseEvent) {
		if (!isMyTurn) return;
		e.stopPropagation();
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		popover = { target, x: rect.left, y: rect.bottom + 4 };
	}

	function closePopover() {
		popover = null;
	}

	function popoverInclude() {
		if (!popover) return;
		if (popover.target.kind === 'tag') {
			onTag(popover.target.id, true);
		} else {
			onLabel(popover.target.value, true);
		}
		search = '';
		popover = null;
	}

	function popoverExclude() {
		if (!popover) return;
		if (popover.target.kind === 'tag') {
			onTag(popover.target.id, false);
		} else {
			onLabel(popover.target.value, false);
		}
		search = '';
		popover = null;
	}

	function clickTag(tag: TagEntry, e: MouseEvent) {
		if (!isMyTurn) return;
		openPopover({ kind: 'tag', id: tag.id, name: tag.name }, e);
	}

	function clickLabel(entry: LabelEntry, e: MouseEvent) {
		if (!isMyTurn) return;
		if (entry.state !== 'neutral') {
			onRemoveLabel(entry.value);
			return;
		}
		openPopover({ kind: 'label', value: entry.value }, e);
	}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<svelte:window onclick={() => popover && closePopover()} onkeydown={(e) => e.key === 'Escape' && closePopover()} />

<div class="panel">
	<!-- Applied filters strip -->
	{#if hasAppliedFilters}
		<div class="applied-strip">
			{#if onlyFree !== null}
				<span class="applied-chip {onlyFree ? 'include' : 'exclude'}">
					<span class="chip-dot {onlyFree ? 'include' : 'exclude'}"></span>
					<span class="chip-label">{onlyFree ? 'Free only' : 'Paid only'}</span>
					<button
						class="chip-x"
						disabled={!isMyTurn}
						onclick={() => onFree(null)}
						title="Remove filter"
					>✕</button>
				</span>
			{/if}
			{#each includedTags as t (t.id)}
				<span class="applied-chip include">
					<span class="chip-dot include"></span>
					<span class="chip-label">{t.name}</span>
					<button
						class="chip-x"
						disabled={!isMyTurn}
						onclick={() => onRemoveTag(t.id)}
						title="Remove filter"
					>✕</button>
				</span>
			{/each}
			{#each excludedTags as t (t.id)}
				<span class="applied-chip exclude">
					<span class="chip-dot exclude"></span>
					<span class="chip-label">{t.name}</span>
					<button
						class="chip-x"
						disabled={!isMyTurn}
						onclick={() => onRemoveTag(t.id)}
						title="Remove filter"
					>✕</button>
				</span>
			{/each}
		</div>
	{/if}

	<!-- Quick filters: Free + Review labels -->
	<div class="quick-filters" class:dimmed={!isMyTurn}>
		<!-- Free toggle -->
		<div class="filter-row">
			<span class="filter-label">Price</span>
			<div class="toggle-group">
				<button
					class="toggle-btn"
					class:active={onlyFree === null}
					disabled={!isMyTurn}
					onclick={() => onFree(null)}
				>All</button>
				<button
					class="toggle-btn"
					class:active={onlyFree === true}
					disabled={!isMyTurn}
					onclick={() => onFree(true)}
				>Free <span class="toggle-count">{freeCount}</span></button>
				<button
					class="toggle-btn"
					class:active={onlyFree === false}
					disabled={!isMyTurn}
					onclick={() => onFree(false)}
				>Paid <span class="toggle-count">{paidCount}</span></button>
			</div>
		</div>

		<!-- Review labels — state-per-chip toggle -->
		{#if labelEntries.length > 0}
			<div class="filter-row">
				<span class="filter-label">Reviews</span>
				<div class="label-chips">
					{#each labelEntries as entry}
						<button
							class="label-chip"
							class:label-included={entry.state === 'included'}
							class:label-excluded={entry.state === 'excluded'}
							class:label-neutral={entry.state === 'neutral'}
							disabled={!isMyTurn}
							onclick={(e) => clickLabel(entry, e)}
						>
							<span class="label-name" class:struck={entry.state === 'excluded'}>{entry.value}</span>
							{#if entry.count != null}
								<span class="tag-count">{entry.count.toLocaleString()}</span>
							{/if}
						</button>
					{/each}
				</div>
			</div>
		{/if}
	</div>

	<!-- Search -->
	<div class="search-wrap" class:dimmed={!isMyTurn}>
		<input
			type="text"
			placeholder="Search tags…"
			bind:value={search}
			disabled={!isMyTurn}
			spellcheck={false}
		/>
	</div>

	<!-- Tag grid -->
	<div class="tag-grid" class:dimmed={!isMyTurn}>
		{#if tags.length === 0}
			<p class="hint">Loading tags…</p>
		{:else if filtered.length === 0 && search.trim()}
			<p class="hint">No tags match "{search}"</p>
		{:else if filtered.length === 0}
			<p class="hint">All tags have been filtered.</p>
		{:else}
			{#each filtered as tag (tag.id)}
				<button
					class="tag-chip"
					disabled={!isMyTurn}
					onclick={(e) => clickTag(tag, e)}
				>
					<span class="tag-name">{tag.name}</span>
					<span class="tag-count">{tag.count.toLocaleString()}</span>
				</button>
			{/each}
		{/if}
	</div>
</div>

<!-- Popover (portaled to body-level to avoid overflow clipping) -->
{#if popover}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="popover-backdrop" onclick={closePopover}>
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="popover"
			style="left: {popover.x}px; top: {popover.y}px"
			onclick={(e) => e.stopPropagation()}
		>
			<span class="popover-name">
				{popover.target.kind === 'tag' ? popover.target.name : popover.target.value}
			</span>
			<div class="popover-actions">
				<button class="pop-btn pop-include" onclick={popoverInclude}>Include</button>
				<button class="pop-btn pop-exclude" onclick={popoverExclude}>Exclude</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.panel {
		display: flex;
		flex-direction: column;
		background: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		overflow: hidden;
		flex: 1;
		min-height: 0;
	}

	/* ── Applied filters strip ── */
	.applied-strip {
		display: flex;
		flex-wrap: wrap;
		gap: 0.4rem;
		padding: 0.65rem 0.75rem;
		border-bottom: 1px solid var(--border);
		background: rgba(255, 255, 255, 0.015);
	}

	.applied-chip {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.3rem 0.5rem 0.3rem 0.6rem;
		border-radius: 999px;
		font-size: 0.8rem;
		font-weight: 600;
		line-height: 1;
	}

	.applied-chip.include {
		background: rgba(16, 185, 129, 0.12);
		border: 1px solid rgba(16, 185, 129, 0.3);
		color: #34d399;
	}

	.applied-chip.exclude {
		background: rgba(239, 68, 68, 0.1);
		border: 1px solid rgba(239, 68, 68, 0.25);
		color: #f87171;
	}

	.applied-chip.exclude .chip-label {
		text-decoration: line-through;
		text-decoration-thickness: 1.5px;
	}

	.chip-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.chip-dot.include { background: #10b981; }
	.chip-dot.exclude { background: #ef4444; }

	.chip-label {
		white-space: nowrap;
	}

	.chip-x {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 16px;
		height: 16px;
		border-radius: 50%;
		border: none;
		background: rgba(255, 255, 255, 0.08);
		color: inherit;
		font-size: 0.6rem;
		line-height: 1;
		padding: 0;
		flex-shrink: 0;
		opacity: 0.6;
		transition: opacity 0.12s, background 0.12s;
	}

	.chip-x:hover:not(:disabled) {
		opacity: 1;
		background: rgba(255, 255, 255, 0.15);
	}

	.chip-x:disabled {
		cursor: default;
		opacity: 0.25;
	}

	/* ── Quick filters ── */
	.quick-filters {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		padding: 0.6rem 0.75rem;
		border-bottom: 1px solid var(--border);
		transition: opacity 0.2s;
	}

	.quick-filters.dimmed {
		opacity: 0.35;
		pointer-events: none;
	}

	.filter-row {
		display: flex;
		align-items: center;
		gap: 0.6rem;
	}

	.filter-label {
		font-size: 0.72rem;
		font-weight: 600;
		color: var(--text-dim);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		width: 56px;
		flex-shrink: 0;
	}

	.toggle-group {
		display: flex;
		gap: 1px;
		border-radius: 6px;
		overflow: hidden;
		border: 1px solid var(--border);
	}

	.toggle-btn {
		padding: 0.3rem 0.6rem;
		font-size: 0.78rem;
		font-weight: 500;
		background: var(--bg-input);
		border: none;
		color: var(--text-dim);
		transition: background 0.12s, color 0.12s;
		display: flex;
		align-items: center;
		gap: 0.3rem;
	}

	.toggle-btn.active {
		background: var(--accent-dim);
		color: var(--accent-light);
	}

	.toggle-btn:hover:not(:disabled):not(.active) {
		background: var(--bg-card-hover);
		color: var(--text-muted);
	}

	.toggle-btn:disabled {
		cursor: default;
	}

	.toggle-count {
		font-size: 0.68rem;
		font-weight: 600;
		color: var(--text-dim);
		font-variant-numeric: tabular-nums;
	}

	.label-chips {
		display: flex;
		flex-wrap: wrap;
		gap: 0.3rem;
	}

	/* ── Label chips (stateful toggle) ── */
	.label-chip {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.3rem 0.7rem;
		border-radius: 999px;
		font-size: 0.82rem;
		font-weight: 500;
		line-height: 1;
		background: var(--bg-input);
		border: 1px solid var(--border);
		color: var(--text);
		transition: background 0.12s, border-color 0.12s, color 0.12s;
	}

	.label-chip:disabled {
		cursor: default;
	}

	.label-chip.label-included {
		background: rgba(16, 185, 129, 0.12);
		border-color: rgba(16, 185, 129, 0.4);
		color: #34d399;
	}

	.label-chip.label-excluded {
		background: rgba(239, 68, 68, 0.1);
		border-color: rgba(239, 68, 68, 0.35);
		color: #f87171;
	}

	.label-chip.label-neutral:hover:not(:disabled) {
		border-color: var(--accent-dim);
		color: var(--accent-light);
	}

	/* Click-to-remove hint on active labels */
	.label-chip.label-included:hover:not(:disabled),
	.label-chip.label-excluded:hover:not(:disabled) {
		opacity: 0.7;
	}

	.label-name.struck {
		text-decoration: line-through;
		text-decoration-thickness: 1.5px;
	}

	/* ── Search ── */
	.search-wrap {
		padding: 0.5rem 0.75rem;
		border-bottom: 1px solid var(--border);
		transition: opacity 0.2s;
	}

	.search-wrap.dimmed {
		opacity: 0.35;
	}

	input {
		width: 100%;
		padding: 0.5rem 0.75rem;
		background: var(--bg-input);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		color: var(--text);
		font-size: 0.88rem;
		outline: none;
		transition: border-color 0.15s;
	}

	input:focus {
		border-color: var(--border-focus);
	}

	input:disabled {
		opacity: 0.5;
	}

	/* ── Tag grid ── */
	.tag-grid {
		display: flex;
		flex-wrap: wrap;
		align-content: flex-start;
		gap: 0.35rem;
		padding: 0.75rem;
		overflow-y: auto;
		flex: 1;
		transition: opacity 0.2s;
	}

	.tag-grid.dimmed {
		opacity: 0.3;
		pointer-events: none;
	}

	.hint {
		font-size: 0.82rem;
		color: var(--text-dim);
		padding: 0.25rem;
		width: 100%;
	}

	/* ── Tag chips ── */
	.tag-chip {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.3rem 0.7rem;
		border-radius: 999px;
		font-size: 0.82rem;
		font-weight: 500;
		line-height: 1;
		background: var(--bg-input);
		border: 1px solid var(--border);
		color: var(--text);
		transition: background 0.12s, border-color 0.12s, opacity 0.15s;
	}

	.tag-chip:disabled {
		cursor: default;
	}

	.tag-chip:hover:not(:disabled) {
		border-color: var(--accent-dim);
		color: var(--accent-light);
	}

	/* Tag count badge */
	.tag-count {
		font-size: 0.7rem;
		font-weight: 600;
		color: var(--text-dim);
		font-variant-numeric: tabular-nums;
	}

	/* ── Popover ── */
	.popover-backdrop {
		position: fixed;
		inset: 0;
		z-index: 200;
	}

	.popover {
		position: fixed;
		z-index: 201;
		background: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 0.5rem 0.6rem;
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
		box-shadow: 0 8px 30px rgba(0, 0, 0, 0.5);
		min-width: 130px;
	}

	.popover-name {
		font-size: 0.78rem;
		font-weight: 600;
		color: var(--text);
		padding: 0 0.15rem;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.popover-actions {
		display: flex;
		gap: 0.35rem;
	}

	.pop-btn {
		flex: 1;
		padding: 0.35rem 0.5rem;
		border-radius: 6px;
		border: 1px solid transparent;
		font-size: 0.78rem;
		font-weight: 600;
		transition: background 0.12s, border-color 0.12s;
	}

	.pop-include {
		background: rgba(16, 185, 129, 0.12);
		color: #34d399;
		border-color: rgba(16, 185, 129, 0.25);
	}

	.pop-include:hover {
		background: rgba(16, 185, 129, 0.22);
		border-color: rgba(16, 185, 129, 0.5);
	}

	.pop-exclude {
		background: rgba(239, 68, 68, 0.1);
		color: #f87171;
		border-color: rgba(239, 68, 68, 0.2);
	}

	.pop-exclude:hover {
		background: rgba(239, 68, 68, 0.2);
		border-color: rgba(239, 68, 68, 0.45);
	}
</style>
