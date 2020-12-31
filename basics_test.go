package dagpb

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/fluent"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
)

type pbNode struct {
	links []pbLink
	data  []byte
}

func mkcid(t *testing.T, cidStr string) cid.Cid {
	c, err := cid.Decode(cidStr)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func validate(t *testing.T, actual ipld.Node, expected pbNode) {
	mi := actual.MapIterator()
	_, isTyped := actual.(*_PBNode)

	hasLinks := false
	hasData := false
	for !mi.Done() {
		key, val, err := mi.Next()
		if err != nil {
			t.Fatal(err)
		}
		keyStr, err := key.AsString()
		if err != nil {
			t.Fatal(err)
		}
		if keyStr == "Links" {
			if val.Kind() != ipld.Kind_List {
				t.Fatal("Links is not a list")
			}
			if val.IsAbsent() {
				t.Fatal("Links is absent")
			}
			if val.IsNull() {
				t.Fatal("Links is null")
			}
			if val.Length() != int64(len(expected.links)) {
				t.Fatal("non-empty Links list")
			}
			hasLinks = true
		} else if keyStr == "Data" {
			if isTyped && expected.data == nil {
				if !val.IsAbsent() {
					t.Fatalf("Empty Data is not marked as absent")
				}
				if val.Kind() != ipld.Kind_Null {
					t.Fatal("Empty Data is not null")
				}
				if val.IsNull() {
					t.Fatal("Empty Data is null")
				}
			}
			hasData = !isTyped || !val.IsAbsent()
			if hasData {
				if expected.data == nil {
					t.Fatal("Got unexpected Data")
				} else {
					byts, err := val.AsBytes()
					if err != nil {
						t.Fatal(err)
					} else if bytes.Compare(expected.data, byts) != 0 {
						t.Fatal("Got unexpected Data contents")
					}
				}
			}
		} else {
			t.Fatalf("Unexpected map key: %v", keyStr)
		}
	}
	if !hasLinks {
		t.Fatal("Did not find Links")
	}
	if expected.data != nil && !hasData {
		t.Fatal("Did not find Data")
	}
}

func runTest(t *testing.T, bytsHex string, expected pbNode) {
	byts, _ := hex.DecodeString(bytsHex)

	roundTrip := func(t *testing.T, node ipld.Node) {
		var buf bytes.Buffer
		if err := Encoder(node, &buf); err != nil {
			t.Fatal(err)
		}

		// fmt.Printf("CMP\n\tFrom: %v\n\tTo:   %v\n", hex.EncodeToString(byts), hex.EncodeToString(buf.Bytes()))
		if bytes.Compare(buf.Bytes(), byts) != 0 {
			t.Fatal("Round-trip resulted in different bytes")
		}
	}

	t.Run("basicnode", func(t *testing.T) {
		nb := basicnode.Prototype__Map{}.NewBuilder()
		err := Decoder(nb, bytes.NewReader(byts))
		if err != nil {
			t.Fatal(err)
		}

		node := nb.Build()
		validate(t, node, expected)
		roundTrip(t, node)
	})

	t.Run("typed", func(t *testing.T) {
		nb := Type.PBNode.NewBuilder()
		err := Decoder(nb, bytes.NewReader(byts))
		if err != nil {
			t.Fatal(err)
		}
		node := nb.Build()
		validate(t, node, expected)
		roundTrip(t, node)
	})
}

func TestEmptyNode(t *testing.T) {
	runTest(t, "", pbNode{})
}

func TestNodeWithData(t *testing.T) {
	runTest(t, "0a050001020304", pbNode{data: []byte{00, 01, 02, 03, 04}})
}

func TestNodeWithDataZero(t *testing.T) {
	runTest(t, "0a00", pbNode{data: []byte{}})
}

func TestNodeWithLink(t *testing.T) {
	expected := pbNode{}
	expected.links = append(expected.links, pbLink{hash: mkcid(t, "QmWDtUQj38YLW8v3q4A6LwPn4vYKEbuKWpgSm6bjKW6Xfe")})
	runTest(t, "12240a2212207521fe19c374a97759226dc5c0c8e674e73950e81b211f7dd3b6b30883a08a51", expected)
}

func TestNodeWithLinkAndData(t *testing.T) {
	expected := pbNode{data: []byte("some data")}
	expected.links = append(expected.links, pbLink{hash: mkcid(t, "QmWDtUQj38YLW8v3q4A6LwPn4vYKEbuKWpgSm6bjKW6Xfe")})
	runTest(t, "12240a2212207521fe19c374a97759226dc5c0c8e674e73950e81b211f7dd3b6b30883a08a510a09736f6d652064617461", expected)
}

func TestNodeWithTwoUnsortedLinks(t *testing.T) {
	encoded := "12340a2212208ab7a6c5e74737878ac73863cb76739d15d4666de44e5756bf55a2f9e9ab5f431209736f6d65206c696e6b1880c2d72f12370a2212208ab7a6c5e74737878ac73863cb76739d15d4666de44e5756bf55a2f9e9ab5f44120f736f6d65206f74686572206c696e6b18080a09736f6d652064617461"
	expected := pbNode{data: []byte("some data")}
	expected.links = append(expected.links, pbLink{name: "some link", hasName: true, hash: mkcid(t, "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39U"), tsize: 100000000, hasTsize: true})
	expected.links = append(expected.links, pbLink{name: "some other link", hasName: true, hash: mkcid(t, "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39V"), tsize: 8, hasTsize: true})

	runTest(t, encoded, expected)

	// assembled in a rough order, Data coming first, links badly sorted
	node := fluent.MustBuildMap(basicnode.Prototype__Map{}, 2, func(fma fluent.MapAssembler) {
		fma.AssembleEntry("Data").AssignBytes([]byte("some data"))
		fma.AssembleEntry("Links").CreateList(2, func(fla fluent.ListAssembler) {
			fla.AssembleValue().CreateMap(3, func(fma fluent.MapAssembler) {
				fma.AssembleEntry("Name").AssignString("some other link")
				fma.AssembleEntry("Tsize").AssignInt(8)
				fma.AssembleEntry("Hash").AssignLink(cidlink.Link{Cid: mkcid(t, "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39V")})
			})
			fla.AssembleValue().CreateMap(3, func(fma fluent.MapAssembler) {
				fma.AssembleEntry("Name").AssignString("some link")
				fma.AssembleEntry("Tsize").AssignInt(100000000)
				fma.AssembleEntry("Hash").AssignLink(cidlink.Link{Cid: mkcid(t, "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39U")})
			})
		})
	})

	var buf bytes.Buffer
	if err := Encoder(node, &buf); err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(buf.Bytes()) != encoded {
		t.Fatal("did not encode to expected bytes")
	}
}

func TestNodeWithStableSortedLinks(t *testing.T) {
	cids := []string{
		"QmUGhP2X8xo9dsj45vqx1H6i5WqPqLqmLQsHTTxd3ke8mp",
		"QmP7SrR76KHK9A916RbHG1ufy2TzNABZgiE23PjZDMzZXy",
		"QmQg1v4o9xdT3Q14wh4S7dxZkDjyZ9ssFzFzyep1YrVJBY",
		"QmdP6fartWRrydZCUjHgrJ4XpxSE4SAoRsWJZ1zJ4MWiuf",
		"QmNNjUStxtMC1WaSZYiDW6CmAUrvd5Q2e17qnxPgVdwrwW",
		"QmWJwqZBJWerHsN1b7g4pRDYmzGNnaMYuD3KSbnpaxsB2h",
		"QmRXPSdysBS3dbUXe6w8oXevZWHdPQWaR2d3fggNsjvieL",
		"QmTUZAXfws6zrhEksnMqLxsbhXZBQs4FNiarjXSYQqVrjC",
		"QmNNk7dTdh8UofwgqLNauq6N78DPc6LKK2yBs1MFdx7Mbg",
		"QmW5mrJfyqh7B4ywSvraZgnWjS3q9CLiYURiJpCX3aro5i",
		"QmTFHZL5CkgNz19MdPnSuyLAi6AVq9fFp81zmPpaL2amED",
	}

	node := fluent.MustBuildMap(basicnode.Prototype__Map{}, 2, func(fma fluent.MapAssembler) {
		fma.AssembleEntry("Data").AssignBytes([]byte("some data"))
		fma.AssembleEntry("Links").CreateList(int64(len(cids)), func(fla fluent.ListAssembler) {
			for _, cid := range cids {
				fla.AssembleValue().CreateMap(3, func(fma fluent.MapAssembler) {
					fma.AssembleEntry("Name").AssignString("")
					fma.AssembleEntry("Tsize").AssignInt(262158)
					fma.AssembleEntry("Hash").AssignLink(cidlink.Link{Cid: mkcid(t, cid)})
				})
			}
		})
	})

	var buf bytes.Buffer
	if err := Encoder(node, &buf); err != nil {
		t.Fatal(err)
	}
	nb := basicnode.Prototype__Map{}.NewBuilder()
	err := Decoder(nb, bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	reencNode := nb.Build()
	links, _ := reencNode.LookupByString("Links")
	if links.Length() != int64(len(cids)) {
		t.Fatal("Incorrect number of links after round-trip")
	}
	iter := links.ListIterator()
	for !iter.Done() {
		ii, n, _ := iter.Next()
		h, _ := n.LookupByString("Hash")
		l, _ := h.AsLink()
		cl, _ := l.(cidlink.Link)
		if cids[ii] != cl.String() {
			t.Fatal("CIDs did not retain position after round-trip")
		}
	}

	if hex.EncodeToString(buf.Bytes()) != "122a0a2212205822d187bd40b04cc8ae7437888ebf844efac1729e098c8816d585d0fcc42b5b1200188e8010122a0a2212200b79badee10dc3f7781a7a9d0f020cc0f710b328c4975c2dbc30a170cd188e2c1200188e8010122a0a22122022ad631c69ee983095b5b8acd029ff94aff1dc6c48837878589a92b90dfea3171200188e8010122a0a221220df7fd08c4784fe6938c640df473646e4f16c7d0c6567ab79ec6981767fc3f01a1200188e8010122a0a22122000888c815ad7d055377bdb7b7779fc9740e548cb5dac90c71b9af9f51a879c2d1200188e8010122a0a221220766db372d015c5c700f538336556370165c889334791487a5e48d6080f1c99ea1200188e8010122a0a2212202f533004ceed74279b32c58eb0e3d2a23bc27ba14ab07298406c42bab8d543211200188e8010122a0a2212204c50cfdefa0209766f885919ac8ffc258e9253c3001ac23814f875d414d394731200188e8010122a0a22122000894611dfa192853020cbbade1a9a0a3f359d26e0d38caf4d72b9b306ff5a0b1200188e8010122a0a221220730ddba83e3147bbe10780b97ff0718c74c36037b97b3b79b45c4511806545811200188e8010122a0a22122048ea9d5d423f678d83d559d2349be8325527290b070c90fc1acd968f0bf70a061200188e80100a09736f6d652064617461" {
		t.Fatal("Encoded form did not match expected")
	}
}

func TestNodeWithUnnamedLinks(t *testing.T) {
	dataByts, _ := hex.DecodeString("080218cbc1819201208080e015208080e015208080e015208080e015208080e015208080e01520cbc1c10f")
	expected := pbNode{data: dataByts}
	expected.links = []pbLink{
		{name: "", hasName: true, hash: mkcid(t, "QmSbCgdsX12C4KDw3PDmpBN9iCzS87a5DjgSCoW9esqzXk"), tsize: 45623854, hasTsize: true},
		{name: "", hasName: true, hash: mkcid(t, "Qma4GxWNhywSvWFzPKtEswPGqeZ9mLs2Kt76JuBq9g3fi2"), tsize: 45623854, hasTsize: true},
		{name: "", hasName: true, hash: mkcid(t, "QmQfyxyys7a1e3mpz9XsntSsTGc8VgpjPj5BF1a1CGdGNc"), tsize: 45623854, hasTsize: true},
		{name: "", hasName: true, hash: mkcid(t, "QmSh2wTTZT4N8fuSeCFw7wterzdqbE93j1XDhfN3vQHzDV"), tsize: 45623854, hasTsize: true},
		{name: "", hasName: true, hash: mkcid(t, "QmVXsSVjwxMsCwKRCUxEkGb4f4B98gXVy3ih3v4otvcURK"), tsize: 45623854, hasTsize: true},
		{name: "", hasName: true, hash: mkcid(t, "QmZjhH97MEYwQXzCqSQbdjGDhXWuwW4RyikR24pNqytWLj"), tsize: 45623854, hasTsize: true},
		{name: "", hasName: true, hash: mkcid(t, "QmRs6U5YirCqC7taTynz3x2GNaHJZ3jDvMVAzaiXppwmNJ"), tsize: 32538395, hasTsize: true},
	}

	runTest(
		t,
		"122b0a2212203f29086b59b9e046b362b4b19c9371e834a9f5a80597af83be6d8b7d1a5ad33b120018aed4e015122b0a221220ae1a5afd7c770507dddf17f92bba7a326974af8ae5277c198cf13206373f7263120018aed4e015122b0a22122022ab2ebf9c3523077bd6a171d516ea0e1be1beb132d853778bcc62cd208e77f1120018aed4e015122b0a22122040a77fe7bc69bbef2491f7633b7c462d0bce968868f88e2cbcaae9d0996997e8120018aed4e015122b0a2212206ae1979b14dd43966b0241ebe80ac2a04ad48959078dc5affa12860648356ef6120018aed4e015122b0a221220a957d1f89eb9a861593bfcd19e0637b5c957699417e2b7f23c88653a240836c4120018aed4e015122b0a221220345f9c2137a2cd76d7b876af4bfecd01f80b7dd125f375cb0d56f8a2f96de2c31200189bfec10f0a2b080218cbc1819201208080e015208080e015208080e015208080e015208080e015208080e01520cbc1c10f",
		expected)
}

func TestNodeWithNamedLinks(t *testing.T) {
	dataByts, _ := hex.DecodeString("0801")
	expected := pbNode{data: dataByts}
	expected.links = []pbLink{
		{name: "audio_only.m4a", hasName: true, hash: mkcid(t, "QmaUAwAQJNtvUdJB42qNbTTgDpzPYD1qdsKNtctM5i7DGB"), tsize: 23319629, hasTsize: true},
		{name: "chat.txt", hasName: true, hash: mkcid(t, "QmNVrxbB25cKTRuKg2DuhUmBVEK9NmCwWEHtsHPV6YutHw"), tsize: 996, hasTsize: true},
		{name: "playback.m3u", hasName: true, hash: mkcid(t, "QmUcjKzDLXBPmB6BKHeKSh6ZoFZjss4XDhMRdLYRVuvVfu"), tsize: 116, hasTsize: true},
		{name: "zoom_0.mp4", hasName: true, hash: mkcid(t, "QmQqy2SiEkKgr2cw5UbQ93TtLKEMsD8TdcWggR8q9JabjX"), tsize: 306281879, hasTsize: true},
	}

	runTest(
		t,
		"12390a221220b4397c02da5513563d33eef894bf68f2ccdf1bdfc14a976956ab3d1c72f735a0120e617564696f5f6f6e6c792e6d346118cda88f0b12310a221220025c13fcd1a885df444f64a4a82a26aea867b1148c68cb671e83589f971149321208636861742e74787418e40712340a2212205d44a305b9b328ab80451d0daa72a12a7bf2763c5f8bbe327597a31ee40d1e48120c706c61796261636b2e6d3375187412360a2212202539ed6e85f2a6f9097db9d76cffd49bf3042eb2e3e8e9af4a3ce842d49dea22120a7a6f6f6d5f302e6d70341897fb8592010a020801",
		expected)
}
