package timectl

import (
	"strings"
	"time"

	. "github.com/stevegt/goadapt"

	"github.com/teambition/rrule-go"

	"github.com/emersion/go-ical"
)

// Strs2set converts a string slice to an RRuleSet for the given time zone
func Strs2set(strs []string, t time.Time, loc *time.Location) (set *rrule.Set, err error) {
	defer Return(&err)
	// prefix each rule with "RRULE:"
	for i, str := range strs {
		strs[i] = "RRULE:" + str
	}
	// parse the rules into a set
	set, err = rrule.StrSliceToRRuleSet(strs)
	Ck(err)
	// set time and time zone
	set.DTStart(t.In(loc))
	return
}

func ParseIcal(txt string) *ical.Calendar {
	r := strings.NewReader(txt)
	cal, err := ical.NewDecoder(r).Decode()
	Ck(err)
	return cal
}
