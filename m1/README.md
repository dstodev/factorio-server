# Manage version 1 "m1"

`m1` is a prototype server management environment written in Python. It receives
limited updates, mostly to test new ideas.

m1 tests are unstable and the code leaves much to be desired.

This tool is being rewritten in Go(lang) as `m2`. Python dependency `docker-py`
does not well-support controls required by this project, such as arbitrary
stream redirection. Docker itself is written in Go, so its Go libraries are more
complete.
