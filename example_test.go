package spansqlx_test

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/spanner"
	"github.com/reiot101/spansqlx"
)

type Singer struct {
	SingerID  int64
	FirstName string
	LastName  string
}

type Album struct {
	SingerID   int64
	AlbumID    int64
	AlbumTitle string
}

var (
	database = "projects/sandbox/instances/sandbox/databases/sandbox"

	allSingers = []Singer{
		{SingerID: 1, FirstName: "Marc", LastName: "Richards"},
		{SingerID: 2, FirstName: "Catalina", LastName: "Smith"},
		{SingerID: 3, FirstName: "Alice", LastName: "Trentor"},
		{SingerID: 4, FirstName: "Lea", LastName: "Martin"},
		{SingerID: 5, FirstName: "David", LastName: "Lomond"},
	}
	allAlbums = []Album{
		{SingerID: 1, AlbumID: 1, AlbumTitle: "Total Junk"},
		{SingerID: 1, AlbumID: 2, AlbumTitle: "Go, Go, Go"},
		{SingerID: 2, AlbumID: 1, AlbumTitle: "Green"},
		{SingerID: 2, AlbumID: 2, AlbumTitle: "Forever Hold Your Peace"},
		{SingerID: 2, AlbumID: 3, AlbumTitle: "Terrified"},
	}
)

func Example_usage() {
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

func ExampleSpanner_usage() {
	// create spanner client.
	client, err := spanner.NewClient(context.Background(), database)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// create spansqlx.DB instance with spanner.Client
	db := spansqlx.NewDb(context.Background(), client)

	// this Pings the spanner database trying to connect
	if err := db.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	// spanner.ReadWriteTarnsaction pipeline.
	if _, err = client.ReadWriteTransaction(context.TODO(), func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
		// setup spanner transaction in context.
		txCtx := spansqlx.SetTxContext(ctx, tx)

		// Get richards singer from Singers table.
		var richards Singer
		err := db.Get(txCtx, &richards,
			`SELECT * FROM Singers WHERE FistName=@first_name AND LastName=@last_name`, "Marc", "Richards")
		if err != nil {
			return err
		}

		// Add an new song for richards singer.
		return db.Exec(
			txCtx,
			`INSERT INTO Albums (SingerID, AlbumID, AlbumTitle) VALUES (@SingerID, @AlbumID, @AlbumTitle)`,
			richards.SingerID,
			3,
			"New Song",
		)
	}); err != nil {
		log.Fatal(err)
	}

}
