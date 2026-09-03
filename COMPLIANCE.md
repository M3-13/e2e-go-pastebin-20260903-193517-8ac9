VERDICT: CHANGES_REQUESTED

Geprüft wurde der übergebene Stand des Go-Backends (REST-Pastebin-API) anhand der für den Projekttyp `go-backend` einschlägigen Pflichten: DSGVO und EU Cyber Resilience Act (CRA). KI-Funktionen, öffentliche Web-UI, Impressums-/Cookie-Pflichten und Barrierefreiheit sind hier nicht einschlägig. Positiv fällt auf, dass die im Sprint definierten Schutzziele AC-08 bis AC-12 im Code umgesetzt sind: Request-Body-Limit, Content-Type-Zwang, JSON-Escaping, minimale Fehlerantworten und logging ohne Request-Bodies. Es bestehen jedoch behebbare Datenschutz- und CRA-Lücken.

## 1. DSGVO

### 1.1 Fehlende Rechtsgrundlage und fehlende Datenschutzdokumentation — **hoch**
**Befund:** Die API speichert beliebige Paste-Inhalte, die personenbezogene Daten enthalten können. Im geprüften Quellcode ist keine Rechtsgrundlage nach Art. 6 DSGVO benannt, und es fehlt eine Datenschutzdokumentation bzw. ein Muster-Hinweis für den Betreiber. Da das Produkt selbst keine Endnutzer-UI hat, muss der Betreiber die Informationen an anderer Stelle bereitstellen; das Produkt sollte ihn dabei unterstützen.  
**Abhilfe:** In `README.md` einen Abschnitt „Datenschutz / Privacy“ ergänzen mit:
- Zweck der Verarbeitung (temporäres Speichern und Abrufen von Paste-Inhalten),
- Rechtsgrundlage: Art. 6 Abs. 1 lit. b DSGVO (Erfüllung des Nutzungsvertrags beim bewussten Upload durch den Nutzer), hilfsweise Art. 6 Abs. 1 lit. f DSGVO,
- Speicherdauer: abhängig von `expires_in_seconds`, sonst bis zum Löschen/Neustart,
- Hinweise zu Betroffenenrechten (Auskunft durch GET, Löschung durch DELETE),
- Pflicht des Betreibers, eine eigene Datenschutzerklärung zu veröffentlichen.

### 1.2 Unverschlüsselter Transport — **hoch**
**Befund:** `main.go` startet den Server ausschließlich mit `srv.ListenAndServe()` ohne TLS. Paste-Inhalte – potenziell personenbezogene Daten – würden im Klartext über das Netz übertragen. Das verletzt die Anforderungen aus Art. 32 DSGVO an die Vertraulichkeit und Integrität der Verarbeitung.  
**Abhilfe:** In `main.go` entweder TLS direkt aktivieren (`srv.ListenAndServeTLS("cert.pem", "key.pem")`, gesteuert über Umgebungsvariablen wie `TLS_CERT`/`TLS_KEY`) oder im `README.md` verbindlich festhalten, dass der Dienst ausschließlich hinter einem TLS-terminierenden Reverse Proxy betrieben werden darf und niemals direkt im Klartext exponiert wird.

### 1.3 `DELETE` ohne Authentifizierung — **hoch**
**Befund:** Jeder, der die zufällige Paste-ID kennt, kann den Paste löschen (`paste_delete.go`). Bei personenbezogenen Inhalten gefährdet das die Integrität und Verfügbarkeit der Daten. Auch `GET` ist ohne Zugriffsschutz möglich; das entspricht zwar einem klassischen Pastebin-Modell, muss datenschutzrechtlich aber klar geregelt sein.  
**Abhilfe:** Für `DELETE` eine Authentifizierung einführen, z. B. einen beim Anlegen zurückgegebenen Lösch-Token, der bei `DELETE /pastes/{id}` im `Authorization`-Header mitgesendet werden muss. Alternativ im `README.md` verbindlich dokumentieren, dass das Produkt nur in einer vertrauenswürdigen Umgebung hinter einem Authentifizierungs-Proxy betrieben werden darf. Ohne diese Schutzmaßnahme sollte der Dienst keine personenbezogenen Daten verarbeiten.

### 1.4 Unbegrenzte Speicherdauer und fehlendes aktives Löschen abgelaufener Pastes — **mittel**
**Befund:** Pastes ohne `expires_in_seconds` erhalten `ExpiresAt = nil` und bleiben bis zum Löschen oder Prozessende im Speicher. Abgelaufene Pastes werden in `store.go` nur bei `Get` oder `List` entfernt. Ein nie wieder abgerufener oder gelisteter Paste bleibt unbegrenzt im RAM. Das widerspricht dem Grundsatz der Speicherbegrenzung nach Art. 5 Abs. 1 lit. e DSGVO.  
**Abhilfe:** In `store.go` eine Methode `CleanupExpired()` ergänzen, die alle abgelaufenen Einträge löscht. In `main.go` einen Hintergrund-Job starten, z. B. mit `time.NewTicker`, der diese Methode regelmäßig aufruft. Zusätzlich eine konfigurierbare maximale Default-/Höchstaufbewahrungsdauer dokumentieren oder implementieren.

