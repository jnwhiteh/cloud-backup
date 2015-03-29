package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	outFilename = flag.String("o", "onedrive_types.go", "The output filename")
	outPackage  = flag.String("p", "main", "The output package name")
	resourceDir = flag.String("i", "resources", "The resources directory")
)

// it is entirely like that I am going to hell
var resourcePattern = regexp.MustCompile("\\Q<!-- { \"blockType\": \"resource\", \"@odata.type\": \"\\E([^\\\"]+?)\"(?s:.+?)```json\n({(?s:.*?))```")

func main() {
	flag.Parse()
	outFile, err := os.Create(*outFilename)
	if err != nil {
		log.Fatalf("Failed when opening output file %s: %s", *outFilename, err)
	}

	resources, err := filepath.Glob(fmt.Sprintf("%s/*.md", *resourceDir))
	if err != nil {
		log.Fatalf("Failed to fetch resources: %s", err)
	}

	fmt.Fprintln(outFile, "package main\n\nimport \"time\"\n\n")

	for _, filename := range resources {
		contents, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("Failed when reading %s: %s", filename, err)
		}

		matches := resourcePattern.FindSubmatch(contents)
		if len(matches) != 3 {
			log.Fatalf("%s: Incorrect number of matches expected %d, got %d", filename, 3, len(matches))
		}

		body := matches[2]
		jsonMap := make(map[string]interface{})
		err = json.Unmarshal(body, &jsonMap)
		if err != nil {
			log.Fatalf("Failed to unmarshal JSON: %s", err)
		}

		typeName := stripTypePrefix(string(matches[1]))

		// strip oneDrive. prefix off"
		fmt.Fprintf(outFile, "type %s struct {\n", strings.Title(string(typeName)))
		for key, typeValue := range jsonMap {
			typeName, ok, comment := discoverType(typeValue)
			if !ok {
				log.Printf("Skipping unknown type %s: %T", key, typeValue)
			} else if comment != "" {
				fmt.Fprintf(outFile, "\t%s %s `json:\"%s\"` // %s\n", transformKey(key), typeName, key, comment)
			} else {
				fmt.Fprintf(outFile, "\t%s %s `json:\"%s\"`\n", transformKey(key), typeName, key)
			}
		}

		fmt.Fprintln(outFile, "}\n")
	}
}

func discoverType(elem interface{}) (string, bool, string) {
	switch elem := elem.(type) {
	case string:
		if len(strings.Split(elem, " | "))+len(strings.Split(elem, "|")) > 2 {
			// string type with specific options
			return "string", true, elem
		} else if elem == "optional string" {
			return "string", true, "optional"
		} else if elem == "string (identifier)" {
			return "string", true, "identifier"
		} else if elem == "string (timestamp)" {
			return "time.Time", true, "string timestamp"
		} else if elem == "string (etag)" {
			return "string", true, "etag"
		} else if elem == "string (path)" {
			return "string", true, "path"
		} else if elem == "string (hex)" {
			return "string", true, "hex"
		} else if elem == "url" {
			return "string", true, "url"
		} else if elem == "timestamp" {
			return "time.Time", true, "timestamp"
		}
		return elem, true, ""
	case float64:
		return "float64", true, ""
	case bool:
		return "bool", true, ""
	case []interface{}:
		// take the first element and take that type
		typeName, ok, comment := discoverType(elem[0])
		if ok {
			typeName = fmt.Sprintf("[]%s", typeName)
		}
		return typeName, ok, comment
	case map[string]interface{}:
		if typeName, ok := elem["@odata.type"].(string); ok {
			typeName := "*" + strings.Title(stripTypePrefix(typeName))
			return typeName, true, ""
		}
	}

	return "", false, ""
}

func transformKey(key string) string {
	if strings.HasPrefix(key, "@") {
		return "Instance" + strings.Replace(key[1:], ".", "_", -1)
	} else {
		return strings.Title(key)
	}
}

func typeLine(keyName, typeName string) string {
	return fmt.Sprintf("\t%s %s `json:\"%s\"`", strings.Title(keyName), typeName, keyName)
}

func stripTypePrefix(typeName string) string {
	prefix := "oneDrive."
	return typeName[len(prefix):]
}

func getIndentedJSON(v interface{}) []byte {
	b, _ := json.MarshalIndent(v, "", "  ")
	return b
}
