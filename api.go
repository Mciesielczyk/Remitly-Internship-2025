package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"awesomeProject/models"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func startAPIServer() {
	client, err := ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/v1/swift-codes/{swiftcode}", GetSwiftCodeHandler(client)).Methods("GET")
	//r.HandleFunc("/v1/swift-codes", GetAllSwiftCodesHandler(client)).Methods("GET")
	r.HandleFunc("/v1/swift-codes/country/{countryISO2code}", GetSwiftCodesByCountryHandler(client)).Methods("GET")
	r.HandleFunc("/v1/swift-codes", AddSwiftCodeHandler(client)).Methods("POST")
	r.HandleFunc("/v1/swift-codes/{swift-code}", DeleteSwiftCodeHandler(client)).Methods("DELETE")

	fmt.Println("🚀 Serwer API działa na porcie 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func GetSwiftCodeHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		swiftcode := vars["swiftcode"] // Ściągamy swiftcode z URL

		// Tworzymy kontekst
		ctx := context.Background()

		// Kolekcje
		hqColl := client.Database("swift_data").Collection("headquarters")
		brColl := client.Database("swift_data").Collection("branches")

		// Sprawdzenie headquarters
		var hq models.Headquarter
		err := hqColl.FindOne(ctx, bson.M{"swiftcode": swiftcode}).Decode(&hq)
		if err == nil {
			// Jeśli jest to headquarters, dodajemy oddziały
			cursor, err := brColl.Find(ctx, bson.M{"bankName": hq.BankName})
			if err == nil {
				var branches []models.Branch
				if err := cursor.All(ctx, &branches); err == nil {
					// Sprawdzamy, czy mamy oddziały
					if len(branches) > 0 {
						hq.Branches = branches
					} else {
						log.Println("Brak oddziałów dla banku:", hq.BankName)
					}
				} else {
					log.Println("Błąd podczas ładowania oddziałów:", err)
				}
			} else {
				log.Println("Błąd podczas wyszukiwania oddziałów:", err)
			}

			// Zwróć HQ + branches w formacie JSON
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(hq)
			return
		}

		// Jeśli to nie jest HQ, sprawdzamy branch
		var branch models.Branch
		err = brColl.FindOne(ctx, bson.M{"swiftcode": swiftcode}).Decode(&branch)
		if err == nil {
			// Zwróć oddział
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(branch)
			return
		}

		// Jeśli nic nie znaleziono
		http.Error(w, "SWIFT code not found", http.StatusNotFound)
	}
}

func GetAllSwiftCodesHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		hqColl := client.Database("swift_data").Collection("headquarters")
		brColl := client.Database("swift_data").Collection("branches")

		// Pobierz wszystkie dane z kolekcji headquarters
		cursor, err := hqColl.Find(ctx, bson.M{}) // Pusty obiekt {}, czyli wszystkie dane
		if err != nil {
			http.Error(w, "Failed to fetch data from headquarters collection", http.StatusInternalServerError)
			return
		}

		var headquarters []models.Headquarter
		if err := cursor.All(ctx, &headquarters); err != nil {
			http.Error(w, "Failed to parse data from headquarters collection", http.StatusInternalServerError)
			return
		}

		// Pobierz wszystkie dane z kolekcji branches
		cursor, err = brColl.Find(ctx, bson.M{}) // Pusty obiekt {}, czyli wszystkie dane
		if err != nil {
			http.Error(w, "Failed to fetch data from branches collection", http.StatusInternalServerError)
			return
		}

		var branches []models.Branch
		if err := cursor.All(ctx, &branches); err != nil {
			http.Error(w, "Failed to parse data from branches collection", http.StatusInternalServerError)
			return
		}

		// Połącz wyniki w jedną strukturę
		result := map[string]interface{}{
			"headquarters": headquarters,
			"branches":     branches,
		}

		// Zwróć dane w formacie JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func GetSwiftCodesByCountryHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		countryiso2 := vars["countryISO2code"]

		// Tworzymy kontekst
		ctx := context.Background()

		// Kolekcje
		hqColl := client.Database("swift_data").Collection("headquarters")
		brColl := client.Database("swift_data").Collection("branches")

		// Przeszukujemy kolekcje headquarters
		var headquarters []models.Headquarter
		cursor, err := hqColl.Find(ctx, bson.M{"countryiso2": countryiso2}) // Używamy 'countryiso2' (małymi literami)
		if err != nil {
			http.Error(w, fmt.Sprintf("Błąd przy wyszukiwaniu headquarters: %s", err), http.StatusInternalServerError)
			return
		}

		// Przeszukujemy kolekcje branches
		var branches []models.Branch
		cursorBranches, err := brColl.Find(ctx, bson.M{"countryiso2": countryiso2}) // Używamy 'countryiso2' (małymi literami)
		if err != nil {
			http.Error(w, fmt.Sprintf("Błąd przy wyszukiwaniu branches: %s", err), http.StatusInternalServerError)
			return
		}

		// Mapowanie na struktury
		if err := cursor.All(ctx, &headquarters); err != nil {
			http.Error(w, fmt.Sprintf("Błąd przy przetwarzaniu headquarters: %s", err), http.StatusInternalServerError)
			return
		}

		if err := cursorBranches.All(ctx, &branches); err != nil {
			http.Error(w, fmt.Sprintf("Błąd przy przetwarzaniu branches: %s", err), http.StatusInternalServerError)
			return
		}

		// Zbieramy wszystkie wyniki (headquarters + branches)
		var swiftCodes []interface{} // Zmieniamy typ na interface{}, bo dane są różne

		// Dodajemy headquarters
		for _, hq := range headquarters {
			swiftCodes = append(swiftCodes, struct {
				Address       string `json:"address"`
				BankName      string `json:"bankName"`
				CountryISO2   string `json:"countryISO2"`
				IsHeadquarter bool   `json:"isHeadquarter"`
				SwiftCode     string `json:"swiftCode"`
			}{
				Address:       hq.Address,
				BankName:      hq.BankName,
				CountryISO2:   hq.CountryISO2,
				IsHeadquarter: true,
				SwiftCode:     hq.SwiftCode,
			})
		}

		// Dodajemy branches
		for _, br := range branches {
			swiftCodes = append(swiftCodes, struct {
				Address       string `json:"address"`
				BankName      string `json:"bankName"`
				CountryISO2   string `json:"countryISO2"`
				IsHeadquarter bool   `json:"isHeadquarter"`
				SwiftCode     string `json:"swiftCode"`
			}{
				Address:       br.Address,
				BankName:      br.BankName,
				CountryISO2:   br.CountryISO2,
				IsHeadquarter: false,
				SwiftCode:     br.SwiftCode,
			})
		}

		// Uzyskujemy nazwę kraju z pierwszego oddziału (jeśli istnieje)
		var countryName string
		if len(branches) > 0 {
			countryName = branches[0].CountryName
		} else {
			countryName = "Nieznany kraj" // Jeśli nie znaleziono żadnego oddziału
		}

		// Struktura odpowiedzi
		response := struct {
			CountryISO2 string        `json:"countryISO2"`
			CountryName string        `json:"countryName"`
			SwiftCodes  []interface{} `json:"swiftCodes"` // Używamy interface{}, aby obsłużyć różne struktury
		}{
			CountryISO2: countryiso2,
			CountryName: countryName, // Zmieniamy na nazwę kraju z pierwszego oddziału
			SwiftCodes:  swiftCodes,
		}

		// Zwrócenie odpowiedzi
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func AddSwiftCodeHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		// Parsowanie request body
		var newCode struct {
			Address       string `json:"address"`
			BankName      string `json:"bankname"`
			CountryISO2   string `json:"countryiso2"`
			CountryName   string `json:"countryname"`
			IsHeadquarter bool   `json:"isheadquarter"`
			SwiftCode     string `json:"swiftcode"`
		}

		if err := json.NewDecoder(r.Body).Decode(&newCode); err != nil {
			http.Error(w, "Nieprawidłowe dane wejściowe", http.StatusBadRequest)
			return
		}

		// W zależności od typu, wybierz kolekcję
		var collection *mongo.Collection
		if newCode.IsHeadquarter {
			collection = client.Database("swift_data").Collection("headquarters")
		} else {
			collection = client.Database("swift_data").Collection("branches")
		}

		// Wstaw dokument do kolekcji
		_, err := collection.InsertOne(ctx, bson.M{
			"address":       newCode.Address,
			"bankname":      newCode.BankName,
			"countryiso2":   newCode.CountryISO2,
			"countryname":   newCode.CountryName,
			"isheadquarter": newCode.IsHeadquarter,
			"swiftcode":     newCode.SwiftCode,
		})
		if err != nil {
			http.Error(w, "Błąd podczas zapisu do bazy danych", http.StatusInternalServerError)
			return
		}

		// Sukces
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "SWIFT code successfully added.",
		})
	}
}

func DeleteSwiftCodeHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get swiftCode from the URL parameters
		vars := mux.Vars(r)
		swiftCode := vars["swift-code"]

		// Create a context for the MongoDB operation
		ctx := context.Background()

		// Access the collections
		hqColl := client.Database("swift_data").Collection("headquarters")
		brColl := client.Database("swift_data").Collection("branches")

		// Delete the swiftCode from the headquarters collection
		deleteResult, err := hqColl.DeleteOne(ctx, bson.M{"swiftcode": swiftCode})
		if err != nil {
			http.Error(w, fmt.Sprintf("Błąd podczas usuwania w headquarters: %s", err), http.StatusInternalServerError)
			return
		}

		// If not found in headquarters, try deleting from branches
		if deleteResult.DeletedCount == 0 {
			deleteResult, err = brColl.DeleteOne(ctx, bson.M{"swiftcode": swiftCode})
			if err != nil {
				http.Error(w, fmt.Sprintf("Błąd podczas usuwania w branches: %s", err), http.StatusInternalServerError)
				return
			}
		}

		// Check if any document was deleted
		if deleteResult.DeletedCount == 0 {
			http.Error(w, "SWIFT code not found", http.StatusNotFound)
			return
		}

		// Return success message
		response := map[string]string{
			"message": "SWIFT code successfully deleted.",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
