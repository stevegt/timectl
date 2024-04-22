package timectl

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/emersion/go-ical"
	grrule "github.com/graham/rrule"
	. "github.com/stevegt/goadapt"
	"github.com/teambition/rrule-go"
)

func TestRruleMinutely(t *testing.T) {
	// get California time zone
	loc, err := time.LoadLocation("America/Los_Angeles")
	Ck(err)

	// rrule that means "every minute between 05:00 and 21:00 on weekdays"
	// onRule := "FREQ=MINUTELY;INTERVAL=1;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20;BYDAY=MO,TU,WE,TH,FR"

	// rrule that means "every minute between 05:00 and 21:00"
	onRule := "FREQ=MINUTELY;INTERVAL=1;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20"
	onRuleSlice := []string{onRule}

	start, err := time.Parse(time.RFC3339, "2020-03-17T00:01:00-07:00")
	Ck(err)
	onSet, err := Strs2set(onRuleSlice, start, loc)
	Tassert(t, err == nil, "strs2set failed: %v", err)

	// verify the rule
	ruleStr := onSet.String()
	expect := "DTSTART;TZID=America/Los_Angeles:20200317T000100\nRRULE:FREQ=MINUTELY;INTERVAL=1;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20"
	Tassert(t, ruleStr == expect, Spf("ruleStr %s != expect %s", ruleStr, expect))

	// get all occurrences in the first 12 hours
	stop := start.Add(12 * time.Hour)
	occurrences := onSet.Between(start, stop, true)
	// should be (12-5)*60 = 420 minutes, and Between inc is true, so
	// add 2 for 05:00 and 12:01 = 422
	Tassert(t, len(occurrences) == 422, Spf("len(occurrences) %d != 422", len(occurrences)))
	first := occurrences[0]
	firstExpect, err := time.Parse(time.RFC3339, "2020-03-17T05:00:00-07:00")
	Ck(err)
	Tassert(t, first.Equal(firstExpect), Spf("first %s != firstExpect %s", first, firstExpect))
	lastI := len(occurrences) - 1
	last := occurrences[lastI]
	Tassert(t, last.Equal(stop), Spf("last %s != stop %s", last, stop))
}

func Test30Minute(t *testing.T) {
	// get California time zone
	loc, err := time.LoadLocation("America/Los_Angeles")
	Ck(err)
	// 30 minute stir rule
	stirRule := "RRULE:FREQ=MINUTELY;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20;BYMINUTE=0,30"
	stirRuleSlice := []string{stirRule}
	start, err := time.Parse(time.RFC3339, "2020-03-17T00:00:00-07:00")
	Ck(err)
	stirSet, err := Strs2set(stirRuleSlice, start, loc)
	Tassert(t, err == nil, "strs2set failed: %v", err)
	// verify the rule
	ruleStr := stirSet.String()
	expect := "DTSTART;TZID=America/Los_Angeles:20200317T000000\nRRULE:FREQ=MINUTELY;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20;BYMINUTE=0,30"
	Tassert(t, ruleStr == expect, Spf("ruleStr %s != expect %s", ruleStr, expect))
	// get all occurrences in the first 12 hours
	stop := start.Add(12 * time.Hour)
	occurrences := stirSet.Between(start, stop, true)
	// should be (12-5)*2 = 14 hours = 14, and add 1 for 12:00 = 15
	Tassert(t, len(occurrences) == 15, Spf("len(occurrences) %d != 15", len(occurrences)))
	first := occurrences[0]
	firstExpect, err := time.Parse(time.RFC3339, "2020-03-17T05:00:00-07:00")
	Ck(err)
	Tassert(t, first.Equal(firstExpect), Spf("first %s != firstExpect %s", first, firstExpect))
	lastI := len(occurrences) - 1
	last := occurrences[lastI]
	Tassert(t, last.Equal(stop), Spf("last %s != stop %s", last, stop))
}

func XXXTestHourly(t *testing.T) {
	// get California time zone
	loc, err := time.LoadLocation("America/Los_Angeles")
	Ck(err)
	runRuleStr := "RRULE:FREQ=HOURLY;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20;BYMINUTE=0"
	runRuleSlice := []string{
		"DURATION:PT1H",
		runRuleStr,
	}
	start, err := time.Parse(time.RFC3339, "2020-03-17T00:00:00-07:00")
	Ck(err)
	runSet, err := Strs2set(runRuleSlice, start, loc)
	Tassert(t, err == nil, "strs2set failed: %v", err)
	ruleStr := runSet.String()
	expect := "DTSTART;TZID=America/Los_Angeles:20200317T000000\nDURATION:PT1H\nRRULE:FREQ=HOURLY;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20;BYMINUTE=0"
	Tassert(t, ruleStr == expect, Spf("ruleStr %s != expect %s", ruleStr, expect))
}

