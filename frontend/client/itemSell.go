package client

import (
	"io"
	"net/http"
	"os"

	"goLive/frontend/common"
	"goLive/frontend/logs"

	"goLive/frontend/validation"
)

var uploadFilePath = "assets/fileuploads/"

func itemSell(res http.ResponseWriter, req *http.Request) {
	// not logged in
	if !isLoggedIn(req) {
		data := templateData{
			Items: sortItemsByDateASC(items),
			User:  getUser(res, req),
		}
		renderTemplate(res, "index.gohtml", data)
		return
	}

	// POST
	if req.Method == http.MethodPost {
		// already logged in
		getUser := getUser(res, req)

		// form elements value
		rawItemName := req.FormValue("itemname")
		rawItemDesc := req.FormValue("itemdesc")
		rawItemStartBid := req.FormValue("itemstartbid")
		rawItemStepBid := req.FormValue("itemstepbid")
		rawItemClosingDate := req.FormValue("itemclosingdate")

		itemName, err := validation.ValidateString(rawItemName, 50)
		if err != nil {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("itemSell name", err)
			return
		}
		itemDesc, err := validation.ValidateString(rawItemDesc, 100)
		if err != nil {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("itemSell desc", err)
			return
		}
		basePrice, err := validation.ValidateString(rawItemStartBid, 50)
		if err != nil {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("itemSell bidprice", err)
			return
		}
		bidIncrement, err := validation.ValidateString(rawItemStepBid, 50)
		if err != nil {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("itemSell bidstep", err)
			return
		}
		itemClosingDate, err := validation.ValidateDate(rawItemClosingDate)
		if err != nil {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("itemSell date", err)
			return
		}

		itemImage, err := uploadFile(req, "itemuploadimage")
		if err != nil {
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			logs.Error("itemSell image", err)
			return
		}

		// catch user by refreshing the page
		// if the page was refreshed,
		// by right, should have the same item name
		// so, search through user selling of items
		found := false
		// check for same item name, because
		// no item id is generated at this moment
		for _, item := range getUser.ItemsSell {
			if item.ItemName == itemName {
				found = true
				break
			}
		}
		// add new item if not found
		if !found {
			// update field values
			p := &item{
				ItemID:    common.GenerateUUID(), // @todo: uuid length is 36, need to change in db
				ItemName:  itemName,
				ItemDesc:  itemDesc,
				ItemImage: uploadFilePath + itemImage,

				ItemBidPrice:     basePrice,
				ItemBidPriceStep: bidIncrement,
				ItemClosingDate:  itemClosingDate,

				ItemBidStatus: 0, // 0:false | 1:true
			}
			item := pointer{
				pointer: &p,
			}

			addItem2db(p.ItemID, *p)

			// add item to slice by updating user map of items
			getUser.ItemsSell = append(getUser.ItemsSell, *item.pointer)

			// also update global items
			items = append(items, *item.pointer)
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

	// welcome user
	getUser := getUser(res, req)
	data := templateData{
		Items: sortItemsByDateASC(items),
		User:  getUser,
	}
	renderTemplate(res, "itemSell.gohtml", data)
	common.Debug("itemSell")
}

func uploadFile(req *http.Request, htmlInputFile string) (string, error) {
	req.ParseMultipartForm(32 << 20)

	// open file stream
	file, fileInfo, err := req.FormFile(htmlInputFile)
	if err != nil {
		return "", err
	}
	defer file.Close() // close file stream

	// save file to path
	f, err := os.OpenFile(uploadFilePath+fileInfo.Filename,
		// O_CREATE create a new file if none exists
		// O_WRONLY open the file write-only
		os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()
	io.Copy(f, file)

	return fileInfo.Filename, nil
}
