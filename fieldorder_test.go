package dagpb

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// Fixtures from IPIP-550 (https://github.com/ipfs/specs/pull/550): the same
// UnixFS Directory and HAMTShard containing hello.txt, encoded in both
// PBNode field orders.
const (
	dirLinksFirstHex  = "12330a24015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03120968656c6c6f2e74787418060a020801"
	dirDataFirstHex   = "0a02080112330a24015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03120968656c6c6f2e7478741806"
	hamtLinksFirstHex = "12350a24015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03120b444668656c6c6f2e74787418060a250805121c800000000000000000000000000000000000000000000000000000002822308002"
	hamtDataFirstHex  = "0a250805121c80000000000000000000000000000000000000000000000000000000282230800212350a24015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03120b444668656c6c6f2e7478741806"
)

func reencodeWithOrder(t *testing.T, encHex string, order FieldOrder) string {
	t.Helper()
	data, err := hex.DecodeString(encHex)
	if err != nil {
		t.Fatal(err)
	}
	nb := Type.PBNode.NewBuilder()
	if err := DecodeBytes(nb, data); err != nil {
		t.Fatal(err)
	}
	enc, err := EncodeOptions{FieldOrder: order}.AppendEncode(nil, nb.Build())
	if err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(enc)
}

// TestFieldOrderRoundTrip decodes each fixture in either order and asserts
// that re-encoding with an explicit FieldOrder reproduces the corresponding
// fixture bytes exactly.
func TestFieldOrderRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name       string
		linksFirst string
		dataFirst  string
	}{
		{"Directory", dirLinksFirstHex, dirDataFirstHex},
		{"HAMTShard", hamtLinksFirstHex, hamtDataFirstHex},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, src := range []string{tc.linksFirst, tc.dataFirst} {
				if got := reencodeWithOrder(t, src, LinksFirst); got != tc.linksFirst {
					t.Fatalf("LinksFirst re-encode:\n got %s\nwant %s", got, tc.linksFirst)
				}
				if got := reencodeWithOrder(t, src, DataFirst); got != tc.dataFirst {
					t.Fatalf("DataFirst re-encode:\n got %s\nwant %s", got, tc.dataFirst)
				}
			}
		})
	}
}

// TestFieldOrderDefaultUnchanged asserts that the package-level entry points
// and the zero-value EncodeOptions still produce the canonical links-first
// bytes.
func TestFieldOrderDefaultUnchanged(t *testing.T) {
	data, err := hex.DecodeString(dirDataFirstHex)
	if err != nil {
		t.Fatal(err)
	}
	nb := Type.PBNode.NewBuilder()
	if err := DecodeBytes(nb, data); err != nil {
		t.Fatal(err)
	}
	node := nb.Build()

	enc, err := AppendEncode(nil, node)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(enc); got != dirLinksFirstHex {
		t.Fatalf("AppendEncode default:\n got %s\nwant %s", got, dirLinksFirstHex)
	}

	var buf bytes.Buffer
	if err := Encode(node, &buf); err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(buf.Bytes()); got != dirLinksFirstHex {
		t.Fatalf("Encode default:\n got %s\nwant %s", got, dirLinksFirstHex)
	}
}
