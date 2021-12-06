package types

import (
	"fmt"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/okex/exchain/libs/cosmos-sdk/store/types"
	"github.com/okex/exchain/libs/tendermint/crypto"
	"github.com/okex/exchain/libs/tendermint/libs/log"
	"github.com/spf13/viper"
	"runtime/debug"
	"time"
)

var (
	maxAccInMap        = 1000000
	deleteAccCount     = 100000
	maxStorageInMap    = 100000
	deleteStorageCount = 10000
)
var (
	FlagMultiCache = "multi-cache"
)

type account interface {
	Copy() interface{}
	GetAddress() AccAddress
	SetAddress(AccAddress) error
	GetPubKey() crypto.PubKey
	SetPubKey(crypto.PubKey) error
	GetAccountNumber() uint64
	SetAccountNumber(uint64) error
	GetSequence() uint64
	SetSequence(uint64) error
	GetCoins() Coins
	SetCoins(Coins) error
	SpendableCoins(blockTime time.Time) Coins
	String() string
}

type storageWithCache struct {
	value []byte
	dirty bool
}

type accountWithCache struct {
	acc     account
	gas     uint64
	isDirty bool
}

type codeWithCache struct {
	code    []byte
	isDirty bool
}

type Cache struct {
	useCache bool
	parent   *Cache

	storageMap map[ethcmn.Address]map[ethcmn.Hash]*storageWithCache
	accMap     map[ethcmn.Address]*accountWithCache

	codeMap map[ethcmn.Hash]*codeWithCache

	gasConfig types.GasConfig
}

var (
	UseCache bool
)

func NewChainCache() *Cache {
	UseCache = viper.GetBool(FlagMultiCache)
	return NewCache(nil, UseCache)
}

func NewCache(parent *Cache, useCache bool) *Cache {
	return &Cache{
		useCache: useCache,
		parent:   parent,

		storageMap: make(map[ethcmn.Address]map[ethcmn.Hash]*storageWithCache, 0),
		accMap:     make(map[ethcmn.Address]*accountWithCache, 0),
		codeMap:    make(map[ethcmn.Hash]*codeWithCache),
		gasConfig:  types.KVGasConfig(),
	}

}

func (c *Cache) skip() bool {
	if c == nil || !c.useCache {
		return true
	}
	return false
}

func (c *Cache) UpdateStorage(addr ethcmn.Address, key ethcmn.Hash, value []byte, isDirty bool) {
	if c.skip() {
		return
	}

	if _, ok := c.storageMap[addr]; !ok {
		c.storageMap[addr] = make(map[ethcmn.Hash]*storageWithCache, 0)
	}
	c.storageMap[addr][key] = &storageWithCache{
		value: value,
		dirty: isDirty,
	}
}

func (c *Cache) UpdateCode(key []byte, value []byte, isdirty bool) {
	if c.skip() {
		return
	}
	hash := ethcmn.BytesToHash(key)
	c.codeMap[hash] = &codeWithCache{
		code:    value,
		isDirty: isdirty,
	}
}

func (c *Cache) GetCode(key []byte) ([]byte, bool) {
	if c.skip() {
		return nil, false
	}

	hash := ethcmn.BytesToHash(key)
	if data, ok := c.codeMap[hash]; ok {
		return data.code, ok
	}

	if c.parent != nil {
		return c.parent.GetCode(hash.Bytes())
	}
	return nil, false
}

func (c *Cache) UpdateAccount(addr AccAddress, acc account, lenBytes int, isDirty bool) {
	if c.skip() {
		return
	}
	ethAddr := ethcmn.BytesToAddress(addr.Bytes())
	if acc != nil {
		
		if acc.GetCoins().String() == "0.000002030496252228aas-d11,0.547561210262965442btc,1343473.043753106013862341btck-e78,39522.597821480591003971dus-959,1132605.038760198798031833ethk-827,2265.450624580574696999ftd-720,19333.166456883550845229ipw-f86,18619.322343938277598605mlx-1f1,40668.323113147134496846oik-575,1032.601005260547909718okb,60.000000000000000000okb-bac,2388955.917713121620359070okt,4548.808364138032726693pha-142,19546.670287555966091290twa-47d,12514.578741142523766885usdk,96126027.183177901010797472usdt-25a,4551.417799438166720356wqc-fe9,19286.488868259919162130wxu-d07,4535.438676266477139488ydj-c4d,0.002161956414037149zat-000,0.003094890981049282zbt-8bc,4.444770525000000000zct-c3e" {
			fmt.Println("settttt", ethAddr.String(), acc.GetCoins().String())
			debug.PrintStack()
		}
	}
	c.accMap[ethAddr] = &accountWithCache{
		acc:     acc,
		isDirty: isDirty,
		gas:     types.Gas(lenBytes)*c.gasConfig.ReadCostPerByte + c.gasConfig.ReadCostFlat,
	}
}

