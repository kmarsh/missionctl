package match

import "testing"

var projects = []Project{
	{"1", "Acme", true},
	{"2", "Acme Redesign", true},
	{"3", "Globex Portal", true},
	{"4", "Globex Portal (2019)", false},
	{"5", "Initech App", true},
	{"6", "Hooli App", true},
	{"7", "Old Site", false},
}

func TestFind(t *testing.T) {
	for query, want := range map[string]string{
		"acme": "1", "GLOBEX PORTAL (2019)": "4", "redesign": "2", "globex": "3", "old": "7",
	} {
		if got, err := Find(query, projects); err != nil || got.ID != want {
			t.Errorf("Find(%q) = %v, %v; want id %s", query, got, err, want)
		}
	}
	for _, query := range []string{"app", "nope"} {
		if got, err := Find(query, projects); err == nil {
			t.Errorf("Find(%q) = %v; want an error", query, got)
		}
	}
}

func TestIsID(t *testing.T) {
	if !IsID("556d297a-a1ed-465d-a4b2-6d4945a0a538") || IsID("Acme") {
		t.Error("IsID misjudged an id or a name")
	}
}
