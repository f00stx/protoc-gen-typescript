package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/f00stx/protoc-gen-typescript/internal"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	flagVerbose               = flag.Int("v", 0, "verbosity level")
	flagAsyncIterators        = flag.Bool("async_iterators", false, "if true, use async iterators")
	flagEnumsAsInts           = flag.Bool("int_enums", false, "if true, generate numeric enums")
	flagOriginalNames         = flag.Bool("original_names", true, "if true, use original proto file field names, otherwise convert to lowerCamelCase")
	flagOutputFilenamePattern = flag.String("outpattern", "{{.Dir}}/{{.Descriptor.GetPackage | default \"none\"}}.{{.BaseName}}.ts", "output filename pattern")
	flagDumpDescriptor        = flag.Bool("dump_request_descriptor", false, "if true, dump request descriptor")
	flagInt64AsString         = flag.Bool("int64_string", true, "if true, use string representation for 64 bit numbers")
)

func main() {
	g := internal.New()
	if terminal.IsTerminal(0) {
		flag.Usage()
		log.Fatalln("stdin appears to be a tty device. This tool is meant to be invoked via the protoc command via a --typescript_out directive.")
	}
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "failed to reading input"))
	}
	if err := proto.Unmarshal(data, g.Request); err != nil {
		log.Fatalln(errors.Wrap(err, "failed to parsing input"))
	}
	if len(g.Request.FileToGenerate) == 0 {
		log.Fatalln(errors.Wrap(err, "no files to generate"))
	}
	parseFlags(g.Request.Parameter)
	g.GenerateAllFiles(&internal.Parameters{
		AsyncIterators:        *flagAsyncIterators,
		Verbose:               *flagVerbose,
		OutputNamePattern:     *flagOutputFilenamePattern,
		EnumsAsInt:            *flagEnumsAsInts,
		OriginalNames:         *flagOriginalNames,
		DumpRequestDescriptor: *flagDumpDescriptor,
		Int64AsString:         *flagInt64AsString,
	})
	data, err = proto.Marshal(g.Response)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "failed to marshal output proto"))
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "failed to write output proto"))
	}
}

func parseFlags(s *string) {
	if s == nil {
		return
	}
	for _, p := range strings.Split(*s, ",") {
		spec := strings.SplitN(p, "=", 2)
		if len(spec) == 1 {
			if err := flag.CommandLine.Set(spec[0], ""); err != nil {
				log.Fatalln("Cannot set flag", p, err)
			}
			continue
		}
		name, value := spec[0], spec[1]
		// TODO: consider supporting package mapping (M flag)
		if err := flag.CommandLine.Set(name, value); err != nil {
			log.Fatalln("Cannot set flag", p)
		}
	}
}
