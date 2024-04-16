package trusted_subnet

import (
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/rs/zerolog"
)

//func TestCheck(t *testing.T) {
//	type args struct {
//		log    *zerolog.Logger
//		subnet *net.IPNet
//	}
//	tests := []struct {
//		name string
//		args args
//		want func(next http.Handler) http.Handler
//	}{
//		{
//			"",
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := Check(tt.args.log, tt.args.subnet); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("Check() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name   string
		subnet string
		IPNet  *net.IPNet
		want   bool
	}{
		{
			name:   "valid subnet",
			subnet: "192.168.3.1/24",
			IPNet: &net.IPNet{
				IP:   net.ParseIP("192.168.3.1"),
				Mask: net.IPMask(net.ParseIP("255.255.255.0").To4()),
			},
			want: true,
		},
	}
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, subnet := IsValid(&logger, tt.subnet)
			if !reflect.DeepEqual(ip, tt.IPNet.IP) {
				t.Errorf("IsValid() ip = %v, IPNet %v", ip, tt.IPNet)
			}
			if !reflect.DeepEqual(subnet, tt.IPNet.Mask) {
				t.Errorf("IsValid() ip = %v, IPNet %v", ip, tt.IPNet)
			}
		})
	}
}
