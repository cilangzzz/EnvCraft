// Package server_command
package server_command

import "tsc/pkg/util/server_command/core"

// New creates a new Executor instance.
func New() *core.Executor {
	return core.NewExecutor()
}
