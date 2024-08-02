package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/tlb"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

const (
	Timeout time.Duration = 5 * time.Second
)

type Config struct {
	LiteServerAddr         string
	LiteServerKey          string
	GlobalNetworkConfigUrl string
}

func LoadConfig() (*Config) {
	var config Config
	flag.StringVar(&config.LiteServerAddr, "address", "135.181.140.212:13206", "Lite server address:port")
	flag.StringVar(&config.LiteServerKey, "key", "K0t3+IWLOXHYMvMcrGZDPs+pn58a17LFbnXoQkKc2xw=", "Lite server public key")
	flag.StringVar(&config.GlobalNetworkConfigUrl, "url", "https://ton.org/global.config.json", "Global config url")
	flag.Parse()
	return &config
}

func main() {
	cfg := LoadConfig()
	log.Printf("Connecting to client %s with key: %v", cfg.LiteServerAddr, cfg.LiteServerKey)

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	client := liteclient.NewConnectionPool()

	// Connect to random working lite server
	// err := client.AddConnectionsFromConfigUrl(cfg.GlobalNetworkConfigUrl)

	err := client.AddConnection(ctx, cfg.LiteServerAddr, cfg.LiteServerKey)
	if err != nil {
		log.Fatalln("Local client connection error: ", err)
		return
	}
	defer cancel()

	api := ton.NewAPIClient(client)

	block, err := api.CurrentMasterchainInfo(context.Background())
	if err != nil {
		log.Fatalln("Get Current Masterchain Info error: ", err)
	}
	fmt.Println("Masterchain block: ", block.SeqNo)

	config, _ := api.GetBlockchainConfig(ctx, block)
	config15Cell := config.Get(15)
	if err != nil {
		log.Fatalf("Failed to get network Config15: %v", err)
	}

	// Convert tonutils-go *cell.Cell to BOC format
	config15BOC := config15Cell.ToBOC()
	if err != nil {
		log.Fatalf("Failed to serialize Config15 cell to BOC: %v", err)
	}

	// Deserialize BOC data to tonkeeper/tongo *boc.Cell format
	cells, err := boc.DeserializeBoc(config15BOC)
	if err != nil {
		log.Fatalf("Failed to deserialize BOC: %v", err)
	}

	if len(cells) == 0 {
		log.Fatalf("No cells found in BOC")
	}

	//Processing parameter parsing with tonkeeper/tongo
	var config15 tlb.ConfigParam15
	err = tlb.Unmarshal(cells[0], &config15)
	if err != nil {
		log.Fatalf("Failed to unmarshal network config Param15: %v", err)
	}

	fmt.Printf("Config15: %+v\n", config15)
}
