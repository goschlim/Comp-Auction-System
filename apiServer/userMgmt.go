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

func getLoginRecord(db *sql.DB, id string) (string, error) {
	var password string
	err := db.QueryRow("select password from users where userID = ?", id).Scan(&password)
	if err != nil {
		if err == sql.ErrNoRows {
			// there were no rows, but otherwise no error occurred
			return "", errors.New("user id not found")
		}
		log.Fatal(err)
	}
	Debug("password retrieved", password)
	return password, nil
}

//Handler serving data request for login validation
func login(w http.ResponseWriter, r *http.Request) {
	if !validKey(r) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("401 - Invalid key"))
		return
	}

	params := mux.Vars(r)

	userID := params["userid"]

	if r.Method == "GET" {
		var loginINFO loginInfo
		password, err := getLoginRecord(db, userID)
		if err == nil {
			loginINFO.Password = password
			json.NewEncoder(w).Encode(
				loginINFO)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - user id not found"))
		}
	}

}

func getUserRecord(db *sql.DB, id string) (userInfo, error) {
	userINFO := userInfo{}
	var userid string
	row := db.QueryRow("select * from users where userID = ?", id)
	err := row.Scan(&userid, &userINFO.UserName, &userINFO.Email, &userINFO.Address, &userINFO.Password, &userINFO.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			// there were no rows, but otherwise no error occurred
			return userINFO, errors.New("user id not found")
		}
		log.Fatal(err)
	}
	Debug("user info retrieved", userINFO)
	return userINFO, nil
}

func deleteUserRecord(db *sql.DB, id string) {
	query := fmt.Sprintf(
		"DELETE FROM users WHERE userID='%s'", id)
	_, err := db.Query(query)
	if err != nil {
		panic(err.Error())
	}
}

func insertUserRecord(db *sql.DB, id string, newUser userInfo) {
	query := fmt.Sprintf("INSERT INTO users VALUES ('%s', '%s', '%s', '%s', '%s', '%d')", id,
		newUser.UserName, newUser.Email, newUser.Address, newUser.Password, newUser.Status)

	_, err := db.Query(query)

	if err != nil {
		panic(err.Error())
	}
}

func editUserRecord(db *sql.DB, id string, newUser userInfo) {
	query := fmt.Sprintf(
		"UPDATE users SET userName='%s', userEmail='%s', address='%s', password='%s', status='%d'  WHERE userID='%s'",
		newUser.UserName, newUser.Email, newUser.Address, newUser.Password, newUser.Status, id)
	_, err := db.Query(query)
	if err != nil {
		panic(err.Error())
	}
}

//Handler for user data management - add, delete, edit, retrieve info
func user(w http.ResponseWriter, r *http.Request) {
	if !validKey(r) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("401 - Invalid key"))
		return
	}

	params := mux.Vars(r)

	userID := params["userid"]

	if r.Method == "GET" {
		var userINFO userInfo
		userINFO, err := getUserRecord(db, userID)

		if err == nil {
			json.NewEncoder(w).Encode(
				userINFO)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - No course found"))
		}
	}

	if r.Method == "DELETE" {
		_, err := getUserRecord(db, userID)

		if err == nil {
			deleteUserRecord(db, userID)
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("202 - user deleted: " +
				params["userid"]))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 - user id not found"))
		}
	}

	if r.Header.Get("Content-type") == "application/json" {

		// POST is for creating new user
		if r.Method == "POST" {

			// read the string sent to the service
			var newUser userInfo
			reqBody, err := ioutil.ReadAll(r.Body)

			if err == nil {
				// convert JSON to object
				json.Unmarshal(reqBody, &newUser)

				if newUser.UserName == "" {
					w.WriteHeader(
						http.StatusUnprocessableEntity)
					w.Write([]byte(
						"422 - Please supply user " +
							"information " + "in JSON format"))
					return
				}

				// check if user exists; add only if
				// user does not exist
				_, err := getUserRecord(db, userID)
				if err != nil {
					insertUserRecord(db, userID, newUser)
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("201 - new user added: " +
						params["userid"]))
				} else {
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte(
						"409 - Duplicate user ID"))
				}
			} else {
				w.WriteHeader(
					http.StatusUnprocessableEntity)
				w.Write([]byte("422 - Please supply user information " +
					"in JSON format"))
			}
		}

		//---PUT is for creating or updating
		// existing user---
		if r.Method == "PUT" {
			var newUser userInfo
			reqBody, err := ioutil.ReadAll(r.Body)

			if err == nil {
				json.Unmarshal(reqBody, &newUser)

				if newUser.UserName == "" {
					w.WriteHeader(
						http.StatusUnprocessableEntity)
					w.Write([]byte(
						"422 - Please supply user " +
							" information " +
							"in JSON format"))
					return
				}

				// check if user exists; add only if
				// user does not exist
				_, err := getUserRecord(db, userID)
				if err != nil {
					//add user
					insertUserRecord(db, userID, newUser)
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("201 - Course added: " +
						params["courseid"]))
				} else {
					// update user
					editUserRecord(db, userID, newUser)
					w.WriteHeader(http.StatusAccepted)
					w.Write([]byte("202 - user updated: " +
						params["userid"]))
				}
			} else {
				w.WriteHeader(
					http.StatusUnprocessableEntity)
				w.Write([]byte("422 - Please supply " +
					"user information " +
					"in JSON format"))
			}
		}

	}

}

//Retrieves all data from the users table from mysql
func getUserRecords(db *sql.DB) []userInfoFull {
	results, err := db.Query("Select * FROM golive_db.users")

	if err != nil {
		panic(err.Error())
	}

	users := []userInfoFull{}

	for results.Next() {
		// map this type to the record in the table

		var userTemp userInfoFull
		err = results.Scan(&userTemp.UserID, &userTemp.UserName, &userTemp.Email,
			&userTemp.Address, &userTemp.Password, &userTemp.Status)

		if err != nil {
			panic(err.Error())
		}

		users = append(users, userTemp)
	}
	return users
}

//Handler to response to API request for all users data
func allusers(w http.ResponseWriter, r *http.Request) {
	if !validKey(r) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("401 - Invalid key"))
		return
	}

	//getting data from mysql db
	users := getUserRecords(db)

	Debug("users:", users)
	//returns all the items in JSON
	json.NewEncoder(w).Encode(users)
}
