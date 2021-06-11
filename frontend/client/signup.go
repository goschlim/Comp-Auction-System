package client

import (
	"net/http"
	"time"

	"goLive/frontend/common"
	"goLive/frontend/logs"

	"goLive/frontend/validation"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// signup for new user.
//
// A http handler that validates data from HTML form inputs
// through a user, generates a hashed password, creates a
// session cookie with token, and login the new user if all
// the data validation checks passed; otherwise, an attacker
// alert is logged in the server logs if the checks failed.
//
// If the new user credentials coincide with an existing user,
// an error message is shown as "Account already taken"; otherwise,
// the new user is signed up and logged in, and redirected to the index page.
func signup(res http.ResponseWriter, req *http.Request) {
	common.Debug("signup")

	// same code as login
	var signupBidItem item

	// check if already logged in
	if isLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	// not logged in
	var getUser user

	// waiting for POST form submit()
	if req.Method == http.MethodPost {
		// [redirected form element value from index landing page]
		rawItemID := req.FormValue("itemidsignup")
		itemID, err := validation.CheckUUIDString(rawItemID)
		common.Debug("signup", itemID)
		if err != nil {
			// invalid uuid error
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("signup item", err)
			return
		}
		// has itemID, update
		if len(itemID) > 0 {
			found := false
			// check whether item exists before updating
			// happens when the itemID was changed
			for _, item := range items {
				if item.ItemID == itemID {
					found = true
					break
				}
			}
			if !found {
				// invalid uuid error
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("signup item not found")
				return
			}
			signupBidItem.ItemID = itemID
		}

		// get HTML form values
		inputUsername := req.FormValue("username")
		inputPassword := req.FormValue("password")
		inputFirstname := req.FormValue("firstname")
		inputLastname := req.FormValue("lastname")

		username, err := validation.ValidateString(inputUsername, 50)
		if err != nil {
			if len(username) == 0 {
				data := templateData{
					Item:      &signupBidItem,
					MaxLength: common.ConstMaxLengthName,
				}
				renderTemplate(res, "signup.gohtml", data)
				return
			}
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("signup username", err)
			return
		}

		// check user input
		if len(username) != 0 {
			if _, ok := mapUsers[username]; ok {
				data := templateData{
					Item:         &signupBidItem,
					MaxLength:    common.ConstMaxLengthName,
					ErrorMessage: "Account already taken",
				}
				renderTemplate(res, "signup.gohtml", data)
				return
			}

			fname, err := validation.ValidateString(inputFirstname, 50)
			if err != nil {
				s, ok := validation.GetErrorStr(err)
				if ok {
					data := templateData{
						Item:         &signupBidItem,
						MaxLength:    common.ConstMaxLengthName,
						ErrorMessage: s,
					}
					renderTemplate(res, "signup.gohtml", data)
					return
				}
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("signup firstname", err)
				return
			}

			lname, err := validation.ValidateString(inputLastname, 50)
			if err != nil {
				s, ok := validation.GetErrorStr(err)
				if ok {
					data := templateData{
						Item:         &signupBidItem,
						MaxLength:    common.ConstMaxLengthName,
						ErrorMessage: s,
					}
					renderTemplate(res, "signup.gohtml", data)
					return
				}
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("signup lastname", err)
				return
			}

			password, err := validation.ValidatePassword(inputPassword)
			if err != nil {
				s, ok := validation.GetErrorStr(err)
				if ok {
					data := templateData{
						Item:         &signupBidItem,
						MaxLength:    common.ConstMaxLengthName,
						ErrorMessage: s,
					}
					renderTemplate(res, "signup.gohtml", data)
					return
				}
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("signup pass")
				return
			}

			// hash password
			hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
			if err != nil {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("signup pass error")
				return
			}

			// [OK]
			// create token
			signedToken, claims, err := generateToken(getUser, res)
			if err != nil {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("<token error>")
				return
			}

			// create cookie
			cookie := &http.Cookie{
				Name:     cookieName,
				Value:    signedToken,
				Expires:  time.Unix(claims.ExpiresAt, 0),
				HttpOnly: true,
				Path:     "/",
				Domain:   "127.0.0.1",
				Secure:   true,
			}
			// current SessionID
			sid := cookie.Value
			// current UserID
			userid := uuid.NewV4().String()
			// issue cookie
			http.SetCookie(res, cookie)
			mapSessions[sid] = username

			// register new user
			mapUsers[username] = user{
				Username:  username,
				Password:  hashedPass,
				FirstName: fname,
				LastName:  lname,
				UserID:    userid,

				// initialize to nothing first
				ItemsBid:     []*item{},
				ItemsSell:    []*item{},
				ItemsAuction: []*item{},
			}
			addUser2db(userid, mapUsers[username])

			logs.Info("<signup new user>")
		}

		// [login or signup to bid]
		// [signup via login page via index page]
		if len(itemID) > 0 {
			var bidItem *item
			// traverse through list of items
			for _, item := range items {
				// found item
				if item.ItemID == itemID {
					bidItem = item
					break
				}
			}
			data := templateData{
				Item: bidItem,
				User: mapUsers[username],
			}
			// [redirect to itemBid page]
			renderTemplate(res, "itemBid.gohtml", data)
			return
		}

		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	data := templateData{
		Item:      &signupBidItem,
		MaxLength: common.ConstMaxLengthName,
	}
	renderTemplate(res, "signup.gohtml", data)
}
