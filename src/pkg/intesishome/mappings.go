package intesishome

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

var (
	_commandMap map[string]interface{}
	_stateMap   map[string]interface{}
	//go:embed "assets/mappingCommand.json"
	_commandMapJSON []byte
	//go:embed "assets/mappingState.json"
	_stateMapJSON []byte
)

func init() {
	if err := json.Unmarshal(_commandMapJSON, &_commandMap); err != nil {
		fmt.Printf("fatal! unable to load in the command map")
		panic(err)
	}
	if err := json.Unmarshal(_stateMapJSON, &_stateMap); err != nil {
		fmt.Printf("fatal! unable to load in the command map")
		panic(err)
	}
}

// TODO: add tests for various unhandled types
func MapCommand(key string, value interface{}) (uid, mValue int, err error) {
	if _, ok := _commandMap[key]; !ok {
		err = fmt.Errorf("key not present in command map: %s", key)
		return
	}
	if i, err := strconv.Atoi(key); err == nil {
		// it's an int already
		uid = i
	} else {
		// map the key to the uid
		uid = int(_commandMap[key].(map[string]interface{})["uid"].(float64))
	}
	// determine what the underlying type for the interface is
	switch value.(type) {
	case float64:
		mValue = int(value.(float64))
		return
	case int:
		mValue = value.(int)
		return
	case string:
		var i int
		i, err = strconv.Atoi(value.(string))
		if err == nil {
			// it's an int so pass it back
			mValue = i
			return
		}
	default:
		err = fmt.Errorf("bad conversion type, expected float / int / string got: %s", reflect.TypeOf(value))
		return
	}
	// otherwise we have to map it, reset the err
	err = nil
	values := _commandMap[key].(map[string]interface{})["values"].(map[string]interface{})
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
	if _, ok := _stateMap[uidS]; !ok {
		return uidS
	}
	if _, ok := _stateMap[uidS].(map[string]interface{})["name"].(string); !ok {
		return uidS
	}
	return _stateMap[uidS].(map[string]interface{})["name"].(string)
}

// returns a string representation of the value, the original value if it cannot be mapped or nil
func DecodeState(name string, value int) interface{} {
	for k := range _stateMap {
		if _stateMap[k].(map[string]interface{})["name"].(string) == name {
			values, ok := _stateMap[k].(map[string]interface{})["values"].(map[string]interface{})
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
