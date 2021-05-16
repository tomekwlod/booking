Install local `migrate`:
- https://github.com/golang-migrate/migrate/tree/master/cmd/migrate

To create new migration file (https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md)
`migrate create -ext sql -dir db/migrations -seq something`

Up:
`migrate -source file://path/to/migrations -database "postgres://user:pass@localhost:5432/database?sslmode=disable" up 2`
or export first (https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md):
`export POSTGRESQL_URL='postgres://user:pass@localhost:5432/database?sslmode=disable'`
and then:
`migrate -source file://path/to/migrations -database $POSTGRESQL_URL up 2`