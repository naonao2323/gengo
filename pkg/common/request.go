package common

type Request int

const (
	DaoPostgresRequest Request = iota
	TestContainerPostgresRequest
	TestFixturePostgresRequest
	FrameworkPostgresRequest
)
