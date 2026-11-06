package i18n

import (
	"regexp"
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"pt_BR.UTF-8":     "pt",
		"zh-Hans-CN":      "zh",
		"de_DE@euro":      "de",
		"fr:en":           "fr",
		"en-US":           "en",
		"sr_RS.UTF-8@lat": "sr",
	}
	for in, want := range cases {
		if got := normalize(in); got != want {
			t.Errorf("%s: got %s want %s", in, got, want)
		}
	}
}

func TestFallsBackToEnglish(t *testing.T) {
	Init("xx")
	if got := T("scan.clean"); got == "scan.clean" {
		t.Fatal("no fallback")
	}
}

var verbs = regexp.MustCompile(`%[sdv]`)

func TestLocalesMatchEnglish(t *testing.T) {
	en := load("en")
	for _, lang := range Languages() {
		m := load(lang)
		for k, v := range en {
			tr, ok := m[k]
			if !ok {
				t.Errorf("%s: missing %s", lang, k)
				continue
			}
			a, b := verbs.FindAllString(v, -1), verbs.FindAllString(tr, -1)
			if len(a) != len(b) {
				t.Errorf("%s: %s has %v, english has %v", lang, k, b, a)
				continue
			}
			for i := range a {
				if a[i] != b[i] {
					t.Errorf("%s: %s reorders verbs", lang, k)
				}
			}
		}
		for k := range m {
			if _, ok := en[k]; !ok {
				t.Errorf("%s: stale key %s", lang, k)
			}
		}
	}
}
