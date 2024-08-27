package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/auth"
)

func main() {
	app := pocketbase.New()
	apis.ActivityLogger(app)

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), false))

		return nil
	})

	app.OnRecordBeforeAuthWithOAuth2Request().Add(func(e *core.RecordAuthWithOAuth2Event) error {
		err := handleBeforeAuthRecord(e, app)
		if err != nil {
			return err
		}

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func handleBeforeAuthRecord(e *core.RecordAuthWithOAuth2Event, app *pocketbase.PocketBase) error {
	/*
	   We need to modify request data prior to record creation to add Smartsheet support.
	   Smartsheet's user field ids do not match with the expected OpenID Connect (OIDC) fields.
	*/
	e.OAuth2User = translateOAuth2UserData(e.OAuth2User)

	if e.Record == nil {
		e.Record = models.NewRecord(e.Collection)
	}

	err := app.Dao().FindById(e.Record, e.OAuth2User.Id)
	if err != nil && err.Error() != "sql: no rows in result set" {
		log.Fatal(err)
		return err
	}

	e.Record.SetId(e.OAuth2User.Id)
	e.Record.SetEmail(e.OAuth2User.Email)
	e.Record.SetTokenKey(e.OAuth2User.AccessToken)
	e.Record.SetVerified(true)
	e.Record.Set("name", e.OAuth2User.Name)
	e.Record.Set("avatarUrl", e.OAuth2User.AvatarUrl)
	e.Record.SetUsername(e.OAuth2User.Email)

	if err := app.Dao().SaveRecord(e.Record); err != nil {
		return err
	}

	/*
		Check and delete any existing externalAuth records matching that user.
		There should only ever be one.
		This value will immediately get reset, but we need to delete it since the underlying
		code seems to only use INSERT and not swap to an UPDATE when needed.
	*/
	extAuths := models.ExternalAuth{}
	err = app.Dao().ExternalAuthQuery().Where(dbx.NewExp("`providerId`={:providerId}", dbx.Params{"providerId": e.Record.Id})).One(&extAuths)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil
		}
		log.Fatal(err)
		return err
	}

	err = app.Dao().DeleteExternalAuth(&extAuths)
	if err != nil {
		return err
	}

	return nil
}

func AssertStringValue(d any) string {
	s, ok := d.(string)
	if !ok {
		log.Fatal("expected value as string")
	}
	return s
}

func translateOAuth2UserData(userData *auth.AuthUser) *auth.AuthUser {
	smartsheetUserData := SmartsheetUser{}
	jsonBytes, err := json.Marshal(userData.RawUser)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	err = json.Unmarshal(jsonBytes, &smartsheetUserData)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	var sb strings.Builder
	if smartsheetUserData.FirstName != nil {
		sb.WriteString(*smartsheetUserData.FirstName)
	}
	if smartsheetUserData.LastName != nil {
		sb.WriteString(" ")
		sb.WriteString(*smartsheetUserData.LastName)
	}

	u := auth.AuthUser{
		Id:           strconv.FormatInt(*smartsheetUserData.Id, 10),
		Name:         sb.String(),
		Email:        smartsheetUserData.Email,
		AvatarUrl:    "https://aws.smartsheet.com/storageProxy/image/images/" + smartsheetUserData.ProfileImage.ImageId,
		AccessToken:  userData.AccessToken,
		RefreshToken: userData.RefreshToken,
		Expiry:       userData.Expiry,
		RawUser:      userData.RawUser,
	}

	return &u
}

type SmartsheetUser struct {
	Id                        *int64            `json:"id"`
	Account                   *Account          `json:"account,omitempty"`
	Admin                     *bool             `json:"admin,omitempty"`
	AlternateEmails           []AlternateEmails `json:"alternateEmails,omitempty"`
	Company                   *string           `json:"company,omitempty"`
	CustomWelcomeScreenViewed *time.Time        `json:"customWelcomeScreenViewed,omitempty"`
	Department                *string           `json:"department,omitempty"`
	Email                     string            `json:"email"`
	FirstName                 *string           `json:"firstName,omitempty"`
	GroupAdmin                bool              `json:"groupAdmin"`
	JiraAdmin                 *bool             `json:"jiraAdmin,omitempty"`
	LastLogin                 *time.Time        `json:"lastLogin,omitempty"`
	LastName                  *string           `json:"lastName,omitempty"`
	LicensedSheetCreator      *bool             `json:"licensedSheetCreator"`
	Locale                    *string           `json:"locale,omitempty"`
	MobilePhone               *string           `json:"mobilePhone,omitempty"`
	ProfileImage              ProfileImage      `json:"profileImage"`
	ResourceViewer            *bool             `json:"resourceViewer,omitempty"`
	Role                      *string           `json:"role,omitempty"`
	SalesforceAdmin           *bool             `json:"salesforceAdmin,omitempty"`
	SalesforceUser            *bool             `json:"salesforceUser,omitempty"`
	SheetCount                *int              `json:"sheetCount,omitempty"`
	TimeZone                  *string           `json:"timeZone,omitempty"`
	Title                     *string           `json:"title,omitempty"`
	WorkPhone                 *string           `json:"workPhone,omitempty"`
}
type Account struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}
type AlternateEmails struct {
	Id        int64  `json:"id"`
	Confirmed bool   `json:"confirmed,omitempty"`
	Email     string `json:"email"`
}
type ProfileImage struct {
	ImageId string `json:"imageId"`
	Height  int    `json:"height"`
	Width   int    `json:"width"`
}
