package main

import (
	"billing/domain"
	"billing/manager/account"
	"billing/manager/asset"
	"billing/manager/banker"
	"billing/manager/history"
	"billing/service"
	"context"
	"github.com/adverax/echo/log"
)

func main() {
	ctx := context.Background()

	dsc := domain.Config.Database.DSC()
	db, err := dsc.Open(nil)
	if err != nil {
		panic(err)
	}
	defer db.Close(ctx)

	err = service.Bootstrap(
		ctx,
		banker.New(
			db,
			account.New(db),
			asset.New(db),
			history.New(db),
		),
		domain.Config.Broker,
		log.NewDebug(""),
	)
	if err != nil {
		panic(err)
	}
}
