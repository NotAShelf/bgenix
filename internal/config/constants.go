package config

const (
	// this is a very stupid way to handle configuration
	// but this is not production code so, suck my ass
	DefaultRulesFile = "./secrets.nix"
	PackageName      = "@pname@"
	Version          = "@version@"
	NixInstantiate   = "@nixInstantiate@"
	JqBinary         = "@jqBin@"
	MktempBinary     = "@mktempBin@"
	DiffBinary       = "@diffBin@"
	AgeBinary        = "@ageBin@"
	AgeVersion       = "@ageVersion@"
)
