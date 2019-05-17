package history

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
		op domain.Operation,
	) error
}

type engine struct {
	sql.Repository
}

func (engine *engine) Append(
	ctx context.Context,
	uid int64,
	account uint32,
	amount float32,
	op domain.Operation,
) error {
	const query = "INSERT INTO history SET uid = ?, account = ?, amount = ?, op = ?"
	_, err := engine.Scope(ctx).Exec(query, uid, account, amount, op)
	return domain.HandleDeprecatedError(err)
}

func New(db sql.DB) Manager {
	return &engine{
		Repository: sql.NewRepository(db),
	}
}
