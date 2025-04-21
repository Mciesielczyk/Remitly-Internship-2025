package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func LoadDataToMongo() {
	// Wczytaj dane z Excela i rozdziel na centrale i oddziały
	hqList, branchList, err := ReadAndOrganizeExcel("Interns_2025_SWIFT_CODES.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	client, err := ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	headquartersCollection := client.Database("swift_data").Collection("headquarters")
	branchesCollection := client.Database("swift_data").Collection("branches")

	// Wyczyść kolekcje przed wczytaniem (opcjonalne)
	headquartersCollection.DeleteMany(context.Background(), map[string]interface{}{})
	branchesCollection.DeleteMany(context.Background(), map[string]interface{}{})

	for _, hq := range hqList {
		result, err := headquartersCollection.InsertOne(context.Background(), hq)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Centrala dodana:", result.InsertedID)
	}

	for _, branch := range branchList {
		result, err := branchesCollection.InsertOne(context.Background(), branch)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Oddział dodany:", result.InsertedID)
	}

	fmt.Println(" Dane zostały załadowane do bazy danych MongoDB.")
}

func ConnectDB() (*mongo.Client, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Testowanie połączenia
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return client, nil
}
