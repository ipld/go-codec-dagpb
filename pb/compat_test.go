package pb

// mirrored in JavaScript @ https://github.com/ipld/js-dag-pb/blob/master/test/test-compat.js

import (
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	cid "github.com/ipfs/go-cid"
)

var dataZero []byte = make([]byte, 0)
var dataSome []byte = []byte{0, 1, 2, 3, 4}
var acid cid.Cid = _mkcid()
var zeroName string = ""
var someName string = "some name"
var zeroTsize uint64 = 0
var someTsize uint64 = 1010
var largeTsize uint64 = 9007199254740991 // JavaScript Number.MAX_SAFE_INTEGER

type testCase struct {
	name          string
	node          *PBNode
	expectedBytes string
	expectedForm  string
	encodeError   string
	decodeError   string
}

var testCases = []testCase{
	{
		name:          "empty",
		node:          &PBNode{},
		expectedBytes: "",
		expectedForm: `{
	"Links": []
}`,
		encodeError: "Links must be an array",
	},
	{
		name:          "Data zero",
		node:          &PBNode{Data: dataZero},
		expectedBytes: "0a00",
		expectedForm: `{
	"Data": "",
	"Links": []
}`,
		encodeError: "Links must be an array",
	},
	{
		name:          "Data some",
		node:          &PBNode{Data: dataSome},
		expectedBytes: "0a050001020304",
		expectedForm: `{
	"Data": "0001020304",
	"Links": []
}`,
		encodeError: "Links must be an array",
	},
	{
		name:          "Links zero",
		node:          &PBNode{Links: make([]*PBLink, 0)},
		expectedBytes: "",
		expectedForm: `{
	"Links": []
}`,
	},
	{
		name:          "Data some Links zero",
		node:          &PBNode{Data: dataSome, Links: make([]*PBLink, 0)},
		expectedBytes: "0a050001020304",
		expectedForm: `{
	"Data": "0001020304",
	"Links": []
}`,
	},
	{
		name:          "Links empty",
		node:          &PBNode{Links: []*PBLink{{}}},
		expectedBytes: "1200",
		encodeError:   "link must have a Hash",
		decodeError:   "expected CID",
	},
	{
		name:          "Data some Links empty",
		node:          &PBNode{Data: dataSome, Links: []*PBLink{{}}},
		expectedBytes: "12000a050001020304",
		encodeError:   "link must have a Hash",
		decodeError:   "expected CID",
	},
	{
		name:          "Links Hash zero",
		expectedBytes: "12020a00",
		decodeError:   "expected CID", // error should come up from go-cid too
	},
	{
		name:          "Links Hash some",
		node:          &PBNode{Links: []*PBLink{{Hash: &acid}}},
		expectedBytes: "120b0a09015500050001020304",
		expectedForm: `{
	"Links": [
		{
			"Hash": "015500050001020304"
		}
	]
}`,
	},
	{
		name:          "Links Name zero",
		node:          &PBNode{Links: []*PBLink{{Name: &zeroName}}},
		expectedBytes: "12021200",
		encodeError:   "link must have a Hash",
		decodeError:   "expected CID",
	},
	{
		name:          "Links Hash some Name zero",
		node:          &PBNode{Links: []*PBLink{{Hash: &acid, Name: &zeroName}}},
		expectedBytes: "120d0a090155000500010203041200",
		expectedForm: `{
	"Links": [
		{
			"Hash": "015500050001020304",
			"Name": ""
		}
	]
}`,
	},
	{
		name:          "Links Name some",
		node:          &PBNode{Links: []*PBLink{{Name: &someName}}},
		expectedBytes: "120b1209736f6d65206e616d65",
		encodeError:   "link must have a Hash",
		decodeError:   "expected CID",
	},
	{
		name:          "Links Hash some Name some",
		node:          &PBNode{Links: []*PBLink{{Hash: &acid, Name: &someName}}},
		expectedBytes: "12160a090155000500010203041209736f6d65206e616d65",
		expectedForm: `{
	"Links": [
		{
			"Hash": "015500050001020304",
			"Name": "some name"
		}
	]
}`,
	},
	{
		name:          "Links Tsize zero",
		node:          &PBNode{Links: []*PBLink{{Tsize: &zeroTsize}}},
		expectedBytes: "12021800",
		encodeError:   "link must have a Hash",
		decodeError:   "expected CID",
	},
	{
		name:          "Links Hash some Tsize zero",
		node:          &PBNode{Links: []*PBLink{{Hash: &acid, Tsize: &zeroTsize}}},
		expectedBytes: "120d0a090155000500010203041800",
		expectedForm: `{
	"Links": [
		{
			"Hash": "015500050001020304",
			"Tsize": 0
		}
	]
}`,
	},
	{
		name:          "Links Tsize some",
		node:          &PBNode{Links: []*PBLink{{Tsize: &someTsize}}},
		expectedBytes: "120318f207",
		encodeError:   "link must have a Hash",
		decodeError:   "expected CID",
	},
	{
		name:          "Links Hash some Tsize some",
		node:          &PBNode{Links: []*PBLink{{Hash: &acid, Tsize: &largeTsize}}},
		expectedBytes: "12140a0901550005000102030418ffffffffffffff0f",
		expectedForm: `{
	"Links": [
		{
			"Hash": "015500050001020304",
			"Tsize": 9007199254740991
		}
	]
}`,
	},
}

