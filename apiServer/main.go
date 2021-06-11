package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type userInfo struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
	Email    string `json:"email"`   //map to first name instead
	Address  string `json:"address"` //map to last name instead
	Status   int    `json:"status"`
}

//This struct includes additional user id info
type userInfoFull struct {
	UserID string `json:"userID"`
	userInfo
}

//This struct intended for login valification purpose
type loginInfo struct {
	Password string `json:"password"`
}

type itemInfo struct {
	UserID          int    `json:"userID"`
	ItemName        string `json:"itemName"`
	ItemDesc        string `json:"itemDesc"`
	BidClosebyOwner int    `json:"bidClosebyOwner"`
	BidCloseDate    string `json:"bidCloseDate"`
	BidIncrement    string `json:"bidIncrement"`
	BidPrice        string `json:"bidPrice"`
	DisplayItem     int    `json:"displayItem"`
	ItemImage       string `json:"itemImage"`
}

//This struct includes addtional item ID field
type itemInfoFull struct {
	ItemID string `json:"itemID"`
	itemInfo
}

var db *sql.DB

// use godot package to load/read the .env file and
// return the value of the key
func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

//This is to validate access key when accessing REST API
func validKey(r *http.Request) bool {
	v := r.URL.Query()
	if key, ok := v["key"]; ok {
		// godotenv package
		chkKey := goDotEnvVariable("STRONGEST_AVENGER")

		if key[0] == chkKey {
			return true
		}
		return false

	}
	return false
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the REST API!")
}

func main() {

	// Use mysql as driverName and a valid DSN as dataSourceName:
	var err error
	db, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/golive_db")

	// handle error
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("Database opened")
	}

	// defer the close till after the main function has finished executing
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/", home)
	router.HandleFunc("/api/v1/login/{userid}", login).Methods("GET")
	router.HandleFunc("/api/v1/user/{userid}", user).Methods("GET", "PUT", "POST", "DELETE")
	router.HandleFunc("/api/v1/item/{itemid}", item).Methods("GET", "PUT", "POST", "DELETE")
	router.HandleFunc("/api/v1/items", allitems)
	router.HandleFunc("/api/v1/users", allusers)

	fmt.Println("Listening at port 8081")

	addr := ":8081"
	if constDebug {
		addr = "localhost" + addr
	}

	err = http.ListenAndServeTLS(addr, "cert.pem", "key.pem", router)
	if err != nil {
		log.Fatal("ListenAndServe : ", err)
	}

}
