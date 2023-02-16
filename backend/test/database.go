package test

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"time"
)

var database_user = "hanko"
var database_password = "hanko"
var database_name = "hanko_test"

type TestDB struct {
	pool        *dockertest.Pool
	resource    *dockertest.Resource
	DatabaseUrl string
	DbCon       *sql.DB
	Dialect     string
}

// StartDB starts a database in a docker container with the specified dialect and name.
// The name is used to name the container, so that multiple container can be started in parallel.
func StartDB(name string, dialect string) (*TestDB, error) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("could not construct pool: %w", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	options, err := getContainerOptions(dialect)
	if err != nil {
		return nil, fmt.Errorf("could not create docker run options: %w", err)
	}

	options.Name = name

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(options, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, fmt.Errorf("could not start resource: %w", err)
	}

	hostAndPort := resource.GetHostPort(getPortID(dialect))
	databaseUrl := fmt.Sprintf("%s://%s:%s@%s/%s?sslmode=disable", dialect, database_user, database_password, hostAndPort, database_name)

	_ = resource.Expire(120)

	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		db, err := sql.Open(dialect, databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	db, err := sql.Open(dialect, databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	return &TestDB{
		pool:        pool,
		resource:    resource,
		DatabaseUrl: databaseUrl,
		DbCon:       db,
		Dialect:     dialect,
	}, nil
}

// PurgeDB stops the docker container.
func PurgeDB(db *TestDB) error {
	if db == nil {
		return nil
	}
	if err := db.pool.Purge(db.resource); err != nil {
		return fmt.Errorf("could not purge resource: %w", err)
	}
	return nil
}

// getContainerOptions returns the options to start a container, which includes the docker image, tag, env variables, ...
func getContainerOptions(dialect string) (*dockertest.RunOptions, error) {
	switch dialect {
	case "postgres":
		return &dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "11",
			Env: []string{
				fmt.Sprintf("POSTGRES_PASSWORD=%s", database_password),
				fmt.Sprintf("POSTGRES_USER=%s", database_user),
				fmt.Sprintf("POSTGRES_DB=%s", database_name),
				"listen_addresses = '*'",
			},
		}, nil
	case "mysql":
		return &dockertest.RunOptions{
			Repository: "mysql",
			Tag:        "8",
			Env: []string{
				fmt.Sprintf("MYSQL_USER=%s", database_user),
				fmt.Sprintf("MYSQL_PASSWORD=%s", database_password),
				fmt.Sprintf("MYSQL_DATABASE=%s", database_name),
				"MYSQL_RANDOM_ROOT_PASSWORD=true",
			},
		}, nil
	default:
		return nil, UnknownDialectError
	}
}

var UnknownDialectError = errors.New("unknown dialect")

func getPortID(dialect string) string {
	switch dialect {
	case "postgres":
		return "5432/tcp"
	case "mysql":
		return "3306/tcp"
	default:
		return ""
	}
}

// LoadFixtures loads predefined data from the path in the database.
func LoadFixtures(db *sql.DB, dialect string, path string) error {
	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect(dialect),
		testfixtures.Directory(path),
	)
	if err != nil {
		return fmt.Errorf("could not create testfixtures: %w", err)
	}

	err = fixtures.Load()
	if err != nil {
		return fmt.Errorf("could not load fixtures: %w", err)
	}

	return nil
}
