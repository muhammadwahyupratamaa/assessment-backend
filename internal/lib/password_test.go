package lib

import "testing"

func TestIsStrongPassword(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "valid strong password",
			password: "Password123!",
			want:     true,
		},
		{
			name:     "missing uppercase letter",
			password: "password123!",
			want:     false,
		},
		{
			name:     "missing lowercase letter",
			password: "PASSWORD123!",
			want:     false,
		},
		{
			name:     "missing number",
			password: "PasswordTest!",
			want:     false,
		},
		{
			name:     "missing special character",
			password: "Password123",
			want:     false,
		},
		{
			name:     "too short",
			password: "Pass1!",
			want:     false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := IsStrongPassword(testCase.password)

			if got != testCase.want {
				t.Fatalf(
					"IsStrongPassword(%q) = %t, want %t",
					testCase.password,
					got,
					testCase.want,
				)
			}
		})
	}
}

func TestHashAndVerifyPassword(t *testing.T) {
	password := "Password123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == password {
		t.Fatal("HashPassword() returned the plain-text password")
	}

	isValid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}

	if !isValid {
		t.Fatal("VerifyPassword() returned false for the correct password")
	}

	isValid, err = VerifyPassword("WrongPassword123!", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}

	if isValid {
		t.Fatal("VerifyPassword() returned true for an incorrect password")
	}
}
