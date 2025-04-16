package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Co chcesz zrobić?")
	fmt.Println("1 - Załaduj dane z Excela do MongoDB")
	fmt.Println("2 - Uruchom serwer API")
	fmt.Print("Wybierz [1/2]: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "1":
		loadDataToMongo()
	case "2":
		startAPIServer()
	default:
		fmt.Println("Niepoprawny wybór, spróbuj ponownie.")
	}
}
