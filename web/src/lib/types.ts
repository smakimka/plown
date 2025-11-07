// --- HTTP ---

export interface Tag {
	id: number;
	name: string;
}

// --- Server → Client ---
// Wire format: {"type":"...", "msg": ...}

export interface JoinedMsg {
	type: 'joined';
	msg: {
		lobby_id: string;
		your_name: string;
	};
}

export interface Game {
	id: number;
	name: string;
	store_url: string;
	icon_url: string;
	review_count: number;
	review_label: string;
	review_percent: number;
	free: boolean;
}

export interface TagCount {
	id: number;
	count: number;
}

export interface LabelCount {
	value: string;
	count: number;
}

export interface StateMsg {
	type: 'state';
	msg: {
		filters: {
			tag: {
				include: number[];
				exclude: number[];
			};
			game: {
				ids: number[];
			};
			free: {
				only_free: boolean | null;
			};
			review_label: {
				is: string[];
				is_not: string[];
			};
		};
		players: string[];
		active_player: string;
		count: number;
		tag_counts: TagCount[];
		label_counts: LabelCount[];
		free_count: number;
		games?: Game[];
	};
}

export interface ErrorMsg {
	type: 'error';
	msg: string;
}

export type ServerMsg = JoinedMsg | StateMsg | ErrorMsg;

// --- Client → Server ---

export interface TagCmd {
	type: 'tag';
	msg: {
		include: boolean;
		id: number;
	};
}

export interface GameCmd {
	type: 'game';
	msg: {
		id: number;
	};
}

export interface RemoveTagCmd {
	type: 'remove_tag';
	msg: {
		id: number;
	};
}

export interface FreeCmd {
	type: 'free';
	msg: {
		is_free: boolean | null;
	};
}

export interface ReviewLabelCmd {
	type: 'review_label';
	msg: {
		is: boolean;
		value: string;
	};
}

export interface RemoveReviewLabelCmd {
	type: 'remove_review_label';
	msg: {
		value: string;
	};
}

// JoinCmd is sent internally by LobbySocket on open — not exported for page use
