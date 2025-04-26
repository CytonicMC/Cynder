package servers

import (
	"fmt"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

// First fallback this way ->
// then V
var fallbackHierarchy = map[string][]string{
	"gilded_gorge": {"instancing_server", "hub"},
	"bedwars":      {"quads", "duos", "solos", "lobby"},
	"cytonic":      {"lobby"},
}

// empty by default
var currentServers = map[string]map[string][]proxy.RegisteredServer{}

// Structure: group -> type -> []servers
// ie: gilded_gorge -> instancing_server -> servers

func AddServerToGroup(group string, serverType string, server proxy.RegisteredServer) {
	// Ensure the group exists
	if _, exists := currentServers[group]; !exists {
		currentServers[group] = make(map[string][]proxy.RegisteredServer)
	}
	// Add server to the specific type
	currentServers[group][serverType] = append(currentServers[group][serverType], server)
	fmt.Printf("Added server %s to group %s (type: %s)\n", server.ServerInfo().Name(), group, serverType)
}

func RemoveServerFromGroup(group string, serverType string, server proxy.RegisteredServer) {
	if serversByType, exists := currentServers[group]; exists {
		if servers, exists := serversByType[serverType]; exists {
			for i, s := range servers {
				if s == nil || s.ServerInfo() == nil || server == nil || server.ServerInfo() == nil {
					return
				}
				if s.ServerInfo().Name() == server.ServerInfo().Name() {
					// Remove the server from the slice
					currentServers[group][serverType] = append(servers[:i], servers[i+1:]...)
					fmt.Printf("Removed server %s from group %s (type: %s)\n", server.ServerInfo().Name(), group, serverType)
					break
				}
			}
		}
	}
}

// todo: get fallbacks in layers. (ie in some random gg server, send to personal gorge, then if that fails to lobby)
func GetLeastLoadedServer(group string, serverType string, excludeIds ...string) proxy.RegisteredServer {
	fmt.Printf("Fetching least loaded server for group %s with type %s\n", group, serverType)

	if _, exists := currentServers[group]; !exists {
		return nil // Group doesn't exist
	}
	if _, exists := currentServers[group][serverType]; !exists {
		return nil // Type doesn't exist in this group
	}

	var leastLoadedServer proxy.RegisteredServer
	leastLoad := int(^uint(0) >> 1) // Max possible int

	for _, server := range currentServers[group][serverType] {
		if contains(excludeIds, server.ServerInfo().Name()) {
			continue
		}

		var b string
		if leastLoadedServer != nil {
			b = leastLoadedServer.ServerInfo().Name()
		} else {
			b = "NONE"
		}

		fmt.Printf("Server %s has %d players! Least load is: %d, which is server %s\n", server.ServerInfo().Name(), server.Players().Len(), leastLoad, b)
		if server.Players().Len() < leastLoad {
			leastLoad = server.Players().Len()
			leastLoadedServer = server
		}
	}

	return leastLoadedServer
}

func GetFallbackServer(currentGroup string, serverType string, excludeIds ...string) (proxy.RegisteredServer, bool) {
	// Find the next fallback group and type
	nextGroup, nextType, found := GetNextFallbackGroup(currentGroup, serverType)
	if !found {
		// Check if the group has any fallback types available
		if len(fallbackHierarchy[currentGroup]) == 0 {
			// If no fallback found and no fallback types exist for this group, return failure
			fmt.Printf("No fallback types available for group %s\n", currentGroup)
			return nil, false
		}

		// If no fallback found, use the "bottom type" of the current group
		nextType = fallbackHierarchy[currentGroup][len(fallbackHierarchy[currentGroup])-1]
	}

	// Try to get the least loaded server from the next group and type
	server := GetLeastLoadedServer(nextGroup, nextType, excludeIds...)
	if server != nil {
		return server, true
	}

	// If the fallback type still doesn't have a server, recursively try the next fallback group
	return GetFallbackServer(nextGroup, serverType, excludeIds...)
}

func GetNextFallbackGroup(currentGroup string, typeToStartFrom string) (string, string, bool) {
	groupKeys := make([]string, 0, len(fallbackHierarchy))
	for key := range fallbackHierarchy {
		groupKeys = append(groupKeys, key)
	}

	// Find the starting group and type
	for groupIndex, group := range groupKeys {
		if group == currentGroup {
			fallbacks := fallbackHierarchy[group]

			// Find the type in the current group and move rightward
			for i, fallback := range fallbacks {
				if fallback == typeToStartFrom {
					// Move to the next fallback in the same group
					if i+1 < len(fallbacks) {
						return group, fallbacks[i+1], true
					}
					// Move to the next group in the hierarchy, if available
					for j := groupIndex + 1; j < len(groupKeys); j++ {
						nextGroup := groupKeys[j]
						if len(fallbackHierarchy[nextGroup]) > 0 {
							return nextGroup, fallbackHierarchy[nextGroup][0], true
						}
					}
					return "", "", false // No further fallbacks
				}
			}
		}
	}

	return "", "", false // Not found
}

func GetFallbackFromServer(currentServer proxy.RegisteredServer, excludeIds ...string) (proxy.RegisteredServer, bool) {
	var currentGroup, serverType string

	// Find the group and type of the current server
	for group, types := range currentServers {
		for typ, servers := range types {
			for _, server := range servers {
				if server.ServerInfo().Name() == currentServer.ServerInfo().Name() {
					currentGroup = group
					serverType = typ
					break
				}
			}
		}
	}

	// If we couldn't determine the group, fallback to the next in hierarchy within the same group
	if currentGroup == "" || serverType == "" {
		fmt.Printf("Server %s not found in any group, attempting fallback to next in hierarchy\n", currentServer.ServerInfo().Name())

		// Check if the current group has fallbacks (hierarchy)
		if currentGroup != "" && len(fallbackHierarchy[currentGroup]) > 0 {
			// Get the next fallback in the same group
			nextTypeIndex := -1
			for i, fallbackType := range fallbackHierarchy[currentGroup] {
				if fallbackType == serverType {
					nextTypeIndex = i + 1
					break
				}
			}

			// If we found a valid next type within the same group
			if nextTypeIndex > 0 && nextTypeIndex < len(fallbackHierarchy[currentGroup]) {
				nextType := fallbackHierarchy[currentGroup][nextTypeIndex]
				// Get the least loaded server from the next fallback type
				fallbackServer := GetLeastLoadedServer(currentGroup, nextType, append(excludeIds, currentServer.ServerInfo().Name())...)
				if fallbackServer != nil {
					fmt.Printf("Fallback server found: %s\n", fallbackServer.ServerInfo().Name())
					return fallbackServer, true
				}
			}
		}

		// If no valid fallback within the current group, fall back to another group
		nextGroup, nextType, found := GetNextFallbackGroup(currentGroup, serverType)
		if found {
			// Get the least loaded server from the next group and type
			fallbackServer := GetLeastLoadedServer(nextGroup, nextType, append(excludeIds, currentServer.ServerInfo().Name())...)
			if fallbackServer != nil {
				fmt.Printf("Fallback server found in next group: %s\n", fallbackServer.ServerInfo().Name())
				return fallbackServer, true
			}
		}

		// If no fallback found, default to the last resolved group and type
		lastGroup, lastType := "", ""
		for group, types := range fallbackHierarchy {
			if len(types) > 0 {
				lastGroup, lastType = group, types[len(types)-1] // Last type in the group
			}
		}

		// Try to get the least loaded server from the last fallback
		fallbackServer := GetLeastLoadedServer(lastGroup, lastType, append(excludeIds, currentServer.ServerInfo().Name())...)
		if fallbackServer != nil {
			fmt.Printf("Fallback to last fallback server: %s\n", fallbackServer.ServerInfo().Name())
			return fallbackServer, true
		}

		fmt.Printf("No fallback found for server %s\n", currentServer.ServerInfo().Name())
		return nil, false
	}

	// Get the fallback server
	fallbackServer, found := GetFallbackServer(currentGroup, serverType, append(excludeIds, currentServer.ServerInfo().Name())...)
	if found {
		return fallbackServer, true
	}

	// Find the last group and type in the fallback hierarchy
	var lastGroup, lastType string
	for group, types := range fallbackHierarchy {
		if len(types) > 0 {
			lastGroup, lastType = group, types[len(types)-1] // Last type in the group
		}
	}

	// If no fallback found, default to the last resolved group and type
	fmt.Printf("No fallback found for server %s, defaulting to %s:%s\n", currentServer.ServerInfo().Name(), lastGroup, lastType)
	fallbackServer = GetLeastLoadedServer(lastGroup, lastType, append(excludeIds, currentServer.ServerInfo().Name())...)
	if fallbackServer != nil {
		return fallbackServer, true
	}

	fmt.Printf("No fallback found for server %s\n", currentServer.ServerInfo().Name())
	return nil, false
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
