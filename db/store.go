package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"sync"

	"game.com/pool/gamer"
	_ "github.com/lib/pq"
)

type GamersStorage interface {
	DeleteGamer(gamer gamer.Gamer)
	AddGamer(gamer gamer.Gamer)
	Err() <-chan error
	ReadGamers() (<-chan gamer.Gamer, <-chan error)
	Run()
}

type PostgresGamersStorage struct {
	addGamer    chan gamer.Gamer
	deleteGamer chan gamer.Gamer
	errors      chan error
	db          *sql.DB
	wg          *sync.WaitGroup
}

type PostgresDBParams struct {
	dbName   string
	host     string
	user     string
	password string
	sslMode  string
}

const (
	TABLE_NAME          = "gamers"
	HOST_DEFAULT        = "localhost"
	DB_NAME             = "gamers"
	USER_DEFAULT        = "test_user"
	PASSWORD_DEFAULT    = "postgres"
	SSL_MODE_DEFAULT    = "disable"
	BUFFER_SIZE_DEFAULT = 16
)

func NewPostgresGamersStorage() (*PostgresGamersStorage, error) {

	bufSize := BUFFER_SIZE_DEFAULT
	s := os.Getenv("BUFFER_SIZE")
	if s != "" {
		i, err := strconv.Atoi(s)
		if err == nil {
			bufSize = i
		}
	}

	addGamer := make(chan gamer.Gamer, bufSize)
	delGamer := make(chan gamer.Gamer, bufSize)
	errors := make(chan error)

	host := os.Getenv("DB_HOST")
	if host == "" {
		host = HOST_DEFAULT
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = USER_DEFAULT
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = PASSWORD_DEFAULT
	}

	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		sslMode = SSL_MODE_DEFAULT
	}

	config := PostgresDBParams{
		host:     host,
		dbName:   DB_NAME,
		user:     user,
		password: password,
		sslMode:  sslMode,
	}

	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=%s", config.host, config.dbName, config.user, config.password, sslMode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при подключении к базе данных: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Ошибка при создании соединения с базой данных: $w", err)
	}

	storage := PostgresGamersStorage{addGamer: addGamer, deleteGamer: delGamer, errors: errors, db: db, wg: &sync.WaitGroup{}}

	exists, err := storage.verifyTableExists()

	if err != nil {
		return nil, fmt.Errorf("Таблица в базе данных не существует: %w", err)
	}
	if !exists {
		if err = storage.createTable(); err != nil {
			return nil, fmt.Errorf("Не удалось создать таблицу: %w", err)
		}
	}

	return &storage, nil

}

func (s *PostgresGamersStorage) AddGamer(gamer gamer.Gamer) {
	s.wg.Add(1)
	s.addGamer <- gamer
}

func (s *PostgresGamersStorage) DeleteGamer(gamer gamer.Gamer) {
	s.wg.Add(1)
	s.deleteGamer <- gamer
}

func (s *PostgresGamersStorage) Err() <-chan error {
	return s.errors
}

func (s *PostgresGamersStorage) ReadGamers() (<-chan gamer.Gamer, <-chan error) {

	outGamer := make(chan gamer.Gamer)
	outError := make(chan error, 1) // нужно для того, чтобы считать ошибку из закрытого канала

	query := "SELECT name, skill, latency, connection_time FROM gamers"

	go func() {
		defer close(outGamer)
		defer close(outError)

		rows, err := s.db.Query(query)
		if err != nil {
			outError <- fmt.Errorf("Ошибка запроса: %w", err)
			return
		}

		defer rows.Close()

		var g gamer.Gamer

		for rows.Next() {

			err = rows.Scan(
				&g.Name, &g.Skill,
				&g.Latency, &g.ConTime)

			if err != nil {
				outError <- err
				return
			}

			outGamer <- g
		}

		err = rows.Err()
		if err != nil {
			outError <- fmt.Errorf("Ошибка чтения из базы данных: %w", err)
		}
	}()

	return outGamer, outError
}

func (s *PostgresGamersStorage) verifyTableExists() (bool, error) {
	var result string

	rows, err := s.db.Query(fmt.Sprintf("SELECT to_regclass('public.%s');", TABLE_NAME))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() && result != TABLE_NAME {
		rows.Scan(&result)
	}

	return result == TABLE_NAME, rows.Err()
}

func (s *PostgresGamersStorage) createTable() error {
	var err error

	createQuery := `CREATE TABLE gamers (
		id      		BIGSERIAL PRIMARY KEY,
		name    		TEXT,
		skill 			FLOAT,
		latency     	FLOAT,
		connection_time TIMESTAMPTZ
	  );`

	_, err = s.db.Exec(createQuery)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresGamersStorage) Run() {

	go func() { // INSERT
		query := `INSERT INTO gamers
			(name, skill, latency, connection_time)
			VALUES ($1, $2, $3, $4)`

		for g := range s.addGamer {
			_, err := s.db.Exec(
				query,
				g.Name, g.Skill, g.Latency, g.ConTime)

			if err != nil {
				s.errors <- err
			}

			s.wg.Done()
		}
	}()

	go func() { // DELETE
		query := `DELETE FROM gamers
			WHERE name = $1`

		for g := range s.deleteGamer {
			_, err := s.db.Exec(query, g.Name)

			if err != nil {
				s.errors <- err
			}

			s.wg.Done()
		}
	}()
}

func (s *PostgresGamersStorage) Wait() {
	s.wg.Wait()
}

func (s *PostgresGamersStorage) Close() error {
	s.wg.Wait()

	if s.addGamer != nil {
		close(s.addGamer)
	}

	if s.deleteGamer != nil {
		close((s.deleteGamer))
	}

	return s.db.Close()
}
