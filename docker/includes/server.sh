#!/usr/bin/env bash
set -euo pipefail

(trap 'kill 0' SIGINT; socat UNIX-CONNECT:/var/run/host_docker.sock UNIX-LISTEN:/var/run/docker.sock,user=root,group=daemon,mode=660 & /skrump/pocketbase serve ${@:1})
