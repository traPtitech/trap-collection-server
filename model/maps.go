package model

var osGameTypeIntsMap = map[string][]uint8{
	"windows": {0, 1, 2},
	"mac":     {0, 1, 3},
}
var osGameTypeIntMap = map[string]uint{
	"windows": 2,
	"mac":     3,
}

var gameTypeIntStrMap = map[uint8]string{
	0: "url",
	1: "jar",
	2: "windows",
	3: "mac",
}

var gameTypeStrIntMap = map[string]uint8{
	"url": 0,
	"jar": 1,
	"windows": 2,
	"mac": 3,
}

var extIntStrMap = map[uint8]string{
	0: "jpeg",
	1: "png",
	2: "gif",
	3: "mp4",
}

var questionTypeIntStrMap = map[uint8]string{
	0: "radio",
	1: "checkbox",
	2: "text",
}
