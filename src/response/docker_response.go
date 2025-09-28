package response

type (
	DockerContainerHistory struct {
		ID      string            `json:"id"`
		Names   []string          `json:"names"`
		Image   string            `json:"image"`
		ImageID string            `json:"imageId"`
		Command string            `json:"command"`
		Created int64             `json:"created"`
		Ports   []Port            `json:"ports"`
		Labels  map[string]string `json:"labels"`

		State      string `json:"state"`
		Status     string `json:"status"`
		HostConfig struct {
			NetworkMode string `json:"networkMode"`
		} `json:"hostConfig"`

		NetworkSettings struct {
			Networks map[string]Network `json:"networks"`
		} `json:"networkSettings"`

		Mounts      []Mount       `json:"mounts"`
		DockerStats []DockerStats `json:"dockerStats,omitempty"`
	}

	Port struct {
		IP          string `json:"ip,omitempty"`
		PrivatePort int    `json:"privatePort,omitempty"`
		PublicPort  int    `json:"publicPort,omitempty"`
		Type        string `json:"type,omitempty"`
	}

	Network struct {
		IPAMConfig          any    `json:"ipamConfig"`
		Links               any    `json:"links"`
		Aliases             any    `json:"aliases"`
		MacAddress          string `json:"macAddress"`
		DriverOpts          any    `json:"driverOpts"`
		GwPriority          int    `json:"gwPriority"`
		NetworkID           string `json:"networkId"`
		EndpointID          string `json:"endpointId"`
		Gateway             string `json:"gateway"`
		IPAddress           string `json:"ipAddress"`
		IPPrefixLen         int    `json:"ipPrefixLen"`
		IPv6Gateway         string `json:"ipv6Gateway"`
		GlobalIPv6Address   string `json:"globalIpv6Address"`
		GlobalIPv6PrefixLen int    `json:"globalIpv6PrefixLen"`
		DNSNames            any    `json:"dnsNames"`
	}

	Mount struct {
		Type        string `json:"type"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
		Mode        string `json:"mode"`
		RW          bool   `json:"rw"`
		Propagation string `json:"propagation"`
	}

	DockerStats struct {
		CPUStats struct {
			CPUUsage struct {
				TotalUsage  uint64   `json:"totalUsage"`
				PercpuUsage []uint64 `json:"percpuUsage"`
			} `json:"cpuUsage"`
			SystemUsage uint64 `json:"systemUsage"`
		} `json:"cpuStats"`
		PreCPUStats struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"totalUsage"`
			} `json:"cpuUsage"`
			SystemUsage uint64 `json:"systemUsage"`
		} `json:"preCpuStats"`
		MemoryStats struct {
			Usage uint64 `json:"usage"`
			Limit uint64 `json:"limit"`
		} `json:"memoryStats"`
	}
)
