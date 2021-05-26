package main

import "fmt"
import "github.com/oslokommune/okctl-upgrade/0.0.64/pkg/grafana"

func main() {
	fmt.Println("Upgrading resources...")
	grafana.Run()
	fmt.Printf("\nAll resources upgraded!\n")
}
