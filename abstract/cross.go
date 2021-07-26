package abstract

import (
	"math"
	"math/big"
)




type decimalMapper struct {
	base *big.Int
}

func NewDecimalMapperFromFloat(amount float64, decimals uint) *decimalMapper {
	qtr := math.Pow(10, float64(decimals))
	base := big.NewInt(int64(amount * qtr))
	return &decimalMapper{ base }
}

func NewDecimalMapperFromBig(amount *big.Int) *decimalMapper {
	return &decimalMapper{ base: amount }
}

func (dm *decimalMapper) MapThrough(originDecimals, destDecimals uint) *big.Int {
	if originDecimals > destDecimals {
		return dm.MapTo(originDecimals - destDecimals)
	} else {
		return dm.MapTo(destDecimals - originDecimals)	
	}
}

func (dm *decimalMapper) MapFrom(decimals uint) *big.Int {
	return mapFromDecimals(dm.base, decimals)
}

func (dm *decimalMapper) MapTo(decimals uint) *big.Int {
	return mapToDecimals(dm.base, decimals)
}


func mapToDecimals(amount *big.Int, decimals uint) *big.Int {
	base, qtr := multiplyParts(amount, decimals)
	return base.Mul(base, qtr)
}

func mapFromDecimals(amount *big.Int, decimals uint) *big.Int {
	base, qtr := multiplyParts(amount, decimals)
	return base.Div(base, qtr)
}

func multiplyParts(amount *big.Int, decimals uint) (*big.Int, *big.Int) {
	base := big.NewInt(0).Set(amount)
	qtr := big.NewInt(10).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	return base, qtr
}