func TestCompat(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			verifyRoundTrip(t, tc)
		})
	}
}

func verifyRoundTrip(t *testing.T, tc testCase) {
	var err error
	var actualBytes string
	var actualForm string
	if tc.node != nil {
		actualBytes, err = nodeToString(t, tc.node)
		if tc.encodeError != "" {
			if err != nil {
				if !strings.Contains(err.Error(), tc.encodeError) {
					t.Fatalf("got unexpeced encode error: [%v] (expected [%v])", err.Error(), tc.encodeError)
				}
			} else {
				t.Fatalf("did not get expected encode error: %v", tc.encodeError)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			} else {
				if actualBytes != tc.expectedBytes {
					t.Logf(
						"Expected bytes: [%v]\nGot: [%v]\n",
						tc.expectedBytes,
						actualBytes)
					t.Error("Did not match")
				}
			}
		}
	}

	actualForm, err = bytesToFormString(t, tc.expectedBytes)
	if tc.decodeError != "" {
		if err != nil {
			if !strings.Contains(err.Error(), tc.decodeError) {
				t.Fatalf("got unexpeced decode error: [%v] (expected [%v])", err.Error(), tc.decodeError)
			}
		} else {
			t.Fatalf("did not get expected decode error: %v", tc.decodeError)
		}
	} else {
		if err != nil {
			t.Fatal(err)
		}
		if actualForm != tc.expectedForm {
			t.Logf(
				"Expected form: [%v]\nGot: [%v]\n",
				tc.expectedForm,
				actualForm)
			t.Error("Did not match")
		}
	}
}

func nodeToString(t *testing.T, n *PBNode) (string, error) {
	bytes, err := n.Marshal()
	if err != nil {
		return "", err
	}
	t.Logf("[%v]\n", hex.EncodeToString(bytes))
	return hex.EncodeToString(bytes), nil
}

func bytesToFormString(t *testing.T, bytesHex string) (string, error) {
	bytes, err := hex.DecodeString(bytesHex)
	if err != nil {
		return "", err
	}
	var rt *PBNode
	if rt, err = UnmarshalPBNode(bytes); err != nil {
		return "", err
	}
	str, err := json.MarshalIndent(cleanPBNode(t, rt), "", "\t")
	if err != nil {
		return "", err
	}
	return string(str), nil
}

// convert a PBLink into a map for clean JSON marshalling
func cleanPBLink(t *testing.T, link *PBLink) map[string]interface{} {
	if link == nil {
		return nil
	}
	nl := make(map[string]interface{})
	if link.Hash != nil {
		nl["Hash"] = hex.EncodeToString(link.Hash.Bytes())
	}
	if link.Name != nil {
		nl["Name"] = link.Name
	}
	if link.Tsize != nil {
		nl["Tsize"] = link.Tsize
	}
	return nl
}

// convert a PBNode into a map for clean JSON marshalling
func cleanPBNode(t *testing.T, node *PBNode) map[string]interface{} {
	nn := make(map[string]interface{})
	if node.Data != nil {
		nn["Data"] = hex.EncodeToString(node.Data)
	}
	if node.Links != nil {
		links := make([]map[string]interface{}, len(node.Links))
		for i, l := range node.Links {
			links[i] = cleanPBLink(t, l)
		}
		nn["Links"] = links
	}
	return nn
}

func _mkcid() cid.Cid {
	_, c, _ := cid.CidFromBytes([]byte{1, 85, 0, 5, 0, 1, 2, 3, 4})
	return c
}
