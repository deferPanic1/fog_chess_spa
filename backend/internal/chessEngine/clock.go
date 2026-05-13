package chessengine

import (
	"fmt"
	"time"
)

type MatchClock struct {
	MatchID    string
	WhiteClock time.Time
	BlackClock time.Time
	TimeStamp  time.Time
}

type ClockServer struct {
	Clocks map[string]*MatchClock
}

func NewClockServer() *ClockServer {
	return &ClockServer{
		Clocks: make(map[string]*MatchClock),
	}
}

func (s *ClockServer) RegisterClock(clock *MatchClock) error {
	_, ok := s.Clocks[clock.MatchID]
	if !ok {
		return fmt.Errorf("the clock for this match{%s} alredy exsist", clock.MatchID)
	}

	s.Clocks[clock.MatchID] = clock

	return nil
}
