package main

import (
	"awesomeProject/models"
	"github.com/xuri/excelize/v2"
	"log"
)

func ReadAndOrganizeExcel(filePath string) ([]models.Headquarter, []models.Branch, error) {
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Fatal(err)
		return nil, nil, err
	}

	rows, err := file.GetRows("Sheet1")
	if err != nil {
		log.Fatal(err)
		return nil, nil, err
	}

	var allSwiftCodes []models.SwiftCode
	for i, row := range rows {
		if i == 0 || len(row) < 8 {
			continue
		}

		swiftCode := models.SwiftCode{
			CountryISO2: row[0],
			SwiftCode:   row[1],
			CodeType:    row[2],
			BankName:    row[3],
			Address:     row[4],
			TownName:    row[5],
			CountryName: row[6],
			TimeZone:    row[7],
		}

		swiftCode.IsHeadquarter = len(swiftCode.SwiftCode) >= 3 && swiftCode.SwiftCode[len(swiftCode.SwiftCode)-3:] == "XXX"
		allSwiftCodes = append(allSwiftCodes, swiftCode)
	}

	// Mapowanie: bankName => Headquarter
	headquarterMap := make(map[string]*models.Headquarter)
	branchMap := make(map[string]models.Branch) // Mapa do usuwania duplikatów w branchach
	var branches []models.Branch

	// 1. Wypełniamy headquarters
	for _, code := range allSwiftCodes {
		if code.IsHeadquarter {
			// Sprawdzamy, czy już mamy taką centralkę
			if _, exists := headquarterMap[code.BankName]; !exists {
				headquarterMap[code.BankName] = &models.Headquarter{
					SwiftCode:     code.SwiftCode,
					BankName:      code.BankName,
					Address:       code.Address,
					CountryISO2:   code.CountryISO2,
					CountryName:   code.CountryName,
					TimeZone:      code.TimeZone,
					IsHeadquarter: true,
					Branches:      []models.Branch{}, // Pusty początkowy slice
				}
			}
		}
	}

	// 2. Dodajemy oddziały do headquarters i listy oddziałów
	for _, code := range allSwiftCodes {
		if !code.IsHeadquarter {
			// Tworzymy branch
			branch := models.Branch{
				SwiftCode:     code.SwiftCode,
				BankName:      code.BankName,
				Address:       code.Address,
				CountryISO2:   code.CountryISO2,
				CountryName:   code.CountryName,
				IsHeadquarter: false,
			}

			// Usuwamy duplikaty branchy
			branchKey := code.SwiftCode + "|" + code.BankName
			if _, exists := branchMap[branchKey]; !exists {
				branchMap[branchKey] = branch
				branches = append(branches, branch)

				// Dodajemy oddział do odpowiadającej mu centrali
				if hq, exists := headquarterMap[code.BankName]; exists {
					hq.Branches = append(hq.Branches, branch)
				}
			}
		}
	}

	// Zbuduj końcową listę headquarters
	var headquarters []models.Headquarter
	for _, hq := range headquarterMap {
		headquarters = append(headquarters, *hq)
	}

	return headquarters, branches, nil
}
