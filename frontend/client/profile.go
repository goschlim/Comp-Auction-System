package client

import (
	"net/http"

	"goLive/frontend/common"
	"goLive/frontend/logs"

	"goLive/frontend/validation"
)

// profile access via authenticated login.
//
// A http handler that checks whether the user is logged in,
// and redirects to the profile template.
func profile(res http.ResponseWriter, req *http.Request) {
	// not logged in, redirect to index
	if !isLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	// already logged in
	getUser := getUser(res, req)

	// form elements value
	rawItemID := req.FormValue("itemid")
	rawItemIDDelete := req.FormValue("itemdeleteid")

	// POST
	// [user click "Yes" to bid, or cancel bid]
	if req.Method == http.MethodPost &&
		len(rawItemID) > 0 {

		itemID, err := validation.CheckUUIDString(rawItemID)
		if err != nil || len(itemID) == 0 {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("profile", err)
			return
		}

		rawItemBidPrice := req.FormValue("newbidprice")
		// "Yes" to bid
		if len(rawItemBidPrice) != 0 {
			bidPrice, bpErr := validation.CheckString(rawItemBidPrice)
			if bpErr != nil {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("profile", err)
				return

			} else if len(itemID) == 0 || len(bidPrice) == 0 {
				data := templateData{
					Items: sortItemsByDateASC(items),
					User:  getUser,
				}
				renderTemplate(res, "index.gohtml", data)
				return
			}

			// catch user by refreshing the page
			// search through user bid items
			found := false
			for _, item := range getUser.ItemsBid {
				if item.ItemID == itemID {
					found = true
					break
				}
			}
			// add new item if not found
			if !found {
				// search through items
				for _, item := range items {
					if item.ItemID == itemID {
						// update item field values
						item.ItemBidPrice = bidPrice
						item.ItemBidStatus = 1
						// add item to slice by updating user map of items
						getUser.ItemsBid = append(getUser.ItemsBid, item)
						break
					}
				}
				common.Debug(bidPrice)
			}

		} else {
			// "Yes" to cancel bid

			// search through items
			for pos, item := range getUser.ItemsBid {
				if item.ItemID == itemID {
					// update item field values
					item.ItemBidStatus = 0
					// remove item from slice by updating user map of items
					getUser.ItemsBid = append(getUser.ItemsBid[:pos], getUser.ItemsBid[pos+1:]...)
					break
				}
			}
		}

	} else if req.Method == http.MethodPost &&
		// [user click "Yes" to delete placement]
		len(rawItemIDDelete) > 0 {

		itemID, err := validation.CheckUUIDString(rawItemIDDelete)
		if err != nil || len(itemID) == 0 {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("profile", err)
			return
		}

		// search through user selling items
		for pos, item := range getUser.ItemsSell {
			if item.ItemID == itemID {
				// remove item found
				getUser.ItemsSell = append(getUser.ItemsSell[:pos], getUser.ItemsSell[pos+1:]...)
				break
			}
		}

		// also update global items list
		for pos, item := range items {
			if item.ItemID == itemID {
				// remove item found
				items = append(items[:pos], items[pos+1:]...)
				break
			}
		}
	}

	// update this user map
	mapUsers[getUser.Username] = user{
		Username:  getUser.Username,
		Password:  getUser.Password,
		FirstName: getUser.FirstName,
		LastName:  getUser.LastName,
		UserID:    getUser.UserID,

		ItemsBid:     sortItemsByDateASC(getUser.ItemsBid),
		ItemsSell:    sortItemsByDateASC(getUser.ItemsSell),
		ItemsAuction: sortItemsByDateASC(getUser.ItemsAuction),
	}

	// first load
	data := templateData{
		User: mapUsers[getUser.Username],
	}
	renderTemplate(res, "profile.gohtml", data)
	common.Debug("profile")
}
