This repo reproduces an issue related to [pgbouncer](https://github.com/pgbouncer/pgbouncer) and [postgres_fdw](https://github.com/postgres/postgres/tree/master/contrib/postgres_fdw).

## Steps to reproduce:

1. `docker-compose up -d`
2. `go run main.go`

You should see error:
```
panic: pq: relation "test1" does not exist
```

## What happens:

1. set up 2 databases: `db1` and `db2`
2. setup pgbouncer that connects to both in `transaction` pooling mode. See [pgbouncer.ini](pgbouncer.ini)
3. create a foreign table in `db2` that connects to `db1` using postgres_fdw
4. open a connection to `db1` using pgbouncer
5. run a query in `db1`
6. error occurs because wrong default schema is used

## Why this happens: (my theory)

When we connect to pgbouncer and uses postgres_fdw to create a foreign link, it sets the search_path to `pg_catalog`: see [related code](https://github.com/postgres/postgres/blob/master/contrib/postgres_fdw/postgres_fdw.c#L3935-L3938). When this is done and the connection is returned to the pool, the search_path is not reset to the default schema. So when our clients opens a connection to pgbouncer, it's very likely it'll pick up the connection that has the search_path set to `pg_catalog` and not the default schema. This is why we see the error `pq: relation "test1" does not exist` because it can't find the table in the pg_catalog schema.

## Questions:

1. this doesn't happens in the pool_model=session
  You can try this by changing the [pgbouncer.ini](pgbouncer.ini) to use `pool_mode=session` and run `docker-compose restart pgbouncer`, followed by `go run main.go`

