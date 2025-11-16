package main

import (
	"encoding/json"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func dump(i interface{}) string { // nolint
	mi, err := json.MarshalIndent(i, "", "    ")
	checkErr(err)

	return string(mi)
}

func unused(i interface{}) {} // nolint
