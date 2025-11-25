package banner

import (
	"fmt"
)

// prints the version message
const version = "v0.0.8"

func PrintVersion() {
	fmt.Printf("Current pvreplace version %s\n", version)
}

// Prints the Colorful banner
func PrintBanner() {
	banner := `
                                     __                 
    ____  _   __ _____ ___   ____   / /____ _ _____ ___ 
   / __ \| | / // ___// _ \ / __ \ / // __  // ___// _ \
  / /_/ /| |/ // /   /  __// /_/ // // /_/ // /__ /  __/
 / .___/ |___//_/    \___// .___//_/ \__,_/ \___/ \___/ 
/_/                      /_/
`
	fmt.Printf("%s\n%55s\n\n", banner, "Current pvreplace version "+version)
}
