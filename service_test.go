package nape_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/muir/nape"
	"github.com/muir/nject/v2"
	"github.com/stretchr/testify/assert"
)

var (
	// Use a custom Transport because httptest "helpfully" kills idle
	// connections on the default transport when a httptest server shuts
	// down.
	tr = &http.Transport{
		// Disable keepalives to avoid the hassle of closing idle
		// connections after each test.
		DisableKeepAlives: true,
	}
	client = &http.Client{Transport: tr}
)

func TestFallibleInjectorFailing(t *testing.T) {
	t.Parallel()
	var initCount int
	var invokeCount int
	var errorsCount int
	multiStartups(
		t,
		"test",
		nil,
		nject.Sequence("hc",
			func(inner func() error, w http.ResponseWriter) {
				t.Logf("wraper (before)")
				err := inner()
				t.Logf("wraper (after, err=%v)", err)
				if err != nil {
					assert.Equal(t, "bailing out", err.Error())
					errorsCount++
					w.WriteHeader(204)
				}
			},
			func() (nject.TerminalError, int) {
				t.Logf("endpoint init")
				initCount++
				return fmt.Errorf("bailing out"), initCount
			},
			func(w http.ResponseWriter, i int) error {
				t.Logf("endpoint invoke")
				w.WriteHeader(204)
				invokeCount++
				return nil
			},
		),
		func(s string) {
			// reset
			t.Logf("reset for %s", s)
			initCount = 0
			invokeCount = 0
			errorsCount = 0
		},
		func(s string) {
			// after register
			assert.Equal(t, 0, initCount, s+" after register endpoint init count")
			assert.Equal(t, 0, invokeCount, s+" after register endpoint invoke count")
			assert.Equal(t, 0, errorsCount, s+" after register endpoint invoke count")
		},
		func(s string) {
			// after start
			assert.Equal(t, 0, initCount, s+" after start endpoint init count")
			assert.Equal(t, 0, invokeCount, s+" after start endpoint invoke count")
			assert.Equal(t, 0, errorsCount, s+" after register endpoint invoke count")
		},
		func(s string) {
			// after 1st call
			assert.Equal(t, 1, initCount, s+" 1st call start init count")
			assert.Equal(t, 0, invokeCount, s+" 1st call start invoke count")
			assert.Equal(t, 1, errorsCount, s+" after register endpoint invoke count")
		},
		func(s string) {
			// after 2nd call
			assert.Equal(t, 2, initCount, s+" 2nd call endpoint init count")
			assert.Equal(t, 0, invokeCount, s+" 2nd call endpoint invoke count")
			assert.Equal(t, 2, errorsCount, s+" after register endpoint invoke count")
		},
	)
}

