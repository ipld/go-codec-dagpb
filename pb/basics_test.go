package pb

import (
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	cid "github.com/ipfs/go-cid"
)

func mkcid(t *testing.T, cidStr string) cid.Cid {
	c, err := cid.Decode(cidStr)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestEmptyNode(t *testing.T) {
	n := NewPBNode()
	byts, err := n.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	rn, err := UnmarshalPBNode(byts)
	if err != nil {
		t.Fatal(err)
	}
	if rn.Data != nil {
		t.Errorf("Expected nil PBNode#Data")
	}
	if rn.Links == nil || len(rn.Links) != 0 {
		t.Errorf("Expected zero-length PBNode#Links")
	}
}

func TestNodeWithData(t *testing.T) {
	n := NewPBNode()
	n.Data = []byte{0, 1, 2, 3, 4}
	byts, err := n.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	rn, err := UnmarshalPBNode(byts)
	if err != nil {
		t.Fatal(err)
	}
	if rn.Data == nil {
		t.Errorf("Expected non-nil PBNode#Data")
	}
	if !reflect.DeepEqual(rn.Data, []byte{0, 1, 2, 3, 4}) {
		t.Errorf("Didn't get expected bytes")
	}
	if rn.Links == nil || len(rn.Links) != 0 {
		t.Errorf("Expected zero-length PBNode#Links")
	}
}

func TestNodeWithLink(t *testing.T) {
	n := NewPBNode()
	n.Links = append(n.Links, NewPBLinkFromCid(mkcid(t, "QmWDtUQj38YLW8v3q4A6LwPn4vYKEbuKWpgSm6bjKW6Xfe")))
	byts, err := n.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	rn, err := UnmarshalPBNode(byts)
	if err != nil {
		t.Fatal(err)
	}
	if rn.Data != nil {
		t.Errorf("Expected nil PBNode#Data")
	}
	if rn.Links == nil || len(rn.Links) != 1 {
		t.Errorf("Expected one PBNode#Links element")
	}
	if rn.Links[0].Tsize != nil || rn.Links[0].Name != nil {
		t.Errorf("Expected Tsize and Name to be nil")
	}
	if rn.Links[0].Hash.String() != "QmWDtUQj38YLW8v3q4A6LwPn4vYKEbuKWpgSm6bjKW6Xfe" {
		t.Errorf("Got unexpected Hash")
	}
}

func TestNodeWithTwoUnsortedLinks(t *testing.T) {
	n := NewPBNodeFromData([]byte("some data"))
	n.Links = append(n.Links, NewPBLink("some other link", mkcid(t, "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39V"), 8))
	n.Links = append(n.Links, NewPBLink("some link", mkcid(t, "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39U"), 100000000))
	byts, err := n.Marshal()
	if err == nil || !strings.Contains(err.Error(), "links must be sorted") || byts != nil {
		t.Fatal("Expected to error from unsorted links")
	}

	n.SortLinks()
	if byts, err = n.Marshal(); err != nil {
		t.Fatal(err)
	}

	expectedBytes := "12340a2212208ab7a6c5e74737878ac73863cb76739d15d4666de44e5756bf55a2f9e9ab5f431209736f6d65206c696e6b1880c2d72f12370a2212208ab7a6c5e74737878ac73863cb76739d15d4666de44e5756bf55a2f9e9ab5f44120f736f6d65206f74686572206c696e6b18080a09736f6d652064617461"
	if hex.EncodeToString(byts) != expectedBytes {
		t.Fatal("Did not get expected bytes")
	}

	rn, err := UnmarshalPBNode(byts)
	if err != nil {
		t.Fatal(err)
	}
	if rn.Data == nil || !reflect.DeepEqual(rn.Data, []byte("some data")) {
		t.Errorf("Did not match expected PBNode#Data")
	}
	if rn.Links == nil || len(rn.Links) != 2 {
		t.Errorf("Expected two PBNode#Links elements")
	}
	if rn.Links[0].Tsize == nil || rn.Links[0].Name == nil || rn.Links[1].Tsize == nil || rn.Links[1].Name == nil {
		t.Errorf("Expected Tsize and Name to not be nil")
	}
	if rn.Links[0].Hash.String() != "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39U" {
		t.Errorf("Got unexpected Hash 0")
	}
	if rn.Links[1].Hash.String() != "QmXg9Pp2ytZ14xgmQjYEiHjVjMFXzCVVEcRTWJBmLgR39V" {
		t.Errorf("Got unexpected Hash 1")
	}
	if *rn.Links[0].Name != "some link" {
		t.Errorf("Got unexpected Name 0")
	}
	if *rn.Links[1].Name != "some other link" {
		t.Errorf("Got unexpected Name 1")
	}
	if *rn.Links[0].Tsize != 100000000 {
		t.Errorf("Got unexpected Tsize 0")
	}
	if *rn.Links[1].Tsize != 8 {
		t.Errorf("Got unexpected Tsize 1")
	}
}

func TestNodeWithStableSortedLinks(t *testing.T) {
	n := NewPBNodeFromData([]byte("some data"))
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

	// name is the same for all, they should remain unsorted
	for _, c := range cids {
		n.Links = append(n.Links, NewPBLink("", mkcid(t, c), 262158))
	}
	byts, err := n.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	n.SortLinks()
	byts2, err := n.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(byts2, byts) {
		t.Errorf("Sort was not stable for unnamed links")
	}

	rn, err := UnmarshalPBNode(byts)
	if err != nil {
		t.Fatal(err)
	}

	for i, link := range rn.Links {
		if link.Hash.String() != cids[i] {
			t.Errorf("Link #%d does not match expected (%v)", i, cids[i])
		}
	}
}

func TestNodeWithUnnamedLinksFixture(t *testing.T) {
	fixture := "122b0a2212203f29086b59b9e046b362b4b19c9371e834a9f5a80597af83be6d8b7d1a5ad33b120018aed4e015122b0a221220ae1a5afd7c770507dddf17f92bba7a326974af8ae5277c198cf13206373f7263120018aed4e015122b0a22122022ab2ebf9c3523077bd6a171d516ea0e1be1beb132d853778bcc62cd208e77f1120018aed4e015122b0a22122040a77fe7bc69bbef2491f7633b7c462d0bce968868f88e2cbcaae9d0996997e8120018aed4e015122b0a2212206ae1979b14dd43966b0241ebe80ac2a04ad48959078dc5affa12860648356ef6120018aed4e015122b0a221220a957d1f89eb9a861593bfcd19e0637b5c957699417e2b7f23c88653a240836c4120018aed4e015122b0a221220345f9c2137a2cd76d7b876af4bfecd01f80b7dd125f375cb0d56f8a2f96de2c31200189bfec10f0a2b080218cbc1819201208080e015208080e015208080e015208080e015208080e015208080e01520cbc1c10f"
	expectedLinks := []*PBLink{
		NewPBLink("", mkcid(t, "QmSbCgdsX12C4KDw3PDmpBN9iCzS87a5DjgSCoW9esqzXk"), 45623854),
		NewPBLink("", mkcid(t, "Qma4GxWNhywSvWFzPKtEswPGqeZ9mLs2Kt76JuBq9g3fi2"), 45623854),
		NewPBLink("", mkcid(t, "QmQfyxyys7a1e3mpz9XsntSsTGc8VgpjPj5BF1a1CGdGNc"), 45623854),
		NewPBLink("", mkcid(t, "QmSh2wTTZT4N8fuSeCFw7wterzdqbE93j1XDhfN3vQHzDV"), 45623854),
		NewPBLink("", mkcid(t, "QmVXsSVjwxMsCwKRCUxEkGb4f4B98gXVy3ih3v4otvcURK"), 45623854),
		NewPBLink("", mkcid(t, "QmZjhH97MEYwQXzCqSQbdjGDhXWuwW4RyikR24pNqytWLj"), 45623854),
		NewPBLink("", mkcid(t, "QmRs6U5YirCqC7taTynz3x2GNaHJZ3jDvMVAzaiXppwmNJ"), 32538395),
	}

	byts, err := hex.DecodeString(fixture)
	rn, err := UnmarshalPBNode(byts)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rn.Links, expectedLinks) {
		t.Errorf("Didn't get expected Links array")
	}
}

