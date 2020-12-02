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

var extIntStrMap = map[uint8]string{
	0: "jpg",
	1: "png",
	2: "gif",
	3: "mp4",
}

var extStrIntMap = map[string]uint8{
	"jpg": 0,
	"png": 1,
	"gif": 2,
	"mp4": 3,
}

var roleStrIntMap = map[string]uint8{
	"image": 0,
	"video": 1,
}

var questionTypeIntStrMap = map[uint8]string{
	0: "radio",
	1: "checkbox",
	2: "text",
}
