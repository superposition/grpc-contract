package impl

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"os"

	"github.com/getamis/grpc-contract/internal/util"
)

type Contract struct {
	Package string
	Name    string
	Imports []string
	Methods []*Method
}

func (c *Contract) IsServerInterface(name string) bool {
	if name == c.Name+"Server" {
		return true
	}
	return false
}

var ContractTemplate string = `package {{ .Package }};
{{ range .Imports }}
import "{{ . }}"{{ end }}

type server struct {
	contract *{{ .Name }}
}

func NewServer(address common.Address, backend bind.ContractBackend) {{ .Name }}Server {
	contract, _ := New{{ .Name }}(address, backend)
	return &server{
		contract: contract,
	}
}

{{ range .Methods }}
{{ . }}
{{ end }}
// TransactOpts converts to bind.TransactOpts
func (m *TransactOpts) TransactOpts() *bind.TransactOpts {
	privateKey, err := crypto.ToECDSA(common.Hex2Bytes(m.PrivateKey))
	if err != nil {
		os.Exit(-1)
	}
	auth := bind.NewKeyedTransactor(privateKey)
	auth.GasLimit = big.NewInt(m.GasLimit)
	auth.GasPrice = big.NewInt(m.GasPrice)
	if m.Nonce < 0 {
		// get system account nonce
		auth.Nonce = nil
	} else {
		auth.Nonce = big.NewInt(m.Nonce)
	}
	auth.Value = big.NewInt(m.Value)
	return auth
}

// BigIntArrayToBytes converts []*big.Int to [][]byte
func BigIntArrayToBytes(ints []*big.Int) (b [][]byte) {
	for _, i := range ints {
		b = append(b, i.Bytes())
	}
	return
}

// BytesToBigIntArray converts [][]byte to []*big.Int
func BytesToBigIntArray(b [][]byte) (ints []*big.Int) {
	for _, i := range b {
		ints = append(ints, new(big.Int).SetBytes(i))
	}
	return
}
`

func (c *Contract) Write(filepath, filename string) {
	implTemplate, err := template.New("contract").Parse(ContractTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template: %v\n", err)
		os.Exit(-1)
	}
	result := new(bytes.Buffer)
	err = implTemplate.Execute(result, c)
	if err != nil {
		fmt.Printf("Failed to render template: %v\n", err)
		os.Exit(-1)
	}
	content := html.UnescapeString(html.UnescapeString(result.String()))
	util.WriteFile(content, filepath, filename)
}