package auth

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Arcturis-Studio/skrump/src/internal/models/smartsheet"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/auth"
)

func TranslateOAuth2UserData(userData *auth.AuthUser) (*auth.AuthUser, error) {
	smartsheetUserData := smartsheet.User{}
	jsonBytes, err := json.Marshal(userData.RawUser)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &smartsheetUserData)
	if err != nil {
		return nil, err
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
		Username:     smartsheetUserData.Email,
		AvatarUrl:    "https://aws.smartsheet.com/storageProxy/image/images/" + smartsheetUserData.ProfileImage.ImageId,
		AccessToken:  userData.AccessToken,
		RefreshToken: userData.RefreshToken,
		Expiry:       userData.Expiry,
		RawUser:      userData.RawUser,
	}

	return &u, nil
}

func mapAuthUser(user *auth.AuthUser) (map[string]any, error) {
	var userMap map[string]any
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, &userMap)
	if err != nil {
		return nil, err
	}

	return userMap, nil
}

/*
Only rare scenarios would we need to call one of these methods without the other.
So this is a simple wrapper to handle both in one method call.
*/
func UpsertFullAuthUser(provider string, authRecord *models.Record, user *auth.AuthUser, dao *daos.Dao) (*models.Record, error) {
	record, err := UpsertAuthUser(authRecord, user, dao)
	if err != nil {
		return nil, err
	}

	err = UpsertExternalAuthUser(provider, record.Collection().Id, user, dao)
	if err != nil {
		return nil, err
	}

	return record, nil
}

/*
Ideally, authRecord would scan into e.Record from the originating event.
This function actually does perform the needs of an UPSERT query. Query for
existing record based on PKEY id, if exists, INSERT, otherwise, UPDATE.
*/
func UpsertAuthUser(authRecord *models.Record, user *auth.AuthUser, dao *daos.Dao) (*models.Record, error) {
	var err error
	err = dao.FindById(authRecord, user.Id)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return nil, err
	}

	var userMap map[string]any
	userMap, err = mapAuthUser(user)
	if err != nil {
		return nil, err
	}

	authRecord.Load(userMap)

	if err = dao.SaveRecord(authRecord); err != nil {
		return nil, err
	}

	return authRecord, nil
}

/*
We call it "upserting" the external auth record, but we are really just deleting it since it seems
like we do not have direct control of when this record is created. Every auth sign-up/sign-in
attempt makes an INSERT query, so we can just delete the record and carry on.
*/
func UpsertExternalAuthUser(provider string, collectionId string, user *auth.AuthUser, dao *daos.Dao) error {
	extAuths := models.ExternalAuth{}
	err := dao.ExternalAuthQuery().Where(dbx.NewExp("`provider`={:provider} AND `recordId`={:recordId} AND `collectionId`={:collectionId}", dbx.Params{"provider": provider, "recordId": user.Id, "collectionId": collectionId})).One(&extAuths)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil
		}
		return err
	}

	err = dao.DeleteExternalAuth(&extAuths)
	if err != nil {
		return err
	}
	return nil
}
