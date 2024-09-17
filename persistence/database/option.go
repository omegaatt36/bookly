package database

import (
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v2"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectOption defines a generic connect option for all dialects.
type ConnectOption struct {
	Dialect  string
	Host     string
	Port     int // optional, if you append port in host, this option is unnecessary.
	DBName   string
	User     string
	Password string
	Config   gorm.Config
	Silence  bool

	Testing bool

	Logger logger.Interface
}

// ConnStr generates connection string.
func (opt *ConnectOption) ConnStr() string {
	switch opt.Dialect {
	case "sqlite3":
		return opt.Host
	case "postgres":
		return fmt.Sprintf("host=%s port=%v user=%v "+
			"dbname=%v password=%v sslmode=disable",
			opt.Host, opt.Port, opt.User, opt.DBName, opt.Password)
	default:
		slog.Warn("bad dialect: " + opt.Dialect)
	}

	return ""
}

// Dialector generates gorm Dialector.
func (opt *ConnectOption) Dialector() gorm.Dialector {
	dsn := opt.ConnStr()
	switch opt.Dialect {
	case "sqlite3":
		return sqlite.Open(dsn)
	case "postgres":
		return postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true})
	default:
		slog.Warn("bad dialect: " + opt.Dialect)
	}

	return nil
}

// CliFlags returns cli flag list.
func (opt *ConnectOption) CliFlags() []cli.Flag {
	var flags []cli.Flag
	flags = append(flags, &cli.StringFlag{
		Name:        "db-dialect",
		Usage:       "[sqlite3|postgres]",
		EnvVars:     []string{"DB_DIALECT"},
		Value:       "postgres",
		Required:    true,
		Destination: &opt.Dialect,
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "db-host",
		Usage:       "postgres -> host, sqlite3 -> filepath",
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
