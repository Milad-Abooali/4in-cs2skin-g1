package provablyfair

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
)

var (
	ServerSeed = "SERVER_SECRET"
	ClientSeed = "PLAYER_SEED"
)

// ProvablyFairRand generates a deterministic "random" number based on server seed, client seed and nonce.
// max: the upper limit of the random number (exclusive)
func ProvablyFairRand(serverSeed, clientSeed string, nonce, max int) int {
	input := fmt.Sprintf("%s:%s:%d", serverSeed, clientSeed, nonce)
	hash := sha256.Sum256([]byte(input))
	num := binary.BigEndian.Uint32(hash[:4])
	return int(num) % max
}

func PickItem(caseData map[string]interface{}, serverSeed, clientSeed string, nonce int) map[string]interface{} {

	// Generate provably fair random number 0..1,000,000
	r := ProvablyFairRand(serverSeed, clientSeed, nonce, 1_000_001)

	itemsRaw, ok := caseData["items"].(map[int]map[string]interface{})
	if !ok {
		log.Println("PickItem > no items")
		return nil
	}

	// Select matching item
	for _, item := range itemsRaw {
		minR, _ := item["min_rand"].(float64)
		maxR, _ := item["max_rand"].(float64)

		if float64(r) >= minR && float64(r) <= maxR {
			return item
		}
	}

	return nil
}

func GenerateServerSeed() (string, string) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	seed := hex.EncodeToString(bytes)
	hash := sha256.Sum256([]byte(seed))
	return seed, hex.EncodeToString(hash[:]) // (serverSeed, serverSeedHash)
}
