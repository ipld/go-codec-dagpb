package dagpb

import (
	"io"

	ipld "github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/multicodec"
	"github.com/ipld/go-ipld-prime/traversal"
)

var (
	_ ipld.Decoder = Decode
	_ ipld.Encoder = Encode
)

func init() {
	multicodec.RegisterDecoder(0x70, Decode)
	multicodec.RegisterEncoder(0x70, Encode)
}

// Decode provides an IPLD codec decode interface for DAG-PB data. Provide a
// compatible NodeAssembler and a byte source to unmarshal a DAG-PB IPLD Node.
// Use the NodeAssembler from the PBNode type for safest construction
// (Type.PBNode.NewBuilder()). A Map assembler will also work.
// This function is registered via the go-ipld-prime link loader for multicodec
// code 0x70 when this package is invoked via init.
func Decode(na ipld.NodeAssembler, r io.Reader) error {
	return Unmarshal(na, r)
}

// Encode provides an IPLD codec encode interface for DAG-PB data. Provide a
// conforming Node and a destination for bytes to marshal a DAG-PB IPLD Node.
// The Node must strictly conform to the DAG-PB schema
// (https://github.com/ipld/specs/blob/master/block-layer/codecs/dag-pb.md).
// For safest use, build Nodes using the Type.PBNode type.
// This function is registered via the go-ipld-prime link loader for multicodec
// code 0x70 when this package is invoked via init.
func Encode(n ipld.Node, w io.Writer) error {
	return Marshal(n, w)
}

// AddSupportToChooser takes an existing node prototype chooser and subs in
// PBNode for the dag-pb multicodec code.
func AddSupportToChooser(existing traversal.LinkTargetNodePrototypeChooser) traversal.LinkTargetNodePrototypeChooser {
	return func(lnk ipld.Link, lnkCtx ipld.LinkContext) (ipld.NodePrototype, error) {
		if lnk, ok := lnk.(cidlink.Link); ok && lnk.Cid.Prefix().Codec == 0x70 {
			return Type.PBNode, nil
		}
		return existing(lnk, lnkCtx)
	}
}

// We switched to simpler API names after v1.0.0, so keep the old names around
// as deprecated forwarding funcs until a future v2+.
// TODO: consider deprecating Marshal/Unmarshal too, since it's a bit
// unnecessary to have two supported names for each API.

// Deprecated: use Decode instead.
func Decoder(na ipld.NodeAssembler, r io.Reader) error {
	return Unmarshal(na, r)
}

// Deprecated: use Encode instead.
func Encoder(n ipld.Node, w io.Writer) error {
	return Marshal(n, w)
}
