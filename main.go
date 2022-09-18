package main

import (
	"os"
	"strconv"

	"github.com/ancat/hypercam/cmd/hypercam"
	"github.com/ancat/hypercam/pkg/freezer"
	"github.com/urfave/cli/v2"
)

func requirePid(cCtx *cli.Context) (int, error) {
	if cCtx.NArg() > 0 {
		pid, err := strconv.Atoi(cCtx.Args().Get(0))
		if err != nil {
			panic(err)
		}

		return pid, nil
	}

	return 0, nil
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
		},
		Commands: []*cli.Command{
			{
				Name:	 "pause",
				Usage:	 "pause an entire cgroup by its pid",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:	"pid",
						Usage:	"process id",
					},
				},
				Action: func(cCtx *cli.Context) error {
					pid, _ := requirePid(cCtx)
					if pid == 0 {
						return cli.Exit("pid missing", 1)
					}

					cgroup_name, _ := freezer.GetFreezerInfo(pid)
					if cgroup_name == "" || cgroup_name == "/" {
						return cli.Exit("no cgroup", 1)
					}

					freezer.UpdateFreezerStateByName(cgroup_name, "FROZEN")
					return nil
				},
			},
			{
				Name:	 "unpause",
				Usage:	 "resume an entire cgroup by its pid",
				Action: func(cCtx *cli.Context) error {
					pid, _ := requirePid(cCtx)
					if pid == 0 {
						return cli.Exit("pid missing", 1)
					}

					cgroup_name, _ := freezer.GetFreezerInfo(pid)
					if cgroup_name == "" || cgroup_name == "/" {
						return cli.Exit("no cgroup", 1)
					}

					freezer.UpdateFreezerStateByName(cgroup_name, "THAWED")
					return nil
				},
			},
			{
				Name:	 "info",
				Usage:	 "view open files and sockets for a given target",
				Action: func(cCtx *cli.Context) error {
					pid, _ := requirePid(cCtx)
					if pid == 0 {
						return cli.Exit("pid missing", 1)
					}

					hypercam.Pr(pid)
					return nil
				},
			},
			{
				Name:	 "splice",
				Usage:	 "splice an interactive shell into the target",
				Flags: []cli.Flag {
					&cli.BoolFlag {
						Name: "no-portal",
						Usage: "don't create a portal into the container",
					},
					&cli.StringFlag {
						Name: "exec",
						Usage: "path to shell executable (default /bin/sh)",
					},
					&cli.StringFlag {
						Name: "exec-from-host",
						Usage: "path to shell executable, copied from the host; useful if the target doesn't contain a usable shell. may need to be statically compiled.",
					},
				},
				Action: func(cCtx *cli.Context) error {
					pid, _ := requirePid(cCtx)
					if pid == 0 {
						return cli.Exit("pid missing", 1)
					}

					use_portals := !cCtx.Bool("no-portal")
					host_executable := cCtx.String("exec-from-host")
					guest_executable := cCtx.String("exec")
					if host_executable != "" && guest_executable != "" {
						return cli.Exit("--exec and --exec-from-host are mutually exclusive", 1)
					} else if host_executable == "" && guest_executable != "" {
						hypercam.SpawnShellInside(pid, use_portals, "", guest_executable)
					} else if host_executable != "" && guest_executable == "" {
						hypercam.SpawnShellInside(pid, use_portals, host_executable, "")
					} else {
						// neither is set
						hypercam.SpawnShellInside(pid, use_portals, "", "/bin/sh")
					}

					return nil
				},
			},
			{
				Name:	"scan",
				Usage:	"scan a process' stack and heap",
				Flags:	[]cli.Flag {
					&cli.BoolFlag {
						Name: "hex",
						Usage: "print a hex dump instead",
					},
				},
				Action: func(cCtx *cli.Context) error {
					pid, _ := requirePid(cCtx)
					if pid == 0 {
						return cli.Exit("pid missing", 1)
					}

					hex_dump := cCtx.Bool("hex")
					hypercam.DumpMaps(pid, hex_dump)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
