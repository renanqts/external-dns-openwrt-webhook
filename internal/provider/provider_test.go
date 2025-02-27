package provider

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/renanqts/external-dns-openwrt-webhook/pkg/logger"
	"github.com/renanqts/external-dns-openwrt-webhook/pkg/openwrt"
	"sigs.k8s.io/external-dns/endpoint"
)

func TestProvider(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Provider Suite")
	defer GinkgoRecover()
}

var _ = BeforeSuite(func() {
	if err := logger.Init(&logger.Config{
		Level:    "debug",
		Encoding: "console",
	}); err != nil {
		panic(err)
	}
})

var _ = AfterSuite(func() {
	_ = logger.Log.Sync()
})

var _ = Describe("Provider Suite", func() {
	Context("endpoints records", func() {
		It("should be converted to dns records", func() {
			records := []struct {
				Name   string `json:"name"`
				Type   string `json:"type"`
				Target string `json:"target"`
			}{
				{
					Name:   "a.foobar.com",
					Type:   "A",
					Target: "1.1.1.1",
				},
				{
					Name:   "b.foobar.com",
					Type:   "CNAME",
					Target: "c.foobar.com",
				},
			}

			var endpoints []*endpoint.Endpoint
			for _, record := range records {
				endpoints = append(endpoints, &endpoint.Endpoint{
					DNSName:    record.Name,
					RecordTTL:  defaultTTL,
					RecordType: record.Type,
					Targets:    []string{record.Target},
				})
			}
			dnsRecords := endpoints2DNSRecords(endpoints)
			for index, dnsRecord := range dnsRecords {
				switch dnsRecord.Type {
				case "A":
					Expect(dnsRecord.Name).To(Equal(records[index].Name))
					Expect(dnsRecord.IP).To(Equal(records[index].Target))
				case "CNAME":
					Expect(dnsRecord.CName).To(Equal(records[index].Name))
					Expect(dnsRecord.Target).To(Equal(records[index].Target))
				}
			}
		})

		It("dns records to endpoint", func() {
			records := []struct {
				Name   string `json:"name"`
				Type   string `json:"type"`
				Target string `json:"target"`
			}{
				{
					Name:   "a.foobar.com",
					Type:   "A",
					Target: "1.1.1.1",
				},
				{
					Name:   "b.foobar.com",
					Type:   "CNAME",
					Target: "c.foobar.com",
				},
			}

			dnsRecords := make(map[string]openwrt.DNSRecord)
			for _, record := range records {
				switch record.Type {
				case "A":
					dnsRecords[record.Name] = openwrt.DNSRecord{
						Name: record.Name,
						Type: record.Type,
						IP:   record.Target,
					}
				case "CNAME":
					dnsRecords[record.Name] = openwrt.DNSRecord{
						Type:   record.Type,
						Target: record.Target,
						CName:  record.Name,
					}
				}
			}

			endpoints := dnsRecords2Endpoints(dnsRecords)
			for index, record := range records {
				Expect(endpoints[index].DNSName).To(Equal(record.Name))
				Expect(endpoints[index].Targets[0]).To(Equal(record.Target))
				Expect(endpoints[index].RecordType).To(Equal(record.Type))
			}
		})
	})
})
