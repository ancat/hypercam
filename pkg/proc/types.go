package proc

type Connection struct {
    LocalAddress string
    RemoteAddress string
}

type process_metadata struct {
    executable string
    comm string
    open_files []string
    network_connections []Connection
}
