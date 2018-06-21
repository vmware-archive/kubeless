/*
Copyright (c) 2016-2017 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kafka

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestGetTLSConfiguration(t *testing.T) {
	caFile := buildCacertTempFile(t)
	defer os.Remove(caFile.Name())
	certFile := buildCertTempFile(t)
	defer os.Remove(certFile.Name())
	keyFile := buildKeyTempFile(t)
	defer os.Remove(keyFile.Name())
	fakeFile := buildFakeTempFile(t)
	defer os.Remove(fakeFile.Name())

	type testInput struct {
		caFile   string
		certFile string
		keyFile  string
		insecure string
	}
	type tlsTest struct {
		input    testInput
		expected string
	}

	var tlsTests = []tlsTest{
		{
			input: testInput{
				caFile:   "",
				certFile: certFile.Name(),
				keyFile:  keyFile.Name(),
				insecure: "",
			},
			expected: "",
		},
		{
			input: testInput{
				caFile:   "",
				certFile: "",
				keyFile:  keyFile.Name(),
				insecure: "",
			},
			expected: ErrorCrtFileMandatory.Error(),
		},
		{
			input: testInput{
				caFile:   "",
				certFile: certFile.Name(),
				keyFile:  "",
				insecure: "",
			},
			expected: ErrorKeyFileMandatory.Error(),
		},
		{
			input: testInput{
				caFile:   "",
				certFile: "",
				keyFile:  "",
				insecure: "true",
			},
			expected: "",
		},
		{
			input: testInput{
				caFile:   "",
				certFile: "",
				keyFile:  "",
				insecure: "",
			},
			expected: "",
		},
		{
			input: testInput{
				caFile:   "",
				certFile: "",
				keyFile:  "",
				insecure: "false",
			},
			expected: "",
		},
		{
			input: testInput{
				caFile:   caFile.Name(),
				certFile: "",
				keyFile:  "",
				insecure: "",
			},
			expected: "",
		},
		{
			input: testInput{
				caFile:   fakeFile.Name(),
				certFile: "",
				keyFile:  "",
				insecure: "",
			},
			expected: "",
		},
		{
			input: testInput{
				caFile:   "a_clearly_non_existent_path",
				certFile: "",
				keyFile:  "",
				insecure: "",
			},
			expected: "no such file or directory",
		},
		{
			input: testInput{
				caFile:   "",
				certFile: certFile.Name(),
				keyFile:  "a_clearly_non_existent_path",
				insecure: "",
			},
			expected: "no such file or directory",
		},
		{
			input: testInput{
				caFile:   "",
				certFile: "a_clearly_non_existent_path",
				keyFile:  keyFile.Name(),
				insecure: "",
			},
			expected: "no such file or directory",
		},
		{
			input: testInput{
				caFile:   "",
				certFile: certFile.Name(),
				keyFile:  fakeFile.Name(),
				insecure: "",
			},
			expected: "failed to find any PEM data in key input",
		},
		{
			input: testInput{
				caFile:   "",
				certFile: fakeFile.Name(),
				keyFile:  keyFile.Name(),
				insecure: "",
			},
			expected: "failed to find any PEM data in certificate input",
		},
	}

	for _, test := range tlsTests {
		_, err := GetTLSConfiguration(test.input.caFile, test.input.certFile, test.input.keyFile, test.input.insecure)
		if err != nil && !strings.Contains(err.Error(), test.expected) {
			t.Errorf("GetSASLConfiguration expected (%s) for input (%+v) got (%s)", test.expected, test.input, err)
		}
	}
}

func TestGetSASLConfiguration(t *testing.T) {
	type testInput struct {
		user     string
		password string
	}

	type testExpected struct {
		user     string
		password string
		err      error
	}

	type saslTest struct {
		input    testInput
		expected testExpected
	}

	var saslTests = []saslTest{
		{
			input: testInput{
				user:     "",
				password: "",
			},
			expected: testExpected{
				user:     "",
				password: "",
				err:      ErrorUsernameOrPasswordMandatory,
			},
		},
		{
			input: testInput{
				user:     "test",
				password: "",
			},
			expected: testExpected{
				user:     "",
				password: "",
				err:      ErrorUsernameOrPasswordMandatory,
			},
		},
		{
			input: testInput{
				user:     "",
				password: "test",
			},
			expected: testExpected{
				user:     "",
				password: "",
				err:      ErrorUsernameOrPasswordMandatory,
			},
		},
		{
			input: testInput{
				user:     "test",
				password: "test",
			},
			expected: testExpected{
				user:     "test",
				password: "test",
				err:      nil,
			},
		},
	}

	for _, test := range saslTests {
		user, password, err := GetSASLConfiguration(test.input.user, test.input.password)
		if user != test.expected.user ||
			password != test.expected.password ||
			err != test.expected.err {
			t.Errorf("GetSASLConfiguration expected (%s, %s, %s) for input (%+v) got (%s, %s, %s)",
				test.expected.user,
				test.expected.password,
				test.expected.err,
				test.input,
				user,
				password,
				err,
			)
		}
	}
}

func buildCacertTempFile(t *testing.T) *os.File {
	caFile, err := ioutil.TempFile(os.TempDir(), "cacert")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.WriteString(caFile, `
-----BEGIN CERTIFICATE-----
MIIEljCCAn4CCQCY1/mCx2HLBzANBgkqhkiG9w0BAQsFADANMQswCQYDVQQGEwJV
UzAeFw0xODA2MTExOTIwMDBaFw0yMzA2MTExOTIwMDBaMA0xCzAJBgNVBAYTAlVT
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAvsDAuV5mWEOEN8JPUsYi
tuJSxGkJoNmXStKHRbIhAI2gDzEyZ09CTZoJoS+gJ6sabYCknU5vbhBFM6+bH+TA
mPgA9jpzS2qTXqz/14sFyKBF8ILl0D3zDDXrmJKVZ7vlrqg5mTjRhIJpC6q5yLCb
vJ+PzuBkCW4Bp3ugblNV9tkb37CDPC8lTyASEieiaayXu281Ee7H5qsjBhxqNDkS
kN1JUZlc2JYCKKXBUnIM2PbxTa4plSTKxR7ctHxYC4uc9fdjh9dUgAlbz6xnFOOH
Ukzq4tYUFbxn1knDIaDpE5BYMCLYLwi36qnpNGk4wsQaOQPVg0QJs8ooN9Ynlwng
IC8J/15KMkUHoi1ewX/S46mWTkJcKBnrzQanGEh4SIk3U1QCSBqYKlqTr0susCdP
dD9KbwM07KnwWsHDd9f7VwlTApQXHUOwrmNrfH8GXMfZU0ftqtGqtbgXLJuo7a36
pVKS2rVK8HUfVOJonFmUyAtct+hhjdBjJQgqa3GzYMKd5gNpGlunmZBUraJjes5F
Fn+hNeI8VUJ5C+eau7SOnWD7ytvqwrCGZ02z3Ov1ZsBc0xIM4/g16vTIiMj9OzXF
Bt2cvUjnmEhhE6eNo1PXJaXkprmAu0+j7MINyc3ZbKdH2I65+pNSpC7yC0X3rj+/
4U5JhG+ZwKQNgFvrv3ouSLcCAwEAATANBgkqhkiG9w0BAQsFAAOCAgEAeLMZehwI
iU/JsR9GP4vl7knGXae/3mr8x/i3suaiXtKf0rt74pv11TztDiqN3+tgRYrgV42I
nrXpIven+omjfFweVxQvun5rz0fAkCvoDOxk7gzQsj3n0IWfNucp/MWLJTMsvAxA
44G49HZbv3zSMBypWcE07aZ7SkJxppJU6Xk7/P07eDzK9xBKN3UoTu10F1svXCnt
EwBuM18f2ObzZg654A/MwxsNfigfrp0sMnDelWZn4eRzuCrRoQVaTTypxVlFJ4eQ
lac1cimLJ6oVodl0DsJfMGzIQPlP5W21rKLex7lQW1KAKpuQvsgmRzJBPD7EjbKa
Kt/I3Gb44gW3yPBdTu5Fot0lqcG8urgosC+r9wcVQhIaXvCFxrsTgwRF5O+S3feg
qa5zA97IIkiqMMjdhNrzYgF94Ous2aQ6ZGS0FTT0/Z6DEdk0Dimq0j1JgG0Fya5s
XNuHR8JMhmlWSCcA1g/C6PC10W5n+59pia19XqZwtmXlLZW8xXOV8P1UtBLWGDFU
W66y6OiGkZyL6UzLiDJStiyzCBjY/SYokEXrAftuQIbqNmnXOvQ3huFeXUpqkTcq
9clKlhFQXmjjdvLWiCPyYrOaAlYp76mEEEjHt8b8PwMLUTc+ZzVnnKfaka9kB4kE
vxkjzUyFO0zKQhVvLRnjdhTLoj3QhfL2jNc=
-----END CERTIFICATE-----
`)
	if err != nil {
		t.Fatal(err)
	}
	return caFile
}

func buildCertTempFile(t *testing.T) *os.File {
	certFile, err := ioutil.TempFile(os.TempDir(), "cert")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.WriteString(certFile, `
-----BEGIN CERTIFICATE-----
MIIEjjCCAnYCAQEwDQYJKoZIhvcNAQEFBQAwDTELMAkGA1UEBhMCVVMwHhcNMTgw
NjExMTkyMDQ2WhcNMjAwNjEwMTkyMDQ2WjANMQswCQYDVQQGEwJVUzCCAiIwDQYJ
KoZIhvcNAQEBBQADggIPADCCAgoCggIBALY7Wk0K7VHwULybgy16icGGieQdkcOp
MmMeTvMkNRjSIIwYevH50xIWcprdfePBTlySYNXIyVQY3ZLuOlvVWzLXIY0TiKil
ZRDM+s4MFP2MtdcpKtRVPnIDAHcdoOdNkPXv+l0XUQ/+AcTY4Bv1Cgyi1CGF8lAY
UPPd/n1Vo1bGDPsK5scqHHcazaHXkWFBnaTwxuyyrS2JSQIz30HDygjv8rU8pCeT
C//weXRn0yLZ+7917zYabBGFubr6YrJFXX383xS2EzDLLp7gvxxVHHF7O9V8gCMN
tYOR87sJq9mrjTM/VttJetwQeNoG6C9OnV3KZ1WZQbhJE/jA6oo6HPwZIsO04AgL
Owb0WpPuDZMpcP7co17fib9MrbnRn/3OZBnw57B3mVGF+wZOmNSjIeKml7Xk3LiS
C8acRlDwtWvi1EP9XMczFq0FIYwjZDl1TKsirjHerJ9vhrxsAhVlPJqiyDA7pdFt
xVclAJLjzk0M0JCjfnytr+jG//1kAqtCjObgfv5Gef8+3GYgWvSkekznCTcqUGcQ
ib5MatF8jA2EK3m7XwfUf9erLIybgYn2q5+I0kMQnn5vX/1JXKR3+Ai8LUO5EZxt
5kSS5r+US+eQbibUZOCgVciTVrhADMZ5cjpU1nPFAVZT9f11u3ZONbmS+hGiIen6
8H828FjEoOZ3AgMBAAEwDQYJKoZIhvcNAQEFBQADggIBALZBcD+zHi4+UMXzihDp
nnkQw+uoI9Q0tv9dF3BsQme4Pfo0cAYvZK0Uez10YJHEqDn/1PgyefU5UctCh1Ht
u3VCJr39I7NecGA4w+QU5rBvXpUJpZyiGFWofa88e5OXoUrINt9ENwIYbH2J7FBl
A604yAz6lWWykkS+weuA/n76+38CjGAcrnjOkittFqK4McYcZafANa5SAUv5U2lz
CXYHdXpbdiRYyvWWySqXP8P6wJ/rYL8yOXVqUWk3KXRHDN21xnAd9laMi15nv4iI
42q65yCAKi95ztdqQgbz5xKbZt5jLikQz/5lLtTv6u3aRlKxLCjRz2gEJqPtnjlr
HNZ6Ae91s0EnZIxAMoPuVsEZWpLrkr2b0vubUbPQ6gHKpTSW81oRmoC+HDs3FPFk
iqbpolAfOr/qvje7aRuMXhhrxLGkJ9dFqsGN4gU544wcjPuSZuQ8Df1BZjK09wCo
L/3rlm+2/P0QS7kCMBZoC0RjBXgGCqfHqvRzc511w8Ek0SEAsNlObeDI2WfvGxjr
ge7JSw24OjqEN4cIHv6iFF/ZEENHyVpOlC8U80P5oY9mm6zSIOJOisSTU76IrUmG
+HxXzRRRFYnYjLYgjOR2CqC/eeT4SjyOVJikQ7xgZSvCxq4kCt/h3BOr76m43K7b
PK4GeoEIzKHtUCFGTnegj9ox
-----END CERTIFICATE-----
`)
	if err != nil {
		t.Fatal(err)
	}
	return certFile
}

func buildKeyTempFile(t *testing.T) *os.File {
	keyFile, err := ioutil.TempFile(os.TempDir(), "key")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.WriteString(keyFile, `
-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`)
	if err != nil {
		t.Fatal(err)
	}
	return keyFile
}

func buildFakeTempFile(t *testing.T) *os.File {
	fakeFile, err := ioutil.TempFile(os.TempDir(), "fake")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.WriteString(fakeFile, `
testing`)
	if err != nil {
		t.Fatal(err)
	}
	return fakeFile
}
