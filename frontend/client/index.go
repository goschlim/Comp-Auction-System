package client

import (
	"net/http"
	"strings"

	"goLive/frontend/common"
	"goLive/frontend/logs"

	"goLive/frontend/validation"
)

// index default landing page, or root page.
func index(res http.ResponseWriter, req *http.Request) {
	getUser := getUser(res, req)

	// POST
	if req.Method == http.MethodPost {
		// form element value
		rawSearch := req.FormValue("search")
		searchStr, err := validation.CheckString(rawSearch)
		//logs.Info(searchStr)
		if err != nil {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("index error", err)
			return

		} else if len(searchStr) > 0 {

			// perform stringSearch based on keywords
			patterns := strings.Split(strings.ToLower(searchStr), " ")

			var searchItems []*item
			for _, item := range items {
				// convert all strings to lower case for easy comparisons
				str := strings.Join([]string{
					strings.ToLower(item.ItemName),
					strings.ToLower(item.ItemDesc),
					strings.ToLower(item.ItemBidPrice),
					strings.ToLower(item.ItemClosingDate),
				}, " ")

				// []string Slice
				matches := search(str, patterns)
				// has matches
				if len(matches) > 0 {
					searchItems = append(searchItems, item)
				}
			}
			//common.Debug(searchItems)

			// show found results
			data := templateData{
				Items:   sortItemsByDateASC(searchItems),
				User:    getUser,
				Results: len(searchItems),

				QueryString: searchStr,
			}
			renderTemplate(res, "index.gohtml", data)
			return
		}

		// not logged in
		if !isLoggedIn(req) {
			data := templateData{
				Items:   sortItemsByDateASC(items),
				User:    getUser,
				Results: len(items),

				QueryString: "",
			}
			renderTemplate(res, "index.gohtml", data)
			return
		}

		// already logged in

		// form element value
		rawItemID := req.FormValue("itemid")
		itemID, err := validation.CheckUUIDString(rawItemID)
		//logs.Info(itemID)
		if err != nil {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("index error", err)
			return

		} else if len(itemID) == 0 {
			data := templateData{
				Items:   sortItemsByDateASC(items),
				User:    getUser,
				Results: len(items),

				QueryString: "",
			}
			renderTemplate(res, "index.gohtml", data)
			return
		}

		var bidItem *item
		// search through items
		for _, item := range items {
			if item.ItemID == itemID {
				bidItem = item // item found
				break
			}
		}

		data := templateData{
			Item: bidItem,
			User: getUser,

			ErrorMessage: "",
		}

		// prevent same user from bidding own selling of item
		for _, itemSell := range getUser.ItemsSell {
			// item found
			if itemSell.ItemID == bidItem.ItemID {
				msg := "Cannot bid self-listed item"
				common.Debug(msg)
				// update data
				data = templateData{
					Item: bidItem,
					User: getUser,

					ErrorMessage: msg,
				}
				break
			}
		}
		// show this page
		renderTemplate(res, "itemBid.gohtml", data)
		common.Debug("index > itemBid", bidItem)
		return
	}

	// welcome user
	data := templateData{
		Items:   sortItemsByDateASC(items),
		User:    getUser,
		Results: len(items),

		QueryString: "",
	}
	renderTemplate(res, "index.gohtml", data)
	common.Debug("index")
}
