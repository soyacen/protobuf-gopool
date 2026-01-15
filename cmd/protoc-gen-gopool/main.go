package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
)

var poolIndent = protogen.GoImportPath("sync").Ident("Pool")

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Fprintf(os.Stdout, "%v %v\n", filepath.Base(os.Args[0]), "v0.0.1")
		os.Exit(0)
	}
	var (
		flags                                 flag.FlagSet
		experimentalStripNonFunctionalCodegen = flags.Bool("experimental_strip_nonfunctional_codegen", false, "experimental_strip_nonfunctional_codegen true means that the plugin will not emit certain parts of the generated code in order to make it possible to compare a proto2/proto3 file with its equivalent (according to proto spec) editions file. Primarily, this is the encoded descriptor.")
	)
	protogen.Options{
		ParamFunc:                    flags.Set,
		InternalStripForEditionsDiff: experimentalStripNonFunctionalCodegen,
	}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if f.Generate {
				GenerateFile(gen, f)
			}
		}
		gen.SupportedFeatures = gengo.SupportedFeatures
		gen.SupportedEditionsMinimum = gengo.SupportedEditionsMinimum
		gen.SupportedEditionsMaximum = gengo.SupportedEditionsMaximum
		return nil
	})
}

func GenerateFile(gen *protogen.Plugin, file *protogen.File) {
	filename := file.GeneratedFilenamePrefix + ".pb.pool.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)

	g.P("package ", file.GoPackageName)
	g.P()

	for _, message := range file.Messages {
		// Generate object pool for each message
		generateMessagePool(g, message)
	}
}

func generateMessagePool(g *protogen.GeneratedFile, message *protogen.Message) {
	messageName := message.Desc.Name()
	// Generate the pool variable
	g.P("// ", messageName, "Pool is a sync.Pool for ", messageName)
	g.P("var ", messageName, "Pool = &", poolIndent, "{")
	g.P("New: func() interface{} {")
	g.P("return &", g.QualifiedGoIdent(message.GoIdent), "{}")
	g.P("},")
	g.P("}")
	g.P()

	// Generate Get method
	g.P("// Get", messageName, " gets a ", messageName, " from the pool")
	g.P("func Get", messageName, "() *", g.QualifiedGoIdent(message.GoIdent), " {")
	g.P("return ", messageName, "Pool.Get().(*", g.QualifiedGoIdent(message.GoIdent), ")")
	g.P("}")
	g.P()

	// Generate Put method
	g.P("// Put", messageName, " puts a ", messageName, " back to the pool")
	g.P("func Put", messageName, "(m *", g.QualifiedGoIdent(message.GoIdent), ") {")
	g.P("// Reset the message before putting it back to the pool")
	g.P("m.Reset()")
	g.P(messageName, "Pool.Put(m)")
	g.P("}")
	g.P()
}
