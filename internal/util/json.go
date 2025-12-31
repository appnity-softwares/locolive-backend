package util

import (
	jsoniter "github.com/json-iterator/go"
)

// JSON is a faster drop-in replacement for encoding/json
var JSON = jsoniter.ConfigCompatibleWithStandardLibrary
