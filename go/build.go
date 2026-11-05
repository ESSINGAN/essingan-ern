package ern

import "strings"

// NewAccount builds E1, the ERN of an account.
func NewAccount(account string) (ERN, error) {
	return build(E1, ERN{Service: "account", Account: account, Path: "account"})
}

// NewAccountResource builds E2, such as iam::<acct>:user/u-….
func NewAccountResource(service, account, resourceType, id string) (ERN, error) {
	return build(E2, ERN{Service: service, Account: account, Path: join(resourceType, id)})
}

// NewOrganization builds E3.
func NewOrganization(account, org string) (ERN, error) {
	return build(E3, ERN{Service: "account", Account: account, Path: join("organization", org)})
}

// NewOrganizationResource builds E4, such as iam::<acct>:o-…/group/g-….
func NewOrganizationResource(service, account, org, resourceType, id string) (ERN, error) {
	return build(E4, ERN{Service: service, Account: account, Path: join(org, resourceType, id)})
}

// NewProject builds E5.
func NewProject(account, org, project string) (ERN, error) {
	return build(E5, ERN{Service: "account", Account: account, Path: join(org, "project", project)})
}

// NewProjectResource builds E6 when location is set, and E7 when it is empty.
func NewProjectResource(service, location, account, org, project, resourceType, id string) (ERN, error) {
	want := E7
	if location != "" {
		want = E6
	}
	return build(want, ERN{
		Service:  service,
		Location: location,
		Account:  account,
		Path:     join(org, project, resourceType, id),
	})
}

// NewManaged builds E9, an object shared by every account, such as
// iam:::policy/AdministratorAccess.
func NewManaged(service, resourceType string, names ...string) (ERN, error) {
	return build(E9, ERN{Service: service, Path: join(append([]string{resourceType}, names...)...)})
}

// NewRoleSession builds E10. scope is empty, an organization, or an
// organization and a project.
func NewRoleSession(account, role, session string, scope ...string) (ERN, error) {
	segs := append(append([]string{}, scope...), roleSession, role, session)
	return build(E10, ERN{Service: "sts", Account: account, Path: join(segs...)})
}

// Sub builds E8, a sub-resource of e, which lives and dies with it.
func (e ERN) Sub(resourceType, id string) (ERN, error) {
	sub := e
	sub.Path = join(e.Path, resourceType, id)
	return build(E8, sub)
}

// InPartition returns e in a customer's own Management Plane.
func (e ERN) InPartition(partition string) (ERN, error) {
	e.Partition = partition
	if err := e.Validate(); err != nil {
		return ERN{}, err
	}
	return e, nil
}

func build(want Pattern, e ERN) (ERN, error) {
	if e.Partition == "" {
		e.Partition = DefaultPartition
	}
	got, err := e.classify()
	if err != nil {
		return ERN{}, err
	}
	if got != want {
		return ERN{}, invalid("%s is a %s, not a %s", e, got, want)
	}
	return e, nil
}

func join(segs ...string) string {
	return strings.Join(segs, "/")
}
