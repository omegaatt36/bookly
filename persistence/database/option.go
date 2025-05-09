package database

import (
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v2"
)

// ConnectOption defines a generic connect option for PostgreSQL.
type ConnectOption struct {
	Dialect  string
	Host     string
	Port     int
	DBName   string
	User     string
	Password string
	Silence  bool
	Testing  bool
}

// ConnStr generates PostgreSQL connection string.
func (opt *ConnectOption) ConnStr() string {
	switch opt.Dialect {
	case "postgres":
		return fmt.Sprintf("postgres://%s:%s@%s:%v/%s?sslmode=disable",
			opt.User, opt.Password, opt.Host, opt.Port, opt.DBName)
	default:
		slog.Warn("bad dialect: " + opt.Dialect)
	}

	return ""
}

// CliFlags returns cli flag list.
func (opt *ConnectOption) CliFlags() []cli.Flag {
	var flags []cli.Flag
	flags = append(flags, &cli.StringFlag{
		Name:        "db-dialect",
		Usage:       "postgres",
		EnvVars:     []string{"DB_DIALECT"},
		Value:       "postgres",
		Required:    true,
		Destination: &opt.Dialect,
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "db-host",
		Usage:       "database host",
		EnvVars:     []string{"DB_HOST"},
		Value:       "localhost",
		Destination: &opt.Host,
	})
	flags = append(flags, &cli.IntFlag{
		Name:        "db-port",
		EnvVars:     []string{"DB_PORT"},
		Value:       5432,
		Destination: &opt.Port,
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "db-name",
		EnvVars:     []string{"DB_NAME"},
		Destination: &opt.DBName,
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "db-user",
		EnvVars:     []string{"DB_USER"},
		Destination: &opt.User,
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "db-password",
		EnvVars:     []string{"DB_PASSWORD"},
		Destination: &opt.Password,
	})
	flags = append(flags, &cli.BoolFlag{
		Name:        "db-silence-logger",
		EnvVars:     []string{"DB_SILENCE_LOGGER"},
		Destination: &opt.Silence,
	})

	return flags
}