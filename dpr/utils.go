package dpr

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func writeTags(tags map[string]string, w http.ResponseWriter) {
	if e := json.NewEncoder(w).Encode(tags); e != nil {
		log.Println("ERROR: " + e.Error())
	}
}
