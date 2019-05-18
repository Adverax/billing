package account

import (
	"billing/domain"
	"context"
	"github.com/adverax/echo/database/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func setUp() (context.Context, sql.DB) {
	ctx := context.Background()
	return ctx, domain.Config.Database.DSC().OpenForTest(ctx)
}

func TestEngine_Credit(t *testing.T) {
	type Src struct {
		account uint32
		amount  float32
		source  float32
	}

	type Dst struct {
		amount float32
		err    error
	}

	type Test struct {
		src Src
		dst Dst
	}

	tests := map[string]Test{
		"Valid payment must be accepted": {
			src: Src{
				account: 1,
				amount:  50,
				source:  100,
			},
			dst: Dst{
				amount: 50,
			},
		},
		"Invalid payer must be skipped": {
			src: Src{
				account: 2,
				amount:  50,
				source:  100,
			},
			dst: Dst{
				err: sql.ErrNoRows,
			},
		},
	}

	ctx, db := setUp()
	defer db.Close(ctx)

	e := New(db)

	const query = `
DELETE FROM account;
INSERT INTO account SET id = 1;`
	_, err := db.Exec(query)
	require.NoError(t, err)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			const query = "UPDATE account SET amount = ?"
			_, err := db.Exec(query, test.src.source)
			require.NoError(t, err)
			err = e.Credit(ctx, test.src.account, test.src.amount)
			require.Equal(t, test.dst.err, err)
			if err != nil {
				return
			}

			amount, err := getAccount(db, test.src.account)
			require.NoError(t, err)
			assert.Equal(t, test.dst.amount, amount)
		})
	}
}

func TestEngine_Debit(t *testing.T) {
	type Src struct {
		account uint32
		amount  float32
		source  float32
	}

	type Dst struct {
		amount float32
		err    error
	}

	type Test struct {
		src Src
		dst Dst
	}

	tests := map[string]Test{
		"Valid payment must be accepted": {
			src: Src{
				account: 1,
				amount:  50,
				source:  100,
			},
			dst: Dst{
				amount: 150,
			},
		},
		"Invalid payer must be skipped": {
			src: Src{
				account: 2,
				amount:  50,
				source:  100,
			},
			dst: Dst{
				err: sql.ErrNoRows,
			},
		},
	}

	ctx, db := setUp()
	defer db.Close(ctx)

	e := New(db)

	const query = `
DELETE FROM account;
INSERT INTO account SET id = 1;`
	_, err := db.Exec(query)
	require.NoError(t, err)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			const query = "UPDATE account SET amount = ?"
			_, err := db.Exec(query, test.src.source)
			require.NoError(t, err)
			err = e.Debit(ctx, test.src.account, test.src.amount)
			require.Equal(t, test.dst.err, err)
			if err != nil {
				return
			}

			amount, err := getAccount(db, test.src.account)
			require.NoError(t, err)
			assert.Equal(t, test.dst.amount, amount)
		})
	}
}

func getAccount(db sql.DB, id uint32) (amount float32, err error) {
	const query = "SELECT amount FROM account WHERE id = ?"
	err = db.QueryRow(query, id).Scan(&amount)
	if err != nil {
		return 0, err
	}
	return
}
