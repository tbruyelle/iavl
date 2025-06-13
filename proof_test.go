// nolint: errcheck
package iavl

import (
	"bytes"
	"fmt"
	"sort"
	"testing"

	"github.com/davecgh/go-spew/spew"
	treed "github.com/m1gwings/treedrawer/tree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iavlrand "github.com/cosmos/iavl/internal/rand"
)

func TestTreeGetProof(t *testing.T) {
	require := require.New(t)
	tree := getTestTree(0)
	for _, ikey := range []byte{0x11, 0x32, 0x50, 0x72, 0x99} {
		key := []byte{ikey}
		tree.Set(key, []byte(iavlrand.RandStr(8)))
	}
	v, i, err := tree.SaveVersion()
	fmt.Printf("SAVE %X %d %v\n", v, i, err)
	fmt.Println()
	var addTree func(*Node, *treed.Tree)
	addTree = func(n *Node, x *treed.Tree) {
		if n == nil {
			return
		}
		x = x.AddChild(treed.NodeString(fmt.Sprintf("%X / %.3X", n.key, n.hash)))
		// x = x.AddChild(treed.NodeString(fmt.Sprintf("%X", n.key)))
		if n.leftNodeKey != nil {
			nn, err := n.getLeftNode(tree.ImmutableTree)
			if err != nil {
				panic(err)
			}
			addTree(nn, x)
		}
		if n.rightNodeKey != nil {
			nn, err := n.getRightNode(tree.ImmutableTree)
			if err != nil {
				panic(err)
			}
			addTree(nn, x)
		}
	}
	x := treed.NewTree(treed.NodeString("start"))
	addTree(tree.root, x)
	fmt.Println(x)

	key := []byte{0x32}
	proof, err := tree.GetMembershipProof(key)
	require.NoError(err)
	require.NotNil(proof)

	res, err := tree.VerifyMembership(proof, key)
	require.NoError(err, "%+v", err)
	require.True(res)

	key = []byte{0x33}
	proof, err = tree.GetNonMembershipProof(key)
	require.NoError(err)
	require.NotNil(proof)
	spew.Config.DisableMethods = true
	spew.Dump(proof)
	fmt.Printf("left=%X right=%X\n", proof.GetNonexist().Left.GetKey(), proof.GetNonexist().Right.GetKey())

	res, err = tree.VerifyNonMembership(proof, key)
	require.NoError(err, "%+v", err)
	require.True(res)
}

func TestTreeKeyExistsProof(t *testing.T) {
	tree := getTestTree(0)

	// should get error
	_, err := tree.GetProof([]byte("foo"))
	assert.Error(t, err)

	// insert lots of info and store the bytes
	allkeys := make([][]byte, 200)
	for i := 0; i < 200; i++ {
		key := iavlrand.RandStr(20)
		value := "value_for_" + key
		tree.Set([]byte(key), []byte(value))
		allkeys[i] = []byte(key)
	}
	sortByteSlices(allkeys) // Sort all keys

	// query random key fails
	_, err = tree.GetMembershipProof([]byte("foo"))
	require.Error(t, err)

	// valid proof for real keys
	for _, key := range allkeys {
		proof, err := tree.GetMembershipProof(key)
		require.NoError(t, err)
		require.Equal(t,
			append([]byte("value_for_"), key...),
			proof.GetExist().Value,
		)

		res, err := tree.VerifyMembership(proof, key)
		require.NoError(t, err)
		require.True(t, res)
	}
}

//----------------------------------------

// Contract: !bytes.Equal(input, output) && len(input) >= len(output)
func MutateByteSlice(bytez []byte) []byte {
	// If bytez is empty, panic
	if len(bytez) == 0 {
		panic("Cannot mutate an empty bytez")
	}

	// Copy bytez
	mBytez := make([]byte, len(bytez))
	copy(mBytez, bytez)
	bytez = mBytez

	// Try a random mutation
	switch iavlrand.RandInt() % 2 {
	case 0: // Mutate a single byte
		bytez[iavlrand.RandInt()%len(bytez)] += byte(iavlrand.RandInt()%255 + 1)
	case 1: // Remove an arbitrary byte
		pos := iavlrand.RandInt() % len(bytez)
		bytez = append(bytez[:pos], bytez[pos+1:]...)
	}
	return bytez
}

func sortByteSlices(src [][]byte) [][]byte {
	bzz := byteslices(src)
	sort.Sort(bzz)
	return bzz
}

type byteslices [][]byte

func (bz byteslices) Len() int {
	return len(bz)
}

func (bz byteslices) Less(i, j int) bool {
	switch bytes.Compare(bz[i], bz[j]) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		panic("should not happen")
	}
}

func (bz byteslices) Swap(i, j int) {
	bz[j], bz[i] = bz[i], bz[j]
}
