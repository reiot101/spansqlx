
# spansqlx
[![GitHub Workflow Status (branch)](https://img.shields.io/github/workflow/status/reiot101/spansqlx/CI/main)](https://github.com/reiot101/spansqlx/actions/workflows/ci.yaml?query=branch%3Amain)
![Supported Go Versions](https://img.shields.io/badge/Go-1.16%2C%201.17-lightgrey.svg)
[![GitHub Release](https://img.shields.io/github/release/reiot101/spansqlx.svg)](https://github.com/reiot101/spansqlx/releases)
<!-- [![Coverage Status](https://coveralls.io/repos/github/reiot101/spansqlx/badge.svg?branch=main)](https://coveralls.io/github/reiot101/spansqlx?branch=main) -->
spanner sql pkgs

## install
```
go get github.com/reiot101/spansqlx
```

## usage
Below is an example which shows some common use cases for spansqlx. Check example_test.go for more usage.
```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/reiot101/spansqlx"
)

func main() {
	var database = "projects/sandbox/instances/sandbox/databases/sandbox"
	// this Pings the spanner database trying to connect
	// use spansqlx.Open() for spanner.NewClient() semantics
	db, err := spansqlx.Open(context.Background(), spansqlx.WithDatabase(database))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// use TxPipeline (spanner.ReadWriteTransation), exec multi statements.
	if err := db.TxPipeline(context.Background(), func(ctx context.Context) error {
		var sqlInsertSingers = `INSERT INTO Singers (SingerId, FirstName, LastName) VALUES(@singer_id, @first_name, @last_name)`
		for _, singer := range allSingers {
			if err := db.Exec(ctx, sqlInsertSingers,
				singer.SingerID,
				singer.FirstName,
				singer.LastName,
			); err != nil {
				return err
			}
		}

		var sqlInsertAlbums = `INSERT INTO Albums (SingerID, AlbumID, AlbumTitle) VALUES (@SingerID, @AlbumID, @AlbumTitle)`
		for _, album := range allAlbums {
			if err := db.NamedExec(ctx, sqlInsertAlbums, album); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		log.Fatalf("failed to spansqlx.TxPipeline %v", err)
	}

	// Query the database, storing results in a []Singer (wrapped in []interface)
	if err := db.Select(context.Background(), nil, `SELECT * FROM Singers ORDER BY FirstName DESC`); err != nil {
		log.Fatal(err)
	}

	// You can also get a a single result
	var david Singer
	db.Get(context.Background(), &david, `SELECT * FROM Singers WHERE FirstName=first_name`, "David")
	fmt.Printf("%#v\n", david)
}
```
