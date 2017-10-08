package discovery

type Location struct {
	InstanceGuid string `json:"instance_guid"`
	IPAddress    string `json:"ip_address"`
	TLSPort      string `json:"tls_port"`
}
