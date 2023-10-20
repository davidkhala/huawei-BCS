//go:binary-only-package

package wrapper

import (
	dto2 "ahe/PSW/lib/zkproof/dto"
	
	"ahe/PSW/lib/zkrprc/crypto"
	dto1 "ahe/PSW/lib/zkrprc/dto"
	"ahe/PSW/lib/zkrprc/genparam"

	
	"ahe/PSW/lib/wrapper/inter"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"math/big"
	"os"
)