package util

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"time"
)

var (
	logChannel = make(chan LogEntry, 100)
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
	Endpoint  string    `json:"endpoint"`
	FriendID  string    `json:"friend_id"`
}

func EnqueueRequestLog(entry LogEntry) {
	logChannel <- entry
}

func GetLogsForDate(date time.Time) (logEntries []LogEntry, err error) {
	filePath := path.Join("temp", "logs", fmt.Sprintf("%s.csv", date.Format("01-02-06")))
	if _, err = os.Stat(filePath); err != nil {
		return []LogEntry{}, nil
	}
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening log file: %s", err)
		return
	}

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		log.Printf("Reading log file: %s", err)
		return
	}

	for _, record := range records {
		timeVal, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			log.Println(err)
			continue
		}
		timestamp := time.Unix(timeVal/1000, 0)
		logEntries = append(logEntries, LogEntry{
			Timestamp: timestamp,
			UserID:    record[1],
			Endpoint:  record[2],
			FriendID:  record[3],
		})
	}

	return
}

func WriteChannelLogsToFile() {
	var prevTimestamp *time.Time
	dayEntries := []LogEntry{}

	ok := true
	var entry LogEntry
	for ok {
		select {
		case entry, ok = <-logChannel:
			if ok {
				if prevTimestamp == nil || datesAreEqual(entry.Timestamp, *prevTimestamp) {
					dayEntries = append(dayEntries, entry)
				} else {
					WriteLogsToFile(*prevTimestamp, dayEntries)
					dayEntries = []LogEntry{}
				}
				prevTimestamp = &entry.Timestamp
			}
		default:
			ok = false
		}
		if !ok && len(dayEntries) > 0 {
			WriteLogsToFile(*prevTimestamp, dayEntries)
		}
	}
}

func datesAreEqual(t1 time.Time, t2 time.Time) bool {
	d1, m1, y1 := t1.Date()
	d2, m2, y2 := t2.Date()
	return d1 == d2 && m1 == m2 && y1 == y2
}

func WriteLogsToFile(date time.Time, entries []LogEntry) {
	filePath := path.Join("temp", "logs", fmt.Sprintf("%s.csv", date.Format("01-02-06")))
	if _, err := os.Stat(filePath); err != nil {
		os.MkdirAll(path.Join("temp", "logs"), 0755)
	}
	csvFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("failed creating file: %s", err)
		return
	}
	defer csvFile.Close()

	for _, entry := range entries {
		csvFile.WriteString(fmt.Sprintf("%d,%s,%s,%s\n", entry.Timestamp.UnixMilli(), entry.UserID, entry.Endpoint, entry.FriendID))
	}
}

func GetLogChannelSize() int {
	return len(logChannel)
}
