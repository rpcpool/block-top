package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
)

const defaultRPC = "http://sfo3.rpcpool.wg"

var (
	flagAddr    = flag.String("addr", defaultRPC, "RPC address")
	flagLeaders ArrayFlag
)

var (
	client          *rpc.Client
	commitment      rpc.CommitmentType = rpc.CommitmentConfirmed
	watched_leaders solana.PublicKeySlice
)

func main() {
	flag.Var(&flagLeaders, "leader", "Leader to watch, if none provided, watch all")
	flag.Parse()

	client = rpc.New(*flagAddr)

	if len(flagLeaders) > 0 {
		watched_leaders = make(solana.PublicKeySlice, len(flagLeaders))
		for i := range flagLeaders {
			our, err := solana.PublicKeyFromBase58(flagLeaders[i])
			if err != nil {
				log.Print("couldn't load public key, filtering disabled:", flagLeaders[i], "err=", err)
				continue
			}
			watched_leaders.UniqueAppend(our)
		}

	}

	// a bit of a hack but ok we're not running this from genesis
	var current_slot uint64 = 0

	for {
		epoch, start_slot, end_slot, leader_slots, err := GetEpochBound(context.Background())

		// if this is the first run, call current slot
		if current_slot == 0 {
			current_slot = epoch.AbsoluteSlot
		}

		if err != nil {
			log.Fatal("can't fetch leader schedule, bailing", err)
		}
		log.Println("loaded epoch=", epoch.Epoch, "start_slot=", start_slot, "end_slot=", end_slot)

		for {
			// polling, maybe better as ws
			slot, err := client.GetSlot(context.TODO(), commitment)
			if err != nil {
				log.Println("couldnt load slot", err)
				time.Sleep(250 * time.Millisecond)
				continue
			}

			if current_slot < slot {
				if len(watched_leaders) == 0 || watched_leaders.Has(leader_slots[current_slot]) {
					block, err := client.GetBlock(context.TODO(), current_slot)
					if err != nil {
						var rpcErr *jsonrpc.RPCError
						if errors.As(err, &rpcErr) && (rpcErr.Code == -32007 /* SLOT_SKIPPED */ || rpcErr.Code == -32004 /* BLOCK_NOT_AVAILABLE */) {
							log.Printf("leader=%s slot=%d skipped=true\n", leader_slots[current_slot].String(), current_slot)
						} else {
							log.Println("error fetching block:", err)
							time.Sleep(250 * time.Millisecond)
							continue
						}
					} else {
						bt := time.Unix(*block.BlockTime, 0)
						fmt.Printf("ts=%s leader=%s slot=%d numTx=%d\n", bt, leader_slots[current_slot].String(), current_slot, len(block.Transactions))
					}
				}
				current_slot++

				// we've reached the end of the epoch
				if current_slot == (end_slot + 1) {
					break
				}
			} else {
				log.Println("waiting for new slot")
			}

			time.Sleep(1 * time.Second)
		}
	}
}

func GetEpochBound(ctx context.Context) (epoch rpc.GetEpochInfoResult, start_slot uint64, end_slot uint64, leader_slots map[uint64]solana.PublicKey, err error) {
	epr, err := client.GetEpochInfo(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return
	}
	if epr == nil {
		err = errors.New("no epoch data returend")
		return
	}

	epoch = *epr
	start_slot = epoch.AbsoluteSlot - epoch.SlotIndex
	end_slot = start_slot + epoch.SlotsInEpoch

	resp, err := client.GetLeaderScheduleWithOpts(ctx, &rpc.GetLeaderScheduleOpts{
		Epoch: &epoch.AbsoluteSlot,
	})

	if err != nil {
		return
	}

	if resp == nil {
		err = errors.New("can't load leader slots")
		return
	}

	leader_slots = make(map[uint64]solana.PublicKey, epoch.SlotsInEpoch)
	for pk, slots := range *resp {
		for slot := range slots {
			absolute_slot := start_slot + slots[slot]
			leader_slots[absolute_slot] = pk
		}
	}

	return
}
