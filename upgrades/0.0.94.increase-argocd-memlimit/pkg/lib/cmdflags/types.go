// Package cmdflags contains command line flags used by the application
package cmdflags

// Flags contains the various flags the application supports
type Flags struct {
	Debug   bool
	DryRun  bool
	Confirm bool
}
