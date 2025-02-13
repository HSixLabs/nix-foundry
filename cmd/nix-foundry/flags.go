package main

// Flag variables used across commands
var (
	// Diagnostic flags
	security bool
	system   bool
	fix      bool
	follow   bool
	tail     int

	// Backup flags
	forceBackup bool

	// Profile flags
	profileName  string
	forceProfile bool

	// Uninstall flags
	keepConfig     bool
	forceUninstall bool

	// Update flags
	forceUpdate bool

	// Switch flags
	forceSwitch bool

	// Init flags
	shell       string
	editor      string
	gitName     string
	gitEmail    string
	autoConfig  bool
	projectInit bool
	forceInit   bool

	// Project flags
	forceProject bool

	// Config flags
	forceConfig bool

	// Shared flags
	teamName string
)
