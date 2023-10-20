package generate

import "github.com/hanyue2020/go-swagger3/generate/adaptor"

type Generator interface {
	Gen(args ...string) any
}

var (
	Gens = map[string]Generator{
		"randomint":   &adaptor.RandomInt{},
		"cn_id":       &adaptor.CnIdNo{},
		"cn_mobile":   &adaptor.CnMobile{},
		"cn_name":     &adaptor.CnName{},
		"cn_bankno":   &adaptor.CnBankNo{},
		"cn_address":  &adaptor.CnAddress{},
		"uuid":        &adaptor.Uuid{},
		"unix_mico":   &adaptor.UnixMicro{},
		"unix_nano":   &adaptor.UnixNano{},
		"unix_milli":  &adaptor.UnixMilli{},
		"unix_second": &adaptor.UnixSecond{},
		"pickup":      &adaptor.PickUp{},
	}
)
