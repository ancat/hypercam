# hypercam

hypercam is a tool that lets you interact with processes and containers at a lower level than most other interfaces, letting you do some neat things that would otherwise not be possible.

## Features

* Pause and unpause containers/processes
* Open a live shell into paused containers, allowing you to inspect container state as is
* Open shells (or any arbitrary program!) inside containers whose images don't come with them
* Mount a directory into running containers to make moving files in and out easier
* Dump strings/a hexdump of the heap and stack of any process

```
$ sudo hypercam -h
NAME:
   hypercam - A new cli application

USAGE:
   hypercam [global options] command [command options] [arguments...]

COMMANDS:
   pause    pause an entire cgroup by its pid
   unpause  resume an entire cgroup by its pid
   info     view open files and sockets for a given target
   splice   splice an interactive shell into the target
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)

$ sudo hypercam info 4106
nginx: master process nginx -g daemon off; (/usr/sbin/nginx) 4106
- Open Socket: 0.0.0.0:80<->0.0.0.0:0
bash (/bin/bash) 24355
bash (/bin/bash) 30267
nginx: worker process (/usr/sbin/nginx) 31689
- Open Socket: 172.17.0.2:80<->151.101.117.140:52200
- Open Socket: 0.0.0.0:80<->0.0.0.0:0

$ sudo hypercam pause 4106

$ sudo hypercam splice 4106
root@nginx:/# ps aux
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root         1  0.0  0.2  10644  2808 ?        Ds   Mar16   0:00 nginx: master process nginx
root       284  0.0  0.0   3868     0 pts/0    Ds+  Mar16   0:00 bash
root      4214  0.0  0.0   3864    96 pts/1    Ds+  Sep05   0:00 bash
nginx     4263  0.0  0.0  11100   604 ?        D    Sep05   0:00 nginx: worker process
root      7226  0.0  0.2   7636  2708 ?        R+   16:36   0:00 ps aux
root@nginx:/# etc...
root@nginx:/# exit

$ sudo hypercam unpause 4106

```
