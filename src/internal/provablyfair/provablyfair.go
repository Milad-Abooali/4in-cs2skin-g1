package provablyfair

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
)

// FairRand generates a deterministic "random" number based on server seed, client seed and nonce.
// max: the upper limit of the random number (exclusive)
func FairRand(serverSeed, clientSeed string, nonce, max int) int {
	input := fmt.Sprintf("%s:%s:%d", serverSeed, clientSeed, nonce)
	hash := sha256.Sum256([]byte(input))
	num := binary.BigEndian.Uint32(hash[:4])
	return int(num) % max
}

func PickItem(HE float64, caseData map[string]interface{}, serverSeed, clientSeed string, nonce int) map[string]interface{} {

	selectedItem := selectItem(caseData, serverSeed, clientSeed, nonce)
	priceStr, _ := selectedItem["price"].(string)
	price, _ := strconv.ParseFloat(priceStr, 64)

	casePricestr, _ := caseData["price"].(string)
	casePrice, _ := strconv.ParseFloat(casePricestr, 64)

	if HE > 8 || HE == 0 {
		return selectedItem
	}

	if HE > 0 && HE < 8 {
		if price > casePrice {
			for {
				selectedItem = selectItem(caseData, serverSeed, clientSeed, nonce)
				priceStr, _ = selectedItem["price"].(string)
				price, _ = strconv.ParseFloat(priceStr, 64)
				log.Println("price_item", price, casePrice)
				if price <= casePrice {
					break
				}
				nonce++
			}
			return selectedItem
		}
		return selectedItem
	}

	if HE < 0 {
		if price > casePrice {
			for {
				selectedItem = selectItem(caseData, serverSeed, clientSeed, nonce)
				priceStr, _ = selectedItem["price"].(string)
				price, _ = strconv.ParseFloat(priceStr, 64)
				log.Println("price_item", price, casePrice)
				if price <= casePrice {
					break
				}
				nonce++
			}
			return selectedItem
		}
		return selectedItem
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

func selectItem(caseData map[string]interface{}, serverSeed, clientSeed string, nonce int) map[string]interface{} {
	// Generate provably fair random number 0..1,000,000
	r := FairRand(serverSeed, clientSeed, nonce, 1_000_001)
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
