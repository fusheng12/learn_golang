package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	xerrors "github.com/pkg/errors"
)

type User struct {
	ID   int
	Name string
	Age  int
}

//  sql.ErrNoRows
func userList(db *sql.DB) ([]*User, error) {
	users := make([]*User, 0)
	rows, err := db.Query("select id, name, age from users ", 1)
	if err != nil {
		return users, xerrors.Wrap(err, "Query users table failed: select id, name, age from users")
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&users)
		if err != nil {
			fmt.Printf("row", err)
		}
	}

	return users, rows.Err()
}

func main() {
	db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/hello")
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	defer db.Close()

	res, err := userList(db)
	if xerrors.Cause(err) == sql.ErrNoRows {
		fmt.Printf("main: users table no rows: %s\n", err)
	} else if err != nil {
		fmt.Printf("main: userList err: %s\n", err)
	}
	fmt.Println(res)
}
