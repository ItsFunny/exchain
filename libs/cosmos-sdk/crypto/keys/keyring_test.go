// nolint: goconst
package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/okex/exchain/libs/tendermint/crypto"
	"github.com/okex/exchain/libs/tendermint/crypto/ed25519"

	"github.com/okex/exchain/libs/cosmos-sdk/crypto/keys/hd"
	"github.com/okex/exchain/libs/cosmos-sdk/tests"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
)

func TestLazyKeyManagementKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)

	algo := Secp256k1
	n1, n2, n3 := "personal", "business", "other"
	p1, p2 := "1234", "really-secure!@#$"

	// Check empty state
	l, err := kb.List()
	require.Nil(t, err)
	assert.Empty(t, l)

	_, _, err = kb.CreateMnemonic(n1, English, p1, Ed25519, "")
	require.Error(t, err, "ed25519 keys are currently not supported by keybase")

	// create some keys
	_, err = kb.Get(n1)
	require.Error(t, err)
	i, _, err := kb.CreateMnemonic(n1, English, p1, algo, "")

	require.NoError(t, err)
	require.Equal(t, n1, i.GetName())
	_, _, err = kb.CreateMnemonic(n2, English, p2, algo, "")
	require.NoError(t, err)

	// we can get these keys
	i2, err := kb.Get(n2)
	require.NoError(t, err)
	_, err = kb.Get(n3)
	require.NotNil(t, err)
	_, err = kb.GetByAddress(accAddr(i2))
	require.NoError(t, err)
	addr, err := sdk.AccAddressFromBech32("cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t")
	require.NoError(t, err)
	_, err = kb.GetByAddress(addr)
	require.NotNil(t, err)

	// list shows them in order
	keyS, err := kb.List()
	require.NoError(t, err)
	require.Equal(t, 2, len(keyS))
	// note these are in alphabetical order
	require.Equal(t, n2, keyS[0].GetName())
	require.Equal(t, n1, keyS[1].GetName())
	require.Equal(t, i2.GetPubKey(), keyS[0].GetPubKey())

	// deleting a key removes it
	err = kb.Delete("bad name", "foo", false)
	require.NotNil(t, err)
	err = kb.Delete(n1, p1, false)
	require.NoError(t, err)
	keyS, err = kb.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(keyS))
	_, err = kb.Get(n1)
	require.Error(t, err)

	// create an offline key
	o1 := "offline"
	priv1 := ed25519.GenPrivKey()
	pub1 := priv1.PubKey()
	i, err = kb.CreateOffline(o1, pub1, Ed25519)
	require.Nil(t, err)
	require.Equal(t, pub1, i.GetPubKey())
	require.Equal(t, o1, i.GetName())
	keyS, err = kb.List()
	require.NoError(t, err)
	require.Equal(t, 2, len(keyS))

	// delete the offline key
	err = kb.Delete(o1, "", false)
	require.NoError(t, err)
	keyS, err = kb.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(keyS))

	// addr cache gets nuked - and test skip flag
	err = kb.Delete(n2, "", true)
	require.NoError(t, err)
}

func TestLazySignVerifyKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)
	algo := Secp256k1

	n1, n2, n3 := "some dude", "a dudette", "dude-ish"
	p1, p2, p3 := "1234", "foobar", "foobar"

	// create two users and get their info
	i1, _, err := kb.CreateMnemonic(n1, English, p1, algo, "")
	require.Nil(t, err)

	i2, _, err := kb.CreateMnemonic(n2, English, p2, algo, "")
	require.Nil(t, err)

	// Import a public key
	armor, err := kb.ExportPubKey(n2)
	require.Nil(t, err)
	kb.ImportPubKey(n3, armor)
	i3, err := kb.Get(n3)
	require.NoError(t, err)
	require.Equal(t, i3.GetName(), n3)

	// let's try to sign some messages
	d1 := []byte("my first message")
	d2 := []byte("some other important info!")
	d3 := []byte("feels like I forgot something...")

	// try signing both data with both ..
	s11, pub1, err := kb.Sign(n1, p1, d1)
	require.Nil(t, err)
	require.Equal(t, i1.GetPubKey(), pub1)

	s12, pub1, err := kb.Sign(n1, p1, d2)
	require.Nil(t, err)
	require.Equal(t, i1.GetPubKey(), pub1)

	s21, pub2, err := kb.Sign(n2, p2, d1)
	require.Nil(t, err)
	require.Equal(t, i2.GetPubKey(), pub2)

	s22, pub2, err := kb.Sign(n2, p2, d2)
	require.Nil(t, err)
	require.Equal(t, i2.GetPubKey(), pub2)

	// let's try to validate and make sure it only works when everything is proper
	cases := []struct {
		key   crypto.PubKey
		data  []byte
		sig   []byte
		valid bool
	}{
		// proper matches
		{i1.GetPubKey(), d1, s11, true},
		// change data, pubkey, or signature leads to fail
		{i1.GetPubKey(), d2, s11, false},
		{i2.GetPubKey(), d1, s11, false},
		{i1.GetPubKey(), d1, s21, false},
		// make sure other successes
		{i1.GetPubKey(), d2, s12, true},
		{i2.GetPubKey(), d1, s21, true},
		{i2.GetPubKey(), d2, s22, true},
	}

	for i, tc := range cases {
		valid := tc.key.VerifyBytes(tc.data, tc.sig)
		require.Equal(t, tc.valid, valid, "%d", i)
	}

	// Now try to sign data with a secret-less key
	_, _, err = kb.Sign(n3, p3, d3)
	require.NotNil(t, err)
}

func TestLazyExportImportKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)

	info, _, err := kb.CreateMnemonic("john", English, "secretcpw", Secp256k1, "")
	require.NoError(t, err)
	require.Equal(t, info.GetName(), "john")

	john, err := kb.Get("john")
	require.NoError(t, err)
	require.Equal(t, info.GetName(), "john")
	johnAddr := info.GetPubKey().Address()

	armor, err := kb.Export("john")
	require.NoError(t, err)

	err = kb.Import("john2", armor)
	require.NoError(t, err)

	john2, err := kb.Get("john2")
	require.NoError(t, err)

	require.Equal(t, john.GetPubKey().Address(), johnAddr)
	require.Equal(t, john.GetName(), "john")
	require.Equal(t, john, john2)
}

func TestLazyExportImportPubKeyKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)
	algo := Secp256k1

	// CreateMnemonic a private-public key pair and ensure consistency
	notPasswd := "n9y25ah7"
	info, _, err := kb.CreateMnemonic("john", English, notPasswd, algo, "")
	require.Nil(t, err)
	require.NotEqual(t, info, "")
	require.Equal(t, info.GetName(), "john")
	addr := info.GetPubKey().Address()
	john, err := kb.Get("john")
	require.NoError(t, err)
	require.Equal(t, john.GetName(), "john")
	require.Equal(t, john.GetPubKey().Address(), addr)

	// Export the public key only
	armor, err := kb.ExportPubKey("john")
	require.NoError(t, err)
	// Import it under a different name
	err = kb.ImportPubKey("john-pubkey-only", armor)
	require.NoError(t, err)
	// Ensure consistency
	john2, err := kb.Get("john-pubkey-only")
	require.NoError(t, err)
	// Compare the public keys
	require.True(t, john.GetPubKey().Equals(john2.GetPubKey()))
	// Ensure the original key hasn't changed
	john, err = kb.Get("john")
	require.NoError(t, err)
	require.Equal(t, john.GetPubKey().Address(), addr)
	require.Equal(t, john.GetName(), "john")

	// Ensure keys cannot be overwritten
	err = kb.ImportPubKey("john-pubkey-only", armor)
	require.NotNil(t, err)
}

func TestLazyExportPrivateKeyObjectKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)

	info, _, err := kb.CreateMnemonic("john", English, "secretcpw", Secp256k1, "")
	require.NoError(t, err)
	require.Equal(t, info.GetName(), "john")

	// export private key object
	exported, err := kb.ExportPrivateKeyObject("john", "secretcpw")
	require.Nil(t, err, "%+v", err)
	require.True(t, exported.PubKey().Equals(info.GetPubKey()))
}

func TestLazyAdvancedKeyManagementKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)

	algo := Secp256k1
	n1, n2 := "old-name", "new name"
	p1 := "1234"

	// make sure key works with initial password
	_, _, err = kb.CreateMnemonic(n1, English, p1, algo, "")
	require.Nil(t, err, "%+v", err)

	_, err = kb.Export(n1 + ".notreal")
	require.NotNil(t, err)
	_, err = kb.Export(" " + n1)
	require.NotNil(t, err)
	_, err = kb.Export(n1 + " ")
	require.NotNil(t, err)
	_, err = kb.Export("")
	require.NotNil(t, err)
	exported, err := kb.Export(n1)
	require.Nil(t, err, "%+v", err)

	// import succeeds
	err = kb.Import(n2, exported)
	require.NoError(t, err)

	// second import fails
	err = kb.Import(n2, exported)
	require.NotNil(t, err)
}

func TestLazySeedPhraseKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()
	kb, err := NewKeyring("keybasename", "test", dir, nil)
	require.NoError(t, err)

	algo := Secp256k1
	n1, n2 := "lost-key", "found-again"
	p1, p2 := "1234", "foobar"

	// make sure key works with initial password
	info, mnemonic, err := kb.CreateMnemonic(n1, English, p1, algo, "")
	require.Nil(t, err, "%+v", err)
	require.Equal(t, n1, info.GetName())
	assert.NotEmpty(t, mnemonic)

	// now, let us delete this key
	err = kb.Delete(n1, p1, false)
	require.Nil(t, err, "%+v", err)
	_, err = kb.Get(n1)
	require.NotNil(t, err)

	// let us re-create it from the mnemonic-phrase
	params := *hd.NewFundraiserParams(0, sdk.CoinType, 0)
	hdPath := params.String()
	newInfo, err := kb.CreateAccount(n2, mnemonic, DefaultBIP39Passphrase, p2, hdPath, Secp256k1)
	require.NoError(t, err)
	require.Equal(t, n2, newInfo.GetName())
	require.Equal(t, info.GetPubKey().Address(), newInfo.GetPubKey().Address())
	require.Equal(t, info.GetPubKey(), newInfo.GetPubKey())
}
