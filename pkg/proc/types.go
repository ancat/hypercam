package proc

type Connection struct {
	LocalAddress string
	RemoteAddress string
}

type MemoryMap struct {
	Name string
	Base uintptr
	Length uint64
	Perms int
}

type process_metadata struct {
	executable string
	comm string
	open_files []string
	network_connections []Connection
}
