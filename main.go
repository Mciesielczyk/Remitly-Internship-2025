package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

// Struktura reprezentująca dane z pliku Excel
type SwiftCode struct {
	CountryISO2   string `json:"countryISO2"`
	SwiftCode     string `json:"swiftCode"`
	CodeType      string `json:"codeType"`
	BankName      string `json:"bankName"`
	Address       string `json:"address"`
	TownName      string `json:"townName"`
	CountryName   string `json:"countryName"`
	TimeZone      string `json:"timeZone"`
	IsHeadquarter bool   `json:"isHeadquarter"`
}

func main() {
	// Otwórz plik Excel
	file, err := excelize.OpenFile("Interns_2025_SWIFT_CODES.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	// Pobierz arkusz (tutaj założyłem, że dane są w "Sheet1")
	rows, err := file.GetRows("Sheet1")
	if err != nil {
		log.Fatal(err)
	}

	// Przechowywanie wyników w tablicy
	var swiftCodes []SwiftCode

	// Iteracja przez wiersze i mapowanie danych do struktury
	for i, row := range rows {
		// Pomiń nagłówek (zakładając, że wiersz 0 to nagłówki)
		if i == 0 {
			continue
		}

		// Przypisanie wartości z wiersza do struktury
		swiftCode := SwiftCode{
			CountryISO2: row[0], // Kolumna 1: Country ISO2 Code
			SwiftCode:   row[1], // Kolumna 2: Swift Code
			CodeType:    row[2], // Kolumna 3: Code Type
			BankName:    row[3], // Kolumna 4: Bank Name
			Address:     row[4], // Kolumna 5: Address
			TownName:    row[5], // Kolumna 6: Town Name
			CountryName: row[6], // Kolumna 7: Country Name
			TimeZone:    row[7], // Kolumna 8: Time Zone
		}

		// Sprawdzamy, czy to jest centrala (jeśli kod SWIFT kończy się na "XXX")
		if len(swiftCode.SwiftCode) >= 3 && swiftCode.SwiftCode[len(swiftCode.SwiftCode)-3:] == "XXX" {
			swiftCode.IsHeadquarter = true
		} else {
			swiftCode.IsHeadquarter = false
		}

		// Dodaj do listy SWIFT Codes
		swiftCodes = append(swiftCodes, swiftCode)
	}

	// Zamień dane na format JSON
	jsonData, err := json.MarshalIndent(swiftCodes, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	// Zapisz dane do pliku JSON
	err = os.WriteFile("swift_codes.json", jsonData, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Plik JSON został zapisany.")

	// Teraz połącz się z MongoDB i załaduj dane z pliku JSON do bazy danych
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Wybierz bazę danych i kolekcję, do której chcesz zaimportować dane
	collection := client.Database("swift_data").Collection("swift_codes")

	// Otwórz plik JSON (teraz używamy jsonFile zamiast file)
	jsonFile, err := os.Open("swift_codes.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	// Zdekoduj dane z pliku JSON
	var data []SwiftCode
	decoder := json.NewDecoder(jsonFile)
	err = decoder.Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	// Przekształć dane na BSON i wstaw do bazy danych MongoDB
	var documents []interface{}
	for _, swiftCode := range data {
		documents = append(documents, swiftCode)
	}

	// Wstaw dokumenty do bazy danych
	_, err = collection.InsertMany(context.Background(), documents)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Dane zostały załadowane do bazy danych MongoDB.")
}
