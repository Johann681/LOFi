package handlers

import "testing"

func TestValidateRegistrationRequiresEmailAndPassword(t *testing.T) {
	request := registerStudentRequest{
		Name:              "Ada",
		Gender:            "woman",
		Orientation:       "straight",
		Age:               22,
		Height:            "170 cm",
		Department:        "Computer Science",
		Class:             "2028",
		VerificationProof: "ada@campus.edu",
		Email:             "",
		Password:          "",
		Preferences:       registerStudentRequestPreferences{TargetGender: "man", Orientation: "straight"},
	}

	if err := validateRegistration(request); err == nil {
		t.Fatal("expected email and password validation to fail")
	}
}

func TestValidateRegistrationAcceptsEmailAndPassword(t *testing.T) {
	request := registerStudentRequest{
		Name:              "Ada",
		Gender:            "woman",
		Orientation:       "straight",
		Age:               22,
		Height:            "170 cm",
		Department:        "Computer Science",
		Class:             "2028",
		VerificationProof: "ada@campus.edu",
		Email:             "ada@campus.edu",
		Password:          "StrongPass123",
		Preferences:       registerStudentRequestPreferences{TargetGender: "man", Orientation: "straight"},
	}

	if err := validateRegistration(request); err != nil {
		t.Fatalf("expected valid registration, got %v", err)
	}
}
