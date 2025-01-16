package servers

import (
	"fmt"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

// empty by default
var currentServers = map[string][]proxy.RegisteredServer{}

func AddServerToGroup(group string, server proxy.RegisteredServer) {
	currentServers[group] = append(currentServers[group], server)
	fmt.Printf("Added server %s to group %s\n", server.ServerInfo().Name(), group)
}

func RemoveServerFromGroup(group string, server proxy.RegisteredServer) {
	servers := currentServers[group]
	for i, s := range servers {
		if s.ServerInfo().Name() == server.ServerInfo().Name() {
			// Remove the server from the slice
			currentServers[group] = append(servers[:i], servers[i+1:]...)
			fmt.Printf("removed server %s from group %s\n", server.ServerInfo().Name(), group)

			break
		}
	}
}

// todo: get fallbacks in layers. (ie in some random gg server, send to personal gorge, then if that fails to lobby)
func GetLeastLoadedServer(serverType string, excludeIds ...string) proxy.RegisteredServer {
	fmt.Printf("Fetching least loaded server for group %s\n", serverType)
	// Find the least loaded server in the specified serverType
	var leastLoadedServer proxy.RegisteredServer
	leastLoad := int(^uint(0) >> 1) // Max possible int

	for _, server := range currentServers[serverType] {
		fmt.Printf("Interatoring")
		if contains(excludeIds, server.ServerInfo().Name()) {
			continue // this server has been exluded
		}
		if server.Players().Len() < leastLoad {
			leastLoad = server.Players().Len()
			leastLoadedServer = server
			fmt.Printf("Least Loaded Server: %s\n", server.ServerInfo().Name())
		}
	}
	fmt.Printf("Returning server: %v", leastLoadedServer)
	return leastLoadedServer
}

// contains checks if a slice contains a specific element
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
