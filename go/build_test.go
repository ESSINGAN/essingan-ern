package ern

import (
	"errors"
	"testing"
)

const (
	acct = "482019375521"
	org  = "o-01jd2v8k4m7q9x3a"
	proj = "p-01jf3w9n5r8t2y6b"
)

func TestBuilders(t *testing.T) {
	must := func(e ERN, err error) ERN {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return e
	}
	instance := must(NewProjectResource("arca", "scaleway-fr-par", acct, org, proj, "instance", "db-01jf4x2n5r8t2y6c"))
	cases := []struct {
		got  ERN
		want string
		form Pattern
	}{
		{must(NewAccount(acct)), "ern:essingan:account::" + acct + ":account", E1},
		{must(NewAccountResource("iam", acct, "user", "u-01jcz7k2m9q4x8n2")), "ern:essingan:iam::" + acct + ":user/u-01jcz7k2m9q4x8n2", E2},
		{must(NewOrganization(acct, org)), "ern:essingan:account::" + acct + ":organization/" + org, E3},
		{must(NewOrganizationResource("iam", acct, org, "group", "g-01je5k7m9p3r6t8d")), "ern:essingan:iam::" + acct + ":" + org + "/group/g-01je5k7m9p3r6t8d", E4},
		{must(NewProject(acct, org, proj)), "ern:essingan:account::" + acct + ":" + org + "/project/" + proj, E5},
		{instance, "ern:essingan:arca:scaleway-fr-par:" + acct + ":" + org + "/" + proj + "/instance/db-01jf4x2n5r8t2y6c", E6},
		{must(NewProjectResource("scu", "", acct, org, proj, "capacity", "cap-01jeq8t2w4y6a9b3")), "ern:essingan:scu::" + acct + ":" + org + "/" + proj + "/capacity/cap-01jeq8t2w4y6a9b3", E7},
		{must(instance.Sub("database", "ldb-01jf4x2n5r8t2y6d")), "ern:essingan:arca:scaleway-fr-par:" + acct + ":" + org + "/" + proj + "/instance/db-01jf4x2n5r8t2y6c/database/ldb-01jf4x2n5r8t2y6d", E8},
		{must(NewManaged("iam", "policy", "AdministratorAccess")), "ern:essingan:iam:::policy/AdministratorAccess", E9},
		{must(NewManaged("catalogue", "artefact", "arca", "0.3.0")), "ern:essingan:catalogue:::artefact/arca/0.3.0", E9},
		{must(NewManaged("catalogue", "provider", "hetzner")), "ern:essingan:catalogue:::provider/hetzner", E9},
		{must(NewRoleSession(acct, "Deployer", "ci-run-7731")), "ern:essingan:sts::" + acct + ":role-session/Deployer/ci-run-7731", E10},
		{must(NewRoleSession(acct, "Deployer", "ci-run-7731", org)), "ern:essingan:sts::" + acct + ":" + org + "/role-session/Deployer/ci-run-7731", E10},
	}
	for _, c := range cases {
		if got := c.got.String(); got != c.want {
			t.Errorf("got %q, want %q", got, c.want)
		}
		if got := c.got.Pattern(); got != c.form {
			t.Errorf("%q: pattern %s, want %s", c.want, got, c.form)
		}
		if parsed, err := Parse(c.want); err != nil || parsed != c.got {
			t.Errorf("Parse(%q) = %+v, %v, want %+v", c.want, parsed, err, c.got)
		}
	}
}

func TestBuildersRefuseBadInput(t *testing.T) {
	cases := map[string]func() (ERN, error){
		"bad account":            func() (ERN, error) { return NewAccount("123") },
		"unknown service":        func() (ERN, error) { return NewAccountResource("s3", acct, "user", "u-01jcz7k2m9q4x8n2") },
		"a name in place of ERI": func() (ERN, error) { return NewAccountResource("iam", acct, "user", "alice") },
		"project id for an org":  func() (ERN, error) { return NewOrganization(acct, proj) },
		"org id for a project":   func() (ERN, error) { return NewProject(acct, org, org) },
		"location without region": func() (ERN, error) {
			return NewProjectResource("arca", "scaleway", acct, org, proj, "instance", "db-01jf4x2n5r8t2y6c")
		},
		"E2 builder making an E3":  func() (ERN, error) { return NewAccountResource("account", acct, "organization", org) },
		"wildcard in a stored ERN": func() (ERN, error) { return NewManaged("iam", "policy", "*") },
		"role name with a slash":   func() (ERN, error) { return NewRoleSession(acct, "a/b", "s") },
	}
	for name, f := range cases {
		if _, err := f(); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: %v, want ErrInvalid", name, err)
		}
	}
}

func TestSubGoesOneLevelDown(t *testing.T) {
	e, err := Parse("ern:essingan:iam::" + acct + ":user/u-01jcz7k2m9q4x8n2/mfa-device/mfa-01jeg9k3n5p7r2t4")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.Sub("x", "y-01jf4x2n5r8t2y6d"); !errors.Is(err, ErrInvalid) {
		t.Errorf("Sub of a sub-resource: %v, want ErrInvalid", err)
	}
}

func TestInPartition(t *testing.T) {
	e, err := NewAccount(acct)
	if err != nil {
		t.Fatal(err)
	}
	moved, err := e.InPartition("essingan-acme01")
	if err != nil {
		t.Fatal(err)
	}
	if want := "ern:essingan-acme01:account::" + acct + ":account"; moved.String() != want {
		t.Errorf("got %q, want %q", moved, want)
	}
	if _, err := e.InPartition("aws"); !errors.Is(err, ErrInvalid) {
		t.Errorf("InPartition(aws): %v, want ErrInvalid", err)
	}
}
