package internal

import "github.com/JensvandeWiel/go-bat/internal/templates/database_pgsql_extra"

type DatabasePgSQLExtra struct {
}

func NewDatabasePgSQL() *DatabasePgSQLExtra {
	return &DatabasePgSQLExtra{}
}

func (i *DatabasePgSQLExtra) Generate(project *Project) error {
	err := project.writeStringTemplateToFile("database/connect.go", database_pgsql_extra.DatabaseConnectTemplate, i)
	if err != nil {
		return err
	}

	return nil
}

func (i *DatabasePgSQLExtra) ModEntries() []string {
	return []string{
		"github.com/jmoiron/sqlx v1.4.0",
		"github.com/lib/pq v1.10.9",
	}
}

func (i *DatabasePgSQLExtra) GitIgnoreEntries() []string {
	return []string{}
}

func (i *DatabasePgSQLExtra) GetExtraPersistentFlags() []string {
	return []string{
		"rootCmd.PersistentFlags().String(\"DBHost\", \"localhost\", \"the database host\")",
		"rootCmd.PersistentFlags().String(\"DBPort\", \"5432\", \"the database port\")",
		"rootCmd.PersistentFlags().String(\"DBUser\", \"user\", \"the database user\")",
		"rootCmd.PersistentFlags().String(\"DBPass\", \"password\", \"the database password\")",
		"rootCmd.PersistentFlags().String(\"DBName\", \"database\", \"the database name\")",
		"viper.BindPFlag(\"DB_HOST\", rootCmd.PersistentFlags().Lookup(\"DBHost\"))",
		"viper.BindPFlag(\"DB_PORT\", rootCmd.PersistentFlags().Lookup(\"DBPort\"))",
		"viper.BindPFlag(\"DB_USER\", rootCmd.PersistentFlags().Lookup(\"DBUser\"))",
		"viper.BindPFlag(\"DB_PASS\", rootCmd.PersistentFlags().Lookup(\"DBPass\"))",
		"viper.BindPFlag(\"DB_NAME\", rootCmd.PersistentFlags().Lookup(\"DBName\"))",
	}
}

func (i *DatabasePgSQLExtra) ExtraType() ExtraType {
	return DatabasePgSQL
}

func (i *DatabasePgSQLExtra) DisallowedExtraTypes() []ExtraType {
	return []ExtraType{}
}

func (i *DatabasePgSQLExtra) ComposerServices() []string {
	return []string{`  postgres:
    image: postgres:17
    environment:
      POSTGRES_DB: 'database'
      POSTGRES_USER: 'user'
      POSTGRES_PASSWORD: 'password'
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data`}
}

func (i *DatabasePgSQLExtra) ComposerVolumes() []string {
	return []string{`  postgres_data:`}
}
