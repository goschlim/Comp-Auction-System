package client

import (
	"net/http"

	"goLive/frontend/common"
	"goLive/frontend/logs"

	"goLive/frontend/validation"
)

func itemEdit(res http.ResponseWriter, req *http.Request) {
	// not logged in
	if !isLoggedIn(req) {
		data := templateData{
			Items: sortItemsByDateASC(items),
			User:  getUser(res, req),
		}
		renderTemplate(res, "index.gohtml", data)
		return
	}

	// already logged in
	getUser := getUser(res, req)

	data := templateData{}

	// POST
	// [user click "Edit Bid Placement" to edit]
	if req.Method == http.MethodPost {
		// form elements value

		// redirected action from "defineItemSell.gohtml" #editBidForm
		rawItemID := req.FormValue("itemeditid")

		// redirected action from "itemEdit.gohtml" #editForm
		rawItemName := req.FormValue("itemname")

		// itemeditid POST value
		if len(rawItemID) > 0 && len(rawItemName) == 0 {
			itemID, err := validation.CheckUUIDString(rawItemID)
			if err != nil || len(itemID) == 0 {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("itemEdit id", err)
				return
			}

			var sellItem *item
			// search through user selling items
			for _, item := range getUser.ItemsSell {
				if item.ItemID == itemID {
					sellItem = item
					break
				}
			}

			data = templateData{
				Item: sellItem,
				User: getUser,
			}

			renderTemplate(res, "itemEdit.gohtml", data)
			common.Debug("itemEdit", sellItem)
			return

		} else if len(rawItemName) > 0 && len(rawItemID) == 0 {
			// itemname POST value

			// form elements value
			rawItemID := req.FormValue("itemeditidold")
			rawItemName := req.FormValue("itemname")
			rawItemDesc := req.FormValue("itemdesc")
			rawItemStartBid := req.FormValue("itemstartbid")
			rawItemStepBid := req.FormValue("itemstepbid")
			rawItemClosingDate := req.FormValue("itemclosingdate")

			itemID, err := validation.CheckUUIDString(rawItemID)
			if err != nil || len(itemID) == 0 {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("itemEdit id", err)
				return
			}
			itemName, err := validation.ValidateString(rawItemName, 50)
			if err != nil {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("itemEdit name", err)
				return
			}
			itemDesc, err := validation.ValidateString(rawItemDesc, 100)
			if err != nil {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("itemEdit desc", err)
				return
			}
			basePrice, err := validation.ValidateString(rawItemStartBid, 50)
			if err != nil {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("itemEdit bidprice", err)
				return
			}
			bidIncrement, err := validation.ValidateString(rawItemStepBid, 50)
			if err != nil {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("itemEdit bidstep", err)
				return
			}
			itemClosingDate, err := validation.ValidateDate(rawItemClosingDate)
			if err != nil {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("itemEdit date", err)
				return
			}

			itemImage, err := uploadFile(req, "itemuploadimage")
			if err != nil {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				logs.Error("itemEdit image", err)
				return
			}

			// search through user selling items, and update
			// found item accordingly
			for _, item := range getUser.ItemsSell {
				if item.ItemID == itemID {
					item.ItemName = itemName
					item.ItemDesc = itemDesc
					item.ItemImage = uploadFilePath + itemImage
					item.ItemBidPrice = basePrice
					item.ItemBidPriceStep = bidIncrement
					item.ItemClosingDate = itemClosingDate
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

			data := templateData{
				User: getUser,
			}
			renderTemplate(res, "profile.gohtml", data)
			return
		}
	}

	data = templateData{
		Items: sortItemsByDateASC(items),
		User:  getUser,
	}
	renderTemplate(res, "index.gohtml", data)
	common.Debug("itemEdit > index")
}
