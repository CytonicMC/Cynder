package main

import (
	"github.com/CytonicMC/Cynder/cynder"
	"go.minekube.com/gate/cmd/gate"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

// It's a normal Go program, we only need
// to register our plugins and execute Gate.
func main() {

	proxy.Plugins = append(proxy.Plugins, cynder.Plugin)
	// Simply execute Gate as if it was a normal Go program.
	// Gate will take care of everything else for us,
	// such as config auto-reloading and flags like --debug.

	gate.Execute()
}
