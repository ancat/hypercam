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
				Action: func(cCtx *cli.Context) error {
					pid, _ := requirePid(cCtx)
					if pid == 0 {
						return cli.Exit("pid missing", 1)
					}

					hypercam.SpawnShellInside(pid)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
