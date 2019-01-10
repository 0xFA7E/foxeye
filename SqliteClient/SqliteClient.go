package SqliteClient

import (
	"database/sql"
	"errors"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type SQLCli struct {
	DB *sql.DB
}

func (db *SQLCli) InitDB(filepath string) {
	var err error
	if filepath == "" {
		log.Fatalf("Database filename not set")
	}
	db.DB, err = sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatalf("Couldnt open DB file:%v", err)
	}
	if db.DB == nil {
		log.Fatalf("DB not a db?")
	}
}

func (db *SQLCli) CreateTable() {
	//create a table if it doesnt exist
	sql_table := `
	CREATE TABLE IF NOT EXISTS ChannelWatch(
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		ChannelLink TEXT,
		ChanID TEXT,
		LastVideo TEXT,
		LastUpdate TEXT
		);
		`
	_, err := db.DB.Exec(sql_table)
	if err != nil {
		log.Fatalf("failed to create ChannelWatch table:%v", err)
	}
}

func (db *SQLCli) AddChannel(channel string, chanID string) error {
	sqlAdditem := `
	INSERT OR REPLACE INTO ChannelWatch(
		ChannelLink,
		ChanID,
		LastVideo,
		LastUpdate
		) values(?,?,?,?);
		`

	stmt, err := db.DB.Prepare(sqlAdditem)
	if err != nil {
		return errors.New("Could not prep add item statement: " + err.Error())
	}
	defer stmt.Close()
	row, err := db.DB.Query("SELECT * from ChannelWatch where ChannelLink=?;", channel)
	if (row.Next()) == true {
		//record already exists, skip
		return nil
	}
	_, err2 := stmt.Exec(channel, chanID, "", "")
	if err2 != nil {
		return errors.New("Failed to insert channel: " + err.Error())
	}
	return nil
}

func (db *SQLCli) UpdateVideo(chanID string, lastvideo string, time string) bool {
	//fmt.Printf("I got called with: \n%v\n%v\n%v\n%v\n", db, chanID, lastvideo, time)
	sqlLastVid, err := db.DB.Query("Select LastVideo from ChannelWatch where ChanID=?;", chanID)
	if err != nil {
		log.Fatalf("Failed to query DB:%v", err)
	}
	defer sqlLastVid.Close()
	for sqlLastVid.Next() {
		var response string
		err = sqlLastVid.Scan(&response)
		if err != nil {
			log.Fatalf("Could not retrieve LastVideo to check recent:%v", err)
		}
		if response == lastvideo {
			//fmt.Printf("No new video\n") //I LIE
			//we havnt actually found a new video, fuck youtube
			return false
		}
	}

	//fmt.Println("Attempting to update sql")
	sqlUpdate, err := db.DB.Prepare("update ChannelWatch set LastVideo=?, LastUpdate=? where ChanID=?;")
	if err != nil {
		log.Fatalf("Failed to update video:%v", err)
	}
	//fmt.Println("Attempting to execute sql with: %v,%v,%v", lastvideo, time, chanID)
	_, err = sqlUpdate.Exec(lastvideo, time, chanID)
	if err != nil {
		log.Fatalf("Failed to execute update statement:%v", err)
	}
	return true
}

func (db *SQLCli) UpdateChanID(channelLink string, channelID string) {
	stmt, err := db.DB.Prepare("update ChannelWatch set ChanID=? where ChannelLink = ?")
	if err != nil {
		log.Fatalf("Failed to update channel ID:%v", err)
	}
	_, err = stmt.Exec(channelID, channelLink)
	if err != nil {
		log.Fatalf("Failed to commit update:%v", err)
	}
}

func (db *SQLCli) RecentVidFromURL(channel string) (lastvid string, updatetime string) {
	sqlCheck := `
	SELECT LastVideo, LastUpdate from ChannelWatch
	WHERE ChanID = ?
	`
	rows, err := db.DB.Query(sqlCheck, channel)
	if err != nil {
		log.Fatalf("Failed to get last video:%v", err)
	}
	defer rows.Close()

	for rows.Next() {
		err2 := rows.Scan(&lastvid, &updatetime)
		if err2 != nil {
			log.Fatalf("Failed to scan last vid info:%v", err)
		}
		return
	}
	return
}

func (db *SQLCli) WatchList() []string {
	sqlReadAll := `
	SELECT ChanID from ChannelWatch
	ORDER BY ID DESC
	`
	rows, err := db.DB.Query(sqlReadAll)
	if err != nil {
		log.Fatalf("Could not read watch channels:%v", err)
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var res string
		err2 := rows.Scan(&res)
		if err2 != nil {
			log.Fatalf("Failed to scan row result:%v", err)
		}
		results = append(results, res)
	}
	return results
}
