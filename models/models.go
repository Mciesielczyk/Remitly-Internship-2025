package models

// Struktura reprezentująca dane SWIFT
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
type Branch struct {
	SwiftCode     string `json:"swiftCode"`
	BankName      string `json:"bankName"`
	CountryISO2   string `json:"countryISO2"`
	Address       string `json:"address"`
	IsHeadquarter bool   `json:"isHeadquarter"`
	CountryName   string `json:"countryName"`
}

type Headquarter struct {
	SwiftCode     string   `json:"swiftCode"`
	BankName      string   `json:"bankName"`
	CountryISO2   string   `json:"countryISO2"`
	CountryName   string   `json:"countryName"`
	Address       string   `json:"address"`
	TimeZone      string   `json:"timeZone"`
	IsHeadquarter bool     `json:"isHeadquarter"`
	Branches      []Branch `json:"branches"`
}
