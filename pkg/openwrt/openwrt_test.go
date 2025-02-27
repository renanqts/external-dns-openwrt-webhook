package openwrt

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mocks "github.com/renanqts/external-dns-openwrt-webhook/internal/mocks/lucirpc"
	"github.com/renanqts/external-dns-openwrt-webhook/pkg/logger"
	"go.uber.org/mock/gomock"
)

func TestOpenWRT(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OpenWRT Suite")
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

var _ = Describe("OpenWRT", func() {
	var (
		ctx         context.Context
		mockCtrl    *gomock.Controller
		mockLuciRPC *mocks.MockLuciRPC
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		mockLuciRPC = mocks.NewMockLuciRPC(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("Get DNS", func() {
		It("get all records", func() {
			expectedJson, err := json.Marshal(map[string]DNSRecord{
				"x": {
					Type: "domain",
					Name: "foobar",
					IP:   "1.1.1.1",
				},
				"y": {
					Type:   "cname",
					CName:  "foobar",
					Target: "bar.foo.com",
				},
				"z": {
					Type: "whatever",
				},
			})
			Expect(err).To(BeNil())
			mockLuciRPC.EXPECT().Uci(ctx, "get_all", []string{"dhcp"}).Return(string(expectedJson), nil)
			o := openWRT{
				lucirpc: mockLuciRPC,
			}
			resultDNS, err := o.GetDNSRecords(ctx)
			Expect(err).To(BeNil())
			Expect(resultDNS).ToNot(BeNil())
			Expect(resultDNS).To(Equal(map[string]DNSRecord{
				"x": {
					Type: "A",
					Name: "foobar",
					IP:   "1.1.1.1",
				},
				"y": {
					Type:   "CNAME",
					CName:  "foobar",
					Target: "bar.foo.com",
				},
			}))
		})
	})

	Context("Set DNS", func() {
		It("set A record with success", func() {
			cfg := "foobar"
			ip := "1.1.1.1"
			name := "foo.bar.com"

			mockLuciRPC.EXPECT().Uci(ctx, "add", []string{"dhcp", "domain"}).Return(cfg, nil)
			mockLuciRPC.EXPECT().Uci(ctx, "set", []string{"dhcp", cfg, "name", name}).Return("", nil)
			mockLuciRPC.EXPECT().Uci(ctx, "set", []string{"dhcp", cfg, "ip", ip}).Return("", nil)
			mockLuciRPC.EXPECT().Uci(ctx, "commit", []string{"dhcp"}).Return("", nil)

			o := openWRT{
				lucirpc: mockLuciRPC,
			}
			err := o.SetDNSRecords(ctx, []DNSRecord{
				{
					Type: "A",
					IP:   ip,
					Name: name,
				},
			})
			Expect(err).To(BeNil())
		})

		It("A without name", func() {
			o := openWRT{}
			err := o.SetDNSRecords(ctx, []DNSRecord{
				{
					Type: "A",
					IP:   "1.1.1.1",
				},
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("name is required"))
		})

		It("A without ip", func() {
			o := openWRT{}
			err := o.SetDNSRecords(ctx, []DNSRecord{
				{
					Type: "A",
					Name: "foobar",
				},
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("ip is required"))
		})

		It("set CNAME record", func() {
			cfg := "foobar"
			cname := "foo.bar.com"
			target := "bar.foo.com"

			mockLuciRPC.EXPECT().Uci(ctx, "add", []string{"dhcp", "cname"}).Return(cfg, nil)
			mockLuciRPC.EXPECT().Uci(ctx, "set", []string{"dhcp", cfg, "cname", cname}).Return("", nil)
			mockLuciRPC.EXPECT().Uci(ctx, "set", []string{"dhcp", cfg, "target", target}).Return("", nil)
			mockLuciRPC.EXPECT().Uci(ctx, "commit", []string{"dhcp"}).Return("", nil)

			o := openWRT{
				lucirpc: mockLuciRPC,
			}
			err := o.SetDNSRecords(ctx, []DNSRecord{
				{
					Type:   "CNAME",
					CName:  cname,
					Target: target,
				},
			})
			Expect(err).To(BeNil())
		})

		It("CNAME without cname", func() {
			o := openWRT{}
			err := o.SetDNSRecords(ctx, []DNSRecord{
				{
					Type:   "CNAME",
					Target: "foo.bar.com",
				},
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("cname is required"))
		})

		It("CNAME without target", func() {
			o := openWRT{}
			err := o.SetDNSRecords(ctx, []DNSRecord{
				{
					Type:  "CNAME",
					CName: "foobar",
				},
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("target is required"))
		})
	})

	Context("Update DNS", func() {
		It("update A record", func() {
			cfg := "x"
			dnsName := "happy.com"
			updatedIP := "2.2.2.2"

			expectedCurrentDNSRecords := map[string]DNSRecord{
				cfg: {
					Type: "domain",
					Name: dnsName,
					IP:   "1.1.1.1",
				},
				"y": {
					Type:   "cname",
					CName:  "foo.bar.com",
					Target: "bar.foo.com",
				},
			}

			expectedCurrentJson, err := json.Marshal(expectedCurrentDNSRecords)
			Expect(err).To(BeNil())
			mockLuciRPC.EXPECT().Uci(ctx, "get_all", []string{"dhcp"}).Return(string(expectedCurrentJson), nil)
			mockLuciRPC.EXPECT().Uci(ctx, "delete", []string{"dhcp", cfg}).Return("", nil)
			mockLuciRPC.EXPECT().Uci(ctx, "add", []string{"dhcp", "domain"}).Return(cfg, nil)
			mockLuciRPC.EXPECT().Uci(ctx, "set", []string{"dhcp", cfg, "name", dnsName}).Return("", nil)
			mockLuciRPC.EXPECT().Uci(ctx, "set", []string{"dhcp", cfg, "ip", updatedIP}).Return("", nil)
			mockLuciRPC.EXPECT().Uci(ctx, "commit", []string{"dhcp"}).Return("", nil)

			o := openWRT{
				lucirpc: mockLuciRPC,
			}
			err = o.UpdateDNSRecords(ctx, []DNSRecord{
				{
					Type: "A",
					Name: dnsName,
					IP:   updatedIP,
				},
			})
			Expect(err).To(BeNil())
		})

		It("update CNAME record", func() {
			cfg := "y"
			cname := "happy.com"
			updatedTarget := "foo.bar.com"

			expectedCurrentDNSRecords := map[string]DNSRecord{
				"x": {
					Type: "domain",
					Name: "happy.com",
					IP:   "1.1.1.1",
				},
				cfg: {
					Type:   "cname",
					CName:  cname,
					Target: "bar.foo.com",
				},
			}

			expectedCurrentJson, err := json.Marshal(expectedCurrentDNSRecords)
			Expect(err).To(BeNil())
			mockLuciRPC.EXPECT().Uci(ctx, "get_all", []string{"dhcp"}).Return(string(expectedCurrentJson), nil)
			mockLuciRPC.EXPECT().Uci(ctx, "delete", []string{"dhcp", cfg}).Return("", nil)
			mockLuciRPC.EXPECT().Uci(ctx, "add", []string{"dhcp", "cname"}).Return(cfg, nil)
			mockLuciRPC.EXPECT().Uci(ctx, "set", []string{"dhcp", cfg, "cname", cname}).Return("", nil)
			mockLuciRPC.EXPECT().Uci(ctx, "set", []string{"dhcp", cfg, "target", updatedTarget}).Return("", nil)
			mockLuciRPC.EXPECT().Uci(ctx, "commit", []string{"dhcp"}).Return("", nil)

			o := openWRT{
				lucirpc: mockLuciRPC,
			}
			err = o.UpdateDNSRecords(ctx, []DNSRecord{
				{
					Type:   "CNAME",
					CName:  cname,
					Target: updatedTarget,
				},
			})
			Expect(err).To(BeNil())
		})

		It("not found", func() {
			expectedCurrentDNSRecords := map[string]DNSRecord{
				"x": {
					Type: "A",
					Name: "happy.com",
					IP:   "1.1.1.1",
				},
				"y": {
					Type:   "CNAME",
					CName:  "foo.bar.com",
					Target: "bar.foo.com",
				},
			}

			expectedCurrentJson, err := json.Marshal(expectedCurrentDNSRecords)
			Expect(err).To(BeNil())
			mockLuciRPC.EXPECT().Uci(ctx, "get_all", []string{"dhcp"}).Return(string(expectedCurrentJson), nil)

			o := openWRT{
				lucirpc: mockLuciRPC,
			}
			err = o.UpdateDNSRecords(ctx, []DNSRecord{
				{
					Type:   "CNAME",
					CName:  "whatever",
					Target: "3.3.3.3",
				},
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("records not found: [{CNAME   whatever 3.3.3.3}]"))
		})
	})

	Context("Delete DNS", func() {
		It("delete A record", func() {
			cfg := "x"
			name := "happy.com"
			ip := "2.2.2.2"

			expectedCurrentDNSRecords := map[string]DNSRecord{
				cfg: {
					Type: "domain",
					Name: name,
					IP:   ip,
				},
				"y": {
					Type:   "cname",
					CName:  "foo.bar.com",
					Target: "bar.foo.com",
				},
			}

			expectedCurrentJson, err := json.Marshal(expectedCurrentDNSRecords)
			Expect(err).To(BeNil())
			mockLuciRPC.EXPECT().Uci(ctx, "get_all", []string{"dhcp"}).Return(string(expectedCurrentJson), nil)
			mockLuciRPC.EXPECT().Uci(ctx, "delete", []string{"dhcp", cfg}).Return("", nil)
			mockLuciRPC.EXPECT().Uci(ctx, "commit", []string{"dhcp"}).Return("", nil)

			o := openWRT{
				lucirpc: mockLuciRPC,
			}
			err = o.DeleteDNSRecords(ctx, []DNSRecord{
				{
					Type: "A",
					Name: name,
					IP:   ip,
				},
			})
			Expect(err).To(BeNil())
		})

		It("delete CNAME record", func() {
			cfg := "y"
			cname := "happy.com"
			target := "foo.bar.com"

			expectedCurrentDNSRecords := map[string]DNSRecord{
				"x": {
					Type: "domain",
					Name: "happy.com",
					IP:   "1.1.1.1",
				},
				cfg: {
					Type:   "cname",
					CName:  cname,
					Target: target,
				},
			}

			expectedCurrentJson, err := json.Marshal(expectedCurrentDNSRecords)
			Expect(err).To(BeNil())
			mockLuciRPC.EXPECT().Uci(ctx, "get_all", []string{"dhcp"}).Return(string(expectedCurrentJson), nil)
			mockLuciRPC.EXPECT().Uci(ctx, "delete", []string{"dhcp", cfg}).Return("", nil)
			mockLuciRPC.EXPECT().Uci(ctx, "commit", []string{"dhcp"}).Return("", nil)

			o := openWRT{
				lucirpc: mockLuciRPC,
			}
			err = o.DeleteDNSRecords(ctx, []DNSRecord{
				{
					Type:   "CNAME",
					CName:  cname,
					Target: target,
				},
			})
			Expect(err).To(BeNil())
		})

		It("not found", func() {
			expectedCurrentDNSRecords := map[string]DNSRecord{
				"x": {
					Type: "domain",
					Name: "happy.com",
					IP:   "1.1.1.1",
				},
				"y": {
					Type:   "cname",
					CName:  "foo.bar.com",
					Target: "bar.foo.com",
				},
			}

			expectedCurrentJson, err := json.Marshal(expectedCurrentDNSRecords)
			Expect(err).To(BeNil())
			mockLuciRPC.EXPECT().Uci(ctx, "get_all", []string{"dhcp"}).Return(string(expectedCurrentJson), nil)

			o := openWRT{
				lucirpc: mockLuciRPC,
			}
			err = o.DeleteDNSRecords(ctx, []DNSRecord{
				{
					Type:   "CNAME",
					CName:  "whatever",
					Target: "3.3.3.3",
				},
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("records not found: [{CNAME   whatever 3.3.3.3}]"))
		})
	})
})
