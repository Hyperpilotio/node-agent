package cgroupfs

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

func parseEntry(line string) (name string, value uint64, err error) {
	fields := strings.Fields(line)
	if len(fields) != 2 {
		return name, value, fmt.Errorf("Invalid format: %s", line)
	}

	value, err = strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return name, value, err
	}

	return fields[0], value, nil
}

func parseIntValue(file string) (uint64, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(strings.TrimSpace(string(raw)), 10, 64)

}

func parseStrValue(file string) (string, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(raw)), nil

}
