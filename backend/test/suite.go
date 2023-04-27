package test

import (
	"fmt"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/pop/v6/logging"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"testing"
)

type Suite struct {
	suite.Suite
	Storage persistence.Storage
	DB      *TestDB
	Name    string // used for database docker container name, so that tests can run in parallel
}

func (s *Suite) SetupSuite() {
	if testing.Short() {
		return
	}
	pop.SetLogger(testLogger)
	//pop.Debug = true
	if s.Name == "" {
		var err error
		id, err := uuid.NewV4()
		if err != nil {
			s.Fail("failed to generate database container name")
		}
		s.Name = id.String()
	}
	dialect := "postgres"
	db, err := StartDB(s.Name, dialect)
	s.NoError(err)
	storage, err := persistence.New(config.Database{
		Url: db.DatabaseUrl,
	})
	s.NoError(err)

	s.Storage = storage
	s.DB = db
}

func (s *Suite) SetupTest() {
	if s.DB != nil {
		err := s.Storage.MigrateUp()
		s.NoError(err)
	}
}

func (s *Suite) TearDownTest() {
	if s.DB != nil {
		err := s.Storage.MigrateDown(-1)
		s.NoError(err)
	}
}

func (s *Suite) TearDownSuite() {
	if s.DB != nil {
		s.NoError(PurgeDB(s.DB))
	}
}

// LoadFixtures loads predefined data from the path in the database.
func (s *Suite) LoadFixtures(path string) error {
	fixtures, err := testfixtures.New(
		testfixtures.Database(s.DB.DbCon),
		testfixtures.Dialect(s.DB.Dialect),
		testfixtures.Directory(path),
		testfixtures.SkipResetSequences(),
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

func testLogger(level logging.Level, s string, args ...interface{}) {

}