func XXXTestHourlyGrrule(t *testing.T) {
	// get time in rfc-5545 format
	nowStr := time.Now().Format("20060102T150405")
	runRuleStr := Spf("DTSTART;TZID=America/Los_Angeles:%s\n", nowStr)
	runRuleStr += "DURATION:PT1H"
	runRuleStr += "RRULE:FREQ=HOURLY;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20;BYMINUTE=0;DURATION=PT1H"
	// Pl("runRuleStr is", runRuleStr)
	// parse using grrule
	runRule, err := grrule.Parse(runRuleStr)
	Ck(err)
	// Pl("show run rule")
	// Pl(runRule)
	// show the next 10 occurrences
	iter := runRule.Iterator()
	var event time.Time
	for i := 0; i < 10; i++ {
		iter.Step(&event)
		fmt.Println(event)
	}

	/*
		Pl("generate random times and check if they are during an occurence")
		for i := 0; i < 10; i++ {
			t := time.Now().In(loc).Add(time.Duration(rand.Intn(86400)) * time.Second)
			Pl("random time", t)
			Pl("is during an occurrence?", isDuringOccurrence(t, runSet))
		}
	*/
}

var icalHourly = `BEGIN:VCALENDAR
PRODID:-//example.com//NONSGML Calendar//EN
VERSION:2.0
BEGIN:VEVENT
DTSTAMP;TZID=Local:20200316T201848
DTSTART;TZID=America/Los_Angeles:20200317T000100
DURATION:PT1800S
RRULE:FREQ=HOURLY;INTERVAL=1;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20
SUMMARY:Hourly recurring event
UID:uid@example.org
END:VEVENT
END:VCALENDAR
`

func TestParseIcal(t *testing.T) {
	cal := ParseIcal(icalHourly)
	show_ical(cal)

	/*
		// get occurrences
		start, err := time.Parse(time.RFC3339, "2020-01-01T00:00:00-07:00")
		Ck(err)
		stop, err := time.Parse(time.RFC3339, "2020-12-31T23:59:59-07:00")
		Ck(err)
		occurrences := cal.Between(start, stop, true)

		// check the first occurrence
		first := occurrences[0]
		Pf("first occurrence %#v\n", first)
	*/

}

func XXXTestIcal(t *testing.T) {
	// try_ical()
	cal := gen_ical()
	show_ical(cal)

	cal = ParseIcal(icalHourly)
	show_ical(cal)
}

// vcalendar entry, America/Los_Angeles time zone, 30 minutes duration for each run

/*
var icalHourlyTmpl = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//github.com/emersion/go-ical
BEGIN:VEVENT
DTSTART;TZID=America/Los_Angeles:%s
RRULE:FREQ=HOURLY;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20;BYMINUTE=0;DURATION=PT1H
END:VEVENT
END:VCALENDAR
`

var icalHourlyTmpl2 = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Example Corp.//CalDAV Client//EN
BEGIN:VEVENT
UID:20230101T050000Z-123@example.com
DTSTAMP:%s
DTSTART:%s
DURATION:PT30M
RRULE:FREQ=HOURLY;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20;UNTIL=20230101T200000Z
SUMMARY:Hourly Recurring Event
END:VEVENT
END:VCALENDAR
`

func try_ical() {

	loc, err := time.LoadLocation("America/Los_Angeles")
	Ck(err)
	Pl("loc is", loc)
	// use github.com/emersion/go-ical
	Pl("trying github.com/emersion/go-ical")
	// ical entry with an hourly run rule, with 1-hour duration for each run
	// get time in rfc-5545 format
	nowStr := time.Now().Format("20060102T150405")
	icalEntry := Spf(icalHourlyTmpl2, nowStr)
	Pl("icalEntry is\n", icalEntry)
	// parse using go-ical
	icalReader := strings.NewReader(icalEntry)
	icalDecoder := ical.NewDecoder(icalReader)
	icalObject, err := icalDecoder.Decode()
	Ck(err)
	Pl("show ical object")
	Pf("%#v\n", icalObject)
	Pl("show the next 10 occurrences")
	for _, event := range icalObject.Events() {
		Pf("event %#v\n", event)
		summary, err := event.Props.Text(ical.PropSummary)
		Ck(err)
		Pf("summary %v\n", summary)
		// Pf("start prop %v\n", event.Props.Get(ical.PropDtStart))
		start, err := event.DateTimeStart(loc)
		Ck(err)
		Pf("start %v\n", start)
		end, err := event.DateTimeEnd(loc)
		Ck(err)
		Pf("end %v\n", end)
	}
}
*/

