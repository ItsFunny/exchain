package types

import (
	"fmt"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/okex/exchain/libs/cosmos-sdk/store/types"
	"github.com/okex/exchain/libs/tendermint/crypto"
	"github.com/okex/exchain/libs/tendermint/libs/log"
	"github.com/spf13/viper"
	"reflect"
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
		fmt.Println("return ")
		return
	}
	fmt.Printf("not return")

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
	hash := ethcmn.BytesToHash(key)
	if c.skip() {
		return nil, false
	}

	if data, ok := c.codeMap[hash]; ok {
		return data.code, ok
	}

	if c.parent != nil {
		return c.parent.GetCode(hash.Bytes())
	}
	return nil, false
}

var (
	TypeMap = make(map[reflect.Type]bool)
)

func (c *Cache) UpdateAccount(addr AccAddress, acc account, lenBytes int, isDirty bool) {
	if c.skip() {
		return
	}
	ethAddr := ethcmn.BytesToAddress(addr.Bytes())
	TypeMap[reflect.TypeOf(acc)] = true
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

func (c *Cache) writeStorage(updateDirty bool) {
	for addr, storages := range c.storageMap {
		if _, ok := c.parent.storageMap[addr]; !ok {
			c.parent.storageMap[addr] = make(map[ethcmn.Hash]*storageWithCache, 0)
		}

		for key, v := range storages {
			if !v.dirty || (updateDirty && v.dirty) {
				c.parent.storageMap[addr][key] = v
			}
		}
	}
	c.storageMap = make(map[ethcmn.Address]map[ethcmn.Hash]*storageWithCache)
}

func (c *Cache) writeAcc(updateDirty bool) {
	for addr, v := range c.accMap {
		if !v.isDirty || (updateDirty && v.isDirty) {
			c.parent.accMap[addr] = v
		}
	}
	c.accMap = make(map[ethcmn.Address]*accountWithCache)
}
func (c *Cache) writeCode(updateDirty bool) {
	for hash, v := range c.codeMap {
		if !v.isDirty || (updateDirty && v.isDirty) {
			c.parent.codeMap[hash] = v
		}
	}
	c.codeMap = make(map[ethcmn.Hash]*codeWithCache)
}

func (c *Cache) Delete(logger log.Logger, height int64) {
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