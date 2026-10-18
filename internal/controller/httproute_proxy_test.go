package controller

import (
	"testing"

	"github.com/netbirdio/netbird/shared/management/http/api"
)

func TestProxyServiceMatches(t *testing.T) {
	target := func(id string) api.ServiceTarget {
		return api.ServiceTarget{
			Enabled:    true,
			TargetId:   id,
			Protocol:   api.ServiceTargetProtocolHttp,
			TargetType: api.ServiceTargetTargetTypeHost,
		}
	}
	req := func(targets ...api.ServiceTarget) api.ServiceRequest {
		return api.ServiceRequest{
			Domain:           "app.example.com",
			Enabled:          true,
			Name:             "app.example.com",
			Mode:             new(api.ServiceRequestModeHttp),
			PassHostHeader:   new(false),
			RewriteRedirects: new(false),
			Targets:          &targets,
		}
	}
	// What the server returns for req("a", "b"): it fills in each
	// target's host and port, metadata and fields the request left unset.
	existing := func() api.Service {
		host := "10.0.0.1"
		a, b := target("a"), target("b")
		a.Host, a.Port = &host, 8080
		b.Host = &host
		return api.Service{
			Id:               "svc",
			Domain:           "app.example.com",
			Enabled:          true,
			Name:             "app.example.com",
			Mode:             new(api.ServiceModeHttp),
			PassHostHeader:   new(false),
			RewriteRedirects: new(false),
			Private:          new(false),
			Targets:          []api.ServiceTarget{b, a},
		}
	}

	tests := []struct {
		name   string
		mutate func(*api.Service)
		req    api.ServiceRequest
		want   bool
	}{
		{name: "unchanged, targets in another order", req: req(target("a"), target("b")), want: true},
		{name: "unset bools read as false", mutate: func(s *api.Service) {
			s.PassHostHeader, s.RewriteRedirects = nil, nil
		}, req: req(target("a"), target("b")), want: true},
		{name: "target added", req: req(target("a"), target("b"), target("c")), want: false},
		{name: "target replaced", req: req(target("a"), target("c")), want: false},
		{name: "target disabled", mutate: func(s *api.Service) {
			s.Targets[0].Enabled = false
		}, req: req(target("a"), target("b")), want: false},
		{name: "pass host header turned on", mutate: func(s *api.Service) {
			s.PassHostHeader = new(true)
		}, req: req(target("a"), target("b")), want: false},
		{name: "service disabled", mutate: func(s *api.Service) {
			s.Enabled = false
		}, req: req(target("a"), target("b")), want: false},
		{name: "renamed", mutate: func(s *api.Service) {
			s.Name = "other"
		}, req: req(target("a"), target("b")), want: false},
		{name: "mode missing", mutate: func(s *api.Service) {
			s.Mode = nil
		}, req: req(target("a"), target("b")), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := existing()
			if tt.mutate != nil {
				tt.mutate(&svc)
			}
			if got := proxyServiceMatches(svc, tt.req); got != tt.want {
				t.Errorf("proxyServiceMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}
