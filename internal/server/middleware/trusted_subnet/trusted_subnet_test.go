package trusted_subnet

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		subnet   *net.IPNet
		expected bool
	}{
		{
			name:   "Trusted IP",
			header: "192.168.3.1",
			subnet: &net.IPNet{
				IP:   net.ParseIP("192.168.3.0"),
				Mask: net.CIDRMask(24, 32),
			},
			expected: true,
		},
		{
			name:   "Untrusted IP",
			header: "192.168.15.1",
			subnet: &net.IPNet{
				IP:   net.ParseIP("192.168.3.0"),
				Mask: net.CIDRMask(24, 32),
			},
			expected: false,
		},
		{
			name:   "Invalid IP in header",
			header: "292.168.3.1",
			subnet: &net.IPNet{
				IP:   net.ParseIP("192.168.3.0"),
				Mask: net.CIDRMask(24, 32),
			},
			expected: false,
		},
		{
			name:   "Header not provided",
			header: "",
			subnet: &net.IPNet{
				IP:   net.ParseIP("192.168.3.0"),
				Mask: net.CIDRMask(24, 32),
			},
			expected: false,
		},
	}
	l := zerolog.Logger{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/test", nil)
			request.Header.Set("X-Real-IP", tt.header)

			handler := Check(&l, tt.subnet)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(w, request)

			if tt.expected {
				if w.Code != http.StatusOK {
					t.Errorf("expected status code %d, go %d", http.StatusOK, w.Code)
				}
			} else {
				if w.Code != http.StatusForbidden {
					t.Errorf("expected status code %d, go %d", http.StatusForbidden, w.Code)
				}
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name   string
		subnet string
		want   *net.IPNet
		valid  bool
	}{
		{
			name:   "valid subnet",
			subnet: "192.168.3.1/24",
			want: &net.IPNet{
				IP:   net.ParseIP("192.168.3.0"),
				Mask: net.CIDRMask(24, 32),
			},
			valid: true,
		},
		{
			name:   "invalid subnet",
			subnet: "292.168.3.1/24",
			want: &net.IPNet{
				IP:   net.ParseIP("192.168.3.0"),
				Mask: net.CIDRMask(24, 32),
			},
			valid: false,
		},
	}
	l := zerolog.Logger{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := IsValid(&l, tt.subnet)
			if ok != tt.valid {
				t.Errorf("validating error, want: %t, got: %t", tt.valid, ok)
			}

			if tt.valid {
				if got.String() != tt.want.String() {
					t.Errorf("got wrong subnet, want: %v, got: %v", tt.want, got)
				}
			}
		})
	}
}
