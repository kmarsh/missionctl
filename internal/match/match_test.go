package match

import "testing"

var projects = []Project{
	{"1", "Apollo", true},
	{"2", "Apollo Guidance", true},
	{"3", "Gemini Mobile", true},
	{"4", "Gemini Mobile (2019)", false},
	{"5", "Mercury App", true},
	{"6", "Voyager App", true},
	{"7", "Skylab", false},
}

func TestFind(t *testing.T) {
	for query, want := range map[string]string{
		"apollo": "1", "GEMINI MOBILE (2019)": "4", "guidance": "2", "gemini": "3", "sky": "7",
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
	if !IsID("556d297a-a1ed-465d-a4b2-6d4945a0a538") || IsID("Apollo") {
		t.Error("IsID misjudged an id or a name")
	}
}
