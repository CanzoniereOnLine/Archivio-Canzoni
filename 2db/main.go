package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// SongMetadata struttura per memorizzare i metadati della canzone
type SongMetadata struct {
	Titolo         string
	Autore         string
	Album          string
	Tonalita       string
	Famiglia       string
	Gruppo         string
	Momenti        string
	Identificatore string
	DataRevisione  string
	Trascrittore   string
	TestoCompleto  string
	TestoSemplice  string
}

func parseLaTeXFile(filename string, regexFile string) (*SongMetadata, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	content := string(data)
	re := regexp.MustCompile(`%([a-zA-Z_]+)\{([^}]*)\}`)
	matches := re.FindAllStringSubmatch(content, -1)

	metadata := SongMetadata{}
	for _, match := range matches {
		key := strings.ToLower(match[1])
		value := match[2]
		switch key {
		case "titolo":
			metadata.Titolo = value
		case "autore":
			metadata.Autore = value
		case "album":
			metadata.Album = value
		case "tonalita":
			metadata.Tonalita = value
		case "famiglia":
			metadata.Famiglia = value
		case "gruppo":
			metadata.Gruppo = value
		case "momenti":
			metadata.Momenti = value
		case "identificatore":
			metadata.Identificatore = value
		case "data_revisione":
			metadata.DataRevisione = value
		case "trascrittore":
			metadata.Trascrittore = value
		}
	}

	metadata.TestoCompleto = content
	metadata.TestoSemplice = cleanLatex(content, regexFile)

	return &metadata, nil
}

func cleanLatex(content, regexFile string) string {
	file, err := os.Open(regexFile)
	if err != nil {
		fmt.Println("Errore nell'aprire il file delle regex:", err)
		return content
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	cleanedText := content

	for scanner.Scan() {
		pattern := scanner.Text()
		re := regexp.MustCompile(pattern)
		cleanedText = re.ReplaceAllString(cleanedText, "")
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Errore nella lettura del file delle regex:", err)
	}

	// Rimuove più di due a capo consecutivi
	reMultipleNewlines := regexp.MustCompile(`\n{3,}`)
	cleanedText = reMultipleNewlines.ReplaceAllString(cleanedText, "\n\n")

	// Rimuove eventuali a capo iniziali
	reLeadingNewlines := regexp.MustCompile(`^\n+`)
	cleanedText = reLeadingNewlines.ReplaceAllString(cleanedText, "")

	return cleanedText
}

func createDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "songs.db")
	if err != nil {
		return nil, err
	}

	dropTableQuery := `DROP TABLE IF EXISTS songs;`
	_, err = db.Exec(dropTableQuery)
	if err != nil {
		db.Close()
		return nil, err
	}

	createTableQuery := `CREATE TABLE IF NOT EXISTS songs (
		titolo TEXT,
		autore TEXT,
		album TEXT,
		tonalita TEXT,
		famiglia TEXT,
		gruppo TEXT,
		momenti TEXT,
		identificatore TEXT PRIMARY KEY,
		data_revisione TEXT,
		trascrittore TEXT,
		testo_completo TEXT,
		testo_semplice TEXT
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func insertMetadata(db *sql.DB, metadata *SongMetadata) error {
	insertQuery := `INSERT INTO songs (
		titolo, autore, album, tonalita, famiglia, gruppo, momenti, identificatore, data_revisione, trascrittore, testo_completo, testo_semplice
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.Exec(insertQuery, metadata.Titolo, metadata.Autore, metadata.Album, metadata.Tonalita, metadata.Famiglia, metadata.Gruppo, metadata.Momenti, metadata.Identificatore, metadata.DataRevisione, metadata.Trascrittore, metadata.TestoCompleto, metadata.TestoSemplice)
	return err
}

func processDirectory(directory string, regexFile string) {
	db, err := createDatabase()
	if err != nil {
		fmt.Println("Errore nella creazione del database:", err)
		return
	}
	defer db.Close()

	files, err := os.ReadDir(directory)
	if err != nil {
		fmt.Println("Errore nella lettura della cartella:", err)
		return
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".tex") {
			continue
		}
		filename := filepath.Join(directory, file.Name())
		metadata, err := parseLaTeXFile(filename, regexFile)
		if err != nil {
			fmt.Println("Errore durante la lettura del file", filename, ":", err)
			continue
		}

		err = insertMetadata(db, metadata)
		if err != nil {
			fmt.Println("Errore nell'inserimento dei dati per", filename, ":", err)
		} else {
			fmt.Println("Dati inseriti con successo per", filename)
		}
	}
}

func main() {
	directory := "../archivio-canzoni" // Cambia con il percorso della cartella contenente i file TeX
	regexFile := "regex.txt"           // File contenente le regex per la pulizia del testo

	processDirectory(directory, regexFile)
}
