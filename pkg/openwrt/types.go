package openwrt

// DNSRecord represents a DNS record in LuciRPC
type DNSRecord struct {
	Type   string `json:".type" validate:"required"`
	IP     string `json:"ip,omitempty"`
	Name   string `json:"name,omitempty"`
	CName  string `json:"cname,omitempty"`
	Target string `json:"target,omitempty"`
}
