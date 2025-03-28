/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package dotnet

import (
	"crypto/rsa"
	"fmt"
	"math"
	"math/big"
)

func NewPublicKey(pubkey *rsa.PublicKey) *PublicKey {
	bigE := big.NewInt(int64(pubkey.E))
	return &PublicKey{
		Exponent: bigE.Bytes(),
		Modulus:  pubkey.N.Bytes(),
	}
}

func (pk *PublicKey) ToRsaPublicKey() (*rsa.PublicKey, error) {
	mod := new(big.Int).SetBytes(pk.Modulus)
	exp := new(big.Int).SetBytes(pk.Exponent)

	var e int64
	if !exp.IsInt64() {
		return nil, fmt.Errorf("%s can be represented as an int64", exp)
	} else {
		e = exp.Int64()
		if e > math.MaxInt {
			return nil, fmt.Errorf("%d integer overflow", e)
		}
		if e <= 0 {
			return nil, fmt.Errorf("%d must be positive number", e)
		}
	}

	pubkey := rsa.PublicKey{
		N: mod,
		E: int(e),
	}
	return &pubkey, nil
}

func NewPrivateKey(key *rsa.PrivateKey) *PrivateKey {
	key.Precompute()

	pub := NewPublicKey(&key.PublicKey)
	return &PrivateKey{
		PublicKey: *pub,
		D:         key.D.Bytes(),
		P:         key.Primes[0].Bytes(),
		Q:         key.Primes[1].Bytes(),
		DP:        key.Precomputed.Dp.Bytes(),
		DQ:        key.Precomputed.Dq.Bytes(),
		InverseQ:  key.Precomputed.Qinv.Bytes(),
	}
}

func (key *PrivateKey) ToRsaPrivateKey() (*rsa.PrivateKey, error) {
	pub, err := key.PublicKey.ToRsaPublicKey()
	if err != nil {
		return nil, fmt.Errorf("can't convert to rsa.PublicKey: %v", err)
	}
	priv := &rsa.PrivateKey{
		PublicKey: *pub,

		D: new(big.Int).SetBytes(key.D),
		Primes: []*big.Int{
			new(big.Int).SetBytes(key.P),
			new(big.Int).SetBytes(key.Q),
		},
	}
	priv.Precompute()
	return priv, nil
}
