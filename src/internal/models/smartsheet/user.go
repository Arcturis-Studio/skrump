package smartsheet

import "time"

type User struct {
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
