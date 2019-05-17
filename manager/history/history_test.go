package history

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
		op      domain.Operation
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
					op:      1,
				},
			},
			dst: Dst{
				Row: &Row{
					id:      2,
					uid:     100,
					account: 1,
					amount:  20,
					op:      1,
				},
			},
		},
		"Duplicated row must be rejected": {
			src: Src{
				Row: Row{
					uid:     1,
					account: 1,
					amount:  20,
					op:      1,
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
DELETE FROM history; 
ALTER TABLE history AUTO_INCREMENT=1;
INSERT INTO history SET uid = 1, account = 1, amount = 10, op = 1;
` + test.src.init
			_, err := db.Exec(query)
			require.NoError(t, err)

			err = e.Append(ctx,
				test.src.uid,
				test.src.account,
				test.src.amount,
				test.src.op,
			)
			require.Equal(t, test.dst.err, err)
			if test.dst.Row == nil {
				return
			}

			const query2 = "SELECT uid, account, amount, op FROM history WHERE id = ?"
			row := Row{id: test.dst.id}
			err = db.QueryRow(query2, test.dst.id).Scan(&row.uid, &row.account, &row.amount, &row.op)
			require.NoError(t, err)
			assert.Equal(t, test.dst.Row, &row)
		})
	}
}
