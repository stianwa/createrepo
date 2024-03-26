package createrepo

import (
	"fmt"
	"strings"
	"testing"
)

func TestCompareLibC(t *testing.T) {
	var list = []string{
		"libc.so.6(GLIBC_2.3)(64bit)",
		"libc.so.6()(64bit)",
		"libc.so.6(GLIBC_2.14)(64bit)",
		"libc.so.6(GLIBC_2.25)(64bit)",
		"libc.so.6(GLIBC_2.2.5)(64bit)",
		"libc.so.6(GLIBC_2.3.4)(64bit)",
		"libc.so.6(GLIBC_2.33)(64bit)",
		"libc.so.6(GLIBC_2.32)(64bit)",
		"libc.so.6(GLIBC_2.34)(64bit)",
		"libc.so.6(GLIBC_2.4)(64bit)",
		"libc.so.6(GLIBC_2.7)(64bit)",
	}

	a := ""

	for _, b := range list {
		if a == "" {
			a = b
			continue
		}
		c := compareLibC(a, b)
		fmt.Printf("compareLibC(%s, %s) = %d\n", a, b, c)

		if c == -1 {
			t.Fatalf("compareLibC failed: %s cmp %s: %d", a, b, c)
		}
		if c == 2 {
			a = b
		}
	}

	fmt.Printf("\nLast: %s\n", a)

	/*
		if fmt.Sprintf("%s", a) != e {
		        t.Fatalf("compose failed: got %s, expected %s", a, e)
		}
	*/
}

func TestReadParenthesis(t *testing.T) {
	type Expect struct {
		Elements []string
		OK       bool
	}
	m := map[string]Expect{
		"(GLIBC_2.3.4)(64bit)":  {Elements: []string{"GLIBC_2.3.4", "64bit"}, OK: true},
		"(GLIBC_2.3.4(64bit)":   {Elements: []string{}, OK: false},
		"(GLIBC_2.3.4)(64bit)a": {Elements: []string{}, OK: false},
		"(GLIBC_2.3.4(64bit))":  {Elements: []string{}, OK: false},
		"()":                    {Elements: []string{""}, OK: true},
		"(64bit)":               {Elements: []string{"64bit"}, OK: true},
	}

	for testCase, expect := range m {
		l, ok := readParenthesis(testCase)
		if expect.OK != ok {
			t.Fatalf("readParenthesis faild: test %s: got %t, expected %t", testCase, ok, expect.OK)

		}
		if strings.Join(l, ", ") != strings.Join(expect.Elements, ", ") {
			t.Fatalf("readParenthesis faild: test %s: got %q, expected %q", testCase, strings.Join(l, ", "), strings.Join(expect.Elements, ", "))
		}
	}
}
