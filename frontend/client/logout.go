package client

import (
	"net/http"
	"time"

	"goLive/frontend/common"
)

// logout from the server.
//
// A http handler that handles logout session by invalidating
// the current user session by deleting the session cookie.
func logout(res http.ResponseWriter, req *http.Request) {
	common.Debug("logout")

	// request cookie
	cookie, _ := req.Cookie(cookieName)

	// [delete session entry from map]
	delete(mapSessions, cookie.Value)

	// remove the cookie by updating cookie
	cookie = &http.Cookie{
		Name:    cookieName,
		Value:   "none",
		Expires: time.Now(),
	}
	http.SetCookie(res, cookie)

	http.Redirect(res, req, "/", http.StatusSeeOther)
}
