# SWIFT Code API 

Aplikacja umożliwia:

- Parsowanie danych SWIFT z pliku Excel (.xlsx),
- Przechowywanie ich w bazie danych MongoDB,
- Udostępnienie danych przez REST API (pobieranie, dodawanie, usuwanie),
- Operacje poprzez CLI: ładowanie danych, uruchamianie serwera API, wysyłanie i usuwanie danych SWIFT.


##  Instalacja i uruchomienie

### 1. Klonowanie repozytorium

```bash
git clone https://github.com/Mciesielczyk/Remitly-Internship-2025.git

go mod tidy - sprawdzi czy masz wszsystkie zaleznosci pobrane

go run awesomeProject - uruchomi program
```

Używałen MongoDB Compass, w aplikacji dodac connection na url mongodb://localhost:27017
Najpierw należy uruchomić program i załadować excela do bazy MongoDB. (Powinna się utworzyć baza swift_data,
a wniej brach oraz headqurters.)
Następnie w innym terminalu uruchomić serwer API.
Na koncu można wysyłać zapytania do bazy (np. przez Postmana)

### Containerize the application
Niestety nie udało mi się zmusić do współpracy dockera, ale setup aplikacji powinien być możliwy na wszystkic srodowiskach.

Przykładowe zapytania: 
### 1. 
http://localhost:8080/v1/swift-codes/country/PL

