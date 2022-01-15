package db

import (
	"context"
	"fmt"
	"regexp"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	databasepb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
	instancepb "google.golang.org/genproto/googleapis/spanner/admin/instance/v1"
	"google.golang.org/grpc/codes"
)

// CreateInstance is create spanner instance with database uri
func CreateInstance(ctx context.Context, uri string) error {
	matches := regexp.MustCompile("projects/(.*)/instances/(.*)/databases/.*").FindStringSubmatch(uri)
	if matches == nil || len(matches) != 3 {
		return fmt.Errorf("invalid instance %s", uri)
	}

	instanceName := "projects/" + matches[1] + "/instances/" + matches[2]

	client, err := instance.NewInstanceAdminClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.GetInstance(ctx, &instancepb.GetInstanceRequest{
		Name: instanceName,
	})
	if err != nil && spanner.ErrCode(err) != codes.NotFound {
		return err
	}
	if err == nil {
		// instance already exists
		return nil
	}

	_, err = client.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     "projects/" + matches[1],
		InstanceId: matches[2],
	})
	if err != nil {
		return err
	}

	fmt.Printf("Create instance %s done.\n", matches[1])
	return nil
}

// CreateDatabase is create spanner database with database uri
func CreateDatabase(ctx context.Context, uri string, drop bool) error {
	matches := regexp.MustCompile("^(.*)/databases/(.*)$").FindStringSubmatch(uri)
	if matches == nil || len(matches) != 3 {
		return fmt.Errorf("invalid database %s", uri)
	}

	client, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.GetDatabase(ctx, &databasepb.GetDatabaseRequest{Name: uri})
	if err != nil && spanner.ErrCode(err) != codes.NotFound {
		return err
	}
	if err == nil {
		// Database already exists
		if drop {
			if err = client.DropDatabase(ctx, &databasepb.DropDatabaseRequest{Database: uri}); err != nil {
				return err
			}
		} else {
			return nil
		}
	}

	op, err := client.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          matches[1],
		CreateStatement: "CREATE DATABASE `" + matches[2] + "`",
		ExtraStatements: []string{},
	})
	if err != nil {
		return err
	}
	if _, err = op.Wait(ctx); err != nil {
		return err
	}

	fmt.Printf("Create database %s done.\n", matches[2])
	return nil
}
