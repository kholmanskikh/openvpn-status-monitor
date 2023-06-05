# OpenVPN status monitor

Display the content from the OpenVPN server status file via HTTP.

Currently only the OpenVPN `--status-version 2` is supported.

```
Usage of ./openvpn-status-monitor:
  -interval duration
        read interval (default 5s)
  -listen string
        listen address as in go http.ListenAndServe (default ":8080")
  -status-file string
        path to the OpenVPN status file

```
