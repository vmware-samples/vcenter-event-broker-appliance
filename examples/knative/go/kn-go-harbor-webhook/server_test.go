package main

import (
	"context"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	nethttp "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/embano1/vsphere/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"gotest.tools/v3/assert"
)

type mockClient struct {
	ce.Client

	t        *testing.T
	requests int32
	code     int
}

func (m *mockClient) Send(_ context.Context, event ce.Event) protocol.Result {
	assert.NilError(m.t, event.Validate())

	var inputEvent map[string]interface{}
	err := json.Unmarshal([]byte(harborEvent), &inputEvent)
	assert.NilError(m.t, err)

	assert.Equal(m.t, inputEvent["occur_at"], float64(event.Time().Unix()))

	eventType := strings.ToLower(inputEvent["type"].(string))
	assert.Assert(m.t, strings.Contains(event.Type(), eventType))

	// send mock response
	atomic.AddInt32(&m.requests, 1)
	if m.code != 200 {
		return http.NewResult(m.code, "")
	}
	return nil // ACK
}

func Test_run(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctx = logger.Set(ctx, zaptest.NewLogger(t))
	dir, err := ioutil.TempDir("", "secret")
	assert.NilError(t, err)

	err = ioutil.WriteFile(filepath.Join(dir, userFileKey), []byte("user"), fs.ModePerm)
	assert.NilError(t, err)

	err = ioutil.WriteFile(filepath.Join(dir, passwordFileKey), []byte("pass"), fs.ModePerm)
	assert.NilError(t, err)

	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			logger.Get(ctx).Error("could not clean up temp directory", zap.Error(err))
		}
	})

	t.Setenv("WEBHOOK_SECRET_PATH", dir)

	cfg := config{
		Address:    "127.0.0.1",
		Path:       "/webhook",
		Port:       8080,
		Service:    "testservice",
		Sink:       "somesink",
		Debug:      true,
		SecretPath: dir,
	}

	err = run(ctx, cfg)
	assert.NilError(t, err)
}

func Test_eventhandler(t *testing.T) {
	type basicAuth struct {
		username string
		password string
	}

	basicAuthCredentials := basicAuth{
		username: "user",
		password: "pass",
	}

	tests := []struct {
		name      string
		auth      *basicAuth
		code      int
		method    string
		wantCount int32
		wantCode  int
	}{
		{
			name:      "successfully sends event (no auth)",
			code:      200,
			method:    nethttp.MethodPost,
			wantCount: 1,
			wantCode:  200,
		},
		{
			name:      "fails with 405 if method is not POST",
			code:      0,
			method:    nethttp.MethodGet,
			wantCount: 0,
			wantCode:  405,
		},
		{
			name: "fails with 401 not authorized",
			auth: &basicAuth{
				username: "userrrrr",
				password: "passssssss",
			},
			code:      0,
			method:    nethttp.MethodPost,
			wantCount: 0,
			wantCode:  401,
		},
		{
			name: "successfully sends event (basic auth)",
			auth: &basicAuth{
				username: "user",
				password: "pass",
			},
			code:      200,
			method:    nethttp.MethodPost,
			wantCount: 1,
			wantCode:  200,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := logger.Set(context.Background(), zaptest.NewLogger(t))
			t.Setenv("K_SERVICE", "testservice")

			// overwrite retry config
			retries = 3
			retryDelay = time.Millisecond

			mc := mockClient{
				t:    t,
				code: tc.code,
			}

			testHandler := eventHandler(ctx, &mc)
			req := httptest.NewRequest(tc.method, "/webhook", strings.NewReader(harborEvent))

			if tc.auth != nil {
				t.Setenv("WEBHOOK_SECRET_PATH", "/somesecret")
				testHandler = withBasicAuth(ctx, eventHandler(ctx, &mc), basicAuthCredentials.username, basicAuthCredentials.password)
				req.SetBasicAuth(tc.auth.username, tc.auth.password)
			}

			rec := httptest.NewRecorder()
			testHandler.ServeHTTP(rec, req)

			count := atomic.LoadInt32(&mc.requests)
			assert.Equal(t, tc.wantCount, count)
			assert.Equal(t, tc.wantCode, rec.Code)
		})
	}
}

const harborEvent = `
{
  "type": "PULL_ARTIFACT",
  "occur_at": 1655887788,
  "operator": "admin",
  "event_data": {
    "resources": [
      {
        "digest": "sha256:3b465cbcadf7d437fc70c3b6aa2c93603a7eef0a3f5f1e861d91f303e4aabdee",
        "tag": "sha256:3b465cbcadf7d437fc70c3b6aa2c93603a7eef0a3f5f1e861d91f303e4aabdee",
        "resource_url": "harbor-app.jarvis.tanzu/veba-test/csi-test@sha256:3b465cbcadf7d437fc70c3b6aa2c93603a7eef0a3f5f1e861d91f303e4aabdee"
      }
    ],
    "repository": {
      "date_created": 1655887764,
      "name": "csi-test",
      "namespace": "veba-test",
      "repo_full_name": "veba-test/csi-test",
      "repo_type": "public"
    }
  }
}`