func TestNodeWithNamedLinksFixture(t *testing.T) {
	fixture := "12390a221220b4397c02da5513563d33eef894bf68f2ccdf1bdfc14a976956ab3d1c72f735a0120e617564696f5f6f6e6c792e6d346118cda88f0b12310a221220025c13fcd1a885df444f64a4a82a26aea867b1148c68cb671e83589f971149321208636861742e74787418e40712340a2212205d44a305b9b328ab80451d0daa72a12a7bf2763c5f8bbe327597a31ee40d1e48120c706c61796261636b2e6d3375187412360a2212202539ed6e85f2a6f9097db9d76cffd49bf3042eb2e3e8e9af4a3ce842d49dea22120a7a6f6f6d5f302e6d70341897fb8592010a020801"
	expectedLinks := []*PBLink{
		NewPBLink("audio_only.m4a", mkcid(t, "QmaUAwAQJNtvUdJB42qNbTTgDpzPYD1qdsKNtctM5i7DGB"), 23319629),
		NewPBLink("chat.txt", mkcid(t, "QmNVrxbB25cKTRuKg2DuhUmBVEK9NmCwWEHtsHPV6YutHw"), 996),
		NewPBLink("playback.m3u", mkcid(t, "QmUcjKzDLXBPmB6BKHeKSh6ZoFZjss4XDhMRdLYRVuvVfu"), 116),
		NewPBLink("zoom_0.mp4", mkcid(t, "QmQqy2SiEkKgr2cw5UbQ93TtLKEMsD8TdcWggR8q9JabjX"), 306281879),
	}

	byts, err := hex.DecodeString(fixture)
	rn, err := UnmarshalPBNode(byts)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rn.Links, expectedLinks) {
		t.Errorf("Didn't get expected Links array")
	}
}
