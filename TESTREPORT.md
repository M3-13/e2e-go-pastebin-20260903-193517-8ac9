VERDICT: BUGS_FOUND

**Bug 1**
- **Titel:** Test `TestCreateListDeleteRoutesRegistered` erwartet fälschlich keinen 404 für DELETE auf unbekannte ID
- **Symptom:** `go test ./...` bricht mit FAIL ab und blockiert die Abnahme/CI, obwohl das tatsächliche Verhalten des DELETE-Endpunkts (404 für eine unbekannte ID) der Spezifikation AC-04 entspricht. Der Test prüft die Routen-Registrierung anhand einer nicht existierenden ID und bewertet das spezifikationskonforme 404 als Fehler.
- **Repro:** `go test ./...` im Projektstamm ausführen.
- **Evidence:**  
  `--- FAIL: TestCreateListDeleteRoutesRegistered (0.00s)`  
  `    --- FAIL: TestCreateListDeleteRoutesRegistered/delete (0.00s)`  
  `        handler_behavior_test.go:112: DELETE /pastes/some-id answered 404; route must be registered`
- **Suspected file(s):** `handler_behavior_test.go` – der Untertest `delete` in `TestCreateListDeleteRoutesRegistered`. Er sollte vor dem DELETE eine gültige Paste anlegen oder die Erwartung an das laut AC-04 korrekte 404 für unbekannte IDs anpassen.
- **Severity:** high