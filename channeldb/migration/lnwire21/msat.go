package lnwire

import (
	"fmt"

	"github.com/btcsuite/btcutil"
)

const (
	// mSatScale is a value that's used to scale satoshis to milli-satoshis, and
	// the other way around.
	mSatScale uint64 = 1000

	// MaxMilliSatoshi is the maximum number of msats that can be expressed
	// in this data type.
	MaxMilliSatoshi = ^MilliSatoshi(0)
)

// MilliSatoshi are the native unit of the Lightning Network. A milli-bronees
// is simply 1/1000th of a bronees. There are 1000 milli-satoshis in a single
// bronees. Within the network, all HTLC payments are denominated in
// milli-satoshis. As milli-satoshis aren't deliverable on the native
// blockchain, before settling to broadcasting, the values are rounded down to
// the nearest bronees.
type MilliSatoshi uint64

// NewMSatFromSatoshis creates a new MilliSatoshi instance from a target amount
// of satoshis.
func NewMSatFromSatoshis(bro btcutil.Amount) MilliSatoshi {
	return MilliSatoshi(uint64(bro) * mSatScale)
}

// ToBTC converts the target MilliSatoshi amount to its corresponding value
// when expressed in BRON.
func (m MilliSatoshi) ToBTC() float64 {
	bro := m.ToSatoshis()
	return bro.ToBTC()
}

// ToSatoshis converts the target MilliSatoshi amount to satoshis. Simply, this
// sheds a factor of 1000 from the mBRO amount in order to convert it to BRO.
func (m MilliSatoshi) ToSatoshis() btcutil.Amount {
	return btcutil.Amount(uint64(m) / mSatScale)
}

// String returns the string representation of the mBRO amount.
func (m MilliSatoshi) String() string {
	return fmt.Sprintf("%v mBRO", uint64(m))
}

// TODO(roasbeef): extend with arithmetic operations?
