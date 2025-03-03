package integrationtests

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	postgresMigrate "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/twmb/franz-go/pkg/kgo"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
)

type IntegrationTestSuite struct {
	suite.Suite
	postgresClient    *sql.DB
	postgresContainer *testcontainers.Container
	kafkaContainer    *kafka.KafkaContainer
}

// SetupSuite runs once before all tests.
func (s *IntegrationTestSuite) SetupSuite() {
	// Initialize dependencies (DB, Redis, etc.)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	log.Println("Setting up containers")
	s.createPostgresContainer(ctx)
	s.createPostgresClient(ctx)
	s.startPostgresMigration(ctx)
	s.createKafkaContainer(ctx)

	cancel()
}

// TearDownSuite runs once after all tests.
func (s *IntegrationTestSuite) TearDownSuite() {
	// Cleanup resources
}

// Run the test suite
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) createKafkaClient(ctx context.Context, opt ...kgo.Opt) (*kgo.Client, error) {
	host, err := s.kafkaContainer.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get Kafka container host: %v", err)
	}

	port, err := s.kafkaContainer.MappedPort(ctx, "9093") // Kafka default port
	if err != nil {
		log.Fatalf("Failed to get Kafka container port: %v", err)
	}

	// Construct the Kafka bootstrap server address
	bootstrapServer := fmt.Sprintf("%s:%s", host, port.Port())

	seeds := []string{bootstrapServer}
	kafkaClient, err := kgo.NewClient(append([]kgo.Opt{kgo.SeedBrokers(seeds...), kgo.AllowAutoTopicCreation()}, opt...)...)

	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %v", err)
	}

	return kafkaClient, nil
}

func (s *IntegrationTestSuite) createKafkaContainer(ctx context.Context) (*kafka.KafkaContainer, error) {
	kafkaContainer, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0")
	if err != nil {
		return nil, fmt.Errorf("failed to start kafka container: %v", err)
	}
	s.kafkaContainer = kafkaContainer
	return kafkaContainer, nil
}

func (s *IntegrationTestSuite) createPostgresContainer(ctx context.Context) (*testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17.2-alpine3.21",
		ExposedPorts: []string{"5432:5432"},
		Env: map[string]string{
			"POSTGRES_DB":       "public",
			"POSTGRES_USER":     "user",
			"POSTGRES_PASSWORD": "password",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(5 * time.Second),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start PostgreSQL container: %v", err)
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %v", err)
	}

	mappedPort, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %v", err)
	}

	connectionString := fmt.Sprintf("postgres://user:password@%s:%s/public?sslmode=disable", host, mappedPort.Port())
	os.Setenv("DB_CONNECTION_STRING", connectionString)

	fmt.Println("PostgreSQL is running at:", connectionString)

	s.postgresContainer = &postgresContainer
	return &postgresContainer, nil
}

func (s *IntegrationTestSuite) createPostgresClient(ctx context.Context) (*sql.DB, error) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")

	postgresClient, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to establish database connection: %v", err)
	}

	s.postgresClient = postgresClient
	return postgresClient, nil
}

func (s *IntegrationTestSuite) startPostgresMigration(ctx context.Context) {

	driver, err := postgresMigrate.WithInstance(s.postgresClient, &postgresMigrate.Config{})
	if err != nil {
		log.Fatalf("could not init driver: %s", err)
	}

	migrate, err := migrate.NewWithDatabaseInstance(
		"file://./database/migrations",
		"pgx", driver)

	if err != nil {
		log.Fatalf("could not apply the migration: %s", err)
	}

	migrateDownErr := migrate.Down()

	if migrateDownErr != nil {
		log.Printf("could not apply the migration: %s", migrateDownErr)
	}

	migrateUpErr := migrate.Up()

	if migrateUpErr != nil {
		log.Fatalf("could not apply the migration: %s", migrateUpErr)
	}

	log.Println("Migration applied successfully")
}
