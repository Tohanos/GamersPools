package groups

import (
	"math"
	"time"

	"game.com/pool/gamer"
)

type Group struct {
	Number   int                     `json:"number"`
	Gamers   map[string]*gamer.Gamer `json:"gamers"`
	FormTime time.Time               `json:"form_time"`
}

type GamersGroups struct {
	Groups []*Group `json:"groups"`
}

type GroupStatistics struct {
	GroupNumber  int      `json:"group_number"`
	MinSkill     float64  `json:"min_skill"`
	MaxSkill     float64  `json:"max_skill"`
	AvgSkill     float64  `json:"avg_skill"`
	MinLatency   float64  `json:"min_latency"`
	MaxLatency   float64  `json:"max_latency"`
	AvgLatency   float64  `json:"avg_latency"`
	MinTimeSpent string   `json:"min_time_spent"`
	MaxTimeSpent string   `json:"max_time_spent"`
	AvgTimeSpent string   `json:"avg_time_spent"`
	PlayerNames  []string `json:"player_names"`
}

var (
	maxGroupSize int
	queue        map[string]*gamer.Gamer
)

func NewGamersGroups(maxSize int) *GamersGroups {

	maxGroupSize = maxSize
	gg := make([]*Group, 0, 1)
	return &GamersGroups{gg}
}

func initQueue(gamers map[string]*gamer.Gamer) map[string]*gamer.Gamer {
	queue = make(map[string]*gamer.Gamer)
	for k, v := range gamers {
		queue[k] = v
	}
	return queue
}

func (gg *GamersGroups) CalculateGroups(gm map[string]*gamer.Gamer) {
	for len(queue) >= maxGroupSize {
		gamers := make(map[string]*gamer.Gamer)
		for i := 0; i < maxGroupSize; i++ {
			minSkillDiff := math.MaxFloat64
			minLatencyDiff := math.MaxFloat64
			var gamerFit *gamer.Gamer

			for _, gamer := range queue {
				if gamerFit == nil {
					gamerFit = gamer
				}
				skillDiff := abs(gamer.Skill - gamerFit.Skill)
				latencyDiff := abs(gamer.Latency - gamerFit.Latency)
				if skillDiff <= minSkillDiff && latencyDiff <= minLatencyDiff {
					minSkillDiff = skillDiff
					minLatencyDiff = latencyDiff
					gamerFit = gamer
				}
			}
			gamers[gamerFit.Name] = gamerFit
			delete(queue, gamerFit.Name)
		}

		gg.Groups = append(gg.Groups, &Group{Gamers: gamers, FormTime: time.Now()})
	}

	for n, g := range gg.Groups {
		g.Number = n
	}
}

func (gg *GamersGroups) RecalculateGroups(gm map[string]*gamer.Gamer) {
	queue = initQueue(gm)
	gg.Groups = make([]*Group, 0, 1)
	gg.CalculateGroups(gm)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func (gg *GamersGroups) CalculateGroupStats(number int) GroupStatistics {
	if number >= len(gg.Groups) {
		return GroupStatistics{}
	}
	var minSkill, maxSkill, sumSkill float64
	var minLatency, maxLatency, sumLatency float64
	var minTimeSpent, maxTimeSpent, sumTimeSpent time.Duration

	gamers := gg.Groups[number].Gamers
	names := make([]string, 0, len(gamers))
	formTime := gg.Groups[number].FormTime

	for _, gamer := range gamers {
		names = append(names, gamer.Name)
		sumSkill += gamer.Skill
		sumLatency += gamer.Latency
		timeSpent := formTime.Sub(gamer.ConTime)
		sumTimeSpent += timeSpent

		if gamer.Skill < minSkill || minSkill == 0 {
			minSkill = gamer.Skill
		}
		if gamer.Skill > maxSkill {
			maxSkill = gamer.Skill
		}

		if gamer.Latency < minLatency || minLatency == 0 {
			minLatency = gamer.Latency
		}
		if gamer.Latency > maxLatency {
			maxLatency = gamer.Latency
		}

		if timeSpent < minTimeSpent || minTimeSpent == 0 {
			minTimeSpent = timeSpent
		}
		if timeSpent > maxTimeSpent {
			maxTimeSpent = timeSpent
		}
	}

	avgSkill := sumSkill / float64(len(gamers))
	avgLatency := sumLatency / float64(len(gamers))
	avgTimeSpent := sumTimeSpent / time.Duration(len(gamers))

	return GroupStatistics{
		GroupNumber:  number,
		MinSkill:     minSkill,
		MaxSkill:     maxSkill,
		AvgSkill:     avgSkill,
		MinLatency:   minLatency,
		MaxLatency:   maxLatency,
		AvgLatency:   avgLatency,
		MinTimeSpent: minTimeSpent.String(),
		MaxTimeSpent: maxTimeSpent.String(),
		AvgTimeSpent: avgTimeSpent.String(),
		PlayerNames:  names,
	}
}
