package utils

import (
	"path"
	"strings"
)

// IsRouteAllowed checks if a given route path is allowed based on the list of allowed routes
func IsRouteAllowed(routePath string, allowedRoutes []string) bool {
	// Clean and standardize the route path
	routePath = path.Clean("/" + routePath)

	for _, allowedRoute := range allowedRoutes {
		// Clean and standardize the allowed route pattern
		allowedRoute = path.Clean("/" + allowedRoute)

		// Check for exact match
		if allowedRoute == routePath {
			return true
		}

		// Check wildcard patterns
		if strings.Contains(allowedRoute, "*") {
			if matchWildcardPattern(allowedRoute, routePath) {
				return true
			}
		}

		// Check prefix patterns
		if strings.HasSuffix(allowedRoute, "/**") {
			prefix := strings.TrimSuffix(allowedRoute, "/**")
			if strings.HasPrefix(routePath, prefix) {
				return true
			}
		}
	}

	return false
}

// matchWildcardPattern checks if a route matches a wildcard pattern
func matchWildcardPattern(pattern, route string) bool {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	routeParts := strings.Split(strings.Trim(route, "/"), "/")

	if len(patternParts) != len(routeParts) {
		return false
	}

	for i, part := range patternParts {
		if part != "*" && part != routeParts[i] {
			return false
		}
	}

	return true
}
