package sqlite

import (
	"database/sql"
	"github.com/OkDenAl/mbtiles_converter/pkg/logging"
)

type Pool struct {
	conns chan *sql.DB
}

func New(dbName string, maxPoolConns int, log logging.Logger) (*Pool, error) {
	pool := &Pool{conns: make(chan *sql.DB, maxPoolConns)}

	for i := 0; i < maxPoolConns; i++ {
		conn, err := sql.Open("sqlite3", dbName)
		if err != nil {
			return nil, err
		}
		pool.conns <- conn
	}
	return pool, nil
}

func (p *Pool) Checkout() *sql.DB {
	return <-p.conns
}

func (p *Pool) Checkin(c *sql.DB) {
	p.conns <- c
}
