package BackEnd

import (
    "database/sql"
	"time"
    "github.com/lib/pq"	
)

func GetDb()(*sql.DB, error){
	config, _ := GetConfiguration()
	db, err := sql.Open("postgres",config.DBCockroachConnection)
    if err != nil {
        return nil, err
    }

	// Create the "urlCache" table.
    if _, err := db.Exec(
        "CREATE TABLE IF NOT EXISTS urlCache (url VARCHAR(512), lastTimeChecked TIMESTAMP, jsonResponse TEXT, CONSTRAINT url_primary_key_constraint PRIMARY KEY (url) )"); err != nil {
		return nil, err
    }
	
	// Create the "requestHistory" table.
    if _, err := db.Exec(
        "CREATE TABLE IF NOT EXISTS requestHistory (hashIdentifier VARCHAR(32), url VARCHAR(512), CONSTRAINT request_history_primary_constraint PRIMARY KEY (hashIdentifier, url))"); err != nil {
		return nil, err
    }

	return db,nil	
}

// Return the lastTimeChecked int, jsonResponse string
func GetUrlCache(db *sql.DB, url string)(int64, string, int64, error){
	var lastTimeChecked int64 = 0
	var jsonResponse string = ""
	var currentTime int64 = 0

	var lastTimeCheckedTime time.Time 
	var currentTimeTime time.Time 

	err := db.QueryRow("SELECT lastTimeChecked, jsonResponse, NOW() as currentTime FROM urlCache WHERE url=$1", url).Scan(&lastTimeCheckedTime, &jsonResponse, &currentTimeTime)
	
	if err == sql.ErrNoRows {
		return 0,"",0,nil
	}
	if err != nil {
		return 0,"",0,err
	}

	lastTimeChecked = lastTimeCheckedTime.Unix()
	currentTime = currentTimeTime.Unix()

	return lastTimeChecked,jsonResponse,currentTime,nil
}


// Return the lastTimeChecked int, jsonResponse string
func GetCheckHistory(db *sql.DB, hashIdentifier string)([]string, error){
	var checkHistory []string

	queryStmt, errStmt := db.Prepare("SELECT url FROM requestHistory WHERE hashIdentifier=$1")
	if errStmt != nil {
		return checkHistory, errStmt
	}

	rows, _ := queryStmt.Query(hashIdentifier)
	defer rows.Close()

	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return checkHistory, err
		}
		checkHistory = append(checkHistory, url)
	}

	return checkHistory, errStmt
}


// Add a new request history
func AddCheckHistory(db *sql.DB, hashIdentifier string, url string)(error){

	insertStatement := `INSERT INTO requestHistory (hashIdentifier, url) VALUES ($1,$2)`
	_, err := db.Exec(insertStatement, hashIdentifier, url)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "unique_violation"{//If exists, do nothing
				return nil
			}else{
				return err
			}						
		}
		
	}

	return nil
}


// Add or update url cache
func AddUrlCache(db *sql.DB, url string, jsonResponse string)(error){

	insertStatement := `INSERT INTO urlCache (url,jsonResponse,lastTimeChecked) VALUES ($1,$2,NOW())`
	_, err := db.Exec(insertStatement, url, jsonResponse)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "unique_violation"{//If exists, update
				updateStatement := `UPDATE urlCache SET url=$1, jsonResponse=$2, lastTimeChecked=NOW() WHERE url=$1`
				_, errUpdate := db.Exec(updateStatement, url, jsonResponse)
				if errUpdate != nil {
					return errUpdate
				}else{
					return nil
				}				
			}else{				
				return err
			}						
		}		
	}
	
	return nil
}


