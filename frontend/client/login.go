package client

import (
	"net/http"

	"goLive/frontend/common"
	"goLive/frontend/logs"

	"goLive/frontend/validation"

	"golang.org/x/crypto/bcrypt"
)

// Login for existing Customer.
//
// A http handler that performs validation checks on HTML form inputs
// through a user, generates a hashed password, creates a
// session cookie with token, and login the existing user if all
// the data validation checks passed; otherwise, an attacker
// alert is logged in the server logs if the checks failed.
//
// If the user credentials do not match the database,
// an error message is shown as "No matching username or password"; otherwise,
// the existing user is logged in, and redirected to the index page.
func login(res http.ResponseWriter, req *http.Request) {
	common.Debug("login")

	var loginBidItem item

	// check if already logged in
	if isLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	// POST
	if req.Method == http.MethodPost {
		// [redirected form element value from index landing page]
		rawItemID := req.FormValue("itemidlogin")
		itemID, err := validation.CheckUUIDString(rawItemID)
		common.Debug("login", itemID)
		if err != nil {
			// invalid uuid error
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("login item", err)
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
				logs.Error("login item not found")
				return
			}
			// update found item
			loginBidItem.ItemID = itemID
		}

		// get form elements value
		inputUsername := req.FormValue("username")
		inputPassword := req.FormValue("password")

		username, err := validation.ValidateString(inputUsername, 50)
		if err != nil {
			if len(username) == 0 {
				data := templateData{
					Item:      &loginBidItem,
					MaxLength: common.ConstMaxLengthName,
				}
				renderTemplate(res, "login.gohtml", data)
				return
			}
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("login username", err)
			return
		}

		password, err := validation.ValidatePassword(inputPassword)
		if err != nil {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("login pass")
			return
		}

		getUser, ok := mapUsers[username] // use api to fetch data from database
		if !ok {
			data := templateData{
				Item:         &loginBidItem,
				MaxLength:    common.ConstMaxLengthName,
				ErrorMessage: "No matching username or password",
			}
			renderTemplate(res, "login.gohtml", data)
			return
		}

		err = bcrypt.CompareHashAndPassword(getUser.Password, []byte(password))
		if err != nil {
			data := templateData{
				Item:         &loginBidItem,
				MaxLength:    common.ConstMaxLengthName,
				ErrorMessage: "No matching username or password",
			}
			renderTemplate(res, "login.gohtml", data)
			return
		}

		cookie := setToken(username)
		http.SetCookie(res, cookie)
		logs.Info("cookie set", username)

		// update current session
		mapSessions[cookie.Value] = username

		// [login or signup to bid]
		// [login via index page]
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

		data := templateData{
			User: mapUsers[username],
		}
		renderTemplate(res, "profile.gohtml", data)
		return
	}

	data := templateData{
		Item:      &loginBidItem,
		MaxLength: common.ConstMaxLengthName,
	}
	renderTemplate(res, "login.gohtml", data)
}
