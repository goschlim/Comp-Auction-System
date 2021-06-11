package client

import (
	"html/template"
	"net/http"
	"time"

	"goLive/frontend/common"
	"goLive/frontend/logs"

	"github.com/dgrijalva/jwt-go"
)

type user struct {
	Username  string `json:"userName"`
	Password  []byte `json:"password"`
	FirstName string `json:"email"`
	LastName  string `json:"address"`
	UserID    string // uuid from "github.com/satori/go.uuid"

	ItemsBid     []*item
	ItemsSell    []*item
	ItemsAuction []*item
}

type item struct {
	ItemID    string
	ItemName  string `json:"itemName"`
	ItemDesc  string `json:"itemDesc"`
	ItemImage string `json:"itemImage"`

	ItemBidPrice     string `json:"bidPrice"`
	ItemBidPriceStep string `json:"bidIncrement"`
	ItemClosingDate  string `json:"bidCloseDate"`

	ItemBidStatus int `json:"displayItem"` // 0:false | 1:true
}

type pointer struct {
	pointer **item
}

type templateData struct {
	Items []*item
	User  user

	Item         *item // displays this item details
	Results      int   // show total results of items
	MaxLength    int
	ErrorMessage string

	QueryString string // search query
}

var tp *template.Template

var mapUsers = map[string]user{}

var mapSessions = map[string]string{}

// hardcoded items of stored pointers
var items = []*item{}

func loadItems() {
	//call item data from db via REST API
	arr := getItemFromDB()

	for _, el := range arr {
		p := &item{
			ItemID:    el.ItemID,
			ItemName:  el.ItemName,
			ItemDesc:  el.ItemDesc,
			ItemImage: el.ItemImage,

			ItemBidPrice:     el.ItemBidPrice,
			ItemBidPriceStep: el.ItemBidPriceStep,
			ItemClosingDate:  el.ItemClosingDate,

			ItemBidStatus: el.ItemBidStatus,
		}

		item := pointer{
			pointer: &p,
		}
		items = append(items, *item.pointer)
	}
	//common.Debug(items)
}

func init() {
	// read all templates
	tp = template.Must(template.ParseGlob("templates/*"))

	// get users from store
	loadUsersInfo()
	loadItems()
}

// renderTemplate is a http handler that performs
// error handling on template.
//
// Logs Fatal on template error.
func renderTemplate(res http.ResponseWriter, templateFile string, data interface{}) {
	err := tp.ExecuteTemplate(res, templateFile, data)
	if err != nil {
		http.Error(res, "Internal server error", http.StatusInternalServerError)
		logs.Fatal(err)
	}
}

// loadUsersInfo loads from db via REST API, for server processing later on.
func loadUsersInfo() {

	//call user data from db via REST API
	arr := getAllUsersFromDB()

	for _, el := range arr {
		mapUsers[el.Username] = el
		common.Debug("user added", mapUsers[el.Username])
	}
}

// getUser gets a user by checking the session cookie.
//
// If the user does not exist in the memory of mapSessions,
// a new session cookie is created for the user.
//
// Otherwise, the user is retrieved from the mapUsers by its map key.
func getUser(res http.ResponseWriter, req *http.Request) user {
	// check current session cookie
	cookie, err := req.Cookie(cookieName)
	// create a new cookie if not found
	if err != nil {
		var getUser user
		// create token
		signedToken, claims, err := generateToken(getUser, res)
		if err != nil {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("<token error>")
			return getUser
		}

		// create cookie
		cookie = &http.Cookie{
			Name:     cookieName,
			Value:    signedToken,
			Expires:  time.Unix(claims.ExpiresAt, 0),
			HttpOnly: true,
			Path:     "/",
			Domain:   "127.0.0.1",
			Secure:   true,
		}
	}
	// issue (update) cookie
	http.SetCookie(res, cookie)

	// if user exists, get user by key
	var getUser user
	if username, ok := mapSessions[cookie.Value]; ok {
		getUser = mapUsers[username]
	}
	return getUser
}

// isLoggedIn checks whether user is login based on cookie.
func isLoggedIn(req *http.Request) bool {
	// check cookie
	cookie, err := req.Cookie(cookieName)
	// user not found
	if err != nil {
		return false
	}

	// [otherwise, user found]
	// get JWT signed token from cookie
	signedToken := cookie.Value

	// set new object of claims
	claims := &claims{}

	jwtKey := GetEnvJWT()

	// parse JWT signed token and store result in claims
	// false if invalid token (due to expired time, or signature mismatch)
	token, err := jwt.ParseWithClaims(signedToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})
	if err == nil && token.Valid {
		// user found
		username := mapSessions[cookie.Value]
		_, ok := mapUsers[username]
		return ok
	}

	// user not found
	return false
}
