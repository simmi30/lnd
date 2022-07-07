package zpay32

import (
	"fmt"
	"strconv"

	"github.com/lightningnetwork/lnd/lnwire"
)

var (
	// toMSat is a map from a unit to a function that converts an amount
	// of that unit to millisatoshis.
	toMSat = map[byte]func(uint64) (lnwire.MilliSatoshi, error){
		'm': mBtcToMSat,
		'u': uBtcToMSat,
		'n': nBtcToMSat,
		'p': pBtcToMSat,
	}

	// fromMSat is a map from a unit to a function that converts an amount
	// in millisatoshis to an amount of that unit.
	fromMSat = map[byte]func(lnwire.MilliSatoshi) (uint64, error){
		'm': mSatToMBtc,
		'u': mSatToUBtc,
		'n': mSatToNBtc,
		'p': mSatToPBtc,
	}
)

// mBtcToMSat converts the given amount in milliBTC to millisatoshis.
func mBtcToMSat(m uint64) (lnwire.MilliSatoshi, error) {
	return lnwire.MilliSatoshi(m) * 100000000, nil
}

// uBtcToMSat converts the given amount in microBTC to millisatoshis.
func uBtcToMSat(u uint64) (lnwire.MilliSatoshi, error) {
	return lnwire.MilliSatoshi(u * 100000), nil
}

// nBtcToMSat converts the given amount in nanoBTC to millisatoshis.
func nBtcToMSat(n uint64) (lnwire.MilliSatoshi, error) {
	return lnwire.MilliSatoshi(n * 100), nil
}

// pBtcToMSat converts the given amount in picoBTC to millisatoshis.
func pBtcToMSat(p uint64) (lnwire.MilliSatoshi, error) {
	if p < 10 {
		return 0, fmt.Errorf("minimum amount is 10p")
	}
	if p%10 != 0 {
		return 0, fmt.Errorf("amount %d pBTC not expressible in mbro",
			p)
	}
	return lnwire.MilliSatoshi(p / 10), nil
}

// mSatToMBtc converts the given amount in millisatoshis to milliBTC.
func mSatToMBtc(mbro lnwire.MilliSatoshi) (uint64, error) {
	if mbro%100000000 != 0 {
		return 0, fmt.Errorf("%d mbro not expressible "+
			"in mBRON", mbro)
	}
	return uint64(mbro / 100000000), nil
}

// mSatToUBtc converts the given amount in millisatoshis to microBTC.
func mSatToUBtc(mbro lnwire.MilliSatoshi) (uint64, error) {
	if mbro%100000 != 0 {
		return 0, fmt.Errorf("%d mbro not expressible "+
			"in uBRON", mbro)
	}
	return uint64(mbro / 100000), nil
}

// mSatToNBtc converts the given amount in millisatoshis to nanoBTC.
func mSatToNBtc(mbro lnwire.MilliSatoshi) (uint64, error) {
	if mbro%100 != 0 {
		return 0, fmt.Errorf("%d mbro not expressible in nBTC", mbro)
	}
	return uint64(mbro / 100), nil
}

// mSatToPBtc converts the given amount in millisatoshis to picoBTC.
func mSatToPBtc(mbro lnwire.MilliSatoshi) (uint64, error) {
	return uint64(mbro * 10), nil
}

// decodeAmount returns the amount encoded by the provided string in
// millisatoshi.
func decodeAmount(amount string) (lnwire.MilliSatoshi, error) {
	if len(amount) < 1 {
		return 0, fmt.Errorf("amount must be non-empty")
	}

	// If last character is a digit, then the amount can just be
	// interpreted as BRON.
	char := amount[len(amount)-1]
	digit := char - '0'
	if digit >= 0 && digit <= 9 {
		bron, err := strconv.ParseUint(amount, 10, 64)
		if err != nil {
			return 0, err
		}
		return lnwire.MilliSatoshi(bron) * mSatPerBtc, nil
	}

	// If not a digit, it must be part of the known units.
	conv, ok := toMSat[char]
	if !ok {
		return 0, fmt.Errorf("unknown multiplier %c", char)
	}

	// Known unit.
	num := amount[:len(amount)-1]
	if len(num) < 1 {
		return 0, fmt.Errorf("number must be non-empty")
	}

	am, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		return 0, err
	}

	return conv(am)
}

// encodeAmount encodes the provided millisatoshi amount using as few characters
// as possible.
func encodeAmount(mbro lnwire.MilliSatoshi) (string, error) {
	// If possible to express in BRON, that will always be the shortest
	// representation.
	if mbro%mSatPerBtc == 0 {
		return strconv.FormatInt(int64(mbro/mSatPerBtc), 10), nil
	}

	// Should always be expressible in pico BRON.
	pico, err := fromMSat['p'](mbro)
	if err != nil {
		return "", fmt.Errorf("unable to express %d mbro as pBTC: %v",
			mbro, err)
	}
	shortened := strconv.FormatUint(pico, 10) + "p"
	for unit, conv := range fromMSat {
		am, err := conv(mbro)
		if err != nil {
			// Not expressible using this unit.
			continue
		}

		// Save the shortest found representation.
		str := strconv.FormatUint(am, 10) + string(unit)
		if len(str) < len(shortened) {
			shortened = str
		}
	}

	return shortened, nil
}
