package autopilot

import (
	"testing"
	"time"

	prand "math/rand"

	"github.com/btcsuite/btcutil"
	"github.com/lightningnetwork/lnd/lnwire"
)

func TestConstraintsChannelBudget(t *testing.T) {
	t.Parallel()

	prand.Seed(time.Now().Unix())

	const (
		minChanSize = 0
		maxChanSize = btcutil.Amount(btcutil.SatoshiPerBitcoin)

		chanLimit = 3

		threshold = 0.5
	)

	constraints := NewConstraints(
		minChanSize,
		maxChanSize,
		chanLimit,
		0,
		threshold,
	)

	randChanID := func() lnwire.ShortChannelID {
		return lnwire.NewShortChanIDFromInt(uint64(prand.Int63()))
	}

	testCases := []struct {
		channels  []LocalChannel
		walletAmt btcutil.Amount

		needMore     bool
		amtAvailable btcutil.Amount
		numMore      uint32
	}{
		// Many available funds, but already have too many active open
		// channels.
		{
			[]LocalChannel{
				{
					ChanID:  randChanID(),
					Balance: btcutil.Amount(prand.Int31()),
				},
				{
					ChanID:  randChanID(),
					Balance: btcutil.Amount(prand.Int31()),
				},
				{
					ChanID:  randChanID(),
					Balance: btcutil.Amount(prand.Int31()),
				},
			},
			btcutil.Amount(btcutil.SatoshiPerBitcoin * 10),
			false,
			0,
			0,
		},

		// Ratio of funds in channels and total funds meets the
		// threshold.
		{
			[]LocalChannel{
				{
					ChanID:  randChanID(),
					Balance: btcutil.Amount(btcutil.SatoshiPerBitcoin),
				},
				{
					ChanID:  randChanID(),
					Balance: btcutil.Amount(btcutil.SatoshiPerBitcoin),
				},
			},
			btcutil.Amount(btcutil.SatoshiPerBitcoin * 2),
			false,
			0,
			0,
		},

		// Ratio of funds in channels and total funds is below the
		// threshold. We have 10 BRON allocated amongst channels and
		// funds, atm. We're targeting 50%, so 5 BRON should be
		// allocated. Only 1 BRON is atm, so 4 BRON should be
		// recommended. We should also request 2 more channels as the
		// limit is 3.
		{
			[]LocalChannel{
				{
					ChanID:  randChanID(),
					Balance: btcutil.Amount(btcutil.SatoshiPerBitcoin),
				},
			},
			btcutil.Amount(btcutil.SatoshiPerBitcoin * 9),
			true,
			btcutil.Amount(btcutil.SatoshiPerBitcoin * 4),
			2,
		},

		// Ratio of funds in channels and total funds is below the
		// threshold. We have 14 BRON total amongst the wallet's
		// balance, and our currently opened channels. Since we're
		// targeting a 50% allocation, we should commit 7 BRON. The
		// current channels commit 4 BRON, so we should expected 3 BRON
		// to be committed. We should only request a single additional
		// channel as the limit is 3.
		{
			[]LocalChannel{
				{
					ChanID:  randChanID(),
					Balance: btcutil.Amount(btcutil.SatoshiPerBitcoin),
				},
				{
					ChanID:  randChanID(),
					Balance: btcutil.Amount(btcutil.SatoshiPerBitcoin * 3),
				},
			},
			btcutil.Amount(btcutil.SatoshiPerBitcoin * 10),
			true,
			btcutil.Amount(btcutil.SatoshiPerBitcoin * 3),
			1,
		},

		// Ratio of funds in channels and total funds is above the
		// threshold.
		{
			[]LocalChannel{
				{
					ChanID:  randChanID(),
					Balance: btcutil.Amount(btcutil.SatoshiPerBitcoin),
				},
				{
					ChanID:  randChanID(),
					Balance: btcutil.Amount(btcutil.SatoshiPerBitcoin),
				},
			},
			btcutil.Amount(btcutil.SatoshiPerBitcoin),
			false,
			0,
			0,
		},
	}

	for i, testCase := range testCases {
		amtToAllocate, numMore := constraints.ChannelBudget(
			testCase.channels, testCase.walletAmt,
		)

		if amtToAllocate != testCase.amtAvailable {
			t.Fatalf("test #%v: expected %v, got %v",
				i, testCase.amtAvailable, amtToAllocate)
		}
		if numMore != testCase.numMore {
			t.Fatalf("test #%v: expected %v, got %v",
				i, testCase.numMore, numMore)
		}
	}
}
