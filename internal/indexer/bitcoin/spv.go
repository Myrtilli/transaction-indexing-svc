package bitcoin

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
)

func ValidateHeader(header, prev *BlockHeader) error {
	if header.PreviousHash != prev.BlockHash {
		return errors.New("invalid previous hash")
	}
	if !CheckProofOfWork(header) {
		return errors.New("invalid proof of work")
	}
	if header.Timestamp < prev.Timestamp {
		return errors.New("invalid timestamp")
	}
	return nil
}

func decodeHex(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}

func calculateTarget(difficulty float64) *big.Int {
	maxTarget := new(big.Int).Lsh(big.NewInt(1), 256)
	maxTarget.Sub(maxTarget, big.NewInt(1))

	if difficulty <= 0 {
		return maxTarget
	}

	diffFloat := big.NewFloat(difficulty)
	maxTargetFloat := new(big.Float).SetInt(maxTarget)

	resFloat := new(big.Float).Quo(maxTargetFloat, diffFloat)

	result := new(big.Int)
	resFloat.Int(result)
	return result
}

func CheckProofOfWork(header *BlockHeader) bool {
	hashBytes := decodeHex(header.BlockHash)
	hashInt := new(big.Int).SetBytes(hashBytes)

	target := calculateTarget(header.Difficulty)

	return hashInt.Cmp(target) <= 0
}

func VerifyMerkleProof(txid string, proof [][]byte, root string) bool {
	hash := sha256.Sum256([]byte(txid))
	for _, p := range proof {
		h := sha256.Sum256(append(hash[:], p...))
		hash = h
	}
	return hex.EncodeToString(hash[:]) == root
}
