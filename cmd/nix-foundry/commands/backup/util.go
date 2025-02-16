package backup

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func confirmAction() bool {
	fmt.Print("❗ Are you sure you want to proceed? (y/N): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.EqualFold(scanner.Text(), "y")
}
