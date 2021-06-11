package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"goLive/frontend/common"
	"io/ioutil"
	"log"
	"net/http"
)

const baseURL = "https://localhost:8081/api/v1/user"
const baseURL2 = "https://localhost:8081/api/v1/item"
const baseURL3 = "https://localhost:8081/api/v1/items"
const baseURL4 = "https://localhost:8081/api/v1/users"

//var accessKey string = "2c78afaf-97da-4816-bbee-9ad239abb296"

var accessKey string = common.GetEnv("STRONGEST_AVENGER")

// clientConfig sets http.Client configuration
// for HTTPS connection of self-signed CA cert.
func clientConfig() *http.Client {

	// using self-signed CA cert
	certCA, certErr := ioutil.ReadFile("ssl/cert.pem")
	if certErr != nil {
		log.Fatal(certErr)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certCA)

	// client side configuration
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            certPool,
				InsecureSkipVerify: true,
			},
		},
	}

	return client
}

//add new item data to db via API
func addItem2db(code string, jsonData item) {
	jsonValue, _ := json.Marshal(jsonData)

	request, err := http.NewRequest(http.MethodPost,
		baseURL2+"/"+code+"?key="+accessKey,
		bytes.NewBuffer(jsonValue))

	request.Header.Set("Content-Type", "application/json")

	//call REST API
	client := clientConfig()
	response, err := client.Do(request)

	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		common.Debug(response.StatusCode)
		common.Debug(string(data))
		response.Body.Close()
	}
}

//get all the item data from the db via REST API
func getItemFromDB() []item {
	request, err := http.NewRequest(http.MethodGet,
		baseURL3+"?key="+accessKey, nil)

	client := clientConfig()
	response, err := client.Do(request)

	var items []item

	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		if err == nil {
			// convert JSON to object
			json.Unmarshal(data, &items)
		}

		common.Debug(response.StatusCode)
		common.Debug(string(data))
		response.Body.Close()
	}
	return items
}

//add new user data to db via API
func addUser2db(code string, jsonData user) {
	jsonValue, _ := json.Marshal(jsonData)

	request, err := http.NewRequest(http.MethodPost,
		baseURL+"/"+code+"?key="+accessKey,
		bytes.NewBuffer(jsonValue))

	request.Header.Set("Content-Type", "application/json")

	client := clientConfig()
	response, err := client.Do(request)

	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		common.Debug(response.StatusCode)
		common.Debug(string(data))
		response.Body.Close()
	}
}

//get all the user data from the db via REST API
func getAllUsersFromDB() []user {
	request, err := http.NewRequest(http.MethodGet,
		baseURL4+"?key="+accessKey, nil)

	client := clientConfig()
	response, err := client.Do(request)

	var users []user

	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		if err == nil {
			// convert JSON to object
			json.Unmarshal(data, &users)
		}

		common.Debug(users)

		common.Debug(response.StatusCode)
		common.Debug(string(data))
		response.Body.Close()
	}
	return users
}
