package crawlhttp

import "testing"

func TestValidateURLRejectsPrivateAndOddSchemes(t *testing.T) {
	tests := []struct {
		raw     string
		wantErr bool
	}{
		{"https://darukade.com/products", false},
		{"http://example.com", false},
		{"ftp://example.com", true},
		{"https://127.0.0.1/secret", true},
		{"http://192.168.1.10/", true},
		{"http://10.0.0.5/", true},
		{"file:///etc/passwd", true},
		{"https://", true},
	}
	for _, tc := range tests {
		err := ValidateURL(tc.raw)
		if tc.wantErr && err == nil {
			t.Errorf("ValidateURL(%q) should fail", tc.raw)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("ValidateURL(%q) = %v", tc.raw, err)
		}
	}
}

func TestRobotsAllowsStarAllowSlash(t *testing.T) {
	file := parseRobots("User-agent: *\nAllow: /\n")
	if !file.Allows("/products/x") {
		t.Fatal("Allow: / should permit /products/x")
	}
}

func TestClassifyTransientVsStructural(t *testing.T) {
	if !IsTransient(Transient(errSample("timeout"))) {
		t.Fatal("wrapped timeout should be transient")
	}
	if !IsStructural(Structural(errSample("status 404"))) {
		t.Fatal("404 should be structural")
	}
	if IsTransient(Structural(errSample("bad html"))) {
		t.Fatal("structural must not look transient")
	}
	if !IsStructural(errSample("unexpected token")) {
		t.Fatal("unclassified parse error is structural")
	}
}

type errSample string

func (e errSample) Error() string { return string(e) }

func TestRobotsDisallowPrefix(t *testing.T) {
	file := parseRobots("User-agent: *\nDisallow: /admin\n")
	if file.Allows("/admin/users") {
		t.Fatal("disallow /admin should block /admin/users")
	}
	if !file.Allows("/products") {
		t.Fatal("/products should still be allowed")
	}
}
