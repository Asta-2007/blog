package mysql_auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	model_auth "blog_server/auth/model"
	repository_auth "blog_server/auth/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type authEntity struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) repository_auth.AuthEntity {
	return &authEntity{
		db: db,
	}
}

// Create implements [repository_auth.AuthEntity].
func (a authEntity) Create(ctx context.Context, entity model_auth.Entity) error {
	const query = `
		INSERT INTO auth_entities
			(id, name, status, email, password, created, updated_at)
		VALUES
			(?, ?, ?, ?, ?, ?, ?)
	`

	_, err := a.db.ExecContext(
		ctx,
		query,
		entity.ID,
		entity.Name,
		entity.Status,
		entity.Email,
		entity.Password,
		entity.Created,
		entity.UpdatedAt,
	)

	return err
}

func (a authEntity) GetByEmail(ctx context.Context, email string) (*model_auth.LoginEntity, error) {
	const query = `
		SELECT
			id,
			status,
			email,
			password
		FROM auth_entities
		WHERE email = ?
		LIMIT 1
	`

	var entity model_auth.LoginEntity

	err := a.db.QueryRowContext(ctx, query, email).Scan(
		&entity.ID,
		&entity.Status,
		&entity.Email,
		&entity.Password,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &entity, nil
}

func (a authEntity) GetByID(ctx context.Context, id string) (*model_auth.Entity, error) {
	const query = `
		SELECT
			id,
			name,
			status,
			email,
			password,
			created,
			updated_at
		FROM auth_entities
		WHERE id = ?
	`

	var entity model_auth.Entity

	err := a.db.QueryRowContext(ctx, query, id).Scan(
		&entity.ID,
		&entity.Name,
		&entity.Status,
		&entity.Email,
		&entity.Password,
		&entity.Created,
		&entity.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &entity, nil
}

func (a authEntity) GetAll(ctx context.Context, filter bson.M) ([]*model_auth.Entity, error) {
	query := `
		SELECT
			id,
			name,
			status,
			email,
			password,
			created,
			updated_at
		FROM auth_entities
	`

	var (
		args  []any
		where []string
	)

	for k, v := range filter {
		where = append(where, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*model_auth.Entity

	for rows.Next() {
		var e model_auth.Entity

		err := rows.Scan(
			&e.ID,
			&e.Name,
			&e.Status,
			&e.Email,
			&e.Password,
			&e.Created,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		entities = append(entities, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entities, nil
}

func (a authEntity) Update(ctx context.Context, entity bson.M) error {
	id, ok := entity["id"]
	if !ok {
		return fmt.Errorf("id is required")
	}

	delete(entity, "id")

	if len(entity) == 0 {
		return nil
	}

	var (
		args []any
		sets []string
	)

	for k, v := range entity {
		sets = append(sets, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}

	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE auth_entities SET %s WHERE id = ?",
		strings.Join(sets, ", "),
	)

	_, err := a.db.ExecContext(ctx, query, args...)
	return err
}

func (a authEntity) Delete(ctx context.Context, id string) error {
	const query = `
		DELETE FROM auth_entities
		WHERE id = ?
	`

	_, err := a.db.ExecContext(ctx, query, id)
	return err
}
