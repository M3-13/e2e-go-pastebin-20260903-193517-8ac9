VERDICT: CHANGES_REQUESTED

## Sicherheitsbericht

**Zusammenfassung:** Es wurden keine kritischen oder hohen Sicherheitslücken gefunden. Die API erzwingt die JSON-Content-Type-Prüfung, begrenzt den Request-Body auf 1 MiB, escapt JSON-Inhalte standardmäßig HTML-sicher und loggt keine Paste-Inhalte. Die nachfolgenden Punkte sind überwiegend Härtungsmaßnahmen und ein mittleres DoS-/Log-Risiko, die vor einem Produktivbetrieb umgesetzt werden sollten.

**Scanner-Lücke:** Für dieses Projekt wurde kein automatischer Security-Scanner ausgeführt (`(no applicable security scanners for this project type)`). Die Bewertung beruht daher auf manueller Codeanalyse. Da nur die Go-Standardbibliothek genutzt wird, sind keine bekannten Dependency-Schwachstellen erkennbar; die `go.mod` liegt nicht vollständig im Kontext vor.

### Befunde

1. **[Mittel] Fehlende HTTP-Server-Timeouts**
   - **Datei/Stelle:** `main.go`, `srv := &http.Server{Addr: ..., Handler: ...}`
   - **Risiko:** Ohne Timeouts kann ein Angreifer langsame Verbindungen (Slowloris) aufrechterhalten und Server-Ressourcen blockieren.
   - **Fix:**
     ```go
     srv := &http.Server{
         Addr:              ":" + port,
         Handler:           loggingMiddleware(os.Stdout, newRouter()),
         ReadHeaderTimeout: 5 * time.Second,
         ReadTimeout:       10 * time.Second,
         WriteTimeout:      20 * time.Second,
         IdleTimeout:       120 * time.Second,
     }
     ```
     Das beeinträchtigt keine bestehenden Funktionen oder Tests.

2. **[Mittel] Unbegrenztes Wachstum des In-Memory-Stores**
   - **Datei/Stelle:** `store.go`, `Create` / `paste_create.go`
   - **Risiko:** Massenhaftes Anlegen von Pastes ohne Ablaufdatum füllt den Speicher und führt zu einem Denial of Service.
   - **Fix:** Eine konfigurierbare Obergrenze einführen, z. B. über eine Umgebungsvariable `MAX_PASTES` mit hohem Standardwert (z. B. 10000) und einer Store-Fehlermeldung, die der Handler als `503`/`429` mit JSON-Fehler `{"error":"store full"}` zurückgibt.
     ```go
     var maxPastes = envInt("MAX_PASTES", 10000)

     // in Create unter dem Lock:
     if len(s.pastes) >= maxPastes {
         return Paste{}, errStoreFull
     }
     ```
     Zusätzlich empfiehlt sich ein Standardablauf für Pastes ohne explizites `expires_in_seconds`.

3. **[Niedrig/Mittel] Log-Injection über den URL-Pfad**
   - **Datei/Stelle:** `main.go`, `loggingMiddleware`, `fmt.Fprintf(out, "%s %s %d %s\n", ..., r.URL.Path, ...)`
   - **Risiko:** Prozentkodierte Steuerzeichen (z. B. `%0A`) im Pfad werden von `net/url` dekodiert und erzeugen Zeilenumbrüche im Log. Ein Angreifer kann so gefälschte Logzeilen einschleusen.
   - **Fix:** Den Pfad vor dem Loggen bereinigen, ohne normale Pfade wie `/health` zu verändern:
     ```go
     path := strings.Map(func(r rune) rune {
         if unicode.IsControl(r) {
             return -1
         }
         return r
     }, r.URL.Path)
     fmt.Fprintf(out, "%s %s %d %s\n", r.Method, path, status, time.Since(start))
     ```
     (Import `unicode` ergänzen; `strings` ist bereits vorhanden.)

4. **[Niedrig] Fehlende Security-Header für JSON-Antworten**
   - **Datei/Stelle:** `handler.go`, `writeJSON`
   - **Risiko:** Ältere Browser könnten JSON als anderes Format interpretieren; der Header `X-Content-Type-Options: nosniff` verhindert MIME-Sniffing und unterstützt die XSS-Vermeidung.
   - **Fix:** In `writeJSON` setzen:
     ```go
     w.Header().Set("X-Content-Type-Options", "nosniff")
     ```
     Optional für sensible Inhalte:
     ```go
     w.Header().Set("Cache-Control", "no-store")
     ```

5. **[Niedrig] Kein TLS am HTTP-Server**
   - **Datei/Stelle:** `main.go`, `srv.ListenAndServe()`
   - **Risiko:** Bei direkter Exposition werden IDs und Paste-Inhalte unverschlüsselt übertragen.
   - **Fix:** TLS-Terminierung über einen vorgelagerten Reverse-Proxy (empfohlene Standard-Architektur) oder direkten Einsatz von `ListenAndServeTLS` mit gültigen Zertifikaten. Dokumentieren, falls der Dienst ausschließlich intern erreichbar ist.

### Designhinweis

- **Kein Rate-Limiting / keine Authentifizierung** – Die API ist aktuell vollständig öffentlich und unbegrenzt aufrufbar. Falls sie produktiv und internet-exponiert betrieben wird, sollten zusätzlich eine Rate-Limiting-Middleware und optional ein API-Schlüssel-Mechanismus ergänzt werden. Dies ist nicht Teil der Akzeptanzkriterien, aber für den Betrieb relevant.