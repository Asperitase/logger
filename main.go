package main

import (
	logs "logger/log"
	"os"
)

func main() {
	console := logs.New(logs.Info, true, os.Stdout)

	console.Log(logs.Info, "Logging an integer value: {0}\n", 100)
	console.Log(logs.Info, "Logging a string message: {0}\n", "example string")
	console.Log(logs.Info, "Logging a hexadecimal value: {0}\n", 0x10A)
	console.Log(logs.Info, "Logging a number with leading zeros: {0:05}\n", 100)
	console.Log(logs.Warning, "Logging a float rounded to two decimal places: {0:.2f}\n", 3.14156)
	console.Log(logs.Error, "Logging multiple values: {0:.2f}, {1:05}, {2}, {3}\n", 3.14156, 100, 0x10A, "example string")
}
