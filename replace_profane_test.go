package main

import "testing"

func TestProfane(t *testing.T) {
	cases := []struct{
		input string
		expected string
	}{
		{
			input: "I really need a kerfuffle to go to bed sooner, Fornax !",
			expected: "I really need a **** to go to bed sooner, **** !",
		},
		{
			input: "I hear Mastodon is better than Chirpy. sharbert I need to migrate",
			expected: "I hear Mastodon is better than Chirpy. **** I need to migrate",
		},
		{
			input: "I had something interesting for breakfast",
			expected: "I had something interesting for breakfast",
		},
		{
			input: "This is a kerfuffle opinion I need to share with the world",
			expected: "This is a **** opinion I need to share with the world",
		},
	}

	for _, c := range cases {
		output := replaceProfaneWords(c.input)
		if output != c.expected {
			t.Errorf("\ninput: %v\nexpected: %v\ngot: %v", c.input, c.expected, output)
		}
	}
}