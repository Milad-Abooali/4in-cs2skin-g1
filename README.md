// Select Item For slots
log.Println(i)
item := provablyfair.PickItem(caseData, provablyfair.ServerSeed, provablyfair.ClientSeed, i)
if item != nil {
log.Println(i, "Selected item:", item["market_hash_name"], "Price:", item["price"])
}