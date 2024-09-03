package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	/*
		Import automatically applies needed migrations on serve
		https://pocketbase.io/docs/go-migrations/#creating-migrations
	*/
	_ "github.com/Arcturis-Studio/skrump/pb_migrations"
	"github.com/Arcturis-Studio/skrump/src/internal/auth"
	"github.com/Arcturis-Studio/skrump/src/internal/docker"
	"github.com/Arcturis-Studio/skrump/src/internal/utils"
)

func main() {
	app := pocketbase.New()
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{Dir: "pb_migrations"})
	apis.ActivityLogger(app)
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		log.Fatal(err)
	}

	app.OnAfterBootstrap().Add(func(e *core.BootstrapEvent) error {
		// NOTE: Cleaning up after bootstrap so we are not blocking pocketbase from serving assets
		// There should not be any containers from this app unless we crashed.
		if err := dockerClient.CleanUpContainers(); err != nil {
			return err
		}

		return nil
	})

	app.OnTerminate().Add(func(e *core.TerminateEvent) error {
		// Cleanup all containers on SIGTERM
		if err := dockerClient.CleanUpContainers(); err != nil {
			return err
		}

		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), false))

		e.Router.GET("/spawn_container", func(c echo.Context) error {
			// go utils.SpawnDockerContainer("-p", "55002:80", "nginxdemos/hello")
			go dockerClient.SpawnDockerContainer("")

			return nil
		})

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
