package main

import "testing"

// TestValidDNSServers keeps only what mihomo can actually use. A settings.json
// typo must not cost the user their connection, so unparsable entries are
// dropped rather than handed to the renderer.
func TestValidDNSServers(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"", ""},
		{"  ", ""},
		{"1.1.1.1", "1.1.1.1"},
		{"1.1.1.1, 8.8.8.8", "1.1.1.1, 8.8.8.8"},
		{"1.1.1.1,not-an-ip,8.8.8.8", "1.1.1.1, 8.8.8.8"},
		{"garbage", ""},
		{"2606:4700:4700::1111", "2606:4700:4700::1111"},
		{"1.1.1.1;8.8.8.8", "1.1.1.1, 8.8.8.8"},
		// Duplicates collapse, so the rendered list cannot repeat an entry.
		{"1.1.1.1,1.1.1.1", "1.1.1.1"},
		// YAML / shell punctuation must not survive into the nameserver list.
		{"1.1.1.1\n\t 8.8.8.8", "1.1.1.1, 8.8.8.8"},
	}
	for _, c := range cases {
		if got := validDNSServers(c.raw); got != c.want {
			t.Errorf("validDNSServers(%q) = %q, want %q", c.raw, got, c.want)
		}
	}
}

// TestValidDomainToken pins the accepted suffix shape. Every one of these
// becomes part of a DOMAIN-SUFFIX rule in a generated YAML file, so anything
// outside the narrow set is rejected.
func TestValidDomainToken(t *testing.T) {
	ok := []string{
		"example.com",
		"sub.example.co.uk",
		"*.example.com",
		"a-b_c.example",
		"localhost",
		"xn--punycode.example",
	}
	for _, s := range ok {
		if !validDomainToken(s) {
			t.Errorf("validDomainToken(%q) = false, want true", s)
		}
	}

	bad := []string{
		"",
		".",
		"example.com.",
		".example.com",
		"a..b",
		"*.",
		"||example.com",
		"example.com/path",
		"example:8080",
		"exa mple.com",
		"example,com",
		"\"example.com\"",
		"-example.com",
		"example.com:",
		"exa$mple.com",
		"a\nb",
	}
	for _, s := range bad {
		if validDomainToken(s) {
			t.Errorf("validDomainToken(%q) = true, want false", s)
		}
	}
}

// TestValidDomainList checks the list-level behaviour: bad entries are dropped,
// good ones survive in the renderer's comma+space form.
func TestValidDomainList(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"", ""},
		{"example.com", "example.com"},
		{"example.com,||bad.com,other.org", "example.com, other.org"},
		{"example.com,example.com", "example.com"},
		{"*.a.com,*.b.com", "*.a.com, *.b.com"},
		{"bad/path,also-bad:53", ""},
	}
	for _, c := range cases {
		if got := validDomainList(c.raw); got != c.want {
			t.Errorf("validDomainList(%q) = %q, want %q", c.raw, got, c.want)
		}
	}
}

// TestNormalizePortCollision pins the top-of-range case. The two ports must
// differ, and stepping *up* from 65535 overflows into an invalid port.
func TestNormalizePortCollision(t *testing.T) {
	cases := []struct {
		mixed, api         int
		wantMixed, wantAPI int
	}{
		{7890, 7890, 7890, 7891},
		{65535, 65535, 65535, 65534},
		{7890, 9090, 7890, 9090},
		// Out-of-range values fall back to the defaults first.
		{0, 0, 7890, 9090},
		{-5, 99999, 7890, 9090},
	}
	for _, c := range cases {
		s := clientSettings{MixedPort: c.mixed, APIPort: c.api}
		s.normalize()
		if s.MixedPort != c.wantMixed || s.APIPort != c.wantAPI {
			t.Errorf("normalize(%d,%d) = (%d,%d), want (%d,%d)",
				c.mixed, c.api, s.MixedPort, s.APIPort, c.wantMixed, c.wantAPI)
		}
		if s.MixedPort == s.APIPort {
			t.Errorf("normalize(%d,%d) left both ports equal at %d", c.mixed, c.api, s.MixedPort)
		}
		if s.APIPort < 1 || s.APIPort > 65535 {
			t.Errorf("normalize(%d,%d) produced an out-of-range API port %d", c.mixed, c.api, s.APIPort)
		}
	}
}
