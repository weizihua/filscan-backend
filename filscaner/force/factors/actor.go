package factors

import (
	"reflect"

	"github.com/filecoin-project/go-address"
//	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	in "github.com/filecoin-project/specs-actors/actors/builtin/init"
	"github.com/filecoin-project/specs-actors/actors/builtin/miner"
	"github.com/filecoin-project/specs-actors/actors/builtin/multisig"
	"github.com/filecoin-project/specs-actors/actors/builtin/market"
	"github.com/filecoin-project/specs-actors/actors/builtin/power"
	"github.com/filecoin-project/specs-actors/actors/builtin/paych"
	"github.com/filecoin-project/specs-actors/actors/runtime"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
)

var (
	null = struct{}{}

	actorInfos = map[cid.Cid]ActorInfo{}

	addressToCode = map[address.Address]cid.Cid{
		builtin.SystemActorAddr:	builtin.SystemActorCodeID,
		builtin.InitActorAddr:          builtin.InitActorCodeID,
		builtin.RewardActorAddr:	builtin.RewardActorCodeID,
		builtin.CronActorAddr:		builtin.CronActorCodeID,
		builtin.StoragePowerActorAddr:  builtin.StoragePowerActorCodeID,
		builtin.StorageMarketActorAddr:	builtin.StorageMarketActorCodeID,
		builtin.VerifiedRegistryActorAddr:builtin.VerifiedRegistryActorCodeID,
	}
)

// reflect types
var (
	TypeNull     = reflect.TypeOf(null)
	TypeNil      = reflect.TypeOf(nil)
	TypeActorPtr = reflect.TypeOf((*types.Actor)(nil))
	TypeVMCtx    = reflect.TypeOf(new(runtime.Runtime)).Elem()
)

type actorInterface interface {
	Exports() []interface{}
}

func init() {
	actorInfos[builtin.AccountActorCodeID] = ActorInfo{
		Name:      "AccountActor",
		Methods:   []MethodInfo{},
		methodMap: map[uint64]int{},
	}

	// can add other actorInfo
	actorInfos[builtin.InitActorCodeID] = parseActor(in.Actor{}, builtin.MethodsInit)
	actorInfos[builtin.StorageMinerActorCodeID] = parseActor(miner.Actor{}, builtin.MethodsMiner)
	actorInfos[builtin.MultisigActorCodeID] = parseActor(multisig.Actor{}, builtin.MethodsMultisig)
	actorInfos[builtin.StorageMarketActorCodeID] = parseActor(market.Actor{}, builtin.MethodsMarket)
	actorInfos[builtin.StoragePowerActorCodeID] = parseActor(power.Actor{}, builtin.MethodsPower)
	actorInfos[builtin.PaymentChannelActorCodeID] = parseActor(paych.Actor{}, builtin.MethodsPaych)
}

// LookupByAddress find actor with given code
func LookupByAddress(addr address.Address) (ActorInfo, bool) {
	if code, ok := addressToCode[addr]; ok {
		return Lookup(code)
	}

	return ActorInfo{}, false
}

// Lookup find actor with given code
func Lookup(code cid.Cid) (ActorInfo, bool) {
	act, ok := actorInfos[code]
	return act, ok
}

// ActorInfo is a collection of actor infos
type ActorInfo struct {
	Name      string
	Methods   []MethodInfo
	methodMap map[uint64]int
}

// LookupMethod find method info with given method number
func (a *ActorInfo) LookupMethod(num uint64) (MethodInfo, bool) {
	if idx, ok := a.methodMap[num]; ok {
		return a.Methods[idx], true
	}

	return MethodInfo{}, false
}

// MethodInfo method info
type MethodInfo struct {
	Name      string
	Num       uint64
	paramType reflect.Type
}

// NewParam returns a zero value of the param type
func (m *MethodInfo) NewParam() interface{} {
	if m.paramType == TypeNull {
		return nil
	}

	return reflect.New(m.paramType).Interface()
}

func parseActor(act actorInterface, methods interface{}) ActorInfo {
	methodInfos := []MethodInfo{}
	methodMap := map[uint64]int{}

	methodFuncs := act.Exports()

	mv := reflect.ValueOf(methods)
	mt := mv.Type()
	fnum := mt.NumField()

	for i := 0; i < fnum; i++ {
		mnum := mv.Field(i).Uint()
		methodMap[mnum] = len(methodInfos)

		methodInfos = append(methodInfos, MethodInfo{
			Name:      mt.Field(i).Name,
			Num:       mnum,
			paramType: getMethodParam(methodFuncs[mnum]),
		})
	}

	return ActorInfo{
		Name:      reflect.TypeOf(act).Name(),
		Methods:   methodInfos,
		methodMap: methodMap,
	}
}

func getMethodParam(meth interface{}) reflect.Type {
	if meth == nil {
		return TypeNull
	}

	metht := reflect.TypeOf(meth)
	if metht.Kind() != reflect.Func || metht.NumIn() != 3 {
		return TypeNull
	}

	if metht.In(0) != TypeActorPtr || metht.In(1) != TypeVMCtx {
		return TypeNull
	}

	pt := metht.In(2)
	for pt.Kind() == reflect.Ptr {
		pt = pt.Elem()
	}

	if pt.Kind() != reflect.Struct {
		return TypeNull
	}

	return pt
}
