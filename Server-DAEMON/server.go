package main

import (
	"io"
	"net/http"
	"os"
	"errors"
	"fmt"
	"strconv"
	"crypto/rand"
	"math/big"
	"bytes"
	//"github.com/VividCortex/godaemon"
)

var smallPrimes = []uint8 {
  	3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53,
}

var smallPrimesProduct = new(big.Int).SetUint64(16294579238595022365)

func VPExit(w http.ResponseWriter, r *http.Request) {
	os.Exit(0)
}

func VanityPrime (w http.ResponseWriter, r *http.Request) {
	message := r.URL.RawQuery[5:]
	hexStrings := SplitSubN(message, 2)
	messageArray := make([]byte, len(hexStrings))
	
  	for i := 0; i < len(messageArray); i++ {
  		value,_ := strconv.ParseUint(hexStrings[i], 16, 8)
  		if(value < 16) {
  			value = (value * 16) + RandomFourBitNumber()
  		}
  		messageArray[i] = byte (value)
  	}

	vanity,_ := Prime(rand.Reader, messageArray, 1024)
	fmt.Fprintf(w, "%0x", vanity)
}

func SplitSubN(s string, n int) []string {
    sub := ""
    subs := []string{}

    runes := bytes.Runes([]byte(s))
    l := len(runes)
    for i, r := range runes {
        sub = sub + string(r)
        if (i + 1) % n == 0 {
            subs = append(subs, sub)
            sub = ""
        } else if (i + 1) == l {
            subs = append(subs, sub)
        }
    }

    return subs
}

func RandomFourBitNumber() uint64 {
	bytes := []byte{0}
	if _, err := rand.Reader.Read(bytes); err != nil {
		panic(err)
	}
	number := bytes[0]
	return uint64 (int (number) % 16)
}

func Prime(rand io.Reader, message []byte, bits int) (p *big.Int, err error) {
  	if bits < 2 {
  		err = errors.New("crypto/rand: prime size must be at least 2-bit")
  		return
  	}
  
  	b := uint(bits % 8)
  	if b == 0 {
  		b = 8
  	}
  
  	bytes := make([]byte, (bits+7)/8)
  	p = new(big.Int)
  
  	bigMod := new(big.Int)
  
  	for {
  		_, err = io.ReadFull(rand, bytes)
  		if err != nil {
  			return nil, err
  		}
  
  		// Clear bits in the first byte to make sure the candidate has a size <= bits.
  		bytes[0] &= uint8(int(1<<b) - 1)
  		// Don't let the value be too small, i.e, set the most significant two bits.
  		// Setting the top two bits, rather than just the top bit,
  		// means that when two of these values are multiplied together,
  		// the result isn't ever one bit short.
  		if b >= 2 {
  			bytes[0] |= 3 << (b - 2)
  		} else {
  			// Here b==1, because b cannot be zero.
  			bytes[0] |= 1
  			if len(bytes) > 1 {
  				bytes[1] |= 0x80
  			}
  		}
  		// Make the value odd since an even number this large certainly isn't prime.
  		bytes[len(bytes)-1] |= 1

  		for i := 0; i < len(message); i++ {
  			bytes[i] = message[i]
  		}
  
  		p.SetBytes(bytes)
  
  		// Calculate the value mod the product of smallPrimes. If it's
  		// a multiple of any of these primes we add two until it isn't.
  		// The probability of overflowing is minimal and can be ignored
  		// because we still perform Miller-Rabin tests on the result.
  		bigMod.Mod(p, smallPrimesProduct)
  		mod := bigMod.Uint64()
  
  	NextDelta:
  		for delta := uint64(0); delta < 1<<20; delta += 2 {
  			m := mod + delta
  			for _, prime := range smallPrimes {
  				if m%uint64(prime) == 0 && (bits > 6 || m != uint64(prime)) {
  					continue NextDelta
  				}
  			}
  
  			if delta > 0 {
  				bigMod.SetUint64(delta)
  				p.Add(p, bigMod)
  			}
  			break
  		}
  
  		// There is a tiny possibility that, by adding delta, we caused
  		// the number to be one bit too long. Thus we check BitLen
  		// here.
  		if p.ProbablyPrime(20) && p.BitLen() == bits {
  			return
  		}
  	}
}

func main() {
	//godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	http.HandleFunc("/.well-known/vpexit", VPExit)
	http.HandleFunc("/.well-known/vanityprime", VanityPrime)
	http.ListenAndServe(":8081", nil)
}