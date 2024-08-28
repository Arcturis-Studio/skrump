package main

import (
	"log"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	/*
		Import automatically applies needed migrations on serve
		https://pocketbase.io/docs/go-migrations/#creating-migrations
	*/
	_ "github.com/yourfavoritekyle/skrump/pb_migrations"
	"github.com/yourfavoritekyle/skrump/src/internal/auth"
	"github.com/yourfavoritekyle/skrump/src/internal/utils"
)

func main() {
	app := pocketbase.New()
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{Dir: "pb_migrations"})
	apis.ActivityLogger(app)

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), false))

		return nil
	})

	app.OnRecordBeforeAuthWithOAuth2Request().Add(func(e *core.RecordAuthWithOAuth2Event) error {
		var err error
		e.OAuth2User, err = auth.TranslateOAuth2UserData(e.OAuth2User)
		if err != nil {
			return err
		}

		requestInfo := apis.RequestInfo(e.HttpContext)

		if e.Record == nil {
			e.Record = models.NewRecord(e.Collection)
		}

		e.Record, err = auth.UpsertFullAuthUser(utils.AssertStringValue(requestInfo.Data["provider"]), e.Record, e.OAuth2User, app.Dao())
		if err != nil {
			return err
		}

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
