# Go-Pastebin REST-API

Eine kleine REST-API in Go, die ausschließlich die Standardbibliothek
(`net/http`) nutzt und Pastes in einem thread-sicheren In-Memory-Store verwaltet.
Pastes können angelegt, abgerufen, aufgelistet und gelöscht werden; optional
laufen sie nach einer konfigurierbaren Zeit ab. Alle Antworten sind sauberes JSON
mit korrekten Statuscodes.

## Tech-Stack

- **Sprache**: Go
- **Framework**: `net/http` (Standardbibliothek)
- **Build**: `go build`, `go test`

## Installation & Ausführen

```bash
go run .
```

Der Server startet standardmäßig auf Port `8080`. Der Port lässt sich über die
Umgebungsvariable `PORT` überschreiben:

```bash
PORT=9000 go run .
```

## Endpoints

| Methode | Pfad            | Beschreibung                                             |
|---------|-----------------|----------------------------------------------------------|
| GET     | `/health`       | Liveness-Check, antwortet mit `200`                      |
| POST    | `/pastes`       | Legt einen neuen Paste an                               |
| GET     | `/pastes/{id}`  | Liefert einen einzelnen Paste                          |
| GET     | `/pastes`       | Listet Metadaten aller gültigen Pastes                  |
| DELETE  | `/pastes/{id}`  | Löscht einen Paste                                     |

Alle Antworten sind `application/json`; Fehlerantworten bestehen ausschließlich
aus dem Feld `{"error": "..."}`.

Jede Anfrage erzeugt genau eine Zugriffszeile auf stdout mit Methode, Pfad,
Statuscode und Antwortdauer. Request-Bodys und Paste-Inhalte werden niemals
geloggt.
