# Kessel-Plan README

## About
`kessel-plan` is a lightweight utility for authoring and extending the configuration files that describe the services, resources, and permissions that make up a kessel system. Changes made automatically are always additive, so don't worry about losing any manual changes when running this tool!

## Getting Started
`kessel-plan` requires a working go 1.22.2 or higher toolchain to compile, and all other external dependencies will be fetched automatically. Just clone and `make`.

## Usage
If you don't already have a kessel bootstrap file, use `./kessel-plan new` to generate an empty one (named 'bootstrap.yaml' by default.)

If you do, it can be referenced from any sub-command using the `-input=<path to file>` argument. Otherwise, `kessel-plan` will try to read from a bootstrap.yaml file in the current directory. Likewise, the output will automatically be written to a bootstrap.yaml in the same directory (potentially overwriting the original) unless you override it with the `-output=<path to file>` parameter.

To add permissions to your Kessel system, use the `kessel-plan add-permissions` command (see `kessel-plan add-permissions --help` for details) which allows you to add new resource types and permissions as well as adding permissions to existing resource types for new and existing services. All changes are additive.

For example, to add a concept of users of the space-traffic-control service being able to be granted access to enter, or to approve landings and departures for landing-bays in a given workspace, you could run `./kessel-plan add-permissions -svc=space-traffic-control -res=landing-bay enter approve-landing approve-departure`