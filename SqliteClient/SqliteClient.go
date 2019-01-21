package SqliteClient

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type SQLCli struct {
	DB *sql.DB
}

func InitDB(filename string) *SQLCli {
	var err error
	sqlcli := SQLCli{}
	sqlcli.DB, err = sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatalf("Couldnt open DB file:%v", err)
	}
	if sqlcli.DB == nil {
		log.Fatalf("DB not a db?")
	}

	sqlWatchTable := `
	CREATE TABLE IF NOT EXISTS ChannelWatch(
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		ChannelLink TEXT,
		ChanID TEXT,
		LastVideo TEXT,
		LastUpdate TEXT
		);
		`
	_, err = sqlcli.DB.Exec(sqlWatchTable)
	if err != nil {
		log.Fatalf("failed to create ChannelWatch table:%v", err)
	}
	sqlLinkTable := `
	CREATE TABLE IF NOT EXISTS QuickLinks(
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		LinkID TEXT,
		Link TEXT
		);
		`
	_, err = sqlcli.DB.Exec(sqlLinkTable)
	if err != nil {
		log.Fatalf("Failed to create QuickLinks table:%v", err)
	}
	botMods := `
	CREATE TABLE IF NOT EXISTS BotMods(
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		UserID TEXT);`
	_, err = sqlcli.DB.Exec(botMods)
	if err != nil {
		log.Fatalf("Failed to create BotMods table:%v", err)
	}
	return &sqlcli
}

func (db *SQLCli) IsAuthorized(userID string) bool {
	row, err := db.DB.Query("SELECT ID from BotMods where UserID=?;", userID)
	defer row.Close()
	for row.Next() {
		var response string
		err = row.Scan(&response)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func (db *SQLCli) AddMod(userID string) error {
	sqlAdditem := `
	INSERT OR REPLACE INTO BotMods(
		UserID
		) values(?);
		`
	stmt, err := db.DB.Prepare(sqlAdditem)
	if err != nil {
		return errors.New("Could not prep add item statement: " + err.Error())
	}
	row, err := db.DB.Query("SELECT * from BotMods where UserID=?;", userID)
	if (row.Next()) == true {
		//record already exists, skip
		return errors.New("Mod already exists")
	}
	_, err = stmt.Exec(userID)
	if err != nil {
		return errors.New("Could not add quick link: " + err.Error())
	}
	return nil
}

func (db *SQLCli) RemoveMod(userID string) error {
	sqlRemoveItem := `
	DELETE FROM BotMods WHERE UserID =?;
		`
	stmt, err := db.DB.Prepare(sqlRemoveItem)
	if err != nil {
		return errors.New("Could not prep remove item statement: " + err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(userID)
	if err != nil {
		return errors.New("Could not remove item: " + err.Error())
	}
	return nil
}

func (db *SQLCli) AddLink(linkID string, link string) error {
	sqlAdditem := `
	INSERT OR REPLACE INTO QuickLinks(
		LinkID,
		Link
		) values(?,?);
		`
	stmt, err := db.DB.Prepare(sqlAdditem)
	if err != nil {
		return errors.New("Could not prep add item statement: " + err.Error())
	}
	row, err := db.DB.Query("SELECT * from QuickLinks where LinkID=? OR link=?;", linkID, link)
	if (row.Next()) == true {
		//record already exists, skip
		return errors.New("link already exists")
	}
	_, err = stmt.Exec(linkID, link)
	if err != nil {
		return errors.New("Could not add quick link: " + err.Error())
	}
	return nil
}

func (db *SQLCli) FetchLink(linkID string) (string, error) {

	row, err := db.DB.Query("SELECT Link from QuickLinks where LinkID=?;", linkID)
	defer row.Close()
	for row.Next() {
		var response string
		err = row.Scan(&response)
		if err != nil {
			return "", errors.New("Could not retrieve link: " + err.Error())
		}
		return response, nil
	}
	//shouldnt get here
	return "", nil
}

func (db *SQLCli) RemoveLink(linkID string) error {
	sqlRemoveItem := `
	DELETE FROM QuickLinks WHERE LinkID =?;
		`
	stmt, err := db.DB.Prepare(sqlRemoveItem)
	if err != nil {
		return errors.New("Could not prep remove item statement: " + err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(linkID)
	if err != nil {
		return errors.New("Could not remove item: " + err.Error())
	}
	return nil
}

func (db *SQLCli) AddChannels(idMap map[string]string) error {
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
	for channel, chanid := range idMap {
		row, err := db.DB.Query("SELECT * from ChannelWatch where ChannelLink=?;", channel)
		if (row.Next()) == true {
			//record already exists, skip
			return errors.New("Channel already exists")
		}
		_, err2 := stmt.Exec(channel, chanid, "", "")
		if err2 != nil {
			return errors.New("Failed to insert channel: " + err.Error())
		}
	}
	return nil
}

func (db *SQLCli) RemoveChannels(idMap map[string]string) error {
	sqlRemoveItem := `
	DELETE FROM ChannelWatch WHERE ChanID =?;
		`
	stmt, err := db.DB.Prepare(sqlRemoveItem)
	if err != nil {
		return errors.New("Could not prep remove item statement: " + err.Error())
	}
	defer stmt.Close()
	for _, chanid := range idMap {
		res, err := stmt.Exec(chanid)
		if err != nil {
			return errors.New("Failed to remove channel: " + err.Error())
		}
		fmt.Println(res)
	}
	return nil
}

func (db *SQLCli) UpdateVideo(chanID string, lastvideo string, time string) bool {
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
			//we havnt actually found a new video, fuck youtube
			return false
		}
	}
	sqlUpdate, err := db.DB.Prepare("update ChannelWatch set LastVideo=?, LastUpdate=? where ChanID=?;")
	if err != nil {
		log.Fatalf("Failed to update video:%v", err)
	}
	_, err = sqlUpdate.Exec(lastvideo, time, chanID)
	if err != nil {
		log.Fatalf("Failed to execute update statement:%v", err)
	}
	return true
}

func (db *SQLCli) UpdateChanID(channelLink string, channelID string) error {
	stmt, err := db.DB.Prepare("update ChannelWatch set ChanID=? where ChannelLink = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(channelID, channelLink)
	if err != nil {
		return err
	}
	return nil
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
