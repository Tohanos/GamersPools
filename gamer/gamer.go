package gamer

import (
	"errors"
	"sync"
	"time"
)

type Gamer struct {
	Name    string  `json:"name"`
	Skill   float64 `json:"skill"`
	Latency float64 `json:"latency"`
	ConTime time.Time
}

type Gamerspool struct {
	pool map[string]*Gamer
	sync.RWMutex
}

func NewGamersPool() *Gamerspool {
	gp := make(map[string]*Gamer)
	return &Gamerspool{gp, sync.RWMutex{}}
}

func (gp *Gamerspool) Add(g Gamer) {
	gp.Lock()
	gp.pool[string(g.Name)] = &g
	gp.Unlock()
}

func (gp *Gamerspool) Get(name string) (*Gamer, error) {
	gp.RLock()
	g, ok := gp.pool[name]
	gp.RUnlock()
	if !ok {
		return nil, errors.New("Нет игрока с именем " + name)
	}
	return g, nil
}

func (gp *Gamerspool) Delete(g Gamer) {
	gp.Lock()
	defer gp.Unlock()
	delete(gp.pool, g.Name)
}

func (gp *Gamerspool) GetPoolCopy() map[string]*Gamer {
	gp.RLock()
	defer gp.RUnlock()
	res := make(map[string]*Gamer)
	for k, v := range gp.pool {
		res[k] = v
	}
	return res
}
