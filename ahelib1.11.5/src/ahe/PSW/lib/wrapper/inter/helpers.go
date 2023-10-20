//go:binary-only-package

package inter
import "C"
import (
	dto_zkproof "ahe/PSW/lib/zkproof/dto"
	dto_zkrprc "ahe/PSW/lib/zkrprc/dto"
	"crypto/rand"
	"fmt"
	"math/big"
	"unsafe"
)