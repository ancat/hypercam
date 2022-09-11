# hypercam

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
root@nginx:/# etc...
root@nginx:/#

$ sudo hypercam unpause 4106

```
