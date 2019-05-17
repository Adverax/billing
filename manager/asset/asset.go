package asset

import (
	"billing/domain"
	"context"
	"github.com/adverax/echo/database/sql"
)

type Manager interface {
	Append(
		ctx context.Context,
		uid int64,
		account uint32,
		amount float32,
	) error
	Remove(ctx context.Context,
		uid int64,
		account uint32,
	) (amount float32, err error)
}

type engine struct {
	sql.Repository
}

func (engine *engine) Append(
	ctx context.Context,
	uid int64,
	account uint32,
	amount float32,
) error {
	const query = "INSERT INTO asset SET uid = ?, account = ?, amount = ?"
	_, err := engine.Scope(ctx).Exec(query, uid, account, amount)
	return domain.HandleDeprecatedError(err)
}

func (engine *engine) Remove(
	ctx context.Context,
	uid int64,
	account uint32,
) (amount float32, err error) {
	err = engine.Transaction(
		ctx,
		func(ctx context.Context, scope sql.Scope) error {
			const query1 = "SELECT id, amount FROM asset WHERE uid = ? AND account = ?"
			var id int64
			err := scope.QueryRow(query1, uid, account).Scan(&id, &amount)
			if err != nil {
				return err
			}

			const query2 = "DELETE FROM asset WHERE id = ?"
			_, err = scope.Exec(query2, id)
			return err
		},
	)
	return
}

func New(db sql.DB) Manager {
	return &engine{
		Repository: sql.NewRepository(db),
	}
}
