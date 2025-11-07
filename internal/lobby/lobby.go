package lobby

import (
	"context"
	"errors"
	"log/slog"
	"plown/internal/entity"
	"plown/internal/protocol"
	"plown/internal/repo"
	"slices"
	"strings"
	"sync"

	"github.com/tidwall/gjson"
	"golang.org/x/sync/errgroup"
)

type Config struct {
	SendGamesThreshold int
}

type Lobby struct {
	m   sync.Mutex
	l   *slog.Logger
	cfg Config

	ID    string
	state state

	pgRepo *repo.Postgres
}

type state struct {
	filters      filters
	players      []Player
	activePlayer string
	count        int
	tagCounts    []protocol.TagCount
	labelCounts  []protocol.LabelCount
	freeCount    int
	games        []entity.Game
}

type Player struct {
	name string
	dead bool
	send chan<- protocol.Msg
}

func New(l *slog.Logger, id string, pgRepo *repo.Postgres) (*Lobby, error) {
	lobby := &Lobby{
		m: sync.Mutex{},
		l: l,
		cfg: Config{
			SendGamesThreshold: 150,
		},

		ID: id,
		state: state{
			players: []Player{},
		},

		pgRepo: pgRepo,
	}

	lobby.updateCounts(context.TODO(), []repo.Filter{})

	return lobby, nil
}

func (l *Lobby) Join(send chan<- protocol.Msg, msg protocol.Join) (*Player, error) {
	l.m.Lock()
	defer l.m.Unlock()

	if l.ID != msg.LobbyID {
		return nil, errors.New("wrong lobby id")
	}

	for _, player := range l.state.players {
		if player.name == msg.Name {
			return nil, errors.New("name taken")
		}
	}

	newPlayer := Player{name: msg.Name, send: send}
	l.state.players = append(l.state.players, newPlayer)

	if l.state.activePlayer == "" {
		l.state.activePlayer = newPlayer.name
	}

	l.sendOrClose(&l.state.players[len(l.state.players)-1], protocol.NewOnJoin(l.ID, msg.Name))
	l.broadcastState()

	return &newPlayer, nil
}

func (l *Lobby) Kick(looser *Player) {
	l.m.Lock()
	defer l.m.Unlock()

	if l.state.activePlayer == looser.name {
		l.advanceTurn()
	}

	l.state.players = slices.DeleteFunc(l.state.players, func(p Player) bool {
		return p.name == looser.name
	})

	l.broadcastState()
}

func (l *Lobby) Filter(player *Player, msg *protocol.Msg) error {
	l.m.Lock()
	defer l.m.Unlock()

	if player.name != l.state.activePlayer {
		player.send <- protocol.NewErrorMsg("not your turn")
		return nil
	}

	var err error
	switch msg.Type {
	case protocol.MsgTypeFilterTag:
		err = l.addTagFilter(msg.Msg.(protocol.FilterTag))
	case protocol.MsgTypeRemoveTag:
		err = l.removeTagFilter(msg.Msg.(protocol.RemoveFilterTag))
	case protocol.MsgTypeFilterGame:
		err = l.addGameFilter(msg.Msg.(protocol.FilterGame))
	case protocol.MsgTypeFilterFree:
		err = l.setFreeFilter(msg.Msg.(protocol.FilterFree))
	case protocol.MsgTypeFilterReviewLabel:
		err = l.addReviewLabelFilter(msg.Msg.(protocol.FilterReviewLabel))
	case protocol.MsgTypeRemoveReviewLabel:
		err = l.removeReviewLabelFilter(msg.Msg.(protocol.RemoveReviewLabel))
	default:
		return errors.New("unknown filter message type '" + string(msg.Type) + "'")
	}
	if err != nil {
		return err
	}

	l.updateCounts(context.TODO(), l.state.filters.All())

	if l.state.count <= l.cfg.SendGamesThreshold {
		games, err := l.pgRepo.Games(context.TODO(), l.state.filters.All())
		if err != nil {
			return err
		}

		l.state.count = len(games)
		l.state.games = games
	} else {
		l.state.games = nil
	}

	l.advanceTurn()
	l.broadcastState()

	return nil
}

func (l *Lobby) addTagFilter(msg protocol.FilterTag) error {
	if msg.Include {
		l.state.filters.Tag.IncludeIDs = append(l.state.filters.Tag.IncludeIDs, msg.ID)
	} else {
		l.state.filters.Tag.ExcludeIDs = append(l.state.filters.Tag.ExcludeIDs, msg.ID)
	}

	return nil
}

func (l *Lobby) removeTagFilter(msg protocol.RemoveFilterTag) error {
	l.state.filters.Tag.ExcludeIDs = slices.DeleteFunc(l.state.filters.Tag.ExcludeIDs, func(id int) bool {
		return id == msg.ID
	})

	l.state.filters.Tag.IncludeIDs = slices.DeleteFunc(l.state.filters.Tag.IncludeIDs, func(id int) bool {
		return id == msg.ID
	})

	return nil
}

func (l *Lobby) addGameFilter(msg protocol.FilterGame) error {
	l.state.filters.Games.IDs = append(l.state.filters.Games.IDs, msg.ID)
	return nil
}