func multiStartups(
	t *testing.T,
	name string,
	shc *nject.Collection,
	hc *nject.Collection,
	reset func(string),
	afterRegister func(string), // not called for CreateEnpoint
	afterStart func(string),
	afterCall1 func(string),
	afterCall2 func(string),
) {
	{
		n := name + "-4PreregisterServiceWithMuxRegisterEndpointBeforeStart"
		reset(n)
		ept := "/" + n
		s := nape.PreregisterServiceWithMux(n, shc)
		s.RegisterEndpoint(ept, hc)
		afterRegister(n)
		muxRouter := mux.NewRouter()
		s.Start(muxRouter)
		localServer := httptest.NewServer(muxRouter)
		defer localServer.Close()
		afterStart(n)
		t.Logf("GET %s%s\n", localServer.URL, ept)
		// nolint:noctx
		resp, err := client.Get(localServer.URL + ept)
		assert.NoError(t, err)
		if resp != nil {
			resp.Body.Close()
		}
		afterCall1(n)
		// nolint:noctx
		resp, err = client.Get(localServer.URL + ept)
		assert.NoError(t, err, name)
		if resp != nil {
			assert.Equal(t, 204, resp.StatusCode, name)
			resp.Body.Close()
		}
		afterCall2(n)
	}
	{
		n := name + "-5PreregisterWithMuxRegisterEndpointAfterStartUsingOriginalService"
		reset(n)
		ept := "/" + n
		s := nape.PreregisterServiceWithMux(n, shc)
		muxRouter := mux.NewRouter()
		s.Start(muxRouter)
		localServer := httptest.NewServer(muxRouter)
		defer localServer.Close()
		s.RegisterEndpoint(ept, hc)
		// afterRegister(n)
		afterStart(n)
		t.Logf("GET %s%s\n", localServer.URL, ept)
		// nolint:noctx
		resp, err := client.Get(localServer.URL + ept)
		assert.NoError(t, err)
		if resp != nil {
			resp.Body.Close()
		}
		afterCall1(n)
		// nolint:noctx
		resp, err = client.Get(localServer.URL + ept)
		assert.NoError(t, err, name)
		if resp != nil {
			assert.Equal(t, 204, resp.StatusCode, name)
			resp.Body.Close()
		}
		afterCall2(n)
	}
	{
		n := name + "-6PreregisterWithMuxRegisterEndpointAfterStartUsingStartedService"
		reset(n)
		ept := "/" + n
		s := nape.PreregisterServiceWithMux(n, shc)
		muxRouter := mux.NewRouter()
		sr := s.Start(muxRouter)
		localServer := httptest.NewServer(muxRouter)
		defer localServer.Close()
		sr.RegisterEndpoint(ept, hc)
		// afterRegister(n)
		afterStart(n)
		t.Logf("GET %s%s\n", localServer.URL, ept)
		// nolint:noctx
		resp, err := client.Get(localServer.URL + ept)
		assert.NoError(t, err)
		if resp != nil {
			resp.Body.Close()
		}
		afterCall1(n)
		// nolint:noctx
		resp, err = client.Get(localServer.URL + ept)
		assert.NoError(t, err, name)
		if resp != nil {
			assert.Equal(t, 204, resp.StatusCode, name)
			resp.Body.Close()
		}
		afterCall2(n)
	}
	{
		n := name + "-9RegisterServiceWithMux"
		reset(n)
		ept := "/" + n
		muxRouter := mux.NewRouter()
		s := nape.RegisterServiceWithMux(n, muxRouter, shc)
		s.RegisterEndpoint(ept, hc)
		// afterRegister(n)
		localServer := httptest.NewServer(muxRouter)
		defer localServer.Close()
		afterStart(n)
		t.Logf("GET %s%s\n", localServer.URL, ept)
		// nolint:noctx
		resp, err := client.Get(localServer.URL + ept)
		assert.NoError(t, err)
		if resp != nil {
			resp.Body.Close()
		}
		afterCall1(n)
		// nolint:noctx
		resp, err = client.Get(localServer.URL + ept)
		assert.NoError(t, err, name)
		if resp != nil {
			assert.Equal(t, 204, resp.StatusCode, name)
			resp.Body.Close()
		}
		afterCall2(n)
	}
}

func TestMuxModifiers(t *testing.T) {
	t.Parallel()
	s := nape.PreregisterServiceWithMux("TestCharacterize")

	s.RegisterEndpoint("/x", func(w http.ResponseWriter) {
		w.WriteHeader(204)
	}).Methods("GET")

	s.RegisterEndpoint("/x", func(w http.ResponseWriter) {
		w.WriteHeader(205)
	}).Methods("POST")

	muxRouter := mux.NewRouter()
	s.Start(muxRouter)

	localServer := httptest.NewServer(muxRouter)
	defer localServer.Close()

	// nolint:noctx
	resp, err := client.Get(localServer.URL + "/x")
	if !assert.NoError(t, err) {
		return
	}
	resp.Body.Close()
	assert.Equal(t, 204, resp.StatusCode)

	// nolint:noctx
	resp, err = client.Post(localServer.URL+"/x", "application/json", nil)
	if !assert.NoError(t, err) {
		return
	}
	resp.Body.Close()
	assert.Equal(t, 205, resp.StatusCode)
}
