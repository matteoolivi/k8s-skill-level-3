package pg

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

const (
	usernameEnvVar = "POSTGRES_USER"
	passwordEnvVar = "POSTGRES_PASSWORD"

	// Must match the metadata.name in the service yaml manifest.
	// Also, the service and the pod where this component runs must be in the same K8s namespace.
	host = "pg"

	// Must match one of the spec.ports in the pg service yaml manifest.
	port = "5432"

	// Driver for "database/sql" package.
	driver = "postgres"

	// Name of the db where image URLs are stored.
	// TODO: We should create an ad-hoc DB rather than using "postgres".
	// TODO: Both this and the following consts should be passed in with a config map.
	imageDB = "postgres"

	// Name of the table where image URLs are stored.
	imageTable = "IMAGE"

	imageNameColumn = "name"
	imageURLColumn  = "url"
)

func StoreMetadata(imageName, imageURL string) error {
	// TODO: Factor out to dedicated method.
	connCfg := "user=" + os.Getenv(usernameEnvVar) +
		" dbname=" + imageDB +
		" sslmode=disable" +
		" password=" + os.Getenv(passwordEnvVar) +
		" host=" + host +
		" port=" + port

	db, err := sql.Open(driver, connCfg)
	if err != nil {
		return err
	}
	defer closeConnection(db, imageName)

	insertImageStatement := "INSERT INTO " +
		imageTable +
		" (" +
		imageNameColumn +
		" , " +
		imageURLColumn +
		") VALUES ($1, $2);"
	_, err = db.Query(insertImageStatement, imageName, imageURL)
	if err != nil {
		return err
	}

	return nil
}

func closeConnection(db *sql.DB, imageName string) {
	if err := db.Close(); err != nil {
		log.Printf("failed to close connection to PostgreSQL when processing image %s: %s",
			imageName,
			err)
	}
}
