package dagpb

import (
	"io"

	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/multicodec"
)

var (
	_ ipld.Decoder = Decoder
	_ ipld.Encoder = Encoder
)

func init() {
	multicodec.RegisterDecoder(0x70, Decoder)
	multicodec.RegisterEncoder(0x70, Encoder)
}

// Decoder provides an IPLD codec decode interface for DAG-PB data. Provide a
// compatible NodeAssembler and a byte source to unmarshal a DAG-PB IPLD Node.
// Use the NodeAssembler from the PBNode type for safest construction
// (Type.PBNode.NewBuilder()). A Map assembler will also work.
// This function is registered via the go-ipld-prime link loader for multicodec
// code 0x70 when this package is invoked via init.
func Decoder(na ipld.NodeAssembler, r io.Reader) error {
	return Unmarshal(na, r)
}

// Encoder provides an IPLD codec encode interface for DAG-PB data. Provide a
// conforming Node and a destination for bytes to marshal a DAG-PB IPLD Node.
// The Node must strictly conform to the DAG-PB schema
// (https://github.com/ipld/specs/blob/master/block-layer/codecs/dag-pb.md).
// For safest use, build Nodes using the Type.PBNode type.
// This function is registered via the go-ipld-prime link loader for multicodec
// code 0x70 when this package is invoked via init.
func Encoder(n ipld.Node, w io.Writer) error {
	return Marshal(n, w)
}
