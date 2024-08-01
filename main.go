package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type labels struct {
	YandexCpiFlantComNodeRole string `json:"yandex.cpi.flant.com/node-role,omitempty"`
}
type staticRoutes struct {
	DestinationPrefix string `json:"destinationPrefix"`
	NextHopAddress    string `json:"nextHopAddress"`
	Labels            labels `json:"labels,omitempty"`
}
type routeTable struct {
	StaticRoutes []staticRoutes `json:"staticRoutes"`
	ID           string         `json:"id"`
	FolderID     string         `json:"folderId"`
	CreatedAt    time.Time      `json:"createdAt"`
	Name         string         `json:"name"`
	NetworkID    string         `json:"networkId"`
}

const yc_url = "https://vpc.api.cloud.yandex.net"

func main() {
	vpn_network := os.Getenv("VPN_NETWORK")
	if vpn_network == "" {
		log.Fatal("VPN_NETWORK not set")
	}
	vpn_node_address := os.Getenv("VPN_NODE_ADDRESS")
	if vpn_node_address == "" {
		log.Fatal("VPN_NODE_ADDRESS not set")
	}
	table_id := os.Getenv("ROUTE_TABLE_ID")
	if table_id == "" {
		log.Fatal("ROUTE_TABLE_ID not set")
	}
	token := os.Getenv("API_TOKEN")
	if token == "" {
		log.Fatal("API_TOKEN not set")
	}

	rtable := routeTable{}
	req, _ := http.NewRequest(http.MethodGet, yc_url+"/vpc/v1/routeTables/"+table_id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Route table get failed: status %s\n", resp.Status)
	}

	json.NewDecoder(resp.Body).Decode(&rtable)

	for index, net := range rtable.StaticRoutes {
		if net.DestinationPrefix == vpn_network {
			if net.NextHopAddress == vpn_node_address {
				log.Println("Same route, no action needed")
				os.Exit(0)
			} else {
				log.Printf("VPN node IP neet to change from %s to %s\n", net.NextHopAddress, vpn_node_address)
				rtable.StaticRoutes[index].NextHopAddress = vpn_node_address
				vpn_node_address = ""
				break
			}
		}
	}
	if vpn_node_address != "" {
		rtable.StaticRoutes = append(rtable.StaticRoutes, staticRoutes{DestinationPrefix: vpn_network, NextHopAddress: vpn_node_address})
	}

	postJSON, err := json.Marshal(rtable)
	request, err := http.NewRequest("PATCH", yc_url+"/vpc/v1/routeTables/"+table_id, bytes.NewBuffer(postJSON))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	log.Printf("Status: %s\n", resp.Status)
	if resp.StatusCode != 200 {
		log.Fatal("Route table update failed:", resp.Body)
	}
}
