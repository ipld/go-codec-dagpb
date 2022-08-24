//go:build go1.18

package dagpb_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	dagpb "github.com/ipld/go-codec-dagpb"
	"github.com/ipld/go-ipld-prime/node/basicnode"
)

func FuzzDecodeEncode(f *testing.F) {
	for _, hexInput := range []string{
		// Error cases
		"",     // empty
		"0a00", // zero data

		// Roundtrip cases
		"0a050001020304",                                   // data with empty links
		"120b0a09015500050001020304",                       // links with one hash
		"12160a090155000500010203041209736f6d65206e616d65", // links with one hash and name
		"12140a0901550005000102030418ffffffffffffff0f",     // links with one hash and tsize
	} {
		p, err := hex.DecodeString(hexInput)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(p)
	}
	f.Fuzz(func(t *testing.T, dagpbBytes []byte) {
		builder := basicnode.Prototype.Any.NewBuilder()
		if err := dagpb.DecodeBytes(builder, dagpbBytes); err != nil {
			return // invalid dagpb bytes, do not re-encode
		}
		node := builder.Build()
		var buf bytes.Buffer
		if err := dagpb.Encode(node, &buf); err != nil {
			t.Fatalf("re-encode of valid dagpb failed: %v", err)
		}
		// TODO: do we care that the re-encode matches byte by byte?
	})
}
