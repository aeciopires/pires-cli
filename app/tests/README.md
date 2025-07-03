# pires-cli

<!-- TOC -->

- [pires-cli](#pires-cli)
- [Test commands](#test-commands)

<!-- TOC -->

# Test commands

Run the follow commands to test ``pires-cli``.

```bash
cd pires-cli/app

# Install dependencies
make prepare

go run .
go run . -h
go run . -v
go run . -V
go run . -D

# To see documentation of packages and files
go doc -C internal/getinfo/ -all
go doc -C internal/config/ -all
go doc -C pkg/pireslib/common -all
go doc -C pkg/pireslib/fileeditor -all
go doc -C pkg/pireslib/gcp -all
```
