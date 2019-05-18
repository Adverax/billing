package account

import (
	"billing/domain"
	"context"
	"github.com/adverax/echo/database/sql"
)

type Manager interface {
	Credit(ctx context.Context, account uint32, amount float32) error
	Debit(ctx context.Context, account uint32, amount float32) error
}

type engine struct {
	sql.Repository
}

func (engine *engine) Credit(
	ctx context.Context,
	account uint32,
	amount float32,
) error {
	return engine.upgrade(
		ctx,
		account,
		func(sum float32) (float32, error) {
			if sum < amount {
				return 0, domain.ErrNoMoney
			}
			return sum - amount, nil
		},
	)
}

func (engine *engine) Debit(
	ctx context.Context,
	account uint32,
	amount float32,
) error {
	return engine.upgrade(
		ctx,
		account,
		func(sum float32) (float32, error) {
			return sum + amount, nil
		},
	)
}

func (engine *engine) upgrade(
	ctx context.Context,
	account uint32,
	action func(sum float32) (float32, error),
) error {
	return engine.Transaction(
		ctx,
		func(ctx context.Context, scope sql.Scope) error {
			const query1 = "SELECT amount FROM account WHERE id = ? FOR UPDATE"
			var sum float32
			err := scope.QueryRow(query1, account).Scan(&sum)
			if err != nil {
				return err
			}

			res, err := action(sum)
			if err != nil {
				return err
			}

			const query2 = "UPDATE account SET amount = ? WHERE id = ?"
			_, err = scope.Exec(query2, res, account)
			return err
		},
	)
}

func New(
	db sql.DB,
) Manager {
	return &engine{
		Repository: sql.NewRepository(db),
	}
}
