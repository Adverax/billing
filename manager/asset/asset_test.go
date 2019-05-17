package asset

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

func TestEngine_Append(t *testing.T) {
	type Row struct {
		id      int64
		uid     int64
		account uint32
		amount  float32
	}

	type Src struct {
		Row
		init string
	}

	type Dst struct {
		*Row
		err error
	}

	type Test struct {
		src Src
		dst Dst
	}

	tests := map[string]Test{
		"Unique row must be accepted": {
			src: Src{
				Row: Row{
					uid:     100,
					account: 1,
					amount:  20,
				},
			},
			dst: Dst{
				Row: &Row{
					id:      2,
					uid:     100,
					account: 1,
					amount:  20,
				},
			},
		},
		"Duplicated row must be rejected": {
			src: Src{
				Row: Row{
					uid:     1,
					account: 1,
					amount:  20,
				},
			},
			dst: Dst{
				err: domain.ErrOperationIsDeprecated,
			},
		},
	}

	ctx, db := setUp()
	defer db.Close(ctx)

	e := New(db)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			query := `
DELETE FROM account;
INSERT INTO account SET id = 1; 
DELETE FROM asset; 
ALTER TABLE asset AUTO_INCREMENT=1;
INSERT INTO asset SET uid = 1, account = 1, amount = 10;
` + test.src.init
			_, err := db.Exec(query)
			require.NoError(t, err)

			err = e.Append(ctx,
				test.src.uid,
				test.src.account,
				test.src.amount,
			)
			require.Equal(t, test.dst.err, err)
			if test.dst.Row == nil {
				return
			}

			const query2 = "SELECT uid, account, amount FROM asset WHERE id = ?"
			row := Row{id: test.dst.id}
			err = db.QueryRow(query2, test.dst.id).Scan(&row.uid, &row.account, &row.amount)
			require.NoError(t, err)
			assert.Equal(t, test.dst.Row, &row)
		})
	}
}

func TestEngine_Remove(t *testing.T) {
	type Src struct {
		init    string
		uid     int64
		account uint32
	}

	type Test struct {
		src Src
		dst error
	}

	tests := map[string]Test{
		"Existing row must be deleted": {
			src: Src{
				uid:     1,
				account: 1,
			},
			dst: nil,
		},
		"Invalid row must throw error": {
			src: Src{
				uid:     2,
				account: 1,
			},
			dst: sql.ErrNoRows,
		},
	}

	ctx, db := setUp()
	defer db.Close(ctx)

	e := New(db)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			query := `
DELETE FROM account;
INSERT INTO account SET id = 1; 
DELETE FROM asset; 
ALTER TABLE asset AUTO_INCREMENT=1;
INSERT INTO asset SET uid = 1, account = 1, amount = 10;
` + test.src.init
			_, err := db.Exec(query)
			require.NoError(t, err)

			_, err = e.Remove(ctx,
				test.src.uid,
				test.src.account,
			)
			require.Equal(t, test.dst, err)
		})
	}
}
