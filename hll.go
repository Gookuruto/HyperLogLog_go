package main

import (
	"math"
	"fmt"
	//"hash"
	"hash/fnv"
	"encoding/binary"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"math/rand"
	"github.com/eclesh/hyperloglog"
)

const (
	pow32    float64 = 4294967296
	negpow32 float64 = -4294967296
	alpha16  float64 = 0.673
	alpha32  float64 = 0.697
	alpha64  float64 = 0.709
)

// A HyperLogLog cardinality estimator.
//
// See http://algo.inria.fr/flajolet/Publications/FlFuGaMe07.pdf for
// more information.
type HyperLogLog struct {
	m      uint
	k      float64
	kComp  int
	alphaM float64
	bits   []uint8
}
func (h *HyperLogLog) Clear() {
	h.bits = make([]uint8, h.m)
}

// NewHyperLogLog returns an estimator for counting cardinality to within the given stderr.
//
// Smaller values require more space, but provide more accurate
// results.  For a good time, try 0.001 or so.
func NewHyperLogLog(registers uint) *HyperLogLog {
	rv := &HyperLogLog{}

	m := 1.04/(1.04 / math.Sqrt(float64(registers)))
	rv.m = registers
	rv.k = math.Ceil(math.Log2(float64(m) * float64(m)))
	rv.kComp = int(32 - rv.k)
	fmt.Println("Registers : ")
	//fmt.Println(rv.m)
	//fmt.Println("k : ")
	//fmt.Println(math.Ceil(float64(rv.m)*5/32))

	switch rv.m {
	case 16:
		rv.alphaM = alpha16
	case 32:
		rv.alphaM = alpha32
	case 64:
		rv.alphaM = alpha64
	default:
		rv.alphaM = 0.7213 / (1 + 1.079/float64(rv.m))
	}

	rv.bits = make([]uint8, rv.m)
	fmt.Println(len(rv.bits))
	return rv
}

// Add an item by its hash.
func (h *HyperLogLog) Add(hash uint32) {
	r := 1
	for (hash&1) == 0 && r <= h.kComp {
		r++
		hash >>= 1
	}

	j := hash >> uint(h.kComp)
	if r > int(h.bits[j]) {
		h.bits[j] = uint8(r)
	}
}

// Count returns the current estimate of the number of distinct items seen.
func (h *HyperLogLog) Count() uint64 {
	c := 0.0
	for i := uint(0); i < h.m; i++ {
		c += (1 / math.Pow(2.0, float64(h.bits[i])))
	}
	E := h.alphaM * float64(h.m*h.m) / c

	// -- make corrections

	if E <= 5/2*float64(h.m) {
		V := float64(0)
		for i := uint(0); i < h.m; i++ {
			if h.bits[i] == 0 {
				V++
			}
		}
		if V > 0 {
			E = float64(h.m) * math.Log(float64(h.m)/V)
		}
	} else if E > 1/30*pow32 {
		E = negpow32 * math.Log(1-E/pow32)
	}
	return uint64(E)
}

// Merge another HyperLogLog into this one.
func (h *HyperLogLog) Merge(from *HyperLogLog) {
	if len(h.bits) != len(from.bits) {
		panic("HLLs are incompatible. They must have the same basis")
	}

	for i, v := range from.bits {
		if v > h.bits[i] {
			h.bits[i] = v
		}
	}
}
func generate_M(size,firstelement uint32) []uint32  {
	M :=make([]uint32,0,0)
	j:=firstelement
	for i:=0;i<int(size);i++{
		M=append(M,j)
		j++


	}
	return M
}


func Genrate_many_M(how_many uint32) [][]uint32  {n
	x:=make([][]uint32,how_many,how_many)
	for i:=0;uint32(i)<how_many;i++{
		if i==0 {
			//fmt.Println(i)
			x[i] = make([]uint32, i+1)
			x[i]=generate_M(uint32(i+1),1)
		}else{
			x[i] = make([]uint32, i+1)
			x[i]=generate_M(uint32(i+1),x[i-1][len(x[i-1])-1]+1)
		}

	}
	return x

}


func main() {
	//generate Multisets
	x:=Genrate_many_M(10000)
	//create HLL object
	h,_:=hyperloglog.New(uint(math.Pow(2,10))	)
	//h,_:=hyperloglog.New(16)
	z:=fnv.New32()
	y:=make([]float64,0,0)
	var tmp  uint32
	var rnd uint32
	//fmt.Println(z)
	//fmt.Println("Hello x: ")
	for i:=0;i<len(x);i++{
		//fmt.Println(x[i])
		for j:=0;j<len(x[i]);j++{
			rnd=rand.Uint32()%10
				a := make([]byte, 4)
				binary.LittleEndian.PutUint32(a, x[i][j])
				z.Write(a)
				//tmp = crc32.ChecksumIEEE(a)
				tmp = z.Sum32()
				//fmt.Println(tmp)
			for iter:=uint32(0);iter<=rnd;iter++ {
				h.Add(tmp)
			}
			//`z.Reset()

		}
		//fmt.Println(h.Count())
		//fmt.Println(len(x[i]))
		y=append(y,(float64(h.Count())/float64(len(x[i]))))
		h.Reset()
	}
	//make plot
	pltor:=make(plotter.XYs,len(y))
	for i:=0;i<len(y);i++{
		pltor[i].X=float64(i+1)
		pltor[i].Y=y[i]


	}
	p,err:=plot.New()
	if err!=nil{
		panic(err)
	}
	p.Title.Text = "HyperLogLog"
	p.X.Label.Text="n"
	p.Y.Label.Text="estimator/n"
	err=plotutil.AddScatters(p,"First",pltor)
	//Save plot to png file
	if err :=p.Save(10*vg.Inch,10*vg.Inch,"estimation.png");err!=nil{
		panic(err)
	}
	fmt.Println(y)
}
