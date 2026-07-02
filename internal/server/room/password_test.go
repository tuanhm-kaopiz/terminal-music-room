package room

import "testing"

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		in    string
		want  string
		valid bool
	}{
		{"secret", "secret", true},
		{"  pass  ", "pass", true},
		{"", "", false},
		{"   ", "", false},
		{"a", "a", true},
		{string(make([]byte, 33)), "", false},
	}
	for _, tc := range cases {
		got, err := ValidatePassword(tc.in)
		if tc.valid {
			if err != nil || got != tc.want {
				t.Fatalf("%q: got %q err %v", tc.in, got, err)
			}
		} else if err == nil {
			t.Fatalf("%q: expected error", tc.in)
		}
	}
}

func TestIsEmptyPassword(t *testing.T) {
	if !IsEmptyPassword("") || !IsEmptyPassword("  ") {
		t.Fatal("expected empty")
	}
	if IsEmptyPassword("x") {
		t.Fatal("expected non-empty")
	}
}

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("room-pass")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(hash, "room-pass") {
		t.Fatal("expected match")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatal("expected mismatch")
	}
}
