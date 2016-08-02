package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/satori/go.uuid"
	"github.com/topfreegames/khan/util"
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

// GetUserTokenByToken returns userToken with the given service,token
func GetUserTokenByToken(db DB, app string, service string, token string) (*UserToken, error) {
	var userTokens []UserToken
	tableName := GetTableName(app, service)
	query := fmt.Sprintf("SELECT * FROM %s WHERE token=$1", tableName)
	_, err := db.Select(&userTokens, query, token)

	if err != nil || &userTokens == nil || len(userTokens) == 0 {
		return nil, &ModelNotFoundError{"UserToken", "token", token}
	}
	if len(userTokens) > 1 {
		return nil, &DuplicatedTokenError{tableName, token}
	}
	return &userTokens[0], nil
}

// GetUserTokensBatchByFilters returns userTokens with the given filters starting at offset and limited to limit
func GetUserTokensBatchByFilters(db DB, app string, service string, filters [][]interface{},
	modifiers [][]interface{}) ([]UserToken, error) {
	var userTokens []UserToken
	tableName := GetTableName(app, service)

	// Build params, query filters and query modifiers
	params := []interface{}{}
	queryFilters := []string{}
	for filterIdx, filter := range filters {
		queryFilters = append(
			queryFilters,
			fmt.Sprintf("%s=$%d", filter[0], filterIdx+1),
		)
		params = append(params, filter[1])
	}

	startIdx := len(params) + 1
	queryModifiers := []string{}
	for modifierrIdx, modifier := range modifiers {
		queryModifiers = append(
			queryModifiers,
			fmt.Sprintf("%s $%d", modifier[0], modifierrIdx+startIdx),
		)
		params = append(params, modifier[1])
	}

	// Build query based in the query filters and query modifiers
	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE %s %s",
		tableName,
		strings.Join(queryFilters, " AND "),
		strings.Join(queryModifiers, " "),
	)

	// Execute query based in the given params
	_, err := db.Select(&userTokens, query, params...)
	if err != nil || &userTokens == nil {
		return nil, err
	}
	return userTokens, nil
}

// CountUserTokensByFilters returns userTokens with the given filters starting at offset and limited to limit
func CountUserTokensByFilters(db DB, app string, service string, filters [][]interface{},
	modifiers [][]interface{}) (int64, error) {
	tableName := GetTableName(app, service)

	// Build params, query filters and query modifiers
	params := []interface{}{}
	queryFilters := []string{}
	for filterIdx, filter := range filters {
		queryFilters = append(
			queryFilters,
			fmt.Sprintf("%s=$%d", filter[0], filterIdx+1),
		)
		params = append(params, filter[1])
	}

	startIdx := len(params) + 1
	queryModifiers := []string{}
	for modifierrIdx, modifier := range modifiers {
		queryModifiers = append(
			queryModifiers,
			fmt.Sprintf("%s $%d", modifier[0], modifierrIdx+startIdx),
		)
		params = append(params, modifier[1])
	}

	// Build query based in the query filters and query modifiers
	query := fmt.Sprintf(
		"SELECT COUNT (1) FROM %s WHERE %s %s",
		tableName,
		strings.Join(queryFilters, " AND "),
		strings.Join(queryModifiers, " "),
	)

	// Execute query based in the given params
	userTokensCount, err := db.SelectInt(query, params...)
	if err != nil {
		return -1, err
	}
	return userTokensCount, nil
}

// UpsertToken inserts or updates a Token
func UpsertToken(db DB, app string, service string, userID string, token string, locale string,
	region string, tz string, buildn string, optOut []string) (*UserToken, error) {
	tableName := GetTableName(app, service)

	userToken, err := GetUserTokenByToken(db, app, service, token)
	if err != nil {
		if _, same := err.(*ModelNotFoundError); !same {
			return nil, err
		}
	}

	if userToken != nil && userToken.UserID != userID {
		_, err = db.Delete(userToken)
		if err != nil {
			return nil, err
		}
	}

	params := []interface{}{userID, token, locale, region, tz, buildn, util.NowMilli(), util.NowMilli()}
	startOptOutsAt := 9
	optOuts := []string{}
	for i := range optOut {
		params = append(params, optOut[i])
		optOuts = append(optOuts, fmt.Sprintf("$%d", startOptOutsAt+i))
	}

	query := `INSERT INTO %s (user_id, token, locale, region, tz, build_n, created_at, updated_at, opt_out)
	  VALUES($1, $2, $3, $4, $5, $6, $7, $8, ARRAY[%s])
    ON CONFLICT (user_id, token)
	    DO UPDATE SET locale=$3, region=$4, tz=$5, build_n=$6, created_at=EXCLUDED.created_at,
      updated_at=$8, opt_out=ARRAY[%s]`

	query = fmt.Sprintf(query, tableName, strings.Join(optOuts, ","), strings.Join(optOuts, ","))

	_, err = db.Exec(query, params...)
	if err != nil {
		return nil, err
	}

	userToken, err = GetUserTokenByToken(db, app, service, token)
	if err != nil {
		return nil, err
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
    CREATE INDEX IF NOT EXISTS index_%s_on_user_id ON %s (user_id);
    CREATE UNIQUE INDEX IF NOT EXISTS unique_index_%s_on_user_id ON %s (user_id, token)
    `,
		tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName,
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
