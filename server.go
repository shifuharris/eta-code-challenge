// server.go
//
// REST APIs with Go and MySql.
//
// Usage:
//
//   # run go server in the background
//   $ go run server.go 

package main

import (
	"fmt"
	"log"
	"net/http"
	"database/sql"
	"flag"
	"encoding/json"

	_ "github.com/denisenkom/go-mssqldb"
)

type Title  struct {
    TitleName string
    ReleaseYear int
	Genres []struct {
		Name string
	}
	Participants []struct {
		IsKey string
		IsOnScreen string
		RoleType string
		Name string
		ParticipantType string
    }
	Storylines []struct{
		Description string
        Language string
		Type string
    }
	OtherNames []struct {
	    TitleName string
        TitleNameLanguage string
        TitleNameType string
    }
	Awards []struct {
        Award string
		AwardCompany string
		AwardWon string
		AwardYear int
    }
}

var (
	password = flag.String("P", "sqlpass1234!", "the database password")
	server = flag.String("S", "localhost", "the database server")
	userid = flag.String("U", "etauser", "the database user")
	database= flag.String("D", "Titles", "the database name")
)

// Respond to URLs of the form / and send static content
func Handler(response http.ResponseWriter, request *http.Request){
	http.ServeFile(response, request, request.URL.Path[1:])
}

// Respond to URLs of the form /api/titles
func APIHandler(response http.ResponseWriter, request *http.Request){
	//set mime type to JSON
    response.Header().Set("Content-type", "application/json")

    // Parse flag settings for SQL server
    flag.Parse()

    //Connect to database
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s", *server, *userid, *password, *database)
    db, err := sql.Open("mssql", connString)
    if err != nil {
        log.Fatal("Open connection failed:", err.Error())
    }

	defer db.Close()

	var titles []Title
	var cmd = `Select TitleName, ReleaseYear
	,(SELECT [Name] FROM Genre JOIN TitleGenre ON TitleGenre.GenreId = Genre.ID WHERE TitleGenre.TitleId = Title.TitleId FOR JSON PATH) AS Genres
	,(SELECT CASE tp.IsKey WHEN 'True' THEN 'Yes' ELSE 'No' END AS IsKey, CASE tp.IsOnScreen WHEN 'True' THEN 'Yes' ELSE 'No' END AS IsOnScreen, tp.RoleType, p.[Name], p.ParticipantType FROM Participant p JOIN TitleParticipant tp ON p.Id = tp.ParticipantId WHERE tp.TitleId = Title.TitleId FOR JSON PATH) AS Participants
	,(SELECT [Description], [Language], [Type] FROM Storyline Where StoryLine.TitleId = Title.TitleId FOR JSON PATH) Storylines
	,(SELECT TitleName, TitleNameLanguage, TitleNameType FROM OtherName Where OtherName.TitleId = Title.TitleId FOR JSON PATH) OtherNames
	,(SELECT Award, AwardCompany, CASE AwardWon WHEN 'True' THEN 'Yes' ELSE 'No' END AS AwardWon, AwardYear FROM Award Where Award.TitleId = Title.TitleId FOR JSON PATH) Awards
	FROM Title
	FOR JSON PATH`
	  
	titles, err = exec(db, cmd)
	if err != nil {
		fmt.Println(err)
	}

	json, err := json.Marshal(titles)
    if err != nil {
        log.Fatal(err)
    }

	// Send the json to the client.
    fmt.Fprintln(response, string(json))
}

func exec(db *sql.DB, cmd string) ([]Title, error) {
	rows, err := db.Query(cmd)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if cols == nil {
		return nil, nil
	}

	var titles = make([]Title, 0, 1000)
	var jStr string
	for rows.Next() {
		var tmpString string
		err = rows.Scan(&tmpString)
		if err != nil {
			log.Fatal("Scan: %v", err)
			return nil, nil
		}

		jStr = jStr + tmpString
	}

	if err := json.Unmarshal([]byte(jStr), &titles);
	 err != nil {
        log.Fatal("Unmarshal: %v", err)
    }
	return titles, nil
}

func main(){
    var err string
	http.HandleFunc("/api/titles", APIHandler)
	http.HandleFunc("/", Handler)
	
	// Start listing on a given port with these routes on this server.
	log.Print("Listening on port 8080 ... ")
	errs := http.ListenAndServe("localhost:8080", nil)
	if errs != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}