func (l *Lobby) setFreeFilter(msg protocol.FilterFree) error {
	l.state.filters.Free.isFree = msg.IsFree
	return nil
}

func (l *Lobby) addReviewLabelFilter(msg protocol.FilterReviewLabel) error {
	if msg.Is {
		l.state.filters.ReviewLabel.is = append(l.state.filters.ReviewLabel.is, msg.Value)
	} else {
		l.state.filters.ReviewLabel.isNot = append(l.state.filters.ReviewLabel.isNot, msg.Value)
	}

	return nil
}

func (l *Lobby) removeReviewLabelFilter(msg protocol.RemoveReviewLabel) error {
	l.state.filters.ReviewLabel.is = slices.DeleteFunc(l.state.filters.ReviewLabel.is, func(l string) bool {
		return l == msg.Value
	})

	l.state.filters.ReviewLabel.isNot = slices.DeleteFunc(l.state.filters.ReviewLabel.isNot, func(l string) bool {
		return l == msg.Value
	})

	return nil
}

func (l *Lobby) updateCounts(ctx context.Context, filters []repo.Filter) error {
	eg, eCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		newCount, err := l.pgRepo.GameCount(eCtx, filters)
		if err != nil {
			return err
		}

		l.state.count = newCount

		return nil
	})

	eg.Go(func() error {
		newTagCounts, err := l.pgRepo.TagCount(eCtx, filters)
		if err != nil {
			return err
		}
		l.state.tagCounts = newTagCounts

		return nil
	})

	eg.Go(func() error {
		newLabelCounts, err := l.pgRepo.LabelCount(eCtx, filters)
		if err != nil {
			return err
		}
		l.state.labelCounts = newLabelCounts

		return nil
	})

	eg.Go(func() error {
		newFreeCount, err := l.pgRepo.FreeCount(eCtx, filters)
		if err != nil {
			return err
		}
		l.state.freeCount = newFreeCount

		return nil
	})

	return eg.Wait()
}

func (l *Lobby) advanceTurn() {
	l.l.Info("advancing turn")

	if len(l.state.players) == 0 {
		l.state.activePlayer = ""
		return
	}

	if len(l.state.players) == 1 {
		return
	}

	idx := slices.IndexFunc(l.state.players, func(p Player) bool {
		return p.name == l.state.activePlayer
	})

	if idx+1 == len(l.state.players) {
		idx = -1
	}

	l.state.activePlayer = l.state.players[idx+1].name
}

func (l *Lobby) broadcastState() {
	l.l.Info("broadcasting state")

	state := l.snapshot()

	for i, _ := range l.state.players {
		l.sendOrClose(&l.state.players[i], state)
	}
}

func (l *Lobby) snapshot() protocol.Msg {
	players := make([]string, 0, len(l.state.players))
	for _, player := range l.state.players {
		players = append(players, player.name)
	}

	filters := protocol.Filters{
		Tag: protocol.TagFilter{
			Include: slices.Clone(l.state.filters.Tag.IncludeIDs),
			Exclude: slices.Clone(l.state.filters.Tag.ExcludeIDs),
		},
		Game: protocol.GameFilter{
			IDs: slices.Clone(l.state.filters.Games.IDs),
		},
		Free: protocol.FreeFilter{
			OnlyFree: l.state.filters.Free.isFree,
		},
		ReviewLabel: protocol.ReviewLabelFilter{
			Is:    slices.Clone(l.state.filters.ReviewLabel.is),
			IsNot: slices.Clone(l.state.filters.ReviewLabel.isNot),
		},
	}

	games := make([]protocol.Game, 0, len(l.state.games))
	for _, game := range l.state.games {
		format := gjson.GetBytes(game.Assets, "asset_url_format")
		icon := gjson.GetBytes(game.Assets, "small_capsule")

		iconURL := ""
		if format.Exists() && icon.Exists() {
			iconURL = "https://cdn.cloudflare.steamstatic.com/" + strings.Replace(format.String(), "${FILENAME}", icon.String(), 1)
		}

		games = append(games, protocol.Game{
			ID:            game.ID,
			Name:          game.Name,
			Free:          game.IsFree,
			StoreURL:      "https://store.steampowered.com/" + game.StoreURLPath,
			IconURL:       iconURL,
			ReviewCount:   game.ReviewsFilteredCount,
			ReviewLabel:   game.ReviewsFilteredScoreLabel,
			ReviewPercent: game.ReviewsFilteredPercentPositive,
		})
	}

	state := protocol.State{
		Filters:      filters,
		Players:      players,
		ActivePlayer: l.state.activePlayer,
		Count:        l.state.count,
		TagCounts:    slices.Clone(l.state.tagCounts),
		LabelCounts:  slices.Clone(l.state.labelCounts),
		FreeCount:    l.state.freeCount,
		Games:        games,
	}

	return protocol.NewState(state)
}

func (l *Lobby) sendOrClose(player *Player, msg protocol.Msg) {
	if player.dead {
		return
	}

	select {
	case player.send <- msg:
		return
	default:
		l.l.Warn("player died", "player_name", player.name)
		player.dead = true
		close(player.send)
	}
}
