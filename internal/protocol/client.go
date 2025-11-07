package protocol

type Join struct {
	Name    string `json:"name"`
	LobbyID string `json:"lobby_id"`
}

type FilterTag struct {
	Include bool `json:"include"`
	ID      int  `json:"id"`
}

type RemoveFilterTag struct {
	ID int `json:"id"`
}

type FilterGame struct {
	ID int `json:"id"`
}

type FilterFree struct {
	IsFree *bool `json:"is_free"`
}

type FilterReviewLabel struct {
	Is    bool   `json:"is"`
	Value string `json:"value"`
}

type RemoveReviewLabel struct {
	Value string `json:"value"`
}
