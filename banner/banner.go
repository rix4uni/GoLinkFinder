package banner

import (
	"fmt"
)

// prints the version message
const version = "v0.0.1"

func PrintVersion() {
	fmt.Printf("Current GoLinkFinder version %s\n", version)
}

// Prints the Colorful banner
func PrintBanner() {
	banner := `
   ______        __     _         __    ______ _             __           
  / ____/____   / /    (_)____   / /__ / ____/(_)____   ____/ /___   _____
 / / __ / __ \ / /    / // __ \ / //_// /_   / // __ \ / __  // _ \ / ___/
/ /_/ // /_/ // /___ / // / / // ,<  / __/  / // / / // /_/ //  __// /    
\____/ \____//_____//_//_/ /_//_/|_|/_/    /_//_/ /_/ \__,_/ \___//_/
`
	fmt.Printf("%s\n%70s\n\n", banner, "Current GoLinkFinder version "+version)
}
