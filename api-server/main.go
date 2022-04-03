package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var port int

func init() {
	flag.IntVar(&port, "port", 9090, "Server port")
}

func main() {

	result, err := loadConfigurations()

	flag.Parse()

	cp := NewControlPlane(NewGatewayService([]Gateway{}), NewConfigurationServiceImpl(result))
	s := SetupHttpServer(port, cp)
	err = s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Control plane server started and listening to port %d", port)

}

func loadConfigurations() (map[string][]Configuration, error) {
	// Open our jsonFile
	jsonFile, err := os.Open("configurations.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened configurations.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var configurations map[string][]Configuration

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &configurations)

	for k, v := range configurations {

		fmt.Println(k)
		fmt.Println(v)
	}
	/*fmt.Println("User Type: " + configurations.Users[i].Type)
	fmt.Println("User Age: " + strconv.Itoa(users.Users[i].Age))
	fmt.Println("User Name: " + users.Users[i].Name)
	fmt.Println("Facebook Url: " + users.Users[i].Social.Facebook)
	*/
	return configurations, err
}
