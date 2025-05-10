package user

import (
	"database/sql"
	_ "embed"

	"sirherobrine23.com.br/go-bds/bds/modules/datas/permission"
)

//go:embed sql/sqlite_create.sql
var sqliteTableCreate string

// Create table on start
func CreateSqliteTable(conn *sql.DB) error {
	_, err := conn.Exec(sqliteTableCreate)
	return err
}

func (p *Password) sqliteStorage(passwordHash string) error {
	var count int
	err := p.DB.QueryRow("SELECT count(*) FROM password WHERE user_id == $1 LIMIT 1", p.ID).Scan(&count)
	if err != nil {
		return err
	}

	if count == 1 {
		_, err = p.DB.Exec("UPDATE password SET password = $2 WHERE user_id = $1", p.ID, passwordHash)
		return err
	}

	_, err = p.DB.Exec("INSERT INTO password (user_id, password) VALUES ($1, $2)", p.ID, passwordHash)
	return err
}

func (s *UserSearch) processSqliteRow(row *sql.Row) (*User, error) {
	if row.Err() == nil {
		user := &User{}
		if err := row.Scan(&user.ID, &user.Name, &user.Username, &user.Permission); err != nil {
			return nil, err
		}

		row = s.DB.QueryRow("SELECT password FROM password WHERE user_id = $1", user.ID)
		if row.Err() == nil {
			user.Password = &Password{DB: s.DB, Driver: s.Driver}
			if err := row.Scan(&user.Password.PasswordHash); err != nil {
				return nil, err
			}
		}

		return user, nil
	}
	return nil, row.Err()
}

func (s *UserSearch) sqliteByID(id int64) (*User, error) {
	return s.processSqliteRow(s.DB.QueryRow("SELECT id, \"name\", username, permission FROM user WHERE id = $1", id))
}

func (s *UserSearch) sqliteUsername(username string) (*User, error) {
	return s.processSqliteRow(s.DB.QueryRow("SELECT id, \"name\", username, permission FROM user WHERE username = $1", username))
}

func (s *UserSearch) sqliteCreate(name, username, email, passwordHASH string) (*User, error) {
	res, err := s.DB.Exec("INSERT INTO user(\"name\", username, email, permission) VALUES ($1, $2, $3, $4)", name, username, email, permission.Unknown)
	if err != nil {
		return nil, err
	}

	userID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	} else if _, err = s.DB.Exec("INSERT INTO password(user_id, password) VALUES ($1, $2)", userID, passwordHASH); err != nil {
		return nil, err
	}
	return s.ByID(userID)
}
