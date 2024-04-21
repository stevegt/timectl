package timectl

import (
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/stevegt/goadapt"

	grrule "github.com/graham/rrule"
	"github.com/teambition/rrule-go"

	"github.com/emersion/go-ical"
)

// strs2set converts a string slice to an RRuleSet for the given time zone
func strs2set(strs []string, loc *time.Location) (set *rrule.Set, err error) {
	defer Return(&err)
	// prefix each rule with "RRULE:"
	for i, str := range strs {
		strs[i] = "RRULE:" + str
	}
	// parse the rules into a set
	set, err = rrule.StrSliceToRRuleSet(strs)
	Ck(err)
	// set time zone
	set.DTStart(time.Now().In(loc))
	return
}

func minutely() {
	// get California time zone
	loc, err := time.LoadLocation("America/Los_Angeles")
	Ck(err)

	// rrule that means "every minute between 05:00 and 21:00 on weekdays"
	// onRule := "FREQ=MINUTELY;INTERVAL=1;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20;BYDAY=MO,TU,WE,TH,FR"

	// rrule that means "every minute between 05:00 and 21:00"
	onRule := "FREQ=MINUTELY;INTERVAL=1;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20"

	// rrule with a DTSTART of California time zone
	// onRule := "DTSTART;TZID=America/Los_Angeles:20180101T050000\nRRULE:FREQ=MINUTELY;INTERVAL=1;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20"

	// convert rrule string to rrule struct
	// on, err := rrule.StrToRRule(onRule)
	// Ck(err)

	// set DTSTART to now in California time zone
	// on.Dtstart = time.Now().In(loc)

	// Pl(on)

	onRuleSlice := []string{onRule}

	/*
		// parse the rule into a set
		// onSet, err := rrule.StrSliceToRRuleSetInLoc(onRuleSlice, loc)
		onSet, err := rrule.StrSliceToRRuleSet(onRuleSlice)
		// on, err := rrule.StrToROptionInLocation(onRule, loc)
		Ck(err)

		// set DTSTART to now in California time zone
		onSet.DTStart(time.Now().In(loc))
	*/

	onSet, err := strs2set(onRuleSlice, loc)
	Ck(err)

	Pl("show every-minute rule")
	Pl(onSet)

	// get all occurrences in the next 12 hours
	occurrences := onSet.Between(time.Now(), time.Now().Add(12*time.Hour), true)

	// print them
	Pl("show occurrences")
	for _, occurrence := range occurrences {
		// fmt.Println(occurrence.In(loc))
		fmt.Println(occurrence)
	}
}

func thirty_stir() {
	// get California time zone
	loc, err := time.LoadLocation("America/Los_Angeles")
	Ck(err)
	// 30 minute stir rule
	stirRule := "RRULE:FREQ=MINUTELY;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20;BYMINUTE=0,30"
	stirRuleSlice := []string{stirRule}
	stirSet, err := rrule.StrSliceToRRuleSet(stirRuleSlice)
	Ck(err)
	stirSet.DTStart(time.Now().In(loc))
	Pl("show 30 minute stir rule")
	Pl(stirSet)
	occurrences := stirSet.Between(time.Now(), time.Now().Add(12*time.Hour), true)
	Pl("time now is", time.Now().In(loc))
	Pl("show occurrences")
	for _, occurrence := range occurrences {
		Pl(occurrence.In(loc))
		if occurrence.After(time.Now()) && occurrence.Before(time.Now().Add(25*time.Minute)) {
			fmt.Println("STIR NOW!")
			break
		}
	}
}

func hourly_run() {
	// hourly run rule, with 1-hour duration for each run
	Pl("trying https://github.com/graham/rrule")
	// get time in rfc-5545 format
	nowStr := time.Now().Format("20060102T150405")
	runRuleStr := Spf("DTSTART;TZID=America/Los_Angeles:%s\n", nowStr)
	runRuleStr += "RRULE:FREQ=HOURLY;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20;BYMINUTE=0;DURATION=PT1H"
	Pl("runRuleStr is", runRuleStr)
	// parse using grrule
	runRule, err := grrule.Parse(runRuleStr)
	Ck(err)
	Pl("show run rule")
	Pl(runRule)
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

var icalHourly = `BEGIN:VCALENDAR
PRODID:-//example.com//NONSGML Calendar//EN
VERSION:2.0
BEGIN:VEVENT
DTSTAMP;TZID=Local:20230508T201848
DTSTART;TZID=America/Los_Angeles:20230508T200000
DURATION:PT1800S
RRULE:FREQ=HOURLY;INTERVAL=1;BYHOUR=5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20
SUMMARY:Hourly recurring event
UID:uid@example.org
END:VEVENT
END:VCALENDAR
`

func oldMain() {
	// try_ical()
	cal := gen_ical()
	show_ical(cal)

	cal = parse_ical(icalHourly)
	show_ical(cal)
}

func parse_ical(txt string) *ical.Calendar {
	Pl("parse_ical")
	r := strings.NewReader(txt)
	cal, err := ical.NewDecoder(r).Decode()
	Ck(err)
	return cal
}

// vcalendar entry, America/Los_Angeles time zone, 30 minutes duration for each run

/*
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
			roption.Dtstart = dtstart
			rule, err := rrule.NewRRule(*roption)
			Ck(err)

			Pl("show the occurrences in the next 12 hours")
			occurrences := rule.Between(dtstart, time.Now().Add(12*time.Hour), true)
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
