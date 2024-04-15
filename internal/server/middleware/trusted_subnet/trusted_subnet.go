package trusted_subnet

import (
	"net"
	"net/http"

	"github.com/rs/zerolog"
)

func IsValid(logger *zerolog.Logger, subnet string) (*net.IPNet, bool) {
	l := logger.With().Str("func", "IsValid").Logger()
	_, cidr, err := net.ParseCIDR(subnet)
	if err != nil {
		l.Error().Err(err).Msg("failed to parse CIDR")
		return &net.IPNet{}, false
	}
	return cidr, true
}

func Check(log *zerolog.Logger, subnet *net.IPNet) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := log.With().Str("middleware", "CheckSubnet").Logger()

			stringIP := r.Header.Get("X-Real-IP")
			if stringIP == "" {
				l.Debug().Msg("could not get source IP address")
				http.Error(w, "", http.StatusForbidden)
			}

			ip := net.ParseIP(stringIP)
			if ip == nil {
				l.Debug().Msg("failed to parse IP address")
				http.Error(w, "", http.StatusForbidden)
			}

			if !subnet.Contains(ip) {
				l.Debug().Msgf("IP not allowed: %v", ip)
				http.Error(w, "", http.StatusForbidden)
			}

			next.ServeHTTP(w, r)
		})
	}
}
