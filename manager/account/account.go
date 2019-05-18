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
	return engine.Transaction(
		ctx,
		func(ctx context.Context, scope sql.Scope) error {
			const query1 = "SELECT amount FROM account WHERE id = ? FOR UPDATE"
			var sum float32
			err := scope.QueryRow(query1, account).Scan(&sum)
			if err != nil {
				return err
			}

			if sum < amount {
				return domain.ErrNoMoney
			}

			const query2 = "UPDATE account SET amount = ? WHERE id = ?"
			_, err = scope.Exec(query2, sum-amount, account)
			return err
		},
	)
}

func (engine *engine) Debit(
	ctx context.Context,
	account uint32,
	amount float32,
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

			const query2 = "UPDATE account SET amount = ? WHERE id = ?"
			_, err = scope.Exec(query2, sum+amount, account)
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
