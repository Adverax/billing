package banker

import (
	"billing/domain"
	"context"
	"github.com/adverax/echo/database/sql"
)

type HistoryManager interface {
	Append(ctx context.Context, uid int64, account uint32, amount float32, op domain.Operation) error
}

type AccountManager interface {
	Credit(ctx context.Context, account uint32, amount float32) error
	Debit(ctx context.Context, account uint32, amount float32) error
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
	accounts AccountManager
	assets   AssetManager
	history  HistoryManager
}

func (engine *engine) Credit(
	ctx context.Context,
	uid int64,
	account uint32,
	amount float32,
) error {
	return engine.Transaction(
		ctx,
		func(ctx context.Context) error {
			err := engine.history.Append(ctx, uid, account, amount, domain.OperationCredit)
			if err != nil {
				return err
			}

			return engine.accounts.Credit(ctx, account, amount)
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
		func(ctx context.Context) error {
			err := engine.history.Append(ctx, uid, account, amount, domain.OperationDebit)
			if err != nil {
				return err
			}

			return engine.accounts.Debit(ctx, account, amount)
		},
	)
}

func (engine *engine) Transfer(
	ctx context.Context,
	uid int64,
	src, dst uint32,
	amount float32,
) error {
	return engine.Transaction(
		ctx,
		func(ctx context.Context) error {
			err := engine.history.Append(ctx, uid, src, amount, domain.OperationTransferSrc)
			if err != nil {
				return err
			}

			err = engine.history.Append(ctx, uid, dst, amount, domain.OperationTransferDst)
			if err != nil {
				return err
			}

			err = engine.accounts.Credit(ctx, src, amount)
			if err != nil {
				return err
			}

			return engine.accounts.Debit(ctx, dst, amount)
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
		func(ctx context.Context) error {
			err := engine.history.Append(ctx, uid, account, amount, domain.OperationAcquire)
			if err != nil {
				return err
			}

			err = engine.accounts.Credit(ctx, account, amount)
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
		func(ctx context.Context) error {
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
		func(ctx context.Context) error {
			amount, err := engine.assets.Remove(ctx, uid, account)
			if err != nil {
				return err
			}

			err = engine.accounts.Debit(ctx, account, amount)
			if err != nil {
				return err
			}

			return engine.history.Append(ctx, uid, account, amount, domain.OperationRollback)
		},
	)
}

func New(
	db sql.DB,
	accounts AccountManager,
	assets AssetManager,
	history HistoryManager,
) Manager {
	return &engine{
		Repository: sql.NewRepository(db),
		accounts:   accounts,
		assets:     assets,
		history:    history,
	}
}
