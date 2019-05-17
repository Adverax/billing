package banker

import (
	"billing/domain"
	"context"
	"github.com/adverax/echo/database/sql"
)

type HistoryManager interface {
	Append(ctx context.Context, uid int64, account uint32, amount float32, op domain.Operation) error
}

type AssetManager interface {
	Append(ctx context.Context, uid int64, account uint32, amount float32) error
	Remove(ctx context.Context, uid int64, account uint32) (amount float32, err error)
}

type Manager interface {
	Credit(ctx context.Context, uid int64, account uint32, amount float32) error
	Debit(ctx context.Context, uid int64, account uint32, amount float32) error
	Transfer(ctx context.Context, uid int64, src, dst uint32, amount float32) error
	Acquire(ctx context.Context, uid int64, account uint32, amount float32) error
	Commit(ctx context.Context, uid int64, account uint32) error
	Rollback(ctx context.Context, uid int64, account uint32) error
}

type engine struct {
	sql.Repository
	assets  AssetManager
	history HistoryManager
}

func (engine *engine) Credit(
	ctx context.Context,
	uid int64,
	account uint32,
	amount float32,
) error {
	return engine.Transaction(
		ctx,
		func(ctx context.Context, scope sql.Scope) error {
			err := engine.history.Append(ctx, uid, account, amount, domain.OperationCredit)
			if err != nil {
				return err
			}

			return engine.credit(ctx, account, amount)
		},
	)
}

func (engine *engine) credit(
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
	uid int64,
	account uint32,
	amount float32,
) error {
	return engine.Transaction(
		ctx,
		func(ctx context.Context, scope sql.Scope) error {
			err := engine.history.Append(ctx, uid, account, amount, domain.OperationDebit)
			if err != nil {
				return err
			}

			return engine.debit(ctx, account, amount)
		},
	)
}

func (engine *engine) debit(
	ctx context.Context,
	account uint32,
	amount float32,
) error {
	const query = "UPDATE account SET amount = amount + ? WHERE id = ?"
	res, err := engine.Scope(ctx).Exec(query, amount, account)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return err
}

func (engine *engine) Transfer(
	ctx context.Context,
	uid int64,
	src, dst uint32,
	amount float32,
) error {
	return engine.Transaction(
		ctx,
		func(ctx context.Context, scope sql.Scope) error {
			err := engine.history.Append(ctx, uid, src, amount, domain.OperationTransferSrc)
			if err != nil {
				return err
			}

			err = engine.history.Append(ctx, uid, dst, amount, domain.OperationTransferDst)
			if err != nil {
				return err
			}

			err = engine.credit(ctx, src, amount)
			if err != nil {
				return err
			}

			return engine.debit(ctx, dst, amount)
		},
	)
}

func (engine *engine) Acquire(
	ctx context.Context,
	uid int64,
	account uint32,
	amount float32,
) error {
	return engine.Transaction(
		ctx,
		func(ctx context.Context, scope sql.Scope) error {
			err := engine.history.Append(ctx, uid, account, amount, domain.OperationAcquire)
			if err != nil {
				return err
			}

			err = engine.credit(ctx, account, amount)
			if err != nil {
				return err
			}

			return engine.assets.Append(ctx, uid, account, amount)
		},
	)
}

func (engine *engine) Commit(
	ctx context.Context,
	uid int64,
	account uint32,
) error {
	return engine.Transaction(
		ctx,
		func(ctx context.Context, scope sql.Scope) error {
			amount, err := engine.assets.Remove(ctx, uid, account)
			if err != nil {
				return err
			}

			return engine.history.Append(ctx, uid, account, amount, domain.OperationCommit)
		},
	)
}

func (engine *engine) Rollback(
	ctx context.Context,
	uid int64,
	account uint32,
) error {
	return engine.Transaction(
		ctx,
		func(ctx context.Context, scope sql.Scope) error {
			amount, err := engine.assets.Remove(ctx, uid, account)
			if err != nil {
				return err
			}

			err = engine.debit(ctx, account, amount)
			if err != nil {
				return err
			}

			return engine.history.Append(ctx, uid, account, amount, domain.OperationRollback)
		},
	)
}

func New(
	db sql.DB,
	assets AssetManager,
	history HistoryManager,
) Manager {
	return &engine{
		Repository: sql.NewRepository(db),
		assets:     assets,
		history:    history,
	}
}
