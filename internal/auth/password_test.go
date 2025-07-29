package auth

import (
	"testing"
)

func TestPassword(t *testing.T) {
	cases := []struct{
		input string
		expected string
	}{
		{
			input: "MyNewPassword92347%",
		},
		{
			input: "osd7fnlk32u9ovdffsd",
		},
		{
			input: "sljdjlfjljlk",
		},
		{
			input: "Don't*09234user***ThisPssword",
		},
		{
			input: "Don't*09234user***ThisPssword",
		},
		{
			input: "password123456789",
		},
		{
			input: "qwertyuiop",
		},
	}
	for idx, c := range cases {
		output, err := HashPassword(c.input)
		if err != nil {
			t.Error(err)
		}
		cases[idx].expected = output
	}
	for _, c := range cases {
		err := CheckPasswordHash(c.input, c.expected)
		if err != nil {
			t.Errorf("\ninput: %v\nexpected: %v\nerror: %v", c.input, c.expected, err)
		}
	}
}