package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func getItemRecord(db *sql.DB, id string) (itemInfo, error) {

	itemINFO := itemInfo{}
	var itemid string
	row := db.QueryRow("select * from salesitems where itemID = ?", id)
	err := row.Scan(&itemid, &itemINFO.UserID, &itemINFO.ItemName, &itemINFO.ItemDesc, &itemINFO.BidClosebyOwner,
		&itemINFO.BidCloseDate, &itemINFO.BidIncrement, &itemINFO.BidPrice, &itemINFO.DisplayItem, &itemINFO.ItemImage)
	if err != nil {
		if err == sql.ErrNoRows {
			// there were no rows, but otherwise no error occurred
			return itemINFO, errors.New("item id not found")
		}
		log.Fatal(err)
	}
	Debug("item info retrieved", itemINFO)
	return itemINFO, nil
}

func insertItemRecord(db *sql.DB, id string, newItem itemInfo) {
	query := fmt.Sprintf("INSERT INTO salesitems VALUES ('%s', '%d', '%s', '%s', '%d', '%s', '%s', '%s', '%d', '%s')", id,
		newItem.UserID, newItem.ItemName, newItem.ItemDesc, newItem.BidClosebyOwner, newItem.BidCloseDate,
		newItem.BidIncrement, newItem.BidPrice, newItem.DisplayItem, newItem.ItemImage)

	_, err := db.Query(query)

	if err != nil {
		panic(err.Error())
	}
}

func editItemRecord(db *sql.DB, id string, newItem itemInfo) {
	query := fmt.Sprintf(
		"UPDATE salesitems SET userID='%d', itemName='%s', itemDesc='%s', bidClosebyOwner='%d',"+
			"bidCloseDate='%s', bidIncrement='%s', bidPrice='%s', displayItem='%d', itemImage='%s'  WHERE itemID='%s'",
		newItem.UserID, newItem.ItemName, newItem.ItemDesc, newItem.BidClosebyOwner, newItem.BidCloseDate,
		newItem.BidIncrement, newItem.BidPrice, newItem.DisplayItem, newItem.ItemImage, id)
	_, err := db.Query(query)
	if err != nil {
		panic(err.Error())
	}
}

func deleteItemRecord(db *sql.DB, id string) {
	query := fmt.Sprintf(
		"DELETE FROM salesitems WHERE itemID='%s'", id)
	_, err := db.Query(query)
	if err != nil {
		panic(err.Error())
	}
}

//Handler for item data management - add, delete, edit, retrieve info
func item(w http.ResponseWriter, r *http.Request) {
	if !validKey(r) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("401 - Invalid key"))
		return
	}

	params := mux.Vars(r)
	//itemID, _ := strconv.Atoi(params["itemid"])
	itemID := params["itemid"]

	//userID, _ := strconv.Atoi(params["itemid"])

	if r.Method == "GET" {
		var itemINFO itemInfo
		itemINFO, err := getItemRecord(db, itemID)

		if err == nil {
			json.NewEncoder(w).Encode(
				itemINFO)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - No item found"))
		}
	}

	if r.Method == "DELETE" {

		_, err := getItemRecord(db, itemID)

		if err == nil {
			deleteItemRecord(db, itemID)
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("202 - item deleted: " +
				params["itemid"]))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - item id not found"))
		}
	}

	if r.Header.Get("Content-type") == "application/json" {

		// POST is for creating new item
		if r.Method == "POST" {

			// read the string sent to the service
			var newItem itemInfo
			reqBody, err := ioutil.ReadAll(r.Body)

			if err == nil {
				// convert JSON to object
				json.Unmarshal(reqBody, &newItem)

				if newItem.ItemName == "" {
					w.WriteHeader(
						http.StatusUnprocessableEntity)
					w.Write([]byte(
						"422 - Please supply item " +
							"information " + "in JSON format"))
					return
				}

				// check if item exists; add only if
				// item does not exist
				_, err := getItemRecord(db, itemID)
				if err != nil {
					insertItemRecord(db, itemID, newItem)
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("201 - new item added: " +
						params["itemid"]))
				} else {
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte(
						"409 - Duplicate item ID"))
				}
			} else {
				w.WriteHeader(
					http.StatusUnprocessableEntity)
				w.Write([]byte("422 - Please supply item information " +
					"in JSON format"))
			}
		}

		//---PUT is for creating or updating
		// existing user---
		if r.Method == "PUT" {
			var newItem itemInfo
			reqBody, err := ioutil.ReadAll(r.Body)

			if err == nil {
				json.Unmarshal(reqBody, &newItem)

				if newItem.ItemName == "" {
					w.WriteHeader(
						http.StatusUnprocessableEntity)
					w.Write([]byte(
						"422 - Please supply item " +
							" information " +
							"in JSON format"))
					return
				}

				// check if user exists; add only if
				// user does not exist
				_, err := getItemRecord(db, itemID)
				if err != nil {
					//add user
					insertItemRecord(db, itemID, newItem)
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("201 - Item added: " +
						params["itemid"]))
				} else {
					// update user
					editItemRecord(db, itemID, newItem)
					w.WriteHeader(http.StatusAccepted)
					w.Write([]byte("202 - item updated: " +
						params["userid"]))
				}
			} else {
				w.WriteHeader(
					http.StatusUnprocessableEntity)
				w.Write([]byte("422 - Please supply " +
					"item information " +
					"in JSON format"))
			}
		}

	}

}

func getItemRecords(db *sql.DB) []itemInfoFull {
	results, err := db.Query("Select * FROM golive_db.salesitems")

	if err != nil {
		panic(err.Error())
	}

	items := []itemInfoFull{}
	//var itemid []string

	//i := 0
	for results.Next() {
		// map this type to the record in the table

		//var courseID string
		//var courseINFO courseInfo
		var itemTemp itemInfoFull
		err = results.Scan(&itemTemp.ItemID, &itemTemp.UserID, &itemTemp.ItemName, &itemTemp.ItemDesc, &itemTemp.BidClosebyOwner,
			&itemTemp.BidCloseDate, &itemTemp.BidIncrement, &itemTemp.BidPrice, &itemTemp.DisplayItem, &itemTemp.ItemImage)

		if err != nil {
			panic(err.Error())
		}

		items = append(items, itemTemp)
		//i++
		//courses[courseID] = courseINFO
		//Debug("db:", courses)
		//Debug(person.ID, person.FirstName, person.LastName, person.Age)
	}
	return items
}

//Handlers for requesting all item data
func allitems(w http.ResponseWriter, r *http.Request) {
	if !validKey(r) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("401 - Invalid key"))
		return
	}

	items := getItemRecords(db)

	Debug("items:", items)
	//returns all the items in JSON
	json.NewEncoder(w).Encode(items)
}
