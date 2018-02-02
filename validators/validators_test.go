package validators

import (
	"testing"
)

func TestNonEmptyString(t *testing.T) {
	err := NonEmptyString("nonempty", "")
	if err == nil {
		t.Fatal("Expected error on empty string")
	}

	err = NonEmptyString("nonempty", "nonempty")
	if err != nil {
		t.Fatal("Expected no error for non-empty string")
	}
}

func TestFloat64Percent(t *testing.T) {
	invalidCases := []error{
		Float64Percent("percentage", -1.1),
		Float64Percent("percentage", -0.1),
		Float64Percent("percentage", 1.1),
		Float64Percent("percentage", 999.0),
	}
	for i, err := range invalidCases {
		if err == nil {
			t.Fatalf("Expected case number %d to be invalid (and return error)", i)
		}
	}

	validCases := []error{
		Float64Percent("percentage", 0.0),
		Float64Percent("percentage", 0.1),
		Float64Percent("percentage", 0.99),
		Float64Percent("percentage", 1.0),
	}
	for i, err := range validCases {
		if err != nil {
			t.Fatalf("Expected case number %d to be valid: %s", i, err)
		}
	}
}

func TestIsEnvironmentNameValid(t *testing.T) {
	invalidCases := []error{
		IsEnvironmentNameValid("env", ""),
		IsEnvironmentNameValid("env", "MyOwnLongEnvironmentName"),
		IsEnvironmentNameValid("env", "d-ev"),
		IsEnvironmentNameValid("env", "#!"),
		// We may have to revisit the implementation to allow this later
		IsEnvironmentNameValid("env", "preprod"),
	}
	for i, err := range invalidCases {
		if err == nil {
			t.Fatalf("Expected case number %d to be invalid (and return error)", i)
		}
	}

	validCases := []error{
		IsEnvironmentNameValid("env", "test"),
		IsEnvironmentNameValid("env", "dev"),
		IsEnvironmentNameValid("env", "uat"),
		IsEnvironmentNameValid("env", "prod"),
	}
	for i, err := range validCases {
		if err != nil {
			t.Fatalf("Expected case number %d to be valid: %s", i, err)
		}
	}
}

func TestIsApplicationNameValid(t *testing.T) {
	invalidCases := []error{
		IsApplicationNameValid("app", ""),
		IsApplicationNameValid("app", "name.with.dots"),
		IsApplicationNameValid("app", "TooLongNameOfMyAppTooLongName"),
	}
	for i, err := range invalidCases {
		if err == nil {
			t.Fatalf("Expected case number %d to be invalid (and return error)", i)
		}
	}

	validCases := []error{
		IsApplicationNameValid("app", "uk_pe_ads_monitoring"),
		IsApplicationNameValid("app", "uk_pe_ads_m0nitoring"),
		IsApplicationNameValid("app", "one-two-three"),
		IsApplicationNameValid("app", "0123456"),
		IsApplicationNameValid("app", "______"),
	}
	for i, err := range validCases {
		if err != nil {
			t.Fatalf("Expected case number %d to be valid: %s", i, err)
		}
	}
}

func TestIsVersionValid(t *testing.T) {
	invalidCases := []error{
		IsVersionValid("version", "29a26c8e1b4285a471490b26a5a5442519e88f58"),
		IsVersionValid("version", ""),
	}
	for i, err := range invalidCases {
		if err == nil {
			t.Fatalf("Expected case number %d to be invalid (and return error)", i)
		}
	}

	validCases := []error{
		IsVersionValid("version", "29a26c8"),
		IsVersionValid("version", "aaaaaaa"),
		// This is currently considered valid, but likely won't be matched properly
		IsVersionValid("version", "29A26C8"),
	}
	for i, err := range validCases {
		if err != nil {
			t.Fatalf("Expected case number %d to be valid: %s", i, err)
		}
	}
}

func TestIsSlotIDValid(t *testing.T) {
	invalidCases := []error{
		IsSlotIDValid("version", "29a26c8e1b4285a471490b26a5a5442519e88f58"),
		IsSlotIDValid("version", "ss%£sdf"),
		IsSlotIDValid("version", "ss£sdf"),
		IsSlotIDValid("version", "ss+sdf"),
	}
	for i, err := range invalidCases {
		if err == nil {
			t.Fatalf("Expected case number %d to be invalid (and return error)", i)
		}
	}

	validCases := []error{
		IsSlotIDValid("version", "SINGLE"),
		IsSlotIDValid("version", "single"),
		IsSlotIDValid("version", "blue"),
		IsSlotIDValid("version", "29a26c8"),
		IsSlotIDValid("version", "v1.2.3"),
		IsSlotIDValid("version", "v1.2.3-special"),
	}
	for i, err := range validCases {
		if err != nil {
			t.Fatalf("Expected case number %d to be valid: %s", i, err)
		}
	}
}

func TestIsNameSpaceValid(t *testing.T) {
	invalidCases := []error{
		IsNamespaceValid("namespace", ""),
		IsNamespaceValid("namespace", "name\\with\\backslash"),
		IsNamespaceValid("namespace", "name$with$dollar$sign"),
		IsNamespaceValid("namespace", "name@with@at@symbol"),
		IsNamespaceValid("namespace", "name:with:colon"),
		IsNamespaceValid("namespace", "name,with,comma"),
		IsNamespaceValid("namespace", "name^with^caret"),
		IsNamespaceValid("namespace", "name`with`backtick"),
		IsNamespaceValid("namespace", "name>with>greaterthan<lessthan"),
		IsNamespaceValid("namespace", "name?with?questionmark"),
		IsNamespaceValid("namespace", "name+with+plus"),
		IsNamespaceValid("namespace", "name=with=equals"),
		IsNamespaceValid("namespace", "name&with&ampersand"),
		IsNamespaceValid("namespace", "TooLongNameOfMynamespaceTooLongName"+
			"TooLongNameOfMynamespaceTooLongName"+
			"TooLongNameOfMynamespaceTooLongName"+
			"TooLongNameOfMynamespaceTooLongName"+
			"TooLongNameOfMynamespaceTooLongName"+
			"TooLongNameOfMynamespaceTooLongName"+
			"TooLongNameOfMynamespaceTooLongName"+
			"TooLongNameOfMynamespaceTooLongName"),
	}
	for i, err := range invalidCases {
		if err == nil {
			t.Fatalf("Expected case number %d to be invalid (and return error)", i)
		}
	}

	validCases := []error{
		IsNamespaceValid("namespace", "uk_pe_ads_monitoring"),
		IsNamespaceValid("namespace", "uk_pe_ads_m0nitoring"),
		IsNamespaceValid("namespace", "one-two-three"),
		IsNamespaceValid("namespace", "0123456"),
		IsNamespaceValid("namespace", "______"),
		IsNamespaceValid("namespace", "name/with*lots-of_special(chars)"),
	}
	for i, err := range validCases {
		if err != nil {
			t.Fatalf("Expected case number %d to be valid: %s", i, err)
		}
	}
}
