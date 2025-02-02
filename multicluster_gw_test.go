package multicluster_gw

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func TestMultiClusterGw(t *testing.T) {
	tests := []struct {
		question            string // Corefile data as string
		serviceName         string
		serviceNs           string
		questionType        uint16 // The given request type
		shouldErr           bool   // True if test case is expected to produce an error.
		expectedReturnValue int    // The expected return value.
		expectedErrContent  error  // The expected error
		addToSet            bool
	}{
		// positive
		{
			`myservice.test.svc.clusterset.local.`,
			"myservice",
			"test",
			dns.TypeA,
			false,
			dns.RcodeSuccess,
			nil,
			true,
		},
		// not for the plugin's zone, should foward it and not handle the request:
		{
			`myservice.test.svc.cluster.local.`,
			"myservice",
			"test",
			dns.TypeA,
			false,
			dns.RcodeServerFailure,
			nil,
			true,
		},

		//not in the set, should return NODATA:
		{
			`myservice.test.svc.clusterset.local.`,
			"myservice",
			"test",
			dns.TypeA,
			false,
			dns.RcodeNameError,
			nil,
			false,
		},
	}
	initMcgw()
	ctx := context.TODO()
	r := new(dns.Msg)
	rec := dnstest.NewRecorder((&test.ResponseWriter{}))
	for _, test := range tests {
		initalizeSetForTest(test.question, test.serviceName, test.serviceNs, test.addToSet)
		r.SetQuestion(test.question, test.questionType)

		// call the plugin and check result:
		returnValue, err := Mcgw.ServeDNS(ctx, rec, r)

		assert.Equal(t, test.expectedReturnValue, returnValue)
		if test.shouldErr {
			assert.NotEmpty(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

// Function to initalize our set with a serviceImport for service with name svcName, under Ns svcNs.
// Boolean condition that determine if we do add the service to the set, or not (we add it only if the test wants that this serviceImport will exist).
func initalizeSetForTest(qustion string, svcName string, svcNS string, addToSet bool) {
	// empty the set in each test run:
	Mcgw.SISet = *NewSiSet()

	if addToSet {
		// add the current SI to the set:
		Mcgw.SISet.Add(GenerateNameAsString(svcName, svcNS))
	}
}

func initMcgw() {
	requestsZone := "svc.clusterset.local."
	Mcgw.SISet = *NewSiSet()
	Mcgw.Zones = []string{requestsZone}
	Mcgw.Next = test.ErrorHandler()
}
