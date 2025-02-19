package config

import (
	"log"

	"github.com/Projeto-USPY/uspy-backend/utils"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

var Env Config

type GeneralConfig interface {
	Identify() string
}

// Configuration object, for more info see README.md
type Config struct {
	Domain    string `envconfig:"USPY_DOMAIN" required:"true" default:"localhost"`
	Port      string `envconfig:"USPY_PORT" required:"true" default:"8080"` // careful with this because cloud run must run on port 8080
	JWTSecret string `envconfig:"USPY_JWT_SECRET" required:"true" default:"my_secret"`
	Mode      string `envconfig:"USPY_MODE" required:"true" default:"local"`
	AESKey    string `envconfig:"USPY_AES_KEY" required:"true" default:"71deb5a48500599862d9e2170a60f90194a49fa81c24eacfe9da15cb76ba8b11"` // only used in dev
	RateLimit string `envconfig:"USPY_RATE_LIMIT"`                                                                                         // see github.com/ulule/limiter for more info

	FirestoreKeyPath string `envconfig:"USPY_FIRESTORE_KEY"`

	ProjectID string `envconfig:"USPY_PROJECT_ID"`

	Mailjet // email verification is needed in production
}

func (c Config) IsUsingKey() bool {
	return c.FirestoreKeyPath != ""
}

func (c Config) IsUsingProjectID() bool {
	return c.ProjectID != ""
}

func (c Config) Identify() string {
	if c.IsUsingKey() {
		return c.FirestoreKeyPath
	} else {
		return c.ProjectID
	}
}

func (c Config) IsDev() bool {
	return c.Mode == "dev"
}

func (c Config) IsLocal() bool {
	return c.Mode == "local"
}

// Redact can be used to print the environment config without exposing secret
func (c Config) Redact() Config {
	c.AESKey = "[REDACTED]"
	c.JWTSecret = "[REDACTED]"
	c.Domain = "[REDACTED]"
	c.FirestoreKeyPath = "[REDACTED]"
	c.ProjectID = "[REDACTED]"
	c.Mailjet.APIKey = "[REDACTED]"
	c.Mailjet.Secret = "[REDACTED]"
	return c
}

// TestSetup is used by the emulator, it will only load required defaults, no project-related identifiers
func TestSetup() {
	if err := envconfig.Process("uspy", &Env); err != nil {
		log.Fatal("could not process default env variables: ", err)
	}

	log.Printf("env variables set: %#v\n", Env)
}

// Setup parses the .env file (or uses defaults) to determine environment constants and variables
func Setup() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("did not parse .env file, falling to default env variables")
	}

	if err := envconfig.Process("uspy", &Env); err != nil {
		log.Fatal("could not process default env variables: ", err)
	}

	log.Printf("env variables set: %#v\n", Env.Redact())

	if Env.IsUsingKey() {
		log.Println("Running backend with firestore key")

		if !utils.CheckFileExists(Env.FirestoreKeyPath) {
			log.Fatal("Could not find firestore key path: ", Env.FirestoreKeyPath)
		}
	} else if Env.IsUsingProjectID() {
		log.Println("Running backend with project ID")

		// setup email client
		Env.Mailjet.Setup()
	} else {
		log.Fatal("Could not initialize backend because neither the Firestore Key nor the Project ID were specified")
	}

}
