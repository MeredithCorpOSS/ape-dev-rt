package git

import (
	"reflect"
	"testing"
	"time"
)

func TestParseCommit(t *testing.T) {
	tz, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Errorf("Failed loading timezone: %#v", err)
	}

	sampleCommits := map[string]*GitCommit{
		SampleCommit: &GitCommit{
			AbbreviatedSHA: "28009e4",
			Message:        "Here's my commit message",
			AuthorName:     "Radek Simko",
			AuthorEmail:    "radek.simko@timeinc.com",
			AuthorshipDate: time.Unix(1432810668, 0).In(tz),
			CommitterName:  "Radek Simko",
			CommitterEmail: "radek.simko@timeinc.com",
			CommitDate:     time.Unix(1432810670, 0).In(tz),
		},
	}

	g := NewGit("", "")
	for raw, expectedCommit := range sampleCommits {
		commit, err := g.parseCommit(raw, tz)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(commit, expectedCommit) {
			t.Errorf("Commit not parsed as expected.\ngiven: %#v\nexpected: %#v",
				commit, expectedCommit)
		}
	}
}

func TestSplitRawCommits(t *testing.T) {
	g := NewGit("", "")
	rawCommits := g.splitRawCommits(MultipleCommits)

	if len(rawCommits) != 2 {
		t.Errorf("Unexpected number of splitted commits.\nGiven: %d\nExpected: %d",
			len(rawCommits), 2)
	} else {
		if rawCommits[0] != FirstSeparatedCommit {
			t.Errorf("First splitted commit doesn't match.\nGiven: %s\nExpected: %s",
				rawCommits[0], FirstSeparatedCommit)
		}
		if rawCommits[1] != SecondSeparatedCommit {
			t.Errorf("First splitted commit doesn't match.\nGiven: %s\nExpected: %s",
				rawCommits[1], SecondSeparatedCommit)
		}
	}
}

var MultipleCommits = `28009e4
test
Radek Simko
radek.simko@timeinc.com
2015-05-28 11:57:48 +0100
Radek Simko
radek.simko@timeinc.com
2015-07-01 12:02:15 +0100

4185807
Add an example of application config
Radek Simko
radek.simko@timeinc.com
2015-05-20 14:42:17 -0700
Radek Simko
radek.simko@timeinc.com
2015-05-28 11:57:48 +0100`

var FirstSeparatedCommit = `28009e4
test
Radek Simko
radek.simko@timeinc.com
2015-05-28 11:57:48 +0100
Radek Simko
radek.simko@timeinc.com
2015-07-01 12:02:15 +0100`

var SecondSeparatedCommit = `4185807
Add an example of application config
Radek Simko
radek.simko@timeinc.com
2015-05-20 14:42:17 -0700
Radek Simko
radek.simko@timeinc.com
2015-05-28 11:57:48 +0100`

var SampleCommit = `28009e4
Here's my commit message
Radek Simko
radek.simko@timeinc.com
2015-05-28 11:57:48 +0100
Radek Simko
radek.simko@timeinc.com
2015-05-28 11:57:50 +0100`
