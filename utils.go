package l2gp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type entry struct {
	Message  string   `json:"message"`
	Severity severity `json:"severity,omitempty"`
}

type severity string

const (
	severityInfo  severity = "INFO"
	severityError severity = "ERROR"
)

func outputErrorLog(w http.ResponseWriter, statusCode int, format string, v ...interface{}) {
	e := entry{
		Severity: severityError,
		Message:  fmt.Sprintf(format, v...),
	}
	out, err := json.Marshal(e)
	if err != nil {
		log.Printf("json.Marshal: %v", err)
	}
	fmt.Println(string(out))
	w.WriteHeader(statusCode)
	w.Write(out)
}

func outputLog(format string, v ...interface{}) {
	e := entry{
		Severity: severityInfo,
		Message:  fmt.Sprintf(format, v...),
	}
	out, err := json.Marshal(e)
	if err != nil {
		log.Printf("json.Marshal: %v", err)
	}
	fmt.Println(string(out))
}

func mustGetEnv(envName string) string {
	env := os.Getenv(envName)
	if env == "" {
		log.Fatal("")
	}
	return env
}
