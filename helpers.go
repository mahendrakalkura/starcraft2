package main

import (
	"encoding/json"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func dump(i any) string { // nolint
	mi, err := json.MarshalIndent(i, "", "    ")
	check(err)

	return string(mi)
}

func unused(i any) {} // nolint
