package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Co chcesz zrobić?")
	fmt.Println("1 - Załaduj dane z Excela do MongoDB")
	fmt.Println("2 - Serwer API i operacje SWIFT")
	fmt.Print("Wybierz [1/2]: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "1":
		loadDataToMongo()
	case "2":
		handleServerOptions()
	default:
		fmt.Println("Niepoprawny wybór, spróbuj ponownie.")
	}
}

func handleServerOptions() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n📡 Serwer API i SWIFT operacje:")
	fmt.Println("1 - Uruchom serwer API")
	fmt.Println("2 - Dodaj dane SWIFT (POST)")
	fmt.Println("3 - Usuń dane SWIFT (DELETE)")
	fmt.Print("Wybierz [1/2/3]: ")

	option, _ := reader.ReadString('\n')
	option = strings.TrimSpace(option)

	switch option {
	case "1":
		startAPIServer()
	case "2":
		sendSwiftCode()
	case "3":
		deleteSwiftCode()
	default:
		fmt.Println("Niepoprawny wybór, wracam do głównego menu.")
	}
}

func sendSwiftCode() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("📝 Podaj dane do wysłania:")

	fmt.Print("Adres: ")
	address, _ := reader.ReadString('\n')

	fmt.Print("Nazwa banku: ")
	bankname, _ := reader.ReadString('\n')

	fmt.Print("Kod kraju (ISO2): ")
	countryiso2, _ := reader.ReadString('\n')

	fmt.Print("Nazwa kraju: ")
	countryname, _ := reader.ReadString('\n')

	fmt.Print("Czy to centrala? (true/false): ")
	isHQstr, _ := reader.ReadString('\n')
	isHQstr = strings.TrimSpace(isHQstr)
	isHQ := strings.ToLower(isHQstr) == "true"

	fmt.Print("SWIFT Code: ")
	swiftcode, _ := reader.ReadString('\n')

	data := map[string]interface{}{
		"address":       strings.TrimSpace(address),
		"bankname":      strings.TrimSpace(bankname),
		"countryiso2":   strings.TrimSpace(countryiso2),
		"countryname":   strings.TrimSpace(countryname),
		"isheadquarter": isHQ,
		"swiftcode":     strings.TrimSpace(swiftcode),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("❌ Błąd podczas konwersji danych:", err)
		return
	}

	resp, err := http.Post("http://localhost:8080/v1/swift-codes", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("❌ Błąd podczas wysyłania żądania:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("✅ Odpowiedź serwera:", resp.Status)
}

func deleteSwiftCode() {
	reader := bufio.NewReader(os.Stdin)

	// Ask for the SWIFT code to delete
	fmt.Print("🗑️ Podaj SWIFT Code do usunięcia: ")
	swiftCode, _ := reader.ReadString('\n')
	swiftCode = strings.TrimSpace(swiftCode)

	// Send the DELETE request
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/v1/swift-codes/%s", swiftCode), nil)
	if err != nil {
		fmt.Println("❌ Błąd podczas tworzenia żądania:", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ Błąd podczas wysyłania żądania:", err)
		return
	}
	defer resp.Body.Close()

	// Print server's response
	fmt.Println("✅ Odpowiedź serwera:", resp.Status)
}
