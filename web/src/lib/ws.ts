import type { ServerMsg, TagCmd, GameCmd, RemoveTagCmd, FreeCmd, ReviewLabelCmd, RemoveReviewLabelCmd } from './types';

export type MessageHandler = (msg: ServerMsg) => void;

export class LobbySocket {
	private ws: WebSocket | null = null;
	private handler: MessageHandler;
	private onDisconnect: (() => void) | undefined;
	private lobbyId: string;
	private name: string;
	private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	private closed = false;

	constructor(
		lobbyId: string,
		name: string,
		handler: MessageHandler,
		onDisconnect?: () => void
	) {
		this.lobbyId = lobbyId;
		this.name = name;
		this.handler = handler;
		this.onDisconnect = onDisconnect;
		this.connect();
	}

	private connect() {
		const proto = location.protocol === 'https:' ? 'wss' : 'ws';
		const ws = new WebSocket(`${proto}://${location.host}/api/lobby/${this.lobbyId}/ws`);
		this.ws = ws;
		let opened = false;

		ws.onopen = () => {
			opened = true;
			ws.send(
				JSON.stringify({
					type: 'join',
					msg: { name: this.name, lobby_id: this.lobbyId }
				})
			);
		};

		ws.onmessage = (e) => {
			try {
				const msg = JSON.parse(e.data) as ServerMsg;
				this.handler(msg);
			} catch {
				// ignore malformed
			}
		};

		ws.onclose = () => {
			if (!opened) {
				// Connection failed (e.g. 404) — don't retry
				this.handler({ type: 'error', msg: 'Lobby not found' });
				return;
			}
			this.onDisconnect?.();
			if (!this.closed) {
				this.reconnectTimer = setTimeout(() => this.connect(), 2000);
			}
		};

		ws.onerror = () => ws.close();
	}

	send(cmd: TagCmd | GameCmd | RemoveTagCmd | FreeCmd | ReviewLabelCmd | RemoveReviewLabelCmd) {
		if (this.ws?.readyState === WebSocket.OPEN) {
			this.ws.send(JSON.stringify(cmd));
		}
	}

	close() {
		this.closed = true;
		if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
		this.ws?.close();
	}
}
