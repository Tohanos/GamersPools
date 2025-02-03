package service

import (
	"log"
	"os"
	"strconv"

	"game.com/pool/db"
	"game.com/pool/gamer"
	"game.com/pool/groups"
)

const (
	MAX_GROUP_SIZE_DEFAULT = 3
)

var (
	config ServiceConfig
)

type GPService struct {
	Gp      *gamer.Gamerspool
	Gg      *groups.GamersGroups
	dbStore *db.PostgresGamersStorage
}

type ServiceConfig struct {
	groupSize int
	storeInDB bool
}

func NewGPService() GPService {
	gss := os.Getenv("MAX_GROUP_SIZE")
	config.groupSize = MAX_GROUP_SIZE_DEFAULT
	if gss != "" {
		i, err := strconv.Atoi(gss)
		if err == nil {
			config.groupSize = i
		} else {
			log.Printf("Не задано или задано неверно значение переменной MAX_GROUP_SIZE: %s", err)
		}
	}

	var err error
	config.storeInDB, err = strconv.ParseBool(os.Getenv("STORE_IN_DB"))
	if err != nil {
		log.Printf("Не задано или задано неверно значение переменной STORE_IN_DB: %s", err)
		config.storeInDB = false
	}

	gp := gamer.NewGamersPool()
	gg := groups.NewGamersGroups(config.groupSize)

	var dbStore *db.PostgresGamersStorage

	if config.storeInDB {
		dbStore, err = db.NewPostgresGamersStorage()
		if err != nil {
			log.Printf("Ошибка подключения к базе данных: %s. Дальнейшая работа будет только с хранилищем в памяти", err)
			config.storeInDB = false
			return GPService{Gp: gp, Gg: gg, dbStore: &db.PostgresGamersStorage{}}
		}
		dbStore.Run()
	}

	return GPService{Gp: gp, Gg: gg, dbStore: dbStore}
}

func (gp *GPService) GetGroups() groups.GamersGroups {
	if config.storeInDB {
		err := gp.GetGamersFromDB()
		if err != nil {
			log.Printf("Ошибка чтения из базы данных: %s", err)
		}
		gp.Gg.RecalculateGroups(gp.Gp.GetPoolCopy())
	} else {
		gp.Gg.CalculateGroups(gp.Gp.GetPoolCopy())
	}
	return groups.GamersGroups{Groups: gp.Gg.Groups}
}

func (gp *GPService) ResetGroups() groups.GamersGroups {
	gp.Gg.RecalculateGroups(gp.Gp.GetPoolCopy())
	return groups.GamersGroups{Groups: gp.Gg.Groups}
}

func (gp *GPService) AddGamer(gamer gamer.Gamer) {
	gp.Gp.Add(gamer)
	if config.storeInDB {
		g, _ := gp.Gp.Get(gamer.Name)
		gp.dbStore.AddGamer(*g)
	}
}

func (gp *GPService) DeleteGamer(gamer gamer.Gamer) {
	gp.Gp.Delete(gamer)
	if config.storeInDB {
		gp.dbStore.DeleteGamer(gamer)
	}
}

func (gp *GPService) GetGamersFromDB() error {
	var err error
	var g gamer.Gamer
	gamers := make([]gamer.Gamer, 0, 1)

	chGamers, chErrors := gp.dbStore.ReadGamers()

	ok := true
	for ok && err == nil {
		select {
		case err, ok = <-chErrors:

		case g, ok = <-chGamers:
			if ok { // можем схватить пустого игрока, если закроется канал
				gamers = append(gamers, g)
			}

		}
	}

	if err == nil {
		for _, g = range gamers {
			gp.Gp.Add(g)
		}
	}

	return err
}
