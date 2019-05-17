package service

import (
	"billing/domain"
	"billing/manager/banker"
	"context"
	"github.com/adverax/echo/data"
	"github.com/adverax/echo/log"
	"github.com/nats-io/go-nats"
	"os"
	"os/signal"
	"syscall"
)

func Bootstrap(
	ctx context.Context,
	manager banker.Manager,
	options domain.BrokerOptions,
	logger log.Logger,
) error {
	nc, err := nats.Connect(options.Server)
	if err != nil {
		return err
	}

	c, _ := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		return err
	}
	defer c.Close()

	err = subscribeAll(ctx, c, manager, logger)
	if err != nil {
		return err
	}

	logger.Info("Service is started")
	abort := make(chan os.Signal)
	signal.Notify(abort, syscall.SIGINT, syscall.SIGTERM)
	<-abort

	return nil
}

func subscribeAll(
	ctx context.Context,
	c *nats.EncodedConn,
	manager banker.Manager,
	logger log.Logger,
) error {
	return subscribe(
		c,
		map[string]interface{}{
			"bank.credit": func(subj, reply string, r *CreditRequest) {
				defer handlePanic(logger)
				err := manager.Credit(ctx, r.Uid, r.Account, r.Amount)
				_ = c.Publish(reply,
					CreditResponse{Status: getStatus(err, logger)},
				)
			},
			"bank.debit": func(subj, reply string, r *DebitRequest) {
				defer handlePanic(logger)
				err := manager.Debit(ctx, r.Uid, r.Account, r.Amount)
				_ = c.Publish(reply,
					DebitResponse{Status: getStatus(err, logger)},
				)
			},
			"bank.transfer": func(subj, reply string, r *TransferRequest) {
				defer handlePanic(logger)
				err := manager.Transfer(ctx, r.Uid, r.Src, r.Dst, r.Amount)
				_ = c.Publish(reply,
					TransferResponse{Status: getStatus(err, logger)},
				)
			},
			"bank.acquire": func(subj, reply string, r *AcquireRequest) {
				defer handlePanic(logger)
				err := manager.Acquire(ctx, r.Uid, r.Account, r.Amount)
				_ = c.Publish(reply,
					AcquireResponse{Status: getStatus(err, logger)},
				)
			},
			"bank.commit": func(subj, reply string, r *CommitRequest) {
				defer handlePanic(logger)
				err := manager.Commit(ctx, r.Uid, r.Account)
				_ = c.Publish(reply,
					CommitResponse{Status: getStatus(err, logger)},
				)
			},
			"bank.rollback": func(subj, reply string, r *RollbackRequest) {
				defer handlePanic(logger)
				err := manager.Rollback(ctx, r.Uid, r.Account)
				_ = c.Publish(reply,
					RollbackResponse{Status: getStatus(err, logger)},
				)
			},
		},
	)
}

func subscribe(
	c *nats.EncodedConn,
	handlers map[string]interface{},
) error {
	for key, handler := range handlers {
		_, err := c.QueueSubscribe(
			key,
			key,
			handler,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func getStatus(err error, logger log.Logger) uint8 {
	if err == nil {
		return domain.StatusOk
	}

	switch err {
	case domain.ErrNoMoney:
		return domain.StatusNoMoney
	case domain.ErrOperationIsDeprecated:
		return domain.StatusDeprecated
	case data.ErrNoMatch:
		return domain.StatusNotFound
	default:
		logger.Error(err)
		return domain.StatusUnknownError
	}
}

func handlePanic(logger log.Logger) {
	e := recover()
	if e != nil {
		logger.Error(e)
	}
}
