package harness

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/ttacon/libphonenumber"
)

// DataGen handles generating random data for tests. It ties arbitrary ids to
// generated values so they can be re-used during a test.
type DataGen struct {
	data map[dataGenKey]string
	uniq map[dataGenKey]bool
	mx   sync.Mutex
	g    Generator
	t    *testing.T
	name string
}

type DataGenFunc func() string

type DataGenArgFunc func(string) string

type Generator interface {
	Generate(string) string
}

func (d DataGenFunc) Generate(string) string {
	return d()
}

func (d DataGenArgFunc) Generate(a string) string {
	return d(a)
}

// NewDataGen will create a new data generator. fn should return a new/unique string each time
func NewDataGen(t *testing.T, name string, g Generator) *DataGen {
	return &DataGen{
		data: make(map[dataGenKey]string),
		uniq: make(map[dataGenKey]bool),
		g:    g,
		t:    t,
		name: name,
	}
}

type dataGenKey struct{ arg, id string }

// Get returns the value associated with id. The first time an id is provided,
// a new value is generated. If id is empty, a new value will always be returned.
func (d *DataGen) Get(id string) string {
	return d.GetWithArg("", id)
}

// GetWithArg returns the value associated with id. The first time an id is provided,
// a new value is generated. If id is empty, a new value will always be returned.
func (d *DataGen) GetWithArg(arg, id string) string {
	d.mx.Lock()
	defer d.mx.Unlock()
	key := dataGenKey{arg: arg, id: id}
	val := dataGenKey{arg: arg, id: ""}
	var ok bool
	if id != "" {
		// only return previous value if given an ID
		val.id, ok = d.data[key]
	}
	if !ok {
		val.id = d.g.Generate(arg)
		for d.uniq[val] {
			val.id = d.g.Generate(arg)
		}
		d.uniq[val] = true
		d.t.Logf(`%s("%s") = "%s"`, d.name, id, val.id)
		d.data[key] = val.id
	}

	return val.id
}

// GenUUID will return a random UUID.
func GenUUID() string {
	return uuid.NewV4().String()
}

// GenPhone will return a random phone number
func GenPhone() string {
	return GenPhoneCC("+1")
}

// GenPhoneCC will return a random phone number with supplied country code
func GenPhoneCC(cc string) string {
	ccInt, err := strconv.Atoi(strings.TrimPrefix(cc, "+"))
	if err != nil {
		panic(errors.Wrapf(err, "parse country code '%s'", cc))
	}
	region := libphonenumber.GetRegionCodeForCountryCode(ccInt)
	if region == "" || region == "ZZ" {
		panic(fmt.Sprintf("invalid cc '%s'", cc))
	}
	num := libphonenumber.GetExampleNumber(region)
	*num.NationalNumber = *num.NationalNumber + uint64(rand.Intn(9999))
	return libphonenumber.Format(num, libphonenumber.E164)
}