![image](https://github.com/user-attachments/assets/4e761da5-8a48-45aa-81c5-fb6812c927e0)

### 2.
http://localhost:8080/v1/swift-codes

![image](https://github.com/user-attachments/assets/a1677473-d126-48ab-8505-cfe67db339c6)

### 3.
http://localhost:8080/v1/swift-codes/POLUPLXXX

![image](https://github.com/user-attachments/assets/b680cab5-2395-4464-b67d-00232c9a11e3)

### 4.
http://localhost:8080/v1/swift-codes/BREXPLPWXXX

![image](https://github.com/user-attachments/assets/45385842-6dcc-4c28-9a9b-5d19a90f1821)

# Opis pliku `main.go` 

- **1. Menu główne**: Użytkownik ma dwie opcje:
  - **1**: Załadowanie danych SWIFT z pliku Excel do bazy danych MongoDB.
  - **2**: Uruchomienie serwera API i przeprowadzanie operacji na danych SWIFT.
Po wybraniu jednej z opcji, użytkownik jest przekierowywany do odpowiedniej funkcji.

### 3. **Funkcja `handleServerOptions`**
Funkcja ta obsługuje wybór użytkownika, który zdecydował się na operacje związane z serwerem API:
- **1**: Uruchomienie serwera API (funkcja `startAPIServer`).
- **2**: Dodanie danych SWIFT przez API (funkcja `sendSwiftCode`).
- **3**: Usunięcie danych SWIFT przez API (funkcja `deleteSwiftCode`).
  
 ### 4. **Funkcja `sendSwiftCode`**
Ta funkcja pozwala użytkownikowi dodać dane SWIFT przez serwer API za pomocą żądania HTTP POST.
Po zebraniu danych, funkcja konwertuje je do formatu JSON i wysyła je do serwera API za pomocą metody POST na adres `http://localhost:8080/v1/swift-codes`. Odpowiedź serwera jest wyświetlana na konsoli.

### 5. **Funkcja `deleteSwiftCode`**
Ta funkcja umożliwia usunięcie danych SWIFT na podstawie podanego SWIFT Code. Funkcja wysyła zapytanie DELETE na serwerze API do URL `http://localhost:8080/v1/swift-codes/{swiftCode}`.

### 6. **Obsługa błędów**
Każda operacja wysyłania zapytań HTTP jest opakowana w blok `if` sprawdzający błędy. Jeśli wystąpi błąd (np. podczas konwersji danych na JSON lub wysyłania zapytania), użytkownik otrzyma odpowiedni komunikat o błędzie.


# Opis pliku `excel.go` 

Plik zawiera funkcję `ReadAndOrganizeExcel`, której zadaniem jest odczytanie danych z pliku Excel, przetworzenie ich i zorganizowanie w odpowiednie struktury, które następnie mogą być użyte do dalszego przetwarzania (np. zapis do bazy danych). Funkcja obsługuje dane SWIFT i rozróżnia centrale (headquarters) oraz oddziały (branches), eliminując duplikaty oddziałów.



### 1. **Funkcja `ReadAndOrganizeExcel`**
Główna funkcja w tym pliku, której celem jest przetworzenie danych z pliku Excel i zorganizowanie ich w odpowiednie struktury.

### 2. **Przetwarzanie danych z pliku Excel**
- Funkcja otwiera plik Excel.
- Odczytuje dane z arkusza.
- Iteruje przez wiersze i pomija nagłówki oraz wiersze, które nie zawierają wymaganych danych.

### 3. **Tworzenie struktur `SwiftCode`, `Headquarter`, `Branch`**
- **Dane SWIFT** są mapowane do struktury `SwiftCode` i przechowywane w tablicy `allSwiftCodes`.
- Funkcja rozróżnia kody SWIFT, gdzie kody kończące się na "XXX" są traktowane jako centrala banku, a pozostałe jako oddziały.
  
### 4. **Mapowanie centrali i oddziałów**
- **Headquarters**: Mapowanie centrali banków na podstawie nazwy banku (`BankName`). Dla każdej centrali tworzony jest obiekt `models.Headquarter`, który zawiera dane centrali oraz listę jej oddziałów.
- **Branches**: Oddziały są mapowane i eliminowane są duplikaty. Oddziały są przypisywane do odpowiadającej im centrali na podstawie nazwy banku.

### 5. **Usuwanie duplikatów oddziałów**
- Oddziały są identyfikowane na podstawie unikalnego klucza (kombinacja `SwiftCode` i `BankName`), co pozwala na eliminację duplikatów.
- Jeżeli dany oddział już istnieje w mapie `branchMap`, nie zostaje dodany do listy `branches`.

### 6. **Zbudowanie finalnej listy centrali**
- Na koniec funkcja tworzy finalną listę centrali (`headquarters`), która zawiera wszystkie zidentyfikowane i zorganizowane centrale oraz ich oddziały.

### 7. **Zwracanie wyników**
- Funkcja zwraca dwie listy:
  - **headquarters**: Zorganizowane listy centrali wraz z przypisanymi oddziałami.
  - **branches**: Zorganizowane listy oddziałów banków.
  - W przypadku błędów podczas odczytu pliku lub przetwarzania danych, zwrócony jest błąd.

# Opis pliku `db.go` 
1. **LoadDataToMong`** – funkcja odpowiedzialna za wczytywanie danych z pliku Excel, organizowanie ich w odpowiednie struktury i zapisanie do bazy danych MongoDB.
2. **ConnectDB** – funkcja nawiązująca połączenie z bazą danych MongoDB.

# Opis pliku `models.go` 
Plik zawiera  główne struktury danych używane w aplikacj, (headquarters) (branches).

# Opis pliku `api.go` 

### 1. `startAPIServer`
   Funkcja  uruchamia serwer API na porcie 8080. Łączy się z bazą danych MongoDB i konfiguruje router HTTP. Ustawia trasy do różnych handlerów (np. pobieranie, dodawanie, usuwanie kodów SWIFT).

### 2. `GetSwiftCodeHandler`
   Obsługuje zapytanie HTTP GET, które służy do pobrania danych związanych z danym kodem SWIFT. Funkcja sprawdza, czy kod SWIFT dotyczy centrali (headquarters) lub oddziału (branch) banku, a następnie zwraca odpowiednią strukturę w formacie JSON. W przypadku centrali dodaje również oddziały powiązane z tą centralą.

### 3. `GetAllSwiftCodesHandler`
   Obsługuje zapytanie HTTP GET, które pobiera wszystkie dane z kolekcji `headquarters` oraz `branches` w MongoDB. Łączy je w jedną strukturę i zwraca je w odpowiedzi w formacie JSON. Funkcja umożliwia pobranie pełnej listy centrali i oddziałów.

### 4. `GetSwiftCodesByCountryHandler`
   Obsługuje zapytanie HTTP GET, które filtruje kody SWIFT na podstawie kodu kraju (`countryISO2`). Pobiera dane z kolekcji centrali i oddziałów, które znajdują się w danym kraju, a następnie łączy je w jedną odpowiedź. Zwraca dane w formacie JSON, w tym nazwę kraju oraz powiązane kody SWIFT.

### 5. `AddSwiftCodeHandler`
   Obsługuje zapytanie HTTP POST, które pozwala na dodanie nowego kodu SWIFT do bazy danych. Funkcja oczekuje danych w formacie JSON, a następnie wstawia je do odpowiedniej kolekcji (`headquarters` lub `branches`) w MongoDB. Zwraca komunikat o sukcesie lub błędzie.

### 6. `DeleteSwiftCodeHandler`
   Obsługuje zapytanie HTTP DELETE, które usuwa dany kod SWIFT z bazy danych. Funkcja próbuje usunąć kod z kolekcji centrali (`headquarters`), a jeśli nie zostanie znaleziony, to z kolekcji oddziałów (`branches`). Zwraca odpowiedni komunikat o sukcesie lub błędzie.


# Testy

## 1. Funkcja `testRouter()`

Funkcja `testRouter()` tworzy testowy router HTTP z ustawionymi trasami dla handlerów:

- `GET /v1/swift-codes/{swiftcode}`: Pobiera kod SWIFT na podstawie wartości.
- `GET /v1/swift-codes/country/{countryISO2code}`: Pobiera kody SWIFT na podstawie kraju.
- `POST /v1/swift-codes`: Dodaje nowy kod SWIFT.
- `DELETE /v1/swift-codes/{swift-code}`: Usuwa kod SWIFT.

## 2. Funkcja `insertTestData()`
Funkcja wstawia testowe dane do bazy MongoDB przed uruchomieniem testów. Dodaje dane o centrali banku i oddziale.

## 3. Testy
### Test: `TestServerIntegration`
Testuje odpowiedź serwera na zapytanie GET dla nieistniejącego kodu SWIFT (`TESTCODE123`). Oczekiwane odpowiedzi:
- `404 Not Found` lub
- `200 OK`, w zależności od implementacji handlera.

### Test: `TestDeleteSwiftCodeHandler`
Sprawdza poprawność działania handlera usuwającego kod SWIFT z bazy danych. Test:
- Tworzy testowy wpis w kolekcji.
- Wysyła zapytanie `DELETE`.
- Weryfikuje, że rekord został usunięty z bazy.

### Test: `TestGetSwiftCodeHandler_NotFound`
Testuje zapytanie GET dla nieistniejącego kodu SWIFT. Oczekiwana odpowiedź: 
- `404 Not Found`.
- 
### Test: `TestAddSwiftCodeHandler`
Testuje dodanie nowego kodu SWIFT do bazy danych za pomocą zapytania `POST`. Sprawdza:
- Poprawność odpowiedzi.
- Zapisanie rekordu w kolekcji MongoDB.
- 
### Test: `TestGetSwiftCodesByCountryHandler`
Testuje pobieranie kodów SWIFT na podstawie kraju. Funkcja:
- Tworzy dane testowe dla kraju "Testonia".
- Sprawdza, czy odpowiedź zawiera oba kody SWIFT (dla centrali i oddziału).

## 4. Instrukcja

- **MongoDB**: Testy zakładają działającą bazę MongoDB na porcie `27017`.
- **Testy integracyjne**: Sprawdzają komunikację z bazą danych oraz poprawność odpowiedzi serwera na zapytania HTTP.
- **Użycie `httptest.NewRecorder()`**: Używane do testowania odpowiedzi HTTP bez uruchamiania prawdziwego serwera.

## 5. Podsumowanie

Testy obejmują wszystkie główne przypadki użytkowania API:
- Dodawanie,
- Usuwanie,
- Pobieranie danych,
- Obsługę błędów (np. brak kodu SWIFT).


# Obsługa Przypadków Brzegowych

### 1. **Brak danych w bazie**
Jeśli zapytanie dotyczy kodu SWIFT, który nie istnieje w bazie danych, API zwraca kod błędu HTTP `404 Not Found` z odpowiednim komunikatem:  
*"SWIFT code not found."*

Dla zapytania o kody SWIFT dla kraju, gdy kraj nie jest dostępny w bazie, odpowiedź wygląda następująco:  
*"Country not found in the database."*

### 2. **Niepoprawny kod SWIFT lub kod kraju**
Jeśli zapytanie zawiera niepoprawny kod SWIFT lub kod kraju, API zwraca kod błędu HTTP `400 Bad Request` z następującym komunikatem:  
*"Invalid SWIFT code or country code provided."*

### 3. **Niepoprawne dane wejściowe (np. brak wymaganych pól)**
Jeśli zapytanie `POST` zawiera niepełne dane (np. brak wymaganych pól jak `swiftcode`, `bankname`, `countryiso2`), API zwróci kod błędu HTTP `400 Bad Request` z następującym komunikatem:  
*"Missing required fields in the request body."*

### 4. **Błędy związane z bazą danych**
W przypadku problemów z połączeniem z bazą danych lub zapisaniem danych, API zwróci kod błędu HTTP `500 Internal Server Error` oraz ogólny komunikat o błędzie:  
*"Internal server error, please try again later."*

### 5. **Brak danych w odpowiedzi**
Jeśli zapytanie nie zwróci żadnych wyników (np. brak rekordów dla danego kodu SWIFT lub kraju), API zwróci kod błędu HTTP `404 Not Found` z komunikatem:  
*"No data found for the given criteria."*

### 6. **Zgodność z najlepszymi praktykami RESTful**
Wszystkie błędy są opatrzone odpowiednimi kodami statusu HTTP i komunikatami, zgodnie z zasadami REST API. Przykładem może być komunikat zwracany po usunięciu rekordu:  
*"SWIFT code successfully deleted."*

