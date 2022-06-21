package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/jsonschema"

	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
)

func main() {
	var output string
	flag.StringVar(&output, "out", "", "output filename (\"empty for stdout\")")
	flag.Parse()

	s := jsonschema.Reflect(&config.RouterConfig{})
	b, err := s.MarshalJSON()
	if err != nil {
		log.Fatalf("could not marshal to JSON: %v", err)
	}

	if output != "" {
		f, err := os.Create(output)
		if err != nil {
			log.Fatalf("could not create output file: %v", err)
		}
		_, err = f.Write(b)
		if err != nil {
			log.Fatalf("could not write to output file: %v", err)
		}
	} else {
		fmt.Fprintln(os.Stdout, string(b))
	}
}
