package main

import "encoding/json"

func getIndentedJSON(v interface{}) []byte {
	b, _ := json.MarshalIndent(v, "", "  ")
	return b
}
