VERDICT: BUGS_FOUND

- **Title:** Fehlerhaften Route-Registrierungstest für DELETE korrigieren
- **Symptom:** Die Testsuite ist rot; `go test ./...` bricht mit einem fehlgeschlagenen Test ab. Entwickler und CI können den Stand nicht als grün übernehmen, obwohl der eigentliche Produktcode das geforderte Verhalten zeigt.
- **Repro:** Im Projektverzeichnis `go test ./...` ausführen.
- **Evidence:**
  ```
  --- FAIL: TestCreateListDeleteRoutesRegistered (0.00s)
      --- FAIL: TestCreateListDeleteRoutesRegistered/delete (0.00s)
          handler_behavior_test.go:112: DELETE /pastes/some-id answered 404; route must be registered
  FAIL
  FAIL	pastebin	0.452s
  FAIL
  ```
- **Suspected file(s):** `handler_behavior_test.go` — der Test erwartet, dass `DELETE /pastes/some-id` bei unbekannter ID nicht mit 404 antwortet. Die Spezifikation verlangt für unbekannte oder bereits gelöschte IDs jedoch ausdrücklich 404 (AC-02/AC-04/AC-06), und genau dieses Verhalten setzt `paste_delete.go` korrekt um. Der Registrierungstest muss die Route mit einem existierenden Paste prüfen oder seine 404-Prüfung anpassen.
- **Severity:** high