package main

import "strings"

var hillbillyReplacements = map[string]string{
	"hello":     "howdy",
	"hi":        "well hey there",
	"goodbye":   "y'all come back now",
	"bye":       "y'all come back now",
	"yes":       "darn tootin'",
	"no":        "naw",
	"very":      "plumb",
	"tired":     "tuckered out",
	"angry":     "madder than a wet hen",
	"hungry":    "hungrier than a bear in spring",
	"food":      "vittles",
	"going to":  "fixin' to",
	"about to":  "fixin' to",
	"everyone":  "all y'all",
	"you all":   "y'all",
	"over there": "over yonder",
	"far away":  "a fur piece",
	"nothing":   "nuthin'",
	"something": "somethin'",
	"isn't":     "ain't",
	"aren't":    "ain't",
	"car":       "pickup truck",
	"dog":       "ol' hound dog",
	"fishing":   "fishin'",
	"running":   "runnin'",
	"fixing":    "fixin'",
	"my":        "mah",
	"the":       "th'",
	"them":      "them there",
}

func translateToHillbilly(text string) string {
	result := strings.ToLower(text)
	for eng, hillbilly := range hillbillyReplacements {
		result = strings.ReplaceAll(result, eng, hillbilly)
	}
	return result
}
