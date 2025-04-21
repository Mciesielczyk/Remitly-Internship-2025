package main

import (
	"bytes"
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func testRouter() *mux.Router {
	client, err := ConnectDB()
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/v1/swift-codes/{swiftcode}", GetSwiftCodeHandler(client)).Methods("GET")
	r.HandleFunc("/v1/swift-codes/country/{countryISO2code}", GetSwiftCodesByCountryHandler(client)).Methods("GET")
	r.HandleFunc("/v1/swift-codes", AddSwiftCodeHandler(client)).Methods("POST")
	r.HandleFunc("/v1/swift-codes/{swift-code}", DeleteSwiftCodeHandler(client)).Methods("DELETE")
	return r
}
func insertTestData(client *mongo.Client, t *testing.T) {
	ctx := context.TODO()

	hqColl := client.Database("swift_data").Collection("headquarters")
	brColl := client.Database("swift_data").Collection("branches")

	_, _ = hqColl.DeleteMany(ctx, bson.M{"swiftcode": "BCHICLR10R2"})
	_, _ = brColl.DeleteMany(ctx, bson.M{"bankname": "Bank Testowy"})

	_, err := hqColl.InsertOne(ctx, bson.M{
		"swiftcode":   "BCHICLR10R2",
		"bankname":    "Bank Testowy",
		"countryiso2": "PL",
		"countryname": "Polska",
		"address":     "ul. Główna 1",
	})
	if err != nil {
		t.Fatalf(" Insert HQ error: %v", err)
	}

}

func TestServerIntegration(t *testing.T) {
	router := testRouter()

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/v1/swift-codes/TESTCODE123")
	if err != nil {
		t.Fatalf("Błąd podczas zapytania GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusOK {
		t.Errorf("Oczekiwano 200 lub 404, otrzymano: %d", resp.StatusCode)
	}
}

func TestDeleteSwiftCodeHandler(t *testing.T) {
	client, err := ConnectDB()
	if err != nil {
		t.Fatalf("Nie można połączyć się z MongoDB: %v", err)
	}
	ctx := context.Background()

	swiftCode := "DELETE123"
	_, err = client.Database("swift_data").Collection("branches").InsertOne(ctx, bson.M{
		"swiftcode":     swiftCode,
		"address":       "testowa 1",
		"bankname":      "Bank Testowy",
		"countryiso2":   "PL",
		"countryname":   "Polska",
		"isheadquarter": false,
	})
	if err != nil {
		t.Fatalf("Nie udało się wstawić testowego dokumentu: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/v1/swift-codes/{swift-code}", DeleteSwiftCodeHandler(client)).Methods("DELETE")

	req := httptest.NewRequest("DELETE", "/v1/swift-codes/"+swiftCode, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Oczekiwano kodu 200, otrzymano %d", rr.Code)
	}

	var resp map[string]string
	err = json.NewDecoder(strings.NewReader(rr.Body.String())).Decode(&resp)
	if err != nil {
		t.Errorf("Nie udało się sparsować odpowiedzi JSON: %v", err)
	}

	if resp["message"] != "SWIFT code successfully deleted." {
		t.Errorf("Niepoprawna odpowiedź: %v", resp["message"])
	}

	count, err := client.Database("swift_data").Collection("branches").
		CountDocuments(ctx, bson.M{"swiftcode": swiftCode})
	if err != nil {
		t.Errorf("Błąd podczas zliczania dokumentów: %v", err)
	}
	if count != 0 {
		t.Errorf("Rekord nie został usunięty z bazy danych")
	}
}

func setupTestMongoClient(t *testing.T) *mongo.Client {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		t.Fatalf(" Nie udało się połączyć z MongoDB: %v", err)
	}
	return client
}

func TestGetSwiftCodeHandler_NotFound(t *testing.T) {
	client := setupTestMongoClient(t)

	req := httptest.NewRequest("GET", "/v1/swift-codes/NIEISTNIEJE", nil)
	req = mux.SetURLVars(req, map[string]string{
		"swiftcode": "NIEISTNIEJE",
	})

	rr := httptest.NewRecorder()
	handler := GetSwiftCodeHandler(client)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAddSwiftCodeHandler(t *testing.T) {
	// Tworzymy połączenie z testową bazą danych
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	// Ustawiamy testową bazę danych
	hqColl := client.Database("swift_data").Collection("headquarters")
	brColl := client.Database("swift_data").Collection("branches")

	// Usuwamy dane przed testem
	_, _ = hqColl.DeleteMany(context.Background(), bson.M{})
	_, _ = brColl.DeleteMany(context.Background(), bson.M{})

	// Tworzymy dane do testu
	data := map[string]interface{}{
		"address":       "ul. Testowa 1",
		"bankname":      "Test Bank",
		"countryiso2":   "PL",
		"countryname":   "Polska",
		"isheadquarter": true,
		"swiftcode":     "BCHICLR10R1",
	}

	// Kodujemy dane do JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Error encoding JSON: %v", err)
	}

	// Tworzymy nowy request HTTP
	req, err := http.NewRequest("POST", "/v1/swift-codes", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Tworzymy rekordera odpowiedzi HTTP
	rr := httptest.NewRecorder()

	// Tworzymy mux routera i ustawiamy handler
	router := mux.NewRouter()
	router.HandleFunc("/v1/swift-codes", AddSwiftCodeHandler(client)).Methods("POST")

	// Uruchamiamy handler
	router.ServeHTTP(rr, req)

	// Sprawdzamy odpowiedź
	assert.Equal(t, http.StatusOK, rr.Code, "Oczekiwano statusu 200 OK")

	// Testujemy, czy odpowiedź zawiera odpowiedni komunikat
	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	assert.Equal(t, "SWIFT code successfully added.", response["message"])

	// Sprawdzamy, czy dane zostały zapisane w kolekcji headquarters
	var result bson.M
	err = hqColl.FindOne(context.Background(), bson.M{"swiftcode": "BCHICLR10R1"}).Decode(&result)
	if err != nil {
		t.Fatalf(" Error fetching data from MongoDB: %v", err)
	}

	// Sprawdzamy, czy zapisano poprawnie
	assert.Equal(t, "ul. Testowa 1", result["address"], "Adres powinien się zgadzać")
	assert.Equal(t, "Test Bank", result["bankname"], "Nazwa banku powinna się zgadzać")
	assert.Equal(t, "PL", result["countryiso2"], "Kod kraju powinien się zgadzać")
	assert.Equal(t, "Polska", result["countryname"], "Nazwa kraju powinna się zgadzać")
	assert.True(t, result["isheadquarter"].(bool), "Powinno to być headquarters")
	assert.Equal(t, "BCHICLR10R1", result["swiftcode"], "Kod SWIFT powinien się zgadzać")

	// Po teście usuwamy dane testowe
	_, _ = hqColl.DeleteOne(context.Background(), bson.M{"swiftcode": "BCHICLR10R1"})
	_, _ = brColl.DeleteOne(context.Background(), bson.M{"swiftcode": "BCHICLR10R1"})
}

func TestGetSwiftCodesByCountryHandler(t *testing.T) {
	// Połącz się z MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	hqColl := client.Database("swift_data").Collection("headquarters")
	brColl := client.Database("swift_data").Collection("branches")

	// Wybieramy losowy kraj testowy, np. "XZ" (nieistniejący kod ISO)
	testISO := "XZ"
	testCountry := "Testonia"
	hqSwift := "TESTXZHQ001"
	brSwift := "TESTXZBR001"

	// Upewniamy się, że nie ma tych danych wcześniej
	_ = hqColl.FindOneAndDelete(context.Background(), bson.M{"swiftcode": hqSwift})
	_ = brColl.FindOneAndDelete(context.Background(), bson.M{"swiftcode": brSwift})

	// Dodajemy dane testowe
	_, _ = hqColl.InsertOne(context.Background(), bson.M{
		"address":       "ul. Testowa 1",
		"bankname":      "Test Bank HQ",
		"countryiso2":   testISO,
		"countryname":   testCountry,
		"swiftcode":     hqSwift,
		"isheadquarter": true,
	})

	_, _ = brColl.InsertOne(context.Background(), bson.M{
		"address":       "ul. Oddziałowa 1",
		"bankname":      "Test Bank Branch",
		"countryiso2":   testISO,
		"countryname":   testCountry,
		"swiftcode":     brSwift,
		"isheadquarter": false,
	})

	// Tworzymy zapytanie GET
	req, err := http.NewRequest("GET", "/v1/swift-codes/"+testISO, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Recorder odpowiedzi i router
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/v1/swift-codes/{countryISO2code}", GetSwiftCodesByCountryHandler(client)).Methods("GET")
	router.ServeHTTP(rr, req)

	// Sprawdzamy status odpowiedzi
	assert.Equal(t, http.StatusOK, rr.Code, "Oczekiwano statusu 200 OK")

	// Dekodujemy odpowiedź
	var response struct {
		CountryISO2 string        `json:"countryISO2"`
		CountryName string        `json:"countryName"`
		SwiftCodes  []interface{} `json:"swiftCodes"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	// Sprawdzamy poprawność danych
	assert.Equal(t, testISO, response.CountryISO2)
	assert.Equal(t, testCountry, response.CountryName)
	assert.Len(t, response.SwiftCodes, 2)

	// Wyciągamy kody swift
	swifts := make(map[string]bool)
	for _, entry := range response.SwiftCodes {
		swift := entry.(map[string]interface{})["swiftCode"].(string)
		swifts[swift] = true
	}
	assert.True(t, swifts[hqSwift], "Brak kodu HQ")
	assert.True(t, swifts[brSwift], "Brak kodu Branch")

	// Usuwamy dane po teście
	_, _ = hqColl.DeleteOne(context.Background(), bson.M{"swiftcode": hqSwift})
	_, _ = brColl.DeleteOne(context.Background(), bson.M{"swiftcode": brSwift})
}