func (c *Cache) GetAccount(addr ethcmn.Address) (account, uint64, bool, bool) {
	if c.skip() {
		return nil, 0, false, false
	}

	if data, ok := c.accMap[addr]; ok {
		return data.acc, data.gas, ok, false
	}

	if c.parent != nil {
		addr, gas, ok, _ := c.parent.GetAccount(addr)
		return addr, gas, ok, true
	}
	return nil, 0, false, false

}

func (c *Cache) GetStorage(addr ethcmn.Address, key ethcmn.Hash) ([]byte, bool) {
	if c.skip() {
		return nil, false
	}
	if _, hasAddr := c.storageMap[addr]; hasAddr {
		data, hasKey := c.storageMap[addr][key]
		if hasKey {
			return data.value, hasKey
		}
	}

	if c.parent != nil {
		return c.parent.GetStorage(addr, key)
	}
	return nil, false
}

func (c *Cache) Write(updateDirty bool) {
	if c.skip() {
		return
	}

	if c.parent == nil {
		return
	}
	c.writeStorage(updateDirty)
	c.writeAcc(updateDirty)
	c.writeCode(updateDirty)
}

func needUpdate(updateDirty bool, isDirty bool) bool {
	//Read-Only Data
	if !isDirty {
		return true
	}

	// Dirty Data
	if updateDirty {
		return true
	}
	return false
}

func (c *Cache) writeStorage(updateDirty bool) {
	for addr, storages := range c.storageMap {
		if _, ok := c.parent.storageMap[addr]; !ok {
			c.parent.storageMap[addr] = make(map[ethcmn.Hash]*storageWithCache, 0)
		}

		for key, v := range storages {
			if needUpdate(updateDirty, v.dirty) {
				c.parent.storageMap[addr][key] = v
			}
		}
	}
	c.storageMap = make(map[ethcmn.Address]map[ethcmn.Hash]*storageWithCache)
}

func (c *Cache) writeAcc(updateDirty bool) {
	for addr, v := range c.accMap {
		if needUpdate(updateDirty, v.isDirty) {
			if v != nil && v.acc != nil {
				fmt.Println("update", updateDirty, v.isDirty, addr.String(), v.acc.GetCoins().String())
			}
			c.parent.accMap[addr] = v
		}
	}
	c.accMap = make(map[ethcmn.Address]*accountWithCache)
}
func (c *Cache) writeCode(updateDirty bool) {
	for hash, v := range c.codeMap {
		if needUpdate(updateDirty, v.isDirty) {
			c.parent.codeMap[hash] = v
		}
	}
	c.codeMap = make(map[ethcmn.Hash]*codeWithCache)
}

func (c *Cache) TryDelete(logger log.Logger, height int64) {
	if c.skip() {
		return
	}
	if height%1000 == 0 {
		logger.Info("MultiCache:info", "len(acc)", len(c.accMap), "len(storage)", len(c.storageMap))
	}

	if len(c.accMap) < maxAccInMap && len(c.storageMap) < maxStorageInMap {
		return
	}

	ts := time.Now()
	isDelete := false
	if len(c.accMap) >= maxAccInMap {
		isDelete = true
		cnt := 0
		for key := range c.accMap {
			delete(c.accMap, key)
			cnt++
			if cnt > deleteAccCount {
				break
			}
		}
	}

	if len(c.storageMap) >= maxStorageInMap {
		isDelete = true
		cnt := 0
		for key := range c.storageMap {
			delete(c.storageMap, key)
			cnt++
			if cnt > deleteStorageCount {
				break
			}
		}
	}
	if isDelete {
		logger.Info("MultiCache:info", "time", time.Now().Sub(ts).Seconds(), "len(acc)", len(c.accMap), "len(storage)", len(c.storageMap))
	}
}
