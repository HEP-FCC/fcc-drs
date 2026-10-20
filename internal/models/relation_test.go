package models

import (
	"reflect"
	"strconv"
	"testing"
)

func TestRefPattern(t *testing.T) {
	cases := []struct {
		text string
		want []int
	}{
		{"see DRS-42", []int{42}},
		{"drs42 and Drs 7, DRS#9.", []int{42, 7, 9}},
		{"DRS-1,DRS-2", []int{1, 2}},
		{"(DRS-3)", []int{3}},
		{"fcc-drs-42", nil},
		{"path/DRS-5 x.DRS-6 a_DRS-7 xDRS-8 1DRS-9", nil},
		{"DRS-42abc", nil},
		{"#42", nil},
	}
	for _, c := range cases {
		var got []int
		for _, m := range refPattern.FindAllStringSubmatch(c.text, -1) {
			id, _ := strconv.Atoi(m[2])
			got = append(got, id)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%q: got %v, want %v", c.text, got, c.want)
		}
	}
}

func TestParseRefID(t *testing.T) {
	cases := []struct {
		in string
		id int
		ok bool
	}{
		{"42", 42, true},
		{" DRS-42 ", 42, true},
		{"drs42", 42, true},
		{"DRS#42", 42, true},
		{"#42", 0, false},
		{"DRS-", 0, false},
		{"0", 0, false},
		{"abc", 0, false},
	}
	for _, c := range cases {
		id, ok := ParseRefID(c.in)
		if id != c.id || ok != c.ok {
			t.Errorf("ParseRefID(%q) = %d, %v; want %d, %v", c.in, id, ok, c.id, c.ok)
		}
	}
}
