//nolint:dupl
package jira

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/search", r.URL.Path)
		assert.Equal(t, url.Values{
			"jql": []string{"project=TEST AND status=Done"},
		}, r.URL.Query())
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))

		w.WriteHeader(200)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second), WithInsecureTLS(true))
	resp, err := client.Get(context.Background(), "/search?jql=project=TEST%20AND%20status=Done", Header{
		"Content-Type": "text/plain",
	})

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	_ = resp.Body.Close()
}

func TestGetV1(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/agile/1.0/epic/TEST-1/issue", r.URL.Path)
		assert.Equal(t, url.Values{
			"jql": []string{"project=TEST AND status=Done"},
		}, r.URL.Query())

		w.WriteHeader(200)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))
	resp, err := client.GetV1(context.Background(), "/epic/TEST-1/issue?jql=project=TEST%20AND%20status=Done", nil)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	_ = resp.Body.Close()
}

func TestGetV2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/2/search", r.URL.Path)
		assert.Equal(t, url.Values{
			"jql": []string{"project=TEST AND status=Done"},
		}, r.URL.Query())
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))

		w.WriteHeader(200)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))
	resp, err := client.GetV2(context.Background(), "/search?jql=project=TEST%20AND%20status=Done", Header{
		"Content-Type": "text/plain",
	})

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	_ = resp.Body.Close()
}

func TestPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/issue", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Empty(t, r.Header.Get("X-Requested-By"))

		w.WriteHeader(201)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))
	resp, err := client.Post(context.Background(), "/issue", []byte("hello"), Header{
		"Content-Type":   "application/json",
		"X-Requested-By": "jira-cli",
	})

	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	_ = resp.Body.Close()
}

func TestPostV2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/2/issue", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Empty(t, r.Header.Get("X-Requested-By"))

		w.WriteHeader(201)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))
	resp, err := client.PostV2(context.Background(), "/issue", []byte("hello"), Header{
		"Content-Type": "application/json",
	})

	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	_ = resp.Body.Close()
}

func TestPostV1(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/agile/1.0/issue", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Empty(t, r.Header.Get("X-Requested-By"))

		w.WriteHeader(201)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))
	resp, err := client.PostV1(context.Background(), "/issue", []byte("hello"), Header{
		"Content-Type": "application/json",
	})

	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	_ = resp.Body.Close()
}

func TestPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/issue/TEST-1/assignee", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Empty(t, r.Header.Get("X-Requested-By"))

		w.WriteHeader(204)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))
	resp, err := client.Put(context.Background(), "/issue/TEST-1/assignee", []byte("jon"), Header{
		"Content-Type": "application/json",
	})

	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	_ = resp.Body.Close()
}

func TestPutV2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/2/issue/TEST-1/assignee", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Empty(t, r.Header.Get("X-Requested-By"))

		w.WriteHeader(204)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))
	resp, err := client.PutV2(context.Background(), "/issue/TEST-1/assignee", []byte("jon"), Header{
		"Content-Type": "application/json",
	})

	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	_ = resp.Body.Close()
}

func TestCookieAuthAddsCookieHeader(t *testing.T) {
	cookieAuth := AuthTypeCookie
	cookies := "cloud.session.token=abc; atlassian.xsrf.token=xsrf-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/search", r.URL.Path)
		assert.Equal(t, cookies, r.Header.Get("Cookie"))
		assert.NotEmpty(t, r.Header.Get("User-Agent"))
		assert.Empty(t, r.Header.Get("atl-xsrf-token"))

		w.WriteHeader(200)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL, AuthType: &cookieAuth, Cookies: cookies}, WithTimeout(3*time.Second))
	resp, err := client.Get(context.Background(), "/search", nil)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	_ = resp.Body.Close()
}

func TestCookieAuthAddsXSRFHeaderForMutatingRequests(t *testing.T) {
	cookieAuth := AuthTypeCookie
	cookies := "cloud.session.token=abc; atlassian.xsrf.token=xsrf-123; other=value"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/issue", r.URL.Path)
		assert.Equal(t, cookies, r.Header.Get("Cookie"))
		assert.Equal(t, "xsrf-123", r.Header.Get("atl-xsrf-token"))
		expectedOrigin := "http://" + r.Host
		assert.Equal(t, expectedOrigin, r.Header.Get("Origin"))
		assert.Equal(t, expectedOrigin, r.Header.Get("Referer"))
		assert.Equal(t, browserUserAgent, r.Header.Get("User-Agent"))
		assertBrowserFingerprint(t, r)

		w.WriteHeader(201)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL, AuthType: &cookieAuth, Cookies: cookies}, WithTimeout(3*time.Second))
	resp, err := client.Post(context.Background(), "/issue", []byte("{}"), nil)

	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	_ = resp.Body.Close()
}

func TestCookieAuthWithoutXSRFCookieDoesNotAddXSRFHeader(t *testing.T) {
	cookieAuth := AuthTypeCookie
	cookies := "cloud.session.token=abc"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, cookies, r.Header.Get("Cookie"))
		assert.Empty(t, r.Header.Get("atl-xsrf-token"))
		expectedOrigin := "http://" + r.Host
		assert.Equal(t, expectedOrigin, r.Header.Get("Origin"))
		assert.Equal(t, expectedOrigin, r.Header.Get("Referer"))
		assertBrowserFingerprint(t, r)

		w.WriteHeader(201)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL, AuthType: &cookieAuth, Cookies: cookies}, WithTimeout(3*time.Second))
	resp, err := client.Post(context.Background(), "/issue", []byte("{}"), nil)

	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	_ = resp.Body.Close()
}

func assertBrowserFingerprint(t *testing.T, r *http.Request) {
	t.Helper()
	for header, expected := range map[string]string{
		"Accept-Language":    "en-GB,en-US;q=0.9,en;q=0.8",
		"Sec-Ch-Ua":          `"Chromium";v="148", "Google Chrome";v="148", "Not/A)Brand";v="99"`,
		"Sec-Ch-Ua-Mobile":   "?0",
		"Sec-Ch-Ua-Platform": `"macOS"`,
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-origin",
		"Priority":           "u=1, i",
	} {
		assert.Equalf(t, expected, r.Header.Get(header), "mismatch for header %q", header)
	}
}

func TestDeleteV2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/2/issue/TEST-1", r.URL.Path)
		assert.Empty(t, r.Header.Get("X-Requested-By"))

		w.WriteHeader(204)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))
	resp, err := client.DeleteV2(context.Background(), "/issue/TEST-1", Header{})

	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	_ = resp.Body.Close()
}
