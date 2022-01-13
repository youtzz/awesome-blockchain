package ecc

import (
	"math/big"
)

var Zero = new(big.Rat).SetInt64(0)

func IsEqual(x, y *big.Rat) bool {
	return x.Cmp(y) == 0
}

// Mod 用费马小定理求分数mod
func Mod(rat *big.Rat, q *big.Int) *big.Rat {
	n := rat.Num()
	d := rat.Denom()

	pow := func(num *big.Int, times int64) *big.Int {
		res := big.NewInt(1)
		for i := int64(0); i < times; i++ {
			res.Mul(res, num)
		}
		return res
	}

	p := pow(d, q.Int64()-2)
	p.Mul(p, n)
	p.Mod(p, q)
	return new(big.Rat).SetInt(p)
}

type EllipticCurve struct {
	a, b  *big.Rat
	order *big.Int
}

func NewEllipticCurve(A, B, Order int64) EllipticCurve {
	return EllipticCurve{
		a:     new(big.Rat).SetInt64(A),
		b:     new(big.Rat).SetInt64(B),
		order: big.NewInt(Order),
	}
}

func (f EllipticCurve) Add(p, q Point) Point {
	// def1: p1=O => p1+p2=p2
	// def2: p2=O => p1+p2=p1
	// def3: x1=x2 && (y1+y2)%order=0 => p1+p2=O
	if p.IsO() {
		return q
	} else if q.IsO() {
		return p
	} else if p.X.Cmp(p.X) == 0 {
		yy := new(big.Rat).Add(p.Y, q.Y)
		if IsEqual(f.ModOrder(yy), Zero) {
			return PointO
		}
	}

	k := f.GetSlop(p, q)

	// x = k^2 - p.X - order.X
	x := new(big.Rat).Mul(k, k)
	x.Sub(x, p.X)
	x.Sub(x, q.X)

	// y = k * (x-p.X) + p.Y
	y := new(big.Rat).Sub(x, p.X)
	y.Mul(y, k)
	y.Add(y, p.Y)
	y.Mul(y, new(big.Rat).SetInt64(-1))
	return Point{f.ModOrder(x), f.ModOrder(y)}
}

// GetSlop 计算两点斜率
func (f EllipticCurve) GetSlop(p, q Point) *big.Rat {
	var n, d *big.Rat

	switch {
	case p.Equal(q):
		// k = (3X^2+a)/2Y
		n = new(big.Rat).Add(
			new(big.Rat).Mul(new(big.Rat).SetInt64(3), new(big.Rat).Mul(p.X, p.X)),
			f.a,
		)
		d = new(big.Rat).Mul(new(big.Rat).SetInt64(2), p.Y)
	default:
		n = new(big.Rat).Sub(p.Y, q.Y)
		d = new(big.Rat).Sub(p.X, q.X)
	}

	slop := new(big.Rat).Quo(n, d)

	return f.ModOrder(slop)
}

func (f EllipticCurve) ModOrder(x *big.Rat) *big.Rat {
	return Mod(x, f.order)
}

func (f EllipticCurve) Mul(p Point, n int) Point {
	if n == 0 {
		return NewPoint(0, 0)
	}

	res := p
	for i := 1; i < n; i++ {
		res = f.Add(res, p)
	}
	return res
}

// OnCurve 校验点P是否在曲线上
func (f EllipticCurve) OnCurve(p Point) bool {
	//return p.X^3+f.a*p.X+f.b == p.Y*p.Y
	x3 := new(big.Rat).Mul(p.X, new(big.Rat).Mul(p.X, p.X))

	ax := new(big.Rat).Mul(f.a, p.X)
	b := new(big.Rat).Set(f.b)

	y2 := new(big.Rat).Mul(p.Y, p.Y)

	res := new(big.Rat).Add(x3, new(big.Rat).Add(ax, b))
	res = res.Sub(res, y2)

	return IsEqual(f.ModOrder(res), Zero)
}
