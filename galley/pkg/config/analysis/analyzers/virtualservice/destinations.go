// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package virtualservice

import (
	"fmt"
	"regexp"

	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/galley/pkg/config/analysis"
	"istio.io/istio/galley/pkg/config/analysis/msg"
	"istio.io/istio/galley/pkg/config/collection"
	"istio.io/istio/galley/pkg/config/processor/metadata"
	"istio.io/istio/galley/pkg/config/resource"
)

var (
	fqdnPattern = regexp.MustCompile(`^(.+)\.(.+)\.svc\.cluster\.local$`)
)

// DestinationAnalyzer checks the destinations associated with each virtual service
type DestinationAnalyzer struct{}

var _ analysis.Analyzer = &DestinationAnalyzer{}

type hostAndSubset struct {
	host   resource.Name
	subset string
}

// Metadata implements Analyzer
func (da *DestinationAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name: "virtualservice.DestinationAnalyzer",
		Inputs: collection.Names{
			metadata.IstioNetworkingV1Alpha3Virtualservices,
			metadata.IstioNetworkingV1Alpha3Destinationrules,
		},
	}
}

// Analyze implements Analyzer
func (da *DestinationAnalyzer) Analyze(ctx analysis.Context) {
	// To avoid repeated iteration, precompute the set of existing destination host+subset combinations
	destHostsAndSubsets := initDestHostsAndSubsets(ctx)

	ctx.ForEach(metadata.IstioNetworkingV1Alpha3Virtualservices, func(r *resource.Entry) bool {
		da.analyzeVirtualService(r, ctx, destHostsAndSubsets)
		return true
	})
}

func (da *DestinationAnalyzer) analyzeVirtualService(r *resource.Entry, ctx analysis.Context,
	destHostsAndSubsets map[hostAndSubset]bool) {

	vs := r.Item.(*v1alpha3.VirtualService)
	ns, _ := r.Metadata.Name.InterpretAsNamespaceAndName()

	destinations := getRouteDestinations(vs)

	for _, destination := range destinations {
		// Disabled checkDestinationHost in 1.3 backport since our handling of host discovery is not mature enough in the case
		// where users are running `istioctl experimental analyze` with files only. Leaving this in means
		// a lot of false positives. (Unused code paths removed)
		if !da.checkDestinationSubset(ns, destination, destHostsAndSubsets) {
			ctx.Report(metadata.IstioNetworkingV1Alpha3Virtualservices,
				msg.NewReferencedResourceNotFound(r, "host+subset in destinationrule", fmt.Sprintf("%s+%s", destination.GetHost(), destination.GetSubset())))
		}
	}
}

func (da *DestinationAnalyzer) checkDestinationSubset(vsNamespace string, destination *v1alpha3.Destination, destHostsAndSubsets map[hostAndSubset]bool) bool {
	name := getResourceNameFromHost(vsNamespace, destination.GetHost())

	subset := destination.GetSubset()

	// if there's no subset specified, we're done
	if subset == "" {
		return true
	}

	hs := hostAndSubset{
		host:   name,
		subset: subset,
	}
	if _, ok := destHostsAndSubsets[hs]; ok {
		return true
	}

	return false
}

func initDestHostsAndSubsets(ctx analysis.Context) map[hostAndSubset]bool {
	hostsAndSubsets := make(map[hostAndSubset]bool)
	ctx.ForEach(metadata.IstioNetworkingV1Alpha3Destinationrules, func(r *resource.Entry) bool {
		dr := r.Item.(*v1alpha3.DestinationRule)
		drNamespace, _ := r.Metadata.Name.InterpretAsNamespaceAndName()

		for _, ss := range dr.GetSubsets() {
			hs := hostAndSubset{
				host:   getResourceNameFromHost(drNamespace, dr.GetHost()),
				subset: ss.GetName(),
			}
			hostsAndSubsets[hs] = true
		}
		return true
	})
	return hostsAndSubsets
}

// getResourceNameFromHost figures out the resource.Name to look up from the provided host string
// We need to handle two possible formats: short name and FQDN
// https://istio.io/docs/reference/config/networking/v1alpha3/virtual-service/#Destination
func getResourceNameFromHost(defaultNamespace, host string) resource.Name {

	// First, try to parse as FQDN (which can be cross-namespace)
	namespace, name := getNamespaceAndNameFromFQDN(host)

	//Otherwise, treat this as a short name and use the assumed namespace
	if namespace == "" {
		namespace = defaultNamespace
		name = host
	}
	return resource.NewName(namespace, name)
}

func getNamespaceAndNameFromFQDN(fqdn string) (string, string) {
	result := fqdnPattern.FindAllStringSubmatch(fqdn, -1)
	if len(result) == 0 {
		return "", ""
	}
	return result[0][2], result[0][1]
}

func getRouteDestinations(vs *v1alpha3.VirtualService) []*v1alpha3.Destination {
	destinations := make([]*v1alpha3.Destination, 0)

	for _, r := range vs.GetTcp() {
		for _, rd := range r.GetRoute() {
			destinations = append(destinations, rd.GetDestination())
		}
	}
	for _, r := range vs.GetTls() {
		for _, rd := range r.GetRoute() {
			destinations = append(destinations, rd.GetDestination())
		}
	}
	for _, r := range vs.GetHttp() {
		for _, rd := range r.GetRoute() {
			destinations = append(destinations, rd.GetDestination())
		}
	}

	return destinations
}
