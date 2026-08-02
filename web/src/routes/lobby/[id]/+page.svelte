<script lang="ts">
	import type { PageProps } from './$types';
	import { onDestroy, onMount, untrack } from 'svelte';
	import { fly } from 'svelte/transition';
	import { LobbySocket, ERR_LOBBY_NOT_FOUND } from '$lib/ws';
	import type { Tag, Game, ServerMsg, TagCount, LabelCount } from '$lib/types';
	import PlayerList from '$lib/components/PlayerList.svelte';
	import FilterPanel from '$lib/components/FilterPanel.svelte';
	import Toast from '$lib/components/Toast.svelte';

	// Route-scoped params: id is `string` here, unlike $page.params where it is
	// optional because it is typed across every route.
	// Read once on purpose — the socket, the storage key and the remembered name
	// are all bound to one lobby per mount, so a lobby→lobby navigation would
	// need a full teardown rather than a new id.
	let { params }: PageProps = $props();
	const lobbyId = untrack(() => params.id);

	// --- Remembered name ---
	// There is no GET for a lobby — joining is a ws 'join' event — so a reload
	// always lands on the name prompt unless we remember the name ourselves.
	// sessionStorage, not localStorage: it survives a reload but stays per-tab,
	// so two tabs in the same lobby remain two distinct players.
	const nameKey = `plown:lobby:${lobbyId}:name`;

	function rememberedName(): string {
		try {
			return sessionStorage.getItem(nameKey) ?? '';
		} catch {
			return ''; // storage blocked — fall back to prompting
		}
	}

	function rememberName(name: string) {
		try {
			sessionStorage.setItem(nameKey, name);
		} catch {
			// non-fatal: the user just gets prompted again next reload
		}
	}

	function forgetName() {
		try {
			sessionStorage.removeItem(nameKey);
		} catch {
			// ignore
		}
	}

	// Read once at init so phase, the prefilled input and the auto-join decision
	// can't disagree
	const initialName = rememberedName();

	// --- Tags + labels (loaded once from HTTP) ---
	let tags = $state<Tag[]>([]);
	let labels = $state<string[]>([]);
	onMount(async () => {
		if (initialName) autoJoin();

		const [tagsRes, labelsRes] = await Promise.all([
			fetch('/api/tags'),
			fetch('/api/labels')
		]);
		if (tagsRes.ok) tags = await tagsRes.json();
		if (labelsRes.ok) labels = await labelsRes.json();
	});

	// --- Phase ---
	type Phase = 'name' | 'connecting' | 'active' | 'reconnecting';
	// Skip the prompt flash when we already know who we are
	let phase = $state<Phase>(initialName ? 'connecting' : 'name');

	// --- Name prompt ---
	let nameInput = $state(initialName);
	let nameError = $state('');
	// The server only drops the old player once it notices the closed socket, so
	// a fast reload can race its own previous connection. Retry a rejected
	// auto-join once before falling back to the prompt.
	let autoJoining = false;
	let autoJoinRetried = false;
	let autoJoinTimer: ReturnType<typeof setTimeout> | null = null;

	// --- Lobby state ---
	let myName = $state('');
	let players = $state<string[]>([]);
	let activePlayer = $state('');
	let count = $state(0);
	let initialCount = $state(0);
	let includedIds = $state<number[]>([]);
	let excludedIds = $state<number[]>([]);
	let games = $state<Game[]>([]);
	let rawTagCounts = $state<TagCount[]>([]);
	let labelCounts = $state<LabelCount[]>([]);
	let freeCount = $state(0);
	let onlyFree = $state<boolean | null>(null);
	let includedLabels = $state<string[]>([]);
	let excludedLabels = $state<string[]>([]);
	let gameSearch = $state('');
	let filteredGames = $derived(
		gameSearch.trim()
			? games.filter((g) => g.name.toLowerCase().includes(gameSearch.trim().toLowerCase()))
			: games
	);

	// Narrowing gauge
	const narrowPct = $derived(initialCount > 0 ? Math.round((count / initialCount) * 100) : 100);
	const paidCount = $derived(count - freeCount);
	// Color shifts: green → amber → red as pool depletes
	const gaugeColor = $derived(
		narrowPct > 60 ? '#10b981' : narrowPct > 30 ? '#f59e0b' : '#ef4444'
	);

	// Dot matrix: 100 dots, deterministic death order
	const DOT_COUNT = 100;
	const DOT_PALETTE = ['#10b981', '#6366f1', '#f59e0b', '#ec4899', '#06b6d4', '#8b5cf6'];
	// Seeded shuffle for consistent death order
	const deathOrder = (() => {
		const arr = Array.from({ length: DOT_COUNT }, (_, i) => i);
		// Simple deterministic shuffle (fixed seed)
		let seed = 42;
		for (let i = arr.length - 1; i > 0; i--) {
			seed = (seed * 1664525 + 1013904223) & 0x7fffffff;
			const j = seed % (i + 1);
			[arr[i], arr[j]] = [arr[j], arr[i]];
		}
		return arr;
	})();

	interface Dot { alive: boolean; color: string }
	const dots = $derived.by((): Dot[] => {
		const aliveCount = Math.round((narrowPct / 100) * DOT_COUNT);
		const aliveSet = new Set(deathOrder.slice(0, aliveCount));

		// Assign colors by tag composition (normalized so shares sum to 1)
		const tagColors: { color: string; share: number }[] = [];
		if (rawTagCounts.length > 0 && count > 0) {
			const topN = rawTagCounts.slice(0, DOT_PALETTE.length);
			const totalShares = topN.reduce((s, tc) => s + tc.count, 0);
			for (let i = 0; i < topN.length; i++) {
				tagColors.push({ color: DOT_PALETTE[i], share: topN[i].count / totalShares });
			}
		}

		return Array.from({ length: DOT_COUNT }, (_, i) => {
			const alive = aliveSet.has(i);
			let color = gaugeColor;
			if (alive && tagColors.length > 0) {
				// Assign color based on position within alive dots
				const aliveIdx = deathOrder.indexOf(i);
				const frac = aliveIdx / aliveCount;
				let cumulative = 0;
				for (const tc of tagColors) {
					cumulative += tc.share;
					if (frac < cumulative) { color = tc.color; break; }
				}
			}
			return { alive, color };
		});
	});

	// Pool breakdown: top tags with their palette colors
	const topTags = $derived.by(() => {
		if (rawTagCounts.length === 0 || count === 0) return [];
		return rawTagCounts
			.slice(0, DOT_PALETTE.length)
			.map((tc, i) => {
				const tag = tags.find((t) => t.id === tc.id);
				return tag ? { name: tag.name, color: DOT_PALETTE[i], pct: Math.round((tc.count / count) * 100) } : null;
			})
			.filter((x): x is { name: string; color: string; pct: number } => x !== null);
	});

	// Free/paid proportion for the bar
	const freePct = $derived(count > 0 ? Math.round((freeCount / count) * 100) : 0);

	// --- Last action (hero narration) ---
	interface LastAction {
		player: string;
		target: string;
		included: boolean;
		removed: boolean;
		delta: number; // positive = games removed, negative = games restored
		newCount: number;
	}
	let lastAction = $state<LastAction | null>(null);

	function reviewColor(percent: number, label: string): string {
		if (!label) return 'var(--text-dim)';
		if (percent >= 75) return '#4ade80';
		if (percent >= 40) return '#facc15';
		return '#f87171';
	}

	// --- Resize ---
	let topPx = $state(280);
	let isDragging = false;
	let dragStartY = 0;
	let dragStartH = 0;

	function onHandleMousedown(e: MouseEvent) {
		isDragging = true;
		dragStartY = e.clientY;
		dragStartH = topPx;
		e.preventDefault();
	}

	function onWindowMousemove(e: MouseEvent) {
		if (!isDragging) return;
		const maxH = window.innerHeight * 0.6;
		topPx = Math.min(maxH, Math.max(180, dragStartH + (e.clientY - dragStartY)));
	}

	function onWindowMouseup() {
		isDragging = false;
	}

	let isMyTurn = $derived(myName !== '' && activePlayer === myName);

	// --- Toasts ---
	interface ToastItem {
		id: number;
		target: string;
		included: boolean;
		delta: number;
	}
	let toasts = $state<ToastItem[]>([]);
	let nextId = 0;

	function addToast(item: Omit<ToastItem, 'id'>) {
		toasts = [{ ...item, id: nextId++ }, ...toasts];
	}

	function dismissToast(id: number) {
		toasts = toasts.filter((t) => t.id !== id);
	}

	// --- Socket ---
	let socket: LobbySocket | null = null;

	function handleMsg(msg: ServerMsg) {
		switch (msg.type) {
			case 'joined':
				myName = msg.msg.your_name;
				rememberName(myName);
				autoJoining = false;
				autoJoinRetried = false;
				phase = 'connecting';
				break;

			case 'state': {
				const isUpdate = phase === 'active';
				const actingPlayer = activePlayer; // capture before update

				if (isUpdate) {
					const newIncluded = msg.msg.filters.tag.include ?? [];
					const newExcluded = msg.msg.filters.tag.exclude ?? [];
					const newOnlyFree = msg.msg.filters.free?.only_free ?? null;
					const newIncLabels = msg.msg.filters.review_label?.is ?? [];
					const newExcLabels = msg.msg.filters.review_label?.is_not ?? [];
					const prevCount = count;
					const delta = prevCount - msg.msg.count;
					const playerLabel = actingPlayer === myName ? 'You' : actingPlayer;

					// Detect tag changes
					const newIncId = newIncluded.find((id) => !includedIds.includes(id));
					const newExcId = newExcluded.find((id) => !excludedIds.includes(id));
					const tagId = newIncId ?? newExcId;
					const removedIncId = includedIds.find((id) => !newIncluded.includes(id));
					const removedExcId = excludedIds.find((id) => !newExcluded.includes(id));
					const removedTagId = removedIncId ?? removedExcId;

					// Detect free filter change
					const freeChanged = newOnlyFree !== onlyFree;

					// Detect label changes
					const addedIncLabel = newIncLabels.find((l) => !includedLabels.includes(l));
					const addedExcLabel = newExcLabels.find((l) => !excludedLabels.includes(l));
					const removedIncLabel = includedLabels.find((l) => !newIncLabels.includes(l));
					const removedExcLabel = excludedLabels.find((l) => !newExcLabels.includes(l));

					if (tagId !== undefined && delta >= 0) {
						const tag = tags.find((t) => t.id === tagId);
						if (tag) {
							addToast({ target: tag.name, included: newIncId !== undefined, delta });
							lastAction = { player: playerLabel, target: tag.name, included: newIncId !== undefined, removed: false, delta, newCount: msg.msg.count };
						}
					} else if (removedTagId !== undefined) {
						const tag = tags.find((t) => t.id === removedTagId);
						if (tag) {
							addToast({ target: tag.name, included: true, delta });
							lastAction = { player: playerLabel, target: tag.name, included: removedIncId !== undefined, removed: true, delta, newCount: msg.msg.count };
						}
					} else if (freeChanged) {
						const target = newOnlyFree === true ? 'Free only' : newOnlyFree === false ? 'Paid only' : 'Price filter';
						const isRemoval = newOnlyFree === null;
						const isInclusive = newOnlyFree !== null; // narrowing = not inclusive styling
						addToast({ target, included: isInclusive, delta });
						lastAction = { player: playerLabel, target, included: isInclusive, removed: isRemoval, delta, newCount: msg.msg.count };
					} else if (addedIncLabel) {
						addToast({ target: addedIncLabel, included: true, delta });
						lastAction = { player: playerLabel, target: addedIncLabel, included: true, removed: false, delta, newCount: msg.msg.count };
					} else if (addedExcLabel) {
						addToast({ target: addedExcLabel, included: false, delta });
						lastAction = { player: playerLabel, target: addedExcLabel, included: false, removed: false, delta, newCount: msg.msg.count };
					} else if (removedIncLabel || removedExcLabel) {
						const label = removedIncLabel ?? removedExcLabel!;
						addToast({ target: label, included: true, delta });
						lastAction = { player: playerLabel, target: label, included: true, removed: true, delta, newCount: msg.msg.count };
					} else {
						// Game exclusion fallback
						const newGames = msg.msg.games ?? [];
						const excluded = games.find((g) => !newGames.some((ng) => ng.id === g.id));
						if (excluded) {
							addToast({ target: excluded.name, included: false, delta });
							lastAction = { player: playerLabel, target: excluded.name, included: false, removed: false, delta, newCount: msg.msg.count };
						}
					}

					includedIds = newIncluded;
					excludedIds = newExcluded;
				} else {
					includedIds = msg.msg.filters.tag.include ?? [];
					excludedIds = msg.msg.filters.tag.exclude ?? [];
					if (initialCount === 0) initialCount = msg.msg.count;
				}

				players = msg.msg.players;
				activePlayer = msg.msg.active_player;
				count = msg.msg.count;
				rawTagCounts = msg.msg.tag_counts ?? [];
				labelCounts = msg.msg.label_counts ?? [];
				freeCount = msg.msg.free_count ?? 0;
				onlyFree = msg.msg.filters.free?.only_free ?? null;
				includedLabels = msg.msg.filters.review_label?.is ?? [];
				excludedLabels = msg.msg.filters.review_label?.is_not ?? [];
				games = msg.msg.games ?? [];

				// Phase transition LAST — all data must be ready before
				// the active layout mounts and reads derived values
				if (!isUpdate) phase = 'active';
				break;
			}

			case 'error':
				if (phase === 'connecting' || phase === 'name') {
					socket?.close();
					socket = null;

					// Give a raced auto-join one more shot before bothering the user
					if (autoJoining && !autoJoinRetried && msg.msg !== ERR_LOBBY_NOT_FOUND) {
						autoJoinRetried = true;
						autoJoinTimer = setTimeout(joinLobby, 700);
						break;
					}

					// The remembered name was refused — prompt, prefilled for editing
					forgetName();
					autoJoining = false;
					nameError = msg.msg;
					phase = 'name';
				}
				break;
		}
	}

	function handleDisconnect() {
		if (phase === 'active') phase = 'reconnecting';
	}

	function joinLobby() {
		const name = nameInput.trim();
		if (!name) return;
		nameError = '';
		phase = 'connecting';
		socket = new LobbySocket(lobbyId, name, handleMsg, handleDisconnect);
	}

	/** Join with the name this tab already used for this lobby, no prompt. */
	function autoJoin() {
		autoJoining = true;
		joinLobby();
	}

	function onNameKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') joinLobby();
	}

	onDestroy(() => {
		// A pending retry would otherwise open a socket after teardown and join
		// the lobby as a player nobody can see or close
		if (autoJoinTimer) clearTimeout(autoJoinTimer);
		socket?.close();
	});

	function applyTag(id: number, include: boolean) {
		socket?.send({ type: 'tag', msg: { id, include } });
	}

	function removeTag(id: number) {
		socket?.send({ type: 'remove_tag', msg: { id } });
	}

	function applyGame(id: number) {
		socket?.send({ type: 'game', msg: { id } });
	}

	function setFree(isFree: boolean | null) {
		socket?.send({ type: 'free', msg: { is_free: isFree } });
	}

	function applyLabel(value: string, is: boolean) {
		socket?.send({ type: 'review_label', msg: { is, value } });
	}

	function removeLabel(value: string) {
		socket?.send({ type: 'remove_review_label', msg: { value } });
	}

	function copyLink() {
		navigator.clipboard.writeText(lobbyId);
	}
