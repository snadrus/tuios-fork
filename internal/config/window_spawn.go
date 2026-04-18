package config

// WindowSpawn configures the process that leads a new terminal PTY session.
// When Program is empty, the default interactive shell is used.
// When Program is non-empty, that executable is started with Args (argv after program name).
// Env is merged into the process environment (values replace any existing keys with the same name).
type WindowSpawn struct {
	Program string
	Args    []string
	Env     map[string]string
}
