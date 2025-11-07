package protocol

import (
	"encoding/json"
	"errors"
)

type MsgType string

const (
	// client
	MsgTypeJoin MsgType = "join"

	MsgTypeFilterTag         MsgType = "tag"
	MsgTypeRemoveTag         MsgType = "remove_tag"
	MsgTypeFilterGame        MsgType = "game"
	MsgTypeFilterFree        MsgType = "free"
	MsgTypeFilterReviewLabel MsgType = "review_label"
	MsgTypeRemoveReviewLabel MsgType = "remove_review_label"

	// server
	MsgTypeOnJoin MsgType = "joined"

	MsgTypeState MsgType = "state"
	MsgTypeError MsgType = "error"
)

type Msg struct {
	Type MsgType `json:"type"`
	Msg  any     `json:"msg,omitempty"`
}

type msg struct {
	Type MsgType         `json:"type"`
	Msg  json.RawMessage `json:"msg,omitempty"`
}

func ParseMsg(data []byte) (*Msg, error) {
	var msg msg
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	switch msg.Type {
	case MsgTypeJoin:
		var j Join
		if err := json.Unmarshal(msg.Msg, &j); err != nil {
			return nil, err
		}
		return &Msg{Type: msg.Type, Msg: j}, nil
	case MsgTypeFilterTag:
		var v FilterTag
		if err := json.Unmarshal(msg.Msg, &v); err != nil {
			return nil, err
		}
		return &Msg{Type: msg.Type, Msg: v}, nil
	case MsgTypeRemoveTag:
		var v RemoveFilterTag
		if err := json.Unmarshal(msg.Msg, &v); err != nil {
			return nil, err
		}
		return &Msg{Type: msg.Type, Msg: v}, nil
	case MsgTypeFilterGame:
		var g FilterGame
		if err := json.Unmarshal(msg.Msg, &g); err != nil {
			return nil, err
		}
		return &Msg{Type: msg.Type, Msg: g}, nil
	case MsgTypeFilterFree:
		var f FilterFree
		if err := json.Unmarshal(msg.Msg, &f); err != nil {
			return nil, err
		}
		return &Msg{Type: msg.Type, Msg: f}, nil
	case MsgTypeFilterReviewLabel:
		var r FilterReviewLabel
		if err := json.Unmarshal(msg.Msg, &r); err != nil {
			return nil, err
		}
		return &Msg{Type: msg.Type, Msg: r}, nil
	case MsgTypeRemoveReviewLabel:
		var r RemoveReviewLabel
		if err := json.Unmarshal(msg.Msg, &r); err != nil {
			return nil, err
		}
		return &Msg{Type: msg.Type, Msg: r}, nil
	default:
		return nil, errors.New("unknown msg type")
	}
}
