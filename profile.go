package main

// Tunnel represents a single SSH tunnel configuration within a profile.
type Tunnel struct {
	Name        string `json:"name" yaml:"name"`
	SSHServer   string `json:"ssh_server" yaml:"ssh_server"`
	SSHPort     int    `json:"ssh_port" yaml:"ssh_port"`
	SSHUser     string `json:"ssh_user" yaml:"ssh_user"`
	SSHKeyPath  string `json:"ssh_key_path" yaml:"ssh_key_path"`
	RemoteHost  string `json:"remote_host" yaml:"remote_host"`
	RemotePort  int    `json:"remote_port" yaml:"remote_port"`
	LocalDomain string `json:"local_domain" yaml:"local_domain"`
	LocalPort   int    `json:"local_port" yaml:"local_port"`
}

// Profile groups multiple tunnels under a single IP address.
type Profile struct {
	Name      string   `json:"name" yaml:"name"`
	IPAddress string   `json:"ip_address" yaml:"ip_address"`
	Tunnels   []Tunnel `json:"tunnels" yaml:"tunnels"`
}
