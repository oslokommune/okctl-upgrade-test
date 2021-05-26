package main

import "fmt"
import "github.com/oslokommune/okctl-upgrade/0.0.63/pkg/argocd"

func main() {
	fmt.Printf("Upgrading resources...\n\n")

	argocd.Run()

	fmt.Printf("\nAll resources upgraded!\n")
}
