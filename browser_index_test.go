package nanny

import "testing"

func TestShortValueFormat(t *testing.T) {
	if got, want := shortValueFormat(nil), ""; got != want {
		t.Errorf("got [%v]:%T want [%v]:%T", got, got, want, want)
	}
	if got, want := shortValueFormat(42), "42"; got != want {
		t.Errorf("got [%v]:%T want [%v]:%T", got, got, want, want)
	}
	m := map[string]interface{}{"foo": "bar"}
	if got, want := shortValueFormat(m), `foo="bar" `; got != want {
		t.Errorf("got [%v]:%T want [%v]:%T", got, got, want, want)
	}
	m2 := map[string]interface{}{"foo": nil}
	if got, want := shortValueFormat(m2), `foo=nil `; got != want {
		t.Errorf("got [%v]:%T want [%v]:%T", got, got, want, want)
	}
}
