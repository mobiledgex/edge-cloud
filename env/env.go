package env

import "os"

// Env interface allows us to substitute the system environment with other targets.
// See os package for definition of functions.
type Env interface {
	Getenv(key string) string
	Setenv(key, value string) error
	LookupEnv(key string) (string, bool)
	Environ() []string
}

// EnvOS gets vars from system
type EnvOS struct{}

func (s *EnvOS) Getenv(key string) string {
	return os.Getenv(key)
}

func (s *EnvOS) Setenv(key, value string) error {
	return os.Setenv(key, value)
}

func (s *EnvOS) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func (s *EnvOS) Environ() []string {
	return os.Environ()
}

// EnvMap gets vars from a map
type EnvMap struct {
	Vars map[string]string
}

func (s *EnvMap) Getenv(key string) string {
	return s.Vars[key]
}

func (s *EnvMap) Setenv(key, value string) error {
	s.Vars[key] = value
	return nil
}

func (s *EnvMap) LookupEnv(key string) (string, bool) {
	val, ok := s.Vars[key]
	return val, ok
}

func (s *EnvMap) Environ() []string {
	vars := []string{}
	for k, v := range s.Vars {
		vars = append(vars, k+"="+v)
	}
	return vars
}
