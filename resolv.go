package resolv

const (
	PortDefault = "53"

	ProtoUDP = "udp"
	ProtoTCP = "tcp"
)

type Mode int

const (
	ModeUDP Mode = 1 << iota
	ModeTCP
	ModeMixed
)

var RootServers = []string{
	"a.root-servers.net.",
	"b.root-servers.net.",
	"c.root-servers.net.",
	"d.root-servers.net.",
	"e.root-servers.net.",
	"f.root-servers.net.",
	"g.root-servers.net.",
	"h.root-servers.net.",
	"i.root-servers.net.",
}
