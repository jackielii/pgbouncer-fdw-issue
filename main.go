package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	initDb1Direct()
	createForeignLinkOnDb2()
	// queryDb2Bouncer()
	queryDb1Direct()
	queryDb1Bouncer()
}

// initDb1 connects to postgres directly and creates a table
func initDb1Direct() {
	db1 := try1(sql.Open("postgres", "user=postgres password=password host=127.0.0.1 dbname=db1 sslmode=disable port=54321"))
	try0(db1.Ping())

	try1(db1.Exec("CREATE TABLE IF NOT EXISTS test1 (id SERIAL PRIMARY KEY, name TEXT)"))
	try1(db1.Exec("INSERT INTO test1 (name) VALUES ($1)", "test1"))
}

// createForeignLinkOnDb2 uses pgbouncer to connect to db2 and creates a foreign link to db1
func createForeignLinkOnDb2() {
	db2bouncer := try1(sql.Open("postgres", "user=postgres password=password host=127.0.0.1 dbname=db2 sslmode=disable port=6432"))
	try0(db2bouncer.Ping())
	try1(db2bouncer.Exec(`
create extension if not exists postgres_fdw;

drop server if exists db1 cascade;
create server if not exists db1 FOREIGN DATA WRAPPER postgres_fdw options (host 'pgbouncer', port '6432', dbname 'db1');
create user mapping if not exists for postgres server db1 options (user 'postgres', password 'password');
create schema if not exists db1schema;
import foreign schema public from server db1 into db1schema;`))
}

func queryDb1Bouncer() {
	db1bouncer := try1(sql.Open("postgres", "user=postgres password=password dbname=db1 sslmode=disable port=6432"))
	try0(db1bouncer.Ping())

	var currentSchema string
	try0(db1bouncer.QueryRow("SELECT current_schema()").Scan(&currentSchema))
	log.Printf("current schema: %s", currentSchema) // this will show pg_catalog which is wrong

	rows := try1(db1bouncer.Query("SELECT * FROM test1"))
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		try0(rows.Scan(&id, &name))
		log.Printf("id: %d, name: %s", id, name)
	}
}

func queryDb1Direct() {
	db1 := try1(sql.Open("postgres", "user=postgres password=password host=127.0.0.1 dbname=db1 sslmode=disable port=54321"))
	try0(db1.Ping())
	rows := try1(db1.Query("SELECT * FROM test1"))
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		try0(rows.Scan(&id, &name))
		log.Printf("id: %d, name: %s", id, name)
	}
}

// queryDb2Bouncer queries the foreign link created in createForeignLinkOnDb2 via pgbouncer
func queryDb2Bouncer() {
	db2bouncer := try1(sql.Open("postgres", "user=postgres password=password dbname=db2 sslmode=disable port=6432"))
	try0(db2bouncer.Ping())

	rows := try1(db2bouncer.Query("SELECT * FROM db1schema.test1"))
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		try0(rows.Scan(&id, &name))
		log.Printf("id: %d, name: %s", id, name)
	}
}

func try0(err error) {
	if err != nil {
		panic(err)
	}
}

func try1[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
