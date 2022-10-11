package intesishome

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
)

var (
	commandMap map[string]interface{}
	stateMap   map[string]interface{}
	//go:embed "assets/mappingCommand.json"
	commandMapJSON []byte
	//go:embed "assets/mappingState.json"
	stateMapJSON []byte
)

func init() {
	if err := json.Unmarshal(commandMapJSON, &commandMap); err != nil {
		fmt.Printf("fatal! unable to load in the command map")
		panic(err)
	}
	if err := json.Unmarshal(stateMapJSON, &stateMap); err != nil {
		fmt.Printf("fatal! unable to load in the command map")
		panic(err)
	}
}

func MapCommand(key string, value interface{}) (uid, mValue int, err error) {
	if _, ok := commandMap[key]; !ok {
		err = fmt.Errorf("key not present in command map: %s", key)
		return
	}
	if i, err := strconv.Atoi(key); err == nil {
		// it's an int already
		uid = i
	} else {
		// map the key to the uid
		uid = int(commandMap[key].(map[string]interface{})["uid"].(float64))
	}
	i, err := strconv.Atoi(value.(string))
	if err == nil {
		// it's an int so pass it back
		mValue = i
		return
	}
	// otherwise we have to map it, reset the err
	err = nil
	values := commandMap[key].(map[string]interface{})["values"].(map[string]interface{})
	if _, ok := values[value.(string)]; !ok {
		err = fmt.Errorf("no such value: %v exists for command: %v wanted: %v", value, key, values)
		return
	}
	mValue = int(values[value.(string)].(float64))
	return
}

// for a given uid, decode it's real name as a string
// if we are unable to then return the original uid as a string
func DecodeUid(uid int) string {
	uidS := fmt.Sprint(uid)
	if _, ok := stateMap[uidS]; !ok {
		return uidS
	}
	if _, ok := stateMap[uidS].(map[string]interface{})["name"].(string); !ok {
		return uidS
	}
	return stateMap[uidS].(map[string]interface{})["name"].(string)
}

// returns a string representation of the value, the original value if it cannot be mapped or nil
func DecodeState(name string, value int) interface{} {
	for k := range stateMap {
		if stateMap[k].(map[string]interface{})["name"].(string) == name {
			values, ok := stateMap[k].(map[string]interface{})["values"].(map[string]interface{})
			if !ok {
				// there's no human mapping for the value
				return value
			}
			if _, ok := values[fmt.Sprint(value)]; ok {
				return values[fmt.Sprint(value)]
			}
			// there's no mapping for this value, return nil
			return nil
		}
	}
	// couldn't find it give back the value
	return value
}
