package dagpb

import (
	"io"

	ipld "github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

var (
	_ cidlink.MulticodecDecoder = Decoder
	_ cidlink.MulticodecEncoder = Encoder
)

func init() {
	cidlink.RegisterMulticodecDecoder(0x70, Decoder)
	cidlink.RegisterMulticodecEncoder(0x70, Encoder)
}

func Decoder(na ipld.NodeAssembler, r io.Reader) error {
	return Unmarshal(na, r)
}

func Encoder(n ipld.Node, w io.Writer) error {
	// return Unmarshal(na, w)
	return nil
}
