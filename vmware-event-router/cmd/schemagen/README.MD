# About

`schemagen` generates a compliant JSON [v4
schema](https://json-schema.org/specification.html) for the VMware Event Router
based on struct tags and reflection using the third party package
`https://github.com/alecthomas/jsonschema`.

The resulting file, e.g. `routerconfig.schema.json` in the `vmware-event-router`
directory can be imported into an IDE or code editor, such as VS Code or Goland,
to generate or validate an VMware Event Router YAML configuration file.

# Build

To build the binary Go v1.13+ must be installed. 

```bash
# after cloning this repo
cd cmd/schemagen
go build .

./schemagen -h
Usage of ./schemagen:
  -out string
        output filename ("empty for stdout")
```

# Usage

By default, `schemagen` generates the schema for all files in the
`vmware-event-router/internal/config/v1alpha1` package. Output can be a file or
standard output if `-out` is not specified.

## Update the existing schema file

To update the existing `routerconfig.schema.json` schema definition, run the
following command from the `vmware-event-router` directory.

```bash
go run cmd/schemagen/main.go -out routerconfig.schema.json
```