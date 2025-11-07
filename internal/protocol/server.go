package protocol

type State struct {
	Filters      Filters      `json:"filters"`
	Players      []string     `json:"players"`
	ActivePlayer string       `json:"active_player"`
	Count        int          `json:"count"`
	TagCounts    []TagCount   `json:"tag_counts"`
	LabelCounts  []LabelCount `json:"label_counts"`
	FreeCount    int          `json:"free_count"`
	Games        []Game       `json:"games,omitempty"`
}

type Game struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	StoreURL      string `json:"store_url"`
	IconURL       string `json:"icon_url"`
	ReviewCount   int    `json:"review_count"`
	ReviewLabel   string `json:"review_label"`
	ReviewPercent int    `json:"review_percent"`
	Free          bool   `json:"free"`
}

type TagCount struct {
	ID    int `json:"id"`
	Count int `json:"count"`
}

type LabelCount struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

type Filters struct {
	Tag         TagFilter         `json:"tag"`
	Game        GameFilter        `json:"game"`
	Free        FreeFilter        `json:"free"`
	ReviewLabel ReviewLabelFilter `json:"review_label"`
}

type TagFilter struct {
	Include []int `json:"include"`
	Exclude []int `json:"exclude"`
}

type GameFilter struct {
	IDs []int `json:"ids"`
}

type FreeFilter struct {
	OnlyFree *bool `json:"only_free"`
}

type ReviewLabelFilter struct {
	Is    []string `json:"is"`
	IsNot []string `json:"is_not"`
}

func NewState(state State) Msg {
	return Msg{
		Type: MsgTypeState,
		Msg:  state,
	}
}

type OnJoin struct {
	LobbyID  string `json:"lobby_id"`
	YourName string `json:"your_name"`
}

func NewOnJoin(lobbyID, yourName string) Msg {
	return Msg{
		Type: MsgTypeOnJoin,
		Msg: OnJoin{
			LobbyID:  lobbyID,
			YourName: yourName,
		},
	}
}
