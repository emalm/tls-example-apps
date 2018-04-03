package discovery

type Location struct {
	InstanceGuid string `json:"instance_guid"`
	IPAddress    string `json:"ip_address"`
	TLSPort      string `json:"tls_port"`
}

func (loc Location) Name() string {
	if loc.InstanceGuid != "" {
		return "instance " + loc.InstanceGuid + " at ip " + loc.IPAddress
	}

	return "instance at IP " + loc.IPAddress
}
