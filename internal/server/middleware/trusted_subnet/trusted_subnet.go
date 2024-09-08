package trustedsubnet

import (
	"net"
	"net/http"

	"github.com/rs/zerolog"
)

// IsValid checks if the provided subnet string is a valid CIDR notation.
// It returns the parsed subnet as a *net.IPNet and a boolean indicating whether the subnet is valid.
// If the subnet is not valid, it logs an error message and returns an empty *net.IPNet and false.
func IsValid(logger *zerolog.Logger, subnet string) (*net.IPNet, bool) {
	l := logger.With().Str("func", "IsValid").Logger()
	_, cidr, err := net.ParseCIDR(subnet)
	if err != nil {
		l.Error().Err(err).Msg("failed to parse CIDR")
		return &net.IPNet{}, false
	}
	return cidr, true
}

// Check returns a middleware that checks if the source IP of the request is contained in the provided subnet.
// If the source IP is not contained in the subnet, it logs a debug message and responds with a 403 Forbidden status.
// If the source IP cannot be parsed or is not found in the request headers,
// it also responds with a 403 Forbidden status.
func Check(log *zerolog.Logger, subnet *net.IPNet) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := log.With().Str("middleware", "CheckSubnet").Logger()

			stringIP := r.Header.Get("X-Real-IP")
			if stringIP == "" {
				l.Debug().Msg("could not get source IP address")
				http.Error(w, "", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(stringIP)
			if ip == nil {
				l.Debug().Msg("failed to parse IP address")
				http.Error(w, "", http.StatusForbidden)
				return
			}

			if !subnet.Contains(ip) {
				l.Debug().Msgf("IP not allowed: %v", ip)
				http.Error(w, "", http.StatusForbidden)
				return
			}

			l.Debug().Msgf("X-Real-IP: %s", stringIP)
			next.ServeHTTP(w, r)
		})
	}
}
