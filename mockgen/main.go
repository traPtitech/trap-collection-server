package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	"gopkg.in/yaml.v2"
)

func main() {
	flag.Parse()
	arg := flag.Arg(0)

	file, err := os.OpenFile(arg, os.O_APPEND|os.O_RDWR, os.ModeAppend)
	if err != nil {
		panic(fmt.Errorf("Open File Error: %w", err))
	}
	defer file.Close()

	var yamlMap map[interface{}]interface{}
	err = yaml.NewDecoder(file).Decode(&yamlMap)
	if err != nil {
		panic(fmt.Errorf("Decode Yaml Error: %w", err))
	}

	yamlMap, err = removeSecurity(yamlMap)
	if err != nil {
		panic(fmt.Errorf("Remove Security Error: %w", err))
	}

	err = yaml.NewEncoder(file).Encode(yamlMap)
	if err != nil {
		panic(fmt.Errorf("Write File Error: %w", err))
	}
}

func removeSecurity(yamlMap map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	delete(yamlMap, "security")
	delete(yamlMap, "securitySchemes")
	for key, value := range yamlMap {
		if v := reflect.ValueOf(value); v.Kind() != reflect.Map {
			continue
		}
		partMap, err := removeSecurity(value.(map[interface{}]interface{}))
		if err != nil {
			return map[interface{}]interface{}{}, err
		}

		yamlMap[key] = partMap
	}

	return yamlMap, nil
}