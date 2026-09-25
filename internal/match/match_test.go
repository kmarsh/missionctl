package match

import "testing"

var projects = []Project{
	{"1", "Rev", true},
	{"2", "Rev Refresh", true},
	{"3", "Vintage Aerial", true},
	{"4", "Vintage Aerial (Ken)", false},
	{"5", "Conference App", true},
	{"6", "HSOP App", true},
	{"7", "Old Site", false},
}

func TestFind(t *testing.T) {
	for query, want := range map[string]string{
		"rev": "1", "VINTAGE AERIAL (KEN)": "4", "refresh": "2", "vintage": "3", "old": "7",
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
	if !IsID("556d297a-a1ed-465d-a4b2-6d4945a0a538") || IsID("Rev") {
		t.Error("IsID misjudged an id or a name")
	}
}
