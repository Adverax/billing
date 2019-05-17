package domain

import (
	"errors"
	"github.com/go-sql-driver/mysql"
)

const (
	OperationCredit Operation = iota + 1
	OperationDebit
	OperationTransferSrc
	OperationTransferDst
	OperationAcquire
	OperationCommit
	OperationRollback
)

const (
	StatusOk = iota
	StatusUnknownError
	StatusDeprecated
	StatusNoMoney
	StatusNotFound
)

type Operation uint8

var ErrNoMoney = errors.New("no money")
var ErrOperationIsDeprecated = errors.New("operation is deprecated")

func IsDuplicateKeyError(err error) bool {
	if mysqlError, ok := err.(*mysql.MySQLError); ok {
		return mysqlError.Number == 0x426
	}
	return false
}

func HandleDeprecatedError(err error) error {
	if err == nil {
		return nil
	}
	if IsDuplicateKeyError(err) {
		return ErrOperationIsDeprecated
	}
	return err
}
