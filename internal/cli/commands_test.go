package cli

import (
	"reflect"
	"testing"
)

func TestFindCommand(t *testing.T) {
	command := findCommand("login", false)

	if command == nil {
		t.Fatal("expected login command to exist")
	}

	if command.Name != "login" {
		t.Fatalf("expected login, got %q", command.Name)
	}
}

func TestFindCommandRejectsAuthenticatedCommandBeforeLogin(t *testing.T) {
	command := findCommand("whoami", false)

	if command != nil {
		t.Fatal("expected whoami to be unavailable before login")
	}
}

func TestFindCommandAllowsAuthenticatedCommandAfterLogin(t *testing.T) {
	command := findCommand("whoami", true)

	if command == nil {
		t.Fatal("expected whoami to be available after login")
	}
}

func TestFindCommandTrimsWhitespace(t *testing.T) {
	command := findCommand("  help  ", false)

	if command == nil {
		t.Fatal("expected help command to be found")
	}
}

func TestCommandNames(t *testing.T) {
	expected := []string{
		"register",
		"login",
		"help",
		"exit",
	}

	actual := commandNames(false)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func TestAuthenticatedCommandNames(t *testing.T) {
	expected := []string{
		"whoami",
		"enable-2fa",
		"disable-2fa",
		"logout",
		"help",
	}

	actual := commandNames(true)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
