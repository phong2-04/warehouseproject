package handler

import (
	"net/http/httptest"
	"testing"
)

func TestParseTrustedProxies(t *testing.T) {
	prefixes, err := ParseTrustedProxies(" 10.0.0.0/8, 192.168.1.10/32 ")
	if err != nil {
		t.Fatal(err)
	}
	if len(prefixes) != 2 {
		t.Fatalf("so CIDR = %d, muon 2", len(prefixes))
	}

	empty, err := ParseTrustedProxies(" ")
	if err != nil || len(empty) != 0 {
		t.Fatalf("danh sach rong khong hop le: %v", err)
	}

	for _, invalid := range []string{
		"not-a-cidr",
		"10.0.0.0/8,",
		",10.0.0.0/8",
	} {
		if _, err := ParseTrustedProxies(invalid); err == nil {
			t.Errorf("mong doi loi voi %q", invalid)
		}
	}
}

func TestGetClientIPWithTrustedProxies(t *testing.T) {
	prefixes, err := ParseTrustedProxies("10.0.0.0/8,192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		remote string
		xff    string
		want   string
	}{
		{
			name:   "ket noi truc tiep khong duoc gia mao XFF",
			remote: "203.0.113.8:1234",
			xff:    "198.51.100.9",
			want:   "203.0.113.8",
		},
		{
			name:   "proxy tin cay chuyen tiep IP client",
			remote: "10.0.0.2:8080",
			xff:    "198.51.100.9",
			want:   "198.51.100.9",
		},
		{
			name:   "chuoi nhieu proxy",
			remote: "10.0.0.2:8080",
			xff:    "198.51.100.9, 192.168.1.20",
			want:   "198.51.100.9",
		},
		{
			name:   "XFF sai dinh dang bi bo qua",
			remote: "10.0.0.2:8080",
			xff:    "198.51.100.9, forged-value",
			want:   "10.0.0.2",
		},
		{
			name:   "khong co XFF",
			remote: "10.0.0.2:8080",
			want:   "10.0.0.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/login", nil)
			req.RemoteAddr = tt.remote
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}

			got := getClientIPWithTrustedProxies(req, prefixes)
			if got != tt.want {
				t.Errorf("IP = %q, muon %q", got, tt.want)
			}
		})
	}
}

func TestGetClientIPWithoutTrustedProxiesIgnoresXFF(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/login", nil)
	req.RemoteAddr = "10.0.0.2:8080"
	req.Header.Set("X-Forwarded-For", "198.51.100.9")

	if got := getClientIPWithTrustedProxies(req, nil); got != "10.0.0.2" {
		t.Errorf("IP = %q, muon IP ket noi truc tiep", got)
	}
}
