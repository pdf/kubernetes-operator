package controller

import (
	"github.com/netbirdio/netbird/shared/management/http/api"
)

// proxyServiceMatches reports whether an existing reverse proxy service
// already holds everything req sets, so updating it would change nothing.
// Fields the server fills in itself (target host and port, metadata) and
// fields req leaves unset are not compared.
func proxyServiceMatches(svc api.Service, req api.ServiceRequest) bool {
	if svc.Name != req.Name || svc.Enabled != req.Enabled {
		return false
	}
	if req.Mode != nil && (svc.Mode == nil || string(*svc.Mode) != string(*req.Mode)) {
		return false
	}
	if !boolPtrMatches(svc.PassHostHeader, req.PassHostHeader) ||
		!boolPtrMatches(svc.RewriteRedirects, req.RewriteRedirects) {
		return false
	}
	if req.Targets == nil {
		return true
	}
	return targetsMatch(svc.Targets, *req.Targets)
}

// boolPtrMatches treats an unset want as matching anything, and an unset
// got as false, the API's default.
func boolPtrMatches(got, want *bool) bool {
	if want == nil {
		return true
	}
	return (got != nil && *got) == *want
}

type targetKey struct {
	targetID   string
	targetType api.ServiceTargetTargetType
	protocol   api.ServiceTargetProtocol
	enabled    bool
	path       string
}

func keyOf(t api.ServiceTarget) targetKey {
	k := targetKey{
		targetID:   t.TargetId,
		targetType: t.TargetType,
		protocol:   t.Protocol,
		enabled:    t.Enabled,
	}
	if t.Path != nil {
		k.path = *t.Path
	}
	return k
}

// targetsMatch compares two target lists regardless of order.
func targetsMatch(got, want []api.ServiceTarget) bool {
	if len(got) != len(want) {
		return false
	}
	counts := map[targetKey]int{}
	for _, t := range got {
		counts[keyOf(t)]++
	}
	for _, t := range want {
		k := keyOf(t)
		if counts[k] == 0 {
			return false
		}
		counts[k]--
	}
	return true
}