/*
// isDuringOccurrence returns true if the given time is during an occurrence of the given rule set
func isDuringOccurrence(t time.Time, set *rrule.Set) bool {
	occurrences := set.Between(t, t, true)
	return len(occurrences) > 0
}
*/

func gen_ical() (cal *ical.Calendar) {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		fmt.Println(err)
		return
	}

	event := ical.NewEvent()
	event.Props.SetText(ical.PropUID, "uid@example.org")
	event.Props.SetText(ical.PropSummary, "Hourly recurring event")
	event.Props.SetDateTime(ical.PropDateTimeStamp, time.Now())
	// start at the top of the current hour
	start := time.Now().Truncate(time.Hour).In(loc)
	event.Props.SetDateTime(ical.PropDateTimeStart, start)

	// create a 30-minute duration prop
	duration := ical.NewProp(ical.PropDuration)
	duration.SetDuration(30 * time.Minute)
	event.Props.Set(duration)

	// create a recurrence rule
	roption := &rrule.ROption{
		Dtstart:  start,
		Freq:     rrule.HOURLY,
		Interval: 1,
		Byhour:   []int{5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
	}

	// When we call event.Props.SetRecurrenceRule(roption), that calls
	// call rule.RRuleString(), which generates a recurrence rule
	// string without a DTSTART property. We need to remember to
	// reference the event's DTSTART property when iterating over
	// occurrences later, e.g. in show_ical().  The comment above
	// ical.Props.RecurrenceRule() has more details.
	event.Props.SetRecurrenceRule(roption)

	/*
		// alternative which doesn't work: .String() includes the DTSTART
		// property on a separate line, which is not a valid recurrence
		// rule string.
		rprop := ical.NewProp(ical.PropRecurrenceRule)
		rprop.SetValueType(ical.ValueRecurrence)
		rprop.Value = roption.String()
		event.Props.Set(rprop)
		Pl("rprop is", rprop)
	*/

	cal = ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//example.com//NONSGML Calendar//EN")
	cal.Children = append(cal.Children, event.Component)

	// encode to iCalendar (ics) format
	Pl("show ical")
	enc := ical.NewEncoder(os.Stdout)
	err = enc.Encode(cal)
	Ck(err)

	return

}

func show_ical(cal *ical.Calendar) {
	loc, err := time.LoadLocation("America/Los_Angeles")
	Ck(err)
	for _, event := range cal.Events() {
		summary, err := event.Props.Text(ical.PropSummary)
		Ck(err)
		Pf("summary %v\n", summary)
		dtstart, err := event.Props.DateTime(ical.PropDateTimeStart, loc)
		Ck(err)
		Pf("dtstart %v\n", dtstart)
		duration, err := event.Props.Get(ical.PropDuration).Duration()
		Ck(err)
		Pf("duration %v\n", duration)

		rruleProp := event.Props.Get(ical.PropRecurrenceRule)
		if rruleProp == nil {
			Pl("no rrule prop")
			continue
		} else {

			// We need to remember to reference the event's DTSTART
			// property when iterating over occurrences.  The comment
			// above ical.Props.RecurrenceRule() shows how to do this.
			roption, err := event.Props.RecurrenceRule()
			Ck(err)
			Pl("roption is", roption)
			// Assert(false, "pausing here")
			roption.Dtstart = dtstart
			rule, err := rrule.NewRRule(*roption)
			Ck(err)

			Pl("show the occurrences in the first 12 hours")
			stop := dtstart.Add(12 * time.Hour)
			occurrences := rule.Between(dtstart, stop, true)
			Pl("time now is", time.Now().In(loc))
			for _, occurrence := range occurrences {
				Pf("%s  ", occurrence.In(loc))
				now := time.Now().In(loc)
				if now.After(occurrence.In(loc)) && now.Before(occurrence.In(loc).Add(duration)) {
					Pf("running!")
				} else {
					Pf("not running")
				}
				Pl()
			}

			/*
				dtstartStr := event.Props.Get(ical.PropDateTimeStart).Value
				occurrences := rule.Between(time.Now(), time.Now().AddDate(0, 0, 10), true)
				for _, occurrence := range occurrences {
					Pf("event %#v\n", event)
					summary, err := event.Props.Text(ical.PropSummary)
					Ck(err)
					Pf("summary %v\n", summary)
					Pf("start %v\n", occurrence)
					end := occurrence.Add(ical.Props.Get(ical.PropDuration).Value.(time.Duration))
					Pf("end %v\n", end)
				}
			*/
		}
	}
}
