package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func main() {

	// read file
	// https://developer.mozilla.org/ru/docs/Learn/JavaScript/Objects/JSON
	data, err := ioutil.ReadFile("./file.json")
	if err != nil {
		fmt.Print(err)
	}

	// define data structure
	//Использование тегов в структуре кодируемой в JSON
	//позволяет получить названия полей в результирующем JSON,
	//отличающиеся от названия полей в структуре.

	type Member struct {
		Name           string   `json:"name"`
		Age            int      `json:"age"`
		SecretIdentity string   `json:"secretIdentity"`
		Powers         []string `json:"powers"`
	}

	type Information struct {
		SquadName  string   `json:"squadName"`
		HomeTown   string   `json:"homeTown"`
		Formed     int      `json:"formed"`
		SecretBase string   `json:"secretBase"`
		Active     bool     `json:"active"`
		Members    []Member `json:"members"`
	}

	// json data
	var obj Information

	// unmarshall it
	err = json.Unmarshal(data, &obj)
	if err != nil {
		fmt.Println("error:", err)
	}

	// can access using struct now
	fmt.Printf("%v\n", obj.Members[0])
	fmt.Printf("%v\n", obj.Members[1])
	fmt.Printf("%v\n", obj.Members[2])

}
