package adaptor

import (
	"github.com/mritd/chinaid"
)

type CnName struct{}

func (m *CnName) Gen(args ...string) any {
	return chinaid.Name()
}

type CnMobile struct{}

func (m *CnMobile) Gen(args ...string) any {
	return chinaid.Mobile()
}

type CnBankNo struct{}

func (m *CnBankNo) Gen(args ...string) any {
	return chinaid.BankNo()
}

type CnIdNo struct{}

func (m *CnIdNo) Gen(args ...string) any {
	return chinaid.IDNo()
}

type CnAddress struct{}

func (m *CnAddress) Gen(args ...string) any {
	return chinaid.Address()
}

type CnCity struct{}

func (m *CnCity) Gen(args ...string) any {
	return chinaid.ProvinceAndCity()
}
