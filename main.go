package main

import (
	"fmt"
	"net/http"

	"github.com/Tak1za/go-deck"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println(deck.Card{Rank: deck.Ace, Suit: deck.Spade})
	r := gin.Default()
	r.LoadHTMLFiles("index.html")
	r.GET("/", rootHandler)
	r.GET("/ws", socketHandler)
	r.Run()
}

func rootHandler(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type test struct {
	A string `json:"a"`
	B string `json:"b"`
}

type GameState struct {
	state  map[string](map[string]bool)
	table  []deck.Card
	winner deck.Card
}

type GameInput struct {
	TotalCards int       `json:"totalCards"`
	Round      int       `json:"round"`
	Users      []string  `json:"users"`
	Card       deck.Card `json:"card"`
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Failed to upgrade to websocket: %+v", err)
		return
	}

	var gs GameState
	var input GameInput
	currentDeck := deck.New(deck.Count(52))

	for {
		err := conn.ReadJSON(&input)
		if err != nil {
			break
		}
		processMessage(input, &gs, &currentDeck)
		conn.WriteJSON(gs)
	}
}

func socketHandler(c *gin.Context) {
	wsHandler(c.Writer, c.Request)
}

func getWinner(cards ...deck.Card) deck.Card {
	return deck.Card{Suit: deck.Spade, Rank: deck.Ace}
}

func (gs *GameState) chance(user string, card deck.Card) {
	delete(gs.state[user], card.String())
	gs.table = append(gs.table, card)
	if len(gs.table) == cap(gs.table) {
		gs.winner = getWinner(gs.table...)
	}
}

func (gs *GameState) deal(currentDeck *[]deck.Card) {
	start := 0
	iter := 0
	for k := range gs.state {
		iter++
		end := 8 * iter
		for i := range (*currentDeck)[start:end] {
			gs.state[k][(*currentDeck)[i].String()] = true
		}
		start = end
	}
}

func (gs *GameState) initiate(totalCards int, users ...string) {
	for _, j := range users {
		gs.state[j] = make(map[string]bool, totalCards)
	}
	gs.table = make([]deck.Card, len(users))
}

func processMessage(input GameInput, gs *GameState, newDeck *[]deck.Card) {
	if input.Round == -1 {
		gs.initiate(input.TotalCards, input.Users...)
		return
	}

	if input.Round == 0 {
		gs.deal(newDeck)
		return
	}

	gs.chance(input.Users[0], input.Card)
	return
}
