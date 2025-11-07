package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"plown/internal/protocol"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (h *Handler) LobbyWS(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "lobby_id")

	lobby := h.mgr.GetLobby(id)
	if lobby == nil {
		http.Error(w, "lobby not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.l.Error("upgrading ws", "error", err.Error())
		return
	}

	_, raw, err := conn.ReadMessage()
	if err != nil {
		h.l.Error("reading message", "error", err.Error())
		return
	}

	msg, err := protocol.ParseMsg(raw)
	if err != nil {
		h.l.Error("parsing first message", "error", err.Error())
		return
	}

	if msg.Type != protocol.MsgTypeJoin {
		h.l.Info("first message not of type join")
		conn.WriteJSON(protocol.NewErrorMsg("first message should be of type join"))
		return
	}

	send := make(chan protocol.Msg, 5)
	player, err := lobby.Join(send, msg.Msg.(protocol.Join))
	if err != nil {
		errMsg := protocol.NewErrorMsg(err.Error())
		if err := conn.WriteJSON(errMsg); err != nil {
			h.l.Error("sending join error", "error", err.Error())
		}
		return
	}

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		defer conn.Close()

		for {
			select {
			case msg, ok := <-send:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				if err := conn.WriteJSON(&msg); err != nil {
					return
				}
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	defer lobby.Kick(player)
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}

		msg, err := protocol.ParseMsg(raw)
		if err != nil {
			h.l.Error("parsing message", "error", err.Error())
			if err := conn.WriteJSON(protocol.NewErrorMsg("wrong message")); err != nil {
				break
			}

			continue
		}

		err = nil
		switch msg.Type {
		case protocol.MsgTypeFilterTag, protocol.MsgTypeRemoveTag, protocol.MsgTypeFilterGame,
			protocol.MsgTypeFilterFree, protocol.MsgTypeFilterReviewLabel, protocol.MsgTypeRemoveReviewLabel:
			err = lobby.Filter(player, msg)
		}

		if err != nil {
			h.l.Error("processing message", "msg_type", msg.Type, "msg", msg, "error", err.Error())
			conn.WriteJSON(protocol.NewErrorMsg("server error"))
			return
		}
	}
}
