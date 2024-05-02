package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

type ContactRec struct {
	ID    int
	Name  string
	Phone string
}

func main() {
	urldb := "postgres://db_ouzer:dbouzer_bbass_369@localhost:5432/the_DB"
	conn, err := sql.Open("pgx", urldb)
	if err != nil {
		fmt.Println("connect to db error: ", err)
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err := conn.PingContext(ctx); err != nil {
		fmt.Println(err)
	}
	cancel()
}

func GetContact(ctx context.Context, conn sql.DB, id int) (ContactRec, error) {
	const query = `SELECT "name", "phone" FROM contacts WHERE "user_id" = $1`
	contact := ContactRec{ID: id}
	err := conn.QueryRowContext(ctx, query, id).Scan(&contact)
	return contact, err
}