</script>

<svelte:window onmousemove={onWindowMousemove} onmouseup={onWindowMouseup} />

{#if phase === 'name'}
	<div class="center">
		<div class="join-card">
			<h2>Join Lobby</h2>
			<p class="lobby-code">{lobbyId}</p>
<!--
				The name field is the only control on this card and the sole reason
				the page is showing, so focusing it is the expected behaviour.
				Deferring focus to JS instead would be the same thing hidden from
				the linter, and the browser clears it again after load.
			-->
			<!-- svelte-ignore a11y_autofocus -->
			<input
				type="text"
				placeholder="Your name"
				autofocus
				bind:value={nameInput}
				onkeydown={onNameKeydown}
				maxlength={32}
				spellcheck={false}
			/>
			{#if nameError}
				<p class="error">{nameError}</p>
			{/if}
			<button class="btn-primary" onclick={joinLobby} disabled={!nameInput.trim()}>Join</button>
		</div>
	</div>
{:else if phase === 'connecting'}
	<div class="center">
		<div class="spinner"></div>
		<p class="muted">Connecting…</p>
	</div>
{:else}
	<div class="layout">
		<aside class="sidebar">
			<PlayerList {players} {activePlayer} {myName} />
		</aside>

		<main>
			<header class="topbar">
				<a href="/" class="home-link">Plown</a>
				<div class="lobby-id-top">
					<span class="code-top">{lobbyId}</span>
					<button class="copy-btn-top" onclick={copyLink} title="Copy code">⎘</button>
				</div>
			</header>

			<div class="hero-zone" style={games.length > 0 ? `height: ${topPx}px` : ''}>
				<!-- Turn indicator -->
				<div class="turn-indicator" class:my-turn={isMyTurn}>
					{#if isMyTurn}
						<span class="turn-badge yours">Your turn</span>
					{:else if activePlayer}
						<span class="turn-badge waiting">{activePlayer} is choosing…</span>
					{:else}
						<span class="turn-badge waiting">Waiting for players…</span>
					{/if}
				</div>

				<!-- Dot matrix + count overlay -->
				{#if games.length === 0}
					<div class="dot-section">
						<div class="dot-grid">
							{#each dots as dot, i (i)}
								<span
									class="dot"
									class:dead={!dot.alive}
									style="background: {dot.alive ? dot.color : ''}"
								></span>
							{/each}
						</div>
						<div class="dot-overlay">
							<span class="count-big" style="--gauge-color: {gaugeColor}">{count.toLocaleString('en-US')}</span>
							<span class="count-unit">games · {narrowPct}%</span>
						</div>
					</div>

					<!-- Free/paid bar -->
					{#if count > 0}
						<div class="fp-bar-wrap">
							<div class="fp-bar">
								<div class="fp-free" style="width: {freePct}%"></div>
							</div>
							<div class="fp-labels">
								<span class="fp-label free">{freeCount.toLocaleString()} free</span>
								<span class="fp-label paid">{paidCount.toLocaleString()} paid</span>
							</div>
						</div>
					{/if}

					<!-- Tag color legend -->
					{#if topTags.length > 0}
						<div class="tag-legend">
							{#each topTags as t}
								<span class="legend-item">
									<span class="legend-dot" style="background: {t.color}"></span>
									<span class="legend-name">{t.name}</span>
								</span>
							{/each}
						</div>
					{/if}
				{/if}

				<!-- Last action narration -->
				{#if lastAction}
					{#key lastAction}
						<div
							class="last-action"
							class:inc={lastAction.included || lastAction.removed}
							class:exc={!lastAction.included && !lastAction.removed}
							in:fly={{ y: -8, duration: 200 }}
						>
							<span class="action-player">{lastAction.player}</span>
							<span class="action-verb">{lastAction.removed ? 'removed' : lastAction.included ? 'included' : 'excluded'}</span>
							<strong class="action-tag">{lastAction.target}</strong>
							{#if lastAction.delta > 0}
								<span class="action-delta neg">−{lastAction.delta.toLocaleString()}</span>
							{:else if lastAction.delta < 0}
								<span class="action-delta pos">+{Math.abs(lastAction.delta).toLocaleString()}</span>
							{/if}
						</div>
					{/key}
				{/if}

				<!-- Game panel when pool is small -->
				{#if games.length > 0}
					<div class="game-panel">
						<div class="game-panel-top">
							<input
								class="game-search"
								type="text"
								placeholder="Search games…"
								bind:value={gameSearch}
								spellcheck={false}
							/>
							<span class="count-small">
							{#if gameSearch.trim() && filteredGames.length !== games.length}
								{filteredGames.length} / {games.length}
							{:else}
								{games.length}
							{/if}
							games
						</span>
						</div>
						<div class="game-list">
							{#each filteredGames as game (game.id)}
								<div class="game-row">
									<a class="game-link" href={game.store_url} target="_blank" rel="noopener noreferrer">
										<div class="game-img-wrap">
											{#if game.icon_url}
												<img
													class="game-img"
													src={game.icon_url}
													alt={game.name}
													loading="lazy"
													onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')}
												/>
											{/if}
										</div>
										<div class="game-info">
											<span class="game-name">{game.name}</span>
											<div class="game-badges">
												{#if game.review_label}
													<span class="badge review-badge" style="color: {reviewColor(game.review_percent, game.review_label)}">
														{game.review_label}
														{#if game.review_count > 0}
															· {game.review_count.toLocaleString('en-US')}
														{/if}
													</span>
												{/if}
												{#if game.free}
													<span class="badge free-badge">Free</span>
												{/if}
											</div>
										</div>
									</a>
									<button
										class="game-exclude"
										disabled={!isMyTurn}
										onclick={() => applyGame(game.id)}
										title="Exclude this game"
									>✕</button>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>

		{#if games.length > 0}
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div class="resize-handle" onmousedown={onHandleMousedown}></div>
		{/if}

			<div class="content">
				<FilterPanel
					{tags}
					{includedIds}
					{excludedIds}
					{isMyTurn}
					onTag={applyTag}
					onRemoveTag={removeTag}
					{rawTagCounts}
					{labels}
					{labelCounts}
					{includedLabels}
					{excludedLabels}
					onLabel={applyLabel}
					onRemoveLabel={removeLabel}
					{count}
					{freeCount}
					{onlyFree}
					onFree={setFree}
				/>
			</div>
		</main>
	</div>

	{#if phase === 'reconnecting'}
		<div class="reconnect-overlay">
			<div class="spinner"></div>
			<p>Reconnecting…</p>
		</div>
	{/if}
{/if}

<!-- Toast stack (fixed, always rendered when active) -->
{#if phase === 'active' || phase === 'reconnecting'}
	<div class="toast-stack">
		{#each toasts as toast (toast.id)}
			<div transition:fly={{ x: 280, duration: 280 }}>
				<Toast
					tagName={toast.target}
					included={toast.included}
					delta={toast.delta}
					onDismiss={() => dismissToast(toast.id)}
				/>
			</div>
		{/each}
	</div>
{/if}

<style>
	.center {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		padding: 1rem;
	}

	.join-card {
		background: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 2rem 1.75rem;
		width: 100%;
		max-width: 360px;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.join-card h2 {
		font-size: 1.25rem;
		font-weight: 700;
		letter-spacing: -0.02em;
	}

	.lobby-code {
		font-size: 0.8rem;
		font-family: monospace;
		letter-spacing: 0.12em;
		color: var(--accent-light);
		font-weight: 700;
	}

	.join-card input {
		padding: 0.75rem 1rem;
		background: var(--bg-input);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		color: var(--text);
		font-size: 1rem;
		outline: none;
		transition: border-color 0.15s;
	}

	.join-card input:focus {
		border-color: var(--border-focus);
	}

	.error {
		font-size: 0.85rem;
		color: var(--danger);
	}

	.btn-primary {
		padding: 0.8rem;
		background: var(--accent);
		color: #fff;
		border: none;
		border-radius: var(--radius-sm);
		font-size: 0.95rem;
		font-weight: 600;
		transition: opacity 0.15s, box-shadow 0.15s;
	}

	.btn-primary:hover:not(:disabled) {
		opacity: 0.88;
		box-shadow: 0 0 20px var(--accent-glow);
	}

	.btn-primary:disabled {
		opacity: 0.45;
		cursor: not-allowed;
	}

	.muted {
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.spinner {
		width: 36px;
		height: 36px;
		border: 3px solid var(--border);
		border-top-color: var(--accent);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.layout {
		display: flex;
		height: 100vh;
	}

	.sidebar {
		width: 210px;
		flex-shrink: 0;
		border-right: 1px solid var(--border);
		display: flex;
		flex-direction: column;
		padding-top: 1rem;
	}

	main {
		flex: 1;
		min-width: 0;
		min-height: 0;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	/* ── Topbar ── */
	.topbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.6rem 1.5rem;
		border-bottom: 1px solid var(--border);
	}

	.home-link {
		font-size: 1.1rem;
		font-weight: 800;
		letter-spacing: -0.02em;
		text-decoration: none;
		background: linear-gradient(135deg, #fff 30%, var(--accent-light));
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.lobby-id-top {
		display: flex;
		align-items: center;
		gap: 0.3rem;
	}

	.code-top {
		font-size: 0.78rem;
		font-weight: 700;
		letter-spacing: 0.12em;
		color: var(--text-dim);
		font-family: monospace;
	}

	.copy-btn-top {
		background: none;
		border: none;
		color: var(--text-dim);
		font-size: 0.9rem;
		padding: 0.15rem;
		line-height: 1;
		transition: color 0.15s;
	}

	.copy-btn-top:hover {
		color: var(--text);
	}

	/* ── Hero zone ── */
	.hero-zone {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 1.25rem 1.5rem;
		flex-shrink: 0;
		overflow: hidden;
		gap: 0.75rem;
	}

	/* ── Turn indicator ── */
	.turn-indicator {
		width: 100%;
		text-align: center;
	}

	.turn-indicator.my-turn {
		animation: glow-pulse 2s ease-in-out infinite;
	}

	@keyframes glow-pulse {
		0%, 100% { filter: drop-shadow(0 0 0px transparent); }
		50% { filter: drop-shadow(0 0 12px rgba(99, 102, 241, 0.4)); }
	}

	.turn-badge {
		display: inline-block;
		font-size: 0.82rem;
		font-weight: 700;
		padding: 0.3rem 0.9rem;
		border-radius: 999px;
		letter-spacing: -0.01em;
	}

	.turn-badge.yours {
		background: rgba(99, 102, 241, 0.15);
		border: 1px solid rgba(99, 102, 241, 0.4);
		color: #a5b4fc;
	}

	.turn-badge.waiting {
		color: var(--text-muted);
		animation: pulse 2.5s ease-in-out infinite;
	}

	@keyframes pulse {
		0%, 100% { opacity: 0.55; }
		50% { opacity: 1; }
	}

	/* ── Dot matrix ── */
	.dot-section {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 200px;
		height: 200px;
		flex-shrink: 0;
	}

	.dot-grid {
		display: grid;
		grid-template-columns: repeat(10, 1fr);
		gap: 4px;
		padding: 1rem;
		width: 100%;
		height: 100%;
	}

	.dot {
		border-radius: 3px;
		aspect-ratio: 1;
		transition: opacity 0.4s, transform 0.4s, background 0.4s;
		opacity: 0.85;
	}

	.dot.dead {
		background: rgba(255, 255, 255, 0.04) !important;
		opacity: 1;
		transform: scale(0.7);
	}

	.dot-overlay {
		position: absolute;
		inset: 0;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		pointer-events: none;
	}

	.count-big {
		font-size: 2.4rem;
		font-weight: 800;
		letter-spacing: -0.03em;
		font-variant-numeric: tabular-nums;
		color: #fff;
		text-shadow: 0 2px 20px rgba(0, 0, 0, 0.8), 0 0 40px rgba(0, 0, 0, 0.5);
	}

	.count-unit {
		font-size: 0.78rem;
		font-weight: 500;
		color: var(--text-muted);
		text-shadow: 0 1px 8px rgba(0, 0, 0, 0.8);
	}

	/* ── Last action narration ── */
	.last-action {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		font-size: 0.82rem;
		padding: 0.35rem 0.75rem;
		border-radius: 999px;
		background: var(--bg-card);
		border: 1px solid var(--border);
		flex-shrink: 0;
	}

	.last-action.exc {
		border-color: rgba(239, 68, 68, 0.25);
	}

	.last-action.inc {
		border-color: rgba(16, 185, 129, 0.25);
	}

	.action-player {
		color: var(--text-muted);
		font-weight: 500;
	}

	.action-verb {
		color: var(--text-dim);
	}

	.last-action.exc .action-verb { color: #f87171; }
	.last-action.inc .action-verb { color: #34d399; }

	.action-tag {
		color: var(--text);
	}

	.action-delta {
		font-weight: 700;
		font-variant-numeric: tabular-nums;
	}

	.action-delta.neg { color: #f87171; }
	.action-delta.pos { color: #34d399; }

	/* ── Free/paid bar ── */
	.fp-bar-wrap {
		width: 100%;
		max-width: 320px;
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		flex-shrink: 0;
	}

	.fp-bar {
		height: 4px;
		border-radius: 2px;
		background: rgba(249, 115, 22, 0.25);
		overflow: hidden;
	}

	.fp-free {
		height: 100%;
		background: #10b981;
		border-radius: 2px;
		transition: width 0.4s ease;
	}

	.fp-labels {
		display: flex;
		justify-content: space-between;
	}

	.fp-label {
		font-size: 0.7rem;
		font-weight: 600;
		font-variant-numeric: tabular-nums;
	}

	.fp-label.free { color: #34d399; }
	.fp-label.paid { color: #fb923c; }

	/* ── Tag legend ── */
	.tag-legend {
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 0.5rem 0.75rem;
		flex-shrink: 0;
	}

	.legend-item {
		display: flex;
		align-items: center;
		gap: 0.3rem;
	}

	.legend-dot {
		width: 6px;
		height: 6px;
		border-radius: 2px;
		flex-shrink: 0;
	}

	.legend-name {
		font-size: 0.68rem;
		color: var(--text-muted);
		white-space: nowrap;
	}

	/* ── Resize handle ── */
	.resize-handle {
		height: 5px;
		flex-shrink: 0;
		cursor: ns-resize;
		background: var(--border);
		display: flex;
		align-items: center;
		justify-content: center;
		transition: background 0.15s;
	}

	.resize-handle:hover {
		background: var(--accent-dim);
	}

	.resize-handle::after {
		content: '';
		width: 2rem;
		height: 2px;
		border-radius: 999px;
		background: var(--text-dim);
	}

	/* ── Content ── */
	.content {
		display: flex;
		flex-direction: column;
		padding: 0.75rem 1.5rem 1rem;
		flex: 1;
		min-height: 0;
	}

	.reconnect-overlay {
		position: fixed;
		inset: 0;
		background: rgba(13, 13, 20, 0.8);
		backdrop-filter: blur(4px);
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		z-index: 50;
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	/* Toast stack — fixed bottom-right, newest at bottom */
	.toast-stack {
		position: fixed;
		bottom: 1.5rem;
		right: 1.5rem;
		display: flex;
		flex-direction: column-reverse;
		gap: 0.5rem;
		z-index: 100;
		pointer-events: none;
	}

	/* ── Game panel ── */
	.game-panel {
		width: 100%;
		flex: 1;
		min-height: 0;
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}

	.game-panel-top {
		display: flex;
		align-items: center;
		gap: 0.6rem;
	}

	.game-search {
		flex: 1;
		padding: 0.45rem 0.75rem;
		background: var(--bg-input);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		color: var(--text);
		font-size: 0.88rem;
		outline: none;
		transition: border-color 0.15s;
	}

	.game-search:focus {
		border-color: var(--border-focus);
	}

	.count-small {
		font-size: 0.78rem;
		color: var(--text-muted);
		font-weight: 600;
		white-space: nowrap;
		font-variant-numeric: tabular-nums;
	}

	.game-list {
		display: flex;
		flex-direction: column;
		overflow-y: auto;
		flex: 1;
		min-height: 0;
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
	}

	.game-row {
		display: flex;
		align-items: center;
		border-bottom: 1px solid var(--border);
		transition: background 0.12s;
	}

	.game-row:last-child {
		border-bottom: none;
	}

	.game-row:hover {
		background: var(--bg-card-hover);
	}

	.game-link {
		flex: 1;
		min-width: 0;
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 4px;
		text-decoration: none;
		color: var(--text);
	}

	.game-exclude {
		flex-shrink: 0;
		width: 2rem;
		height: 2rem;
		margin-right: 0.5rem;
		border-radius: 50%;
		border: none;
		background: none;
		color: var(--text-dim);
		font-size: 0.75rem;
		line-height: 1;
		opacity: 0;
		transition: background 0.12s, color 0.12s, opacity 0.12s;
	}

	.game-row:hover .game-exclude {
		opacity: 1;
	}

	.game-exclude:hover:not(:disabled) {
		background: rgba(239, 68, 68, 0.15);
		color: #f87171;
	}

	.game-exclude:disabled {
		cursor: default;
	}

	.game-img-wrap {
		position: relative;
		width: 231px;
		height: 87px;
		flex-shrink: 0;
		border-radius: 2px;
		background: var(--bg-input);
	}

	.game-img {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
		border-radius: 2px;
		display: block;
	}

	.game-info {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		min-width: 0;
	}

	.game-name {
		font-size: 0.9rem;
		font-weight: 700;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.game-badges {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		flex-wrap: wrap;
	}

	.badge {
		font-size: 0.72rem;
		font-weight: 500;
		line-height: 1;
		padding: 0.2rem 0.45rem;
		border-radius: 999px;
	}

	.review-badge {
		background: rgba(255, 255, 255, 0.06);
	}

	.free-badge {
		background: rgba(16, 185, 129, 0.12);
		color: #34d399;
		border: 1px solid rgba(16, 185, 129, 0.25);
	}

	@media (max-width: 640px) {
		.sidebar { display: none; }
		.toast-stack { right: 0.75rem; bottom: 0.75rem; }
	}
</style>
