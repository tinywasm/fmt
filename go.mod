module webtyp.com/fmt

go 1.22

// v0.25.4 contains a local filesystem path (home directory + project folder
// name) accidentally committed in a test file. No credentials or secrets were
// involved. Fixed in v0.25.5; retracted here per Go's standard mechanism so
// `go get`/`go list -m -u` steer consumers away from the affected version.
retract v0.25.4
