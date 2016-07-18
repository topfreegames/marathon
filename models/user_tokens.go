package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/satori/go.uuid"
	"github.com/uber-go/zap"
	"gopkg.in/gorp.v1"
)

// UserToken identifies uniquely one token
type UserToken struct {
	ID        uuid.UUID `db:"id"`
	Token     string    `db:"token"`
	UserID    string    `db:"user_id"`
	Locale    string    `db:"locale"`
	Region    string    `db:"region"`
	Tz        string    `db:"tz"`
	BuildN    string    `db:"build_n"`
	OptOut    []string  `db:"opt_out"`
	CreatedAt int64     `db:"created_at"`
	UpdatedAt int64     `db:"updated_at"`
}

// PreInsert populates fields before inserting a new template
func (o *UserToken) PreInsert(s gorp.SqlExecutor) error {
	o.ID = uuid.NewV4()
	o.CreatedAt = time.Now().Unix()
	o.UpdatedAt = o.CreatedAt
	return nil
}

// PreUpdate populates fields before updating a template
func (o *UserToken) PreUpdate(s gorp.SqlExecutor) error {
	o.UpdatedAt = time.Now().Unix()
	return nil
}

// GetUserTokenByID returns a template by id
func GetUserTokenByID(db DB, app string, service string, id uuid.UUID) (*UserToken, error) {
	var userToken UserToken
	tableName := GetTableName(app, service)
	query := fmt.Sprintf("SELECT * FROM %s WHERE id=$1", tableName)
	err := db.SelectOne(&userToken, query, id)
	if err != nil || &userToken == nil {
		return nil, &ModelNotFoundError{tableName, "id", id}
	}
	return &userToken, nil
}

// GetUserTokenByToken returns templates with the given name
func GetUserTokenByToken(db DB, app string, service string, token string) ([]UserToken, error) {
	var userTokens []UserToken
	tableName := GetTableName(app, service)
	query := fmt.Sprintf("SELECT * FROM %s WHERE token=$1", tableName)
	_, err := db.Select(&userTokens, query, token)

	if err != nil || &userTokens == nil || len(userTokens) == 0 {
		return nil, &ModelNotFoundError{"UserToken", "token", token}
	}
	return userTokens, nil
}

// CreateToken creates a new Token
func CreateToken(db DB, app string, service string, userID string, token string, locale string, region string, tz string, buildN string, optOut []string) (*UserToken, error) {
	tableName := GetTableName(app, service)
	optOutString, marshOptOutErr := json.Marshal(optOut)
	if marshOptOutErr != nil {
		Logger.Error(
			"Could not marshal optOut",
			zap.String("optOut", fmt.Sprintf("%+v", optOut)),
			zap.Error(marshOptOutErr),
		)
		return nil, marshOptOutErr
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (user_id, token, locale, region, tz, build_n, opt_out) VALUES('%s', '%s', '%s', '%s', '%s', '%s', '%s')",
		tableName, userID, token, locale, region, tz, buildN, string(optOutString),
	)

	result, execErr := db.Exec(query)
	if execErr != nil {
		Logger.Error(
			"Could not exec query",
			zap.String("query", query),
			zap.Error(execErr),
		)
		return nil, execErr
	}
	userToken := &UserToken{
		Token:  token,
		UserID: userID,
		Locale: locale,
		Region: region,
		Tz:     tz,
		BuildN: buildN,
		OptOut: optOut,
	}

	Logger.Debug(
		fmt.Sprintf("Inserted userToken to %s", tableName),
		zap.String("query", query),
		zap.String("result", fmt.Sprintf("%+v", result)),
	)

	return userToken, nil
}

// UpdateToken updates a Token
func UpdateToken(db DB, app string, service string, id uuid.UUID, userID string, token string, locale string, region string, tz string, buildn string, optOut []string) (*UserToken, error) {
	userToken, getUserTokenErr := GetUserTokenByID(db, app, service, id)
	if getUserTokenErr != nil {
		return nil, getUserTokenErr
	}

	userToken.UserID = userID
	userToken.Token = token
	userToken.Locale = locale
	userToken.Region = region
	userToken.Tz = tz
	userToken.BuildN = buildn
	userToken.OptOut = optOut

	_, updateErr := db.Update(userToken)
	if updateErr != nil {
		return nil, updateErr
	}

	return userToken, nil
}

// GetTableName retrieves a table name based in the app and service
func GetTableName(app string, service string) string {
	return fmt.Sprintf("%s_%s", app, service)
}

// CreateUserTokensTable creates a table for the model UserToken with the name based on app and service
func CreateUserTokensTable(_db DB, app string, service string) (*gorp.TableMap, error) {
	db := _db.(*gorp.DbMap)

	tableName := GetTableName(app, service)
	Logger.Error(
		"TableName",
		zap.String("name", tableName),
	)

	createQuery := fmt.Sprintf(`
    CREATE TABLE IF NOT EXISTS %s (
      id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
      token varchar(255) NOT NULL CHECK (token <> ''),
      user_id varchar(255) NOT NULL CHECK (user_id <> ''),
      locale varchar(3) NOT NULL CHECK (locale <> ''),
      region varchar(3) NOT NULL CHECK (region <> ''),
      tz varchar(20) NOT NULL CHECK (tz <> ''),
      build_n varchar(255) NOT NULL CHECK (build_n <> ''),
      opt_out varchar[] NOT NULL DEFAULT '{}',
      created_at bigint NOT NULL,
      updated_at bigint NOT NULL
    );
    CREATE INDEX IF NOT EXISTS index_%s_on_locale ON %s (lower(locale));
    CREATE INDEX IF NOT EXISTS index_%s_on_token ON %s (token);
    CREATE INDEX IF NOT EXISTS index_%s_on_user_id ON %s (user_id);`,
		tableName, tableName, tableName, tableName, tableName, tableName, tableName,
	)

	dropQuery := fmt.Sprintf(`
    DROP TABLE IF EXISTS %s;
    DROP INDEX IF EXISTS index_%s_on_locale;
    DROP INDEX IF EXISTS index_%s_on_token;
    DROP INDEX IF EXISTS index_%s_on_user_id;`,
		tableName, tableName, tableName, tableName,
	)
	created, createErr := db.Exec(createQuery)
	if createErr != nil {
		_, dropErr := db.Exec(dropQuery)
		if dropErr != nil {
			Logger.Error(
				"Could not exec queries",
				zap.String("createQuery", createQuery),
				zap.String("dropQuery", dropQuery),
				zap.Error(createErr),
				zap.Error(dropErr),
			)
			return nil, dropErr
		}
		Logger.Error(
			"Could not exec query",
			zap.String("query", createQuery),
			zap.Error(createErr),
		)
		return nil, createErr
	}

	Logger.Info(
		"Created table",
		zap.String("table name", tableName),
		zap.String("table", fmt.Sprintf("%+v", created)),
	)

	tableMap := db.AddTableWithName(UserToken{}, tableName).SetKeys(false, "ID")

	return tableMap, nil
}
