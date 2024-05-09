package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	conn         *sql.DB
	getFileStmt  *sql.Stmt
	insViewsStmt *sql.Stmt
}

func NewStorage(ctx context.Context, conn *sql.DB) *Storage {
	//res_stmt, err := conn.PrepareContext(ctx, `select "file", "view" FROM viewsfile WHERE "file" = $1`)
	sel_stmt, err := conn.PrepareContext(ctx, `select "file", "view", "access", "path", "position", "pathfile" FROM viewsfile WHERE "access" = $1 and "file" = $2`)
	ins_stmt, err := conn.PrepareContext(ctx, `insert into viewsfile (file, view, access, path, position, pathfile)
	values ($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		fmt.Println("[-]: ", err)
		return &Storage{}
	}
	return &Storage{
		conn:         conn,
		getFileStmt:  sel_stmt,
		insViewsStmt: ins_stmt,
	}
}

func (s *Storage) getFileView(ctx context.Context, filename string, access string) ([]ViewsFile, error) {
	recs := []ViewsFile{}
	fmt.Println("<getFileView>")
	fmt.Println("ctx: ", ctx)
	select {
	case <-ctx.Done():
		fmt.Println("ctx is Done")
	default:
		//rows, err := s.getFileStmt.QueryRow(filename).Scan(&f)
		rows, err := s.getFileStmt.QueryContext(ctx, access, filename)
		fmt.Println("ahaaa err: ", err)
		defer rows.Close()
		for rows.Next() {
			fmt.Println("fetching")
			rec := ViewsFile{}
			if err := rows.Scan(&rec.file, &rec.view, &rec.access, &rec.path, &rec.order, &rec.pathfile); err != nil {
				fmt.Println("get err: ", err)
				return nil, err
			}
			recs = append(recs, rec)
		}
		return recs, nil
	}
	return nil, errors.New("closed ctx")
}
func (s *Storage) insertViews(ctx context.Context, p_viewfile ViewsFile) error {
	fmt.Println(p_viewfile)
	_, err := s.insViewsStmt.ExecContext(
		ctx,
		p_viewfile.file,
		p_viewfile.view,
		p_viewfile.access,
		p_viewfile.path,
		p_viewfile.order,
		p_viewfile.pathfile,
	)
	return err
}

type ContactRec struct {
	ID    int
	Name  string
	Phone string
}

type ViewsFile struct {
	file     string
	view     string
	access   string
	path     string
	order    int
	pathfile string
}

func main() {
	urldb := "postgres://db_ouzer:dbouzer_bbass_369@localhost:5432/the_DB"
	conn, err := sql.Open("pgx", urldb)
	if err != nil {
		fmt.Println("connect to db error: ", err)
	}
	defer conn.Close()
	ctx, cancelctx := context.WithTimeout(context.Background(), 15*time.Second)
	if err := conn.PingContext(ctx); err != nil {
		fmt.Println(err)
	}

	contact, err := GetContact(ctx, *conn, 1)
	if err != nil {
		fmt.Fprint(os.Stderr, "[-] aaaa error: ", err)
	}
	fmt.Println(contact)
	store := NewStorage(ctx, conn)

	f := ViewsFile{
		file:     "F100",
		view:     "LBRA-CON-FSB-001",
		access:   "LINKPATH",
		path:     "F100LKS0",
		order:    1,
		pathfile: "F100",
	}
	ret := store.insertViews(ctx, f)
	f.view = "L2"
	ret = store.insertViews(ctx, f)
	f.view = "L3"
	ret = store.insertViews(ctx, f)
	f.view = "L4"
	ret = store.insertViews(ctx, f)
	fmt.Println(ret)

	//store := NewStorage(ctx, conn)
	if err := ctx.Err(); err != nil {
		fmt.Println("err ctx: ", err)
	}
	ret2, err := store.getFileView(ctx, "F300", "LINKPATH")
	if err != nil {
		fmt.Println("err fetching: ", err)
	}
	fmt.Println(ret2, err)

	cancelctx()
}

func GetContact(ctx context.Context, conn sql.DB, id int) (ContactRec, error) {
	const query = `SELECT "contact_name", "phone" FROM contacts WHERE "user_id" = $1`
	contact := ContactRec{ID: id}
	err := conn.QueryRowContext(ctx, query, id).Scan(&contact.Name, &contact.Phone)
	return contact, err
}