### 1.5 Pfad-Logging enthält Paste-ID — **niedrig**
**Befund:** `loggingMiddleware` in `main.go` loggt `r.URL.Path`. Bei `GET /pastes/{id}` und `DELETE /pastes/{id}` erscheint die zufällige ID im Log. Die ID ist pseudonym, kann aber in Kombination mit anderem Wissen personenbeziehbar werden. Positiv ist, dass keine IP-Adresse, kein User-Agent und kein Request-Body geloggt wird.  
**Abhilfe:** Optional in `loggingMiddleware` den Pfad für `/pastes/{id}` maskieren, z. B. `/pastes/:id` loggen. Falls der Pfad aus Wartungsgründen erhalten bleiben soll, die Aufbewahrungsdauer der Logs begrenzen und diese Entscheidung im `README.md` dokumentieren.

### 1.6 Bereits erfüllte DSGVO-Anforderungen
- Datenminimierung: `GET /pastes` liefert ausschließlich Metadaten ohne `content` (`paste_list.go`).
- Fehlerantworten ohne interne Details und ohne Paste-Inhalte (`handler.go`, `paste_create.go`).
- HTML-Escaping von `content` durch `encoding/json` (`handler.go`, `paste_get_test.go`).
- Kryptografisch sichere ID-Erzeugung mit `crypto/rand` (`id.go`).

## 2. EU Cyber Resilience Act (CRA)

### 2.1 Kein SBOM und keine dokumentierte Sicherheits-/Update-Dokumentation — **mittel**
**Befund:** Der Quellcode enthält keine sichtbare Software-Stückliste (SBOM), keine dokumentierten Sicherheitseigenschaften und keinen Update-/Patch-Prozess. Die CRA verlangt für Produkte mit digitalen Elementen ein Mindestmaß an dokumentierter Sicherheits- und Update-Fähigkeit.  
**Abhilfe:** In `README.md` einen Abschnitt „Security & SBOM“ ergänzen:
- Modulname und Go-Version benennen,
- Auflistung der Abhängigkeiten (ausschließlich Go-Standardbibliothek),
- Versionsstand des Produkts,
- Prozess für Sicherheitsupdates (z. B. Rebuild und Redeployment),
- Kontaktadresse für Schwachstellenmeldungen.

### 2.2 Fehlende Server-Timeouts und fehlende Schutzmaßnahmen gegen Missbrauch — **mittel**
**Befund:** Der `http.Server` in `main.go` setzt keine `ReadTimeout`, `ReadHeaderTimeout`, `WriteTimeout` oder `IdleTimeout`. Zudem fehlt ein Rate-Limit. Aus Sicht der CRA (Security by design/default) ist die Standardkonfiguration damit nicht ausreichend widerstandsfähig.  
**Abhilfe:** In `main.go` den `http.Server` ergänzen:
```go
ReadHeaderTimeout: 5 * time.Second,
ReadTimeout:       10 * time.Second,
WriteTimeout:      10 * time.Second,
IdleTimeout:       120 * time.Second,
```
Ein Rate-Limiting entweder als Middleware implementieren oder im `README.md` verbindlich auf einen vorgelagerten Proxy verweisen, der dieses übernimmt.

### 2.3 Bind an alle Netzwerk-Interfaces als unsichere Standardeinstellung — **niedrig**
**Befund:** `Addr: ":" + port` bindet den Dienst an alle Interfaces. Wird die API versehentlich direkt exponiert, ist sie ohne TLS und Authentifizierung erreichbar.  
**Abhilfe:** Standardmäßig nur auf `127.0.0.1:` binden (`Addr: "127.0.0.1:" + port`) und eine explizite Konfiguration für den produktiven Betrieb vorsehen. Alternativ im `README.md` klarstellen, dass der Dienst nur hinter Firewall/Reverse Proxy exponiert werden darf.

### 2.4 Bereits erfüllte CRA-Aspekte
- Begrenzung der Request-Body-Größe (`maxPasteBodyBytes`, `http.MaxBytesReader` in `paste_create.go`).
- Strikte Content-Type-Prüfung.
- Thread-sicherer Store durch `sync.Mutex`.
- Fehler ohne interne Details und minimale Logdaten.
- Kryptografisch sichere ID-Generierung.

## 3. EU AI Act
Nicht anwendbar. Das Produkt enthält keine KI-Funktion.

## 4. Pflichttexte & UI
Nicht anwendbar. Es handelt sich um ein reines Backend ohne Endnutzer-UI. Impressumspflicht, Cookie-Banner und Einwilligungstexte greifen auf dieser Ebene nicht. Die nötige Datenschutzdokumentation ist in Abschnitt 1.1 adressiert.

## 5. Barrierefreiheit
Nicht anwendbar. Es existiert keine öffentliche Web-UI; WCAG/BITV/EAA sind hier nicht einschlägig.

## Fazit
Der Stand erfüllt die funktionalen Sicherheitsanforderungen aus den Acceptance Criteria weitgehend. Es fehlen jedoch zentrale Datenschutz- und CRA-Vorkehrungen: TLS, Schutz des Löschwegs, klare Rechtsgrundlage/Aufbewahrungsregelung für Betreiber sowie ein SBOM-/Security-Abschnitt. Diese Punkte sind behebbar und begründen `CHANGES_REQUESTED`, nicht `BLOCKED`.