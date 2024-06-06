package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type authGenerated struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
	ID      int    `json:"id"`
}

type Config map[string]string

type hostsList struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  []struct {
		Hostid     string `json:"hostid"`
		Host       string `json:"host"`
		Interfaces []struct {
			Interfaceid string `json:"interfaceid"`
			IP          string `json:"ip"`
		} `json:"interfaces"`
	} `json:"result"`
	ID int `json:"id"`
}

type getMacrosResult struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  []struct {
		Hostmacroid string `json:"hostmacroid"`
		Hostid      string `json:"hostid"`
		Macro       string `json:"macro"`
		Value       string `json:"value"`
		Description string `json:"description"`
		Type        string `json:"type"`
	} `json:"result"`
	ID int `json:"id"`
}

type macrosExistsError struct {
	Jsonrpc string `json:"jsonrpc"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	} `json:"error"`
	ID int `json:"id"`
}

var username string
var password string
var hostName string
var zbxServerLink string
var enableLogging bool

func readConfig(filename string) (Config, error) {
	// init with some bogus data
	config := Config{}
	if len(filename) == 0 {
		return config, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if equal := strings.Index(line, "="); equal >= 0 {
			key := strings.TrimSpace(line[:equal])
			value := strings.TrimSpace(line[equal+1:])
			config[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return config, nil
}

func adminAuth(username, userpassword, zabbixServerLink string) (auth string) {
	var result authGenerated

	//Encode the data
	postBody, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "user.login",
		"params": map[string]string{
			"user":     username,
			"password": userpassword,
		},
		"id":   1,
		"auth": nil,
	})
	// This is identical to this query
	// curl --header "Content-Type: application/json" \
	// 	--request POST \
	// 	--data '{"jsonrpc": "2.0", "method": "user.login", "params": {"user": "username", "password": "userpassword"}, "id": 1, "auth": null}' \
	// 	zabbixServerLink

	body, err := makeRequest(postBody, zabbixServerLink)
	if err != nil {
		log.Fatalln(err)
	}

	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can't unmarshal JSON")
	}
	return result.Result
}

func getHosts(auth, hostname, zabbixServerLink string) (hostid string) {
	var result hostsList
	var hostID string
	//Encode the data
	postBody, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "host.get",
		"params": map[string][]string{
			"output":           {"hostid", "host"},
			"selectInterfaces": {"interfaceid", "ip"},
		},
		"id":   2,
		"auth": auth,
	})
	// This is identical to this query
	// curl --header "Content-Type: application/json" \
	// 	--request POST \
	// 	--data '{"jsonrpc": "2.0", "method": "host.get", "params": {"output": ["hostid", "host"], "selectInterfaces": ["interfaceid", "ip"]}, "id": 2, "auth": auth}' \
	// 	zabbixServerLink

	body, err := makeRequest(postBody, zabbixServerLink)
	if err != nil {
		log.Fatalln(err)
	}
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can't unmarshal JSON")
	}

	for _, v := range result.Result {
		if v.Host == hostname {
			hostID = v.Hostid
		}
	}
	return hostID
}

func createMacros(hostid, auth, key, value, zabbixServerLink string) {
	var result macrosExistsError
	//Encode the data
	postBody, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "usermacro.create",
		"params": map[string]string{
			"hostid": hostid,
			"macro":  key,
			"value":  value,
		},
		"id":   1,
		"auth": auth,
	})
	body, err := makeRequest(postBody, zabbixServerLink)
	if err != nil {
		log.Fatalln(err)
	}
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can't unmarshal JSON")
	}
	if result.Error.Code == -32602 {
		hostMacrosId := getMacrosId(auth, hostid, zabbixServerLink, key)
		macrosUpdate(auth, value, zabbixServerLink, hostMacrosId)
	}
}

func getMacrosId(auth, hostid, zabbixServerLink, macroName string) (hostMacroId string) {
	var result getMacrosResult
	postBody, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "usermacro.get",
		"params": map[string]string{
			"hostids": hostid,
			"output":  "extend",
		},
		"auth": auth,
		"id":   1,
	})
	body, err := makeRequest(postBody, zabbixServerLink)
	if err != nil {
		log.Fatalln(err)
	}
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can't unmarshal JSON")
	}
	for _, item := range result.Result {
		if macroName == item.Macro {
			return item.Hostmacroid
		}

	}
	return
}

func macrosUpdate(auth, value, zabbixServerLink, hostMacroId string) {
	postBody, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "usermacro.update",
		"params": map[string]string{
			"hostmacroid": hostMacroId,
			"value":       value,
		},
		"auth": auth,
		"id":   1,
	})
	makeRequest(postBody, zabbixServerLink)
}

func makeRequest(postBody []byte, zabbixServerLink string) (responseBody []byte, err error) {
	//Leverage Go's HTTP Post function to make request
	resp, err := http.Post(zabbixServerLink, "application/json", bytes.NewBuffer(postBody))
	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("An Error Occured: %v", err)
		return nil, err
	}

	log.Println(string(responseBody))
	return responseBody, nil
}

var version string

func main() {
	flag.StringVar(&username, "username", "", "Administrator username")
	flag.StringVar(&password, "password", "", "Administrator password")
	flag.StringVar(&hostName, "hostname", "", "Host name")
	flag.StringVar(&zbxServerLink, "zbxServerLink", "", "Link to zabbix server.")
	flag.BoolVar(&enableLogging, "enableLogging", false, "Set true to enable logs.")
	versionFlag := flag.Bool("version", false, "Print the version of the application")

	flag.Parse()
	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}
	if !enableLogging {
		log.SetFlags(0)
		log.SetOutput(io.Discard)
	}
	config, err := readConfig(`config.txt`)
	if err != nil {
		fmt.Println(err)
	}
	adminLogin := adminAuth(username, password, zbxServerLink)
	gettingAllHosts := getHosts(adminLogin, hostName, zbxServerLink)
	for macrosName, macrosValue := range config {
		createMacros(gettingAllHosts, adminLogin, macrosName, macrosValue, zbxServerLink)
	}

}
