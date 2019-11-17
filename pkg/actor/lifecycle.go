package actor

import (
	"math/rand"
	"sync"
	"time"
)

type Actor interface {
	Start()
	Shutdown()
}

type ScaleGroup struct {
	actors struct {
		a map[uint64]Actor
		sync.Mutex
	}
	constructor func() Actor
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func NewScaleGroup(constructor func() Actor) ScaleGroup {
	return ScaleGroup{
		actors: struct {
			a map[uint64]Actor
			sync.Mutex
		}{
			a: make(map[uint64]Actor),
		},
		constructor: constructor,
	}
}

func (s *ScaleGroup) Add() {
	s.actors.Lock()
	var id uint64
	for _, ok := s.actors.a[id]; ok; {
		id = rand.Uint64()
	}
	actor := s.constructor()
	s.actors.a[id] = actor
	s.actors.Unlock()

	go func() {
		defer func() {
			recover()

			s.actors.Lock()
			delete(s.actors.a, id)
			s.actors.Unlock()
		}()

		actor.Start()
	}()
}

func (s *ScaleGroup) Shutdown() {
	s.actors.Lock()
	for _, actor := range s.actors.a {
		actor.Shutdown()
	}
	s.actors.Unlock()
}
