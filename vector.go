// Copyright ©2013 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mat32

import (
	"github.com/arjunsk/mat32/internal/asm/f32"
	"gonum.org/v1/gonum/blas"
	"gonum.org/v1/gonum/blas/blas32"
)

var (
	vector *VecDense

	_ Matrix  = vector
	_ Vector  = vector
	_ Reseter = vector
)

// Vector is a vector.
type Vector interface {
	Matrix
	AtVec(int) float32
	Len() int
}

// TransposeVec is a type for performing an implicit transpose of a Vector.
// It implements the Vector interface, returning values from the transpose
// of the vector within.
type TransposeVec struct {
	Vector Vector
}

// At returns the value of the element at row i and column j of the transposed
// matrix, that is, row j and column i of the Vector field.
func (t TransposeVec) At(i, j int) float32 {
	return t.Vector.At(j, i)
}

// AtVec returns the element at position i. It panics if i is out of bounds.
func (t TransposeVec) AtVec(i int) float32 {
	return t.Vector.AtVec(i)
}

// Dims returns the dimensions of the transposed vector.
func (t TransposeVec) Dims() (r, c int) {
	c, r = t.Vector.Dims()
	return r, c
}

// T performs an implicit transpose by returning the Vector field.
func (t TransposeVec) T() Matrix {
	return t.Vector
}

// Len returns the number of columns in the vector.
func (t TransposeVec) Len() int {
	return t.Vector.Len()
}

// TVec performs an implicit transpose by returning the Vector field.
func (t TransposeVec) TVec() Vector {
	return t.Vector
}

// Untranspose returns the Vector field.
func (t TransposeVec) Untranspose() Matrix {
	return t.Vector
}

func (t TransposeVec) UntransposeVec() Vector {
	return t.Vector
}

// VecDense represents a column vector.
type VecDense struct {
	mat blas32.Vector
	n   int
	// A BLAS vector can have a negative increment, but allowing this
	// in the mat type complicates a lot of code, and doesn't gain anything.
	// VecDense must have positive increment in this package.
}

// NewVecDense creates a new VecDense of length n. If data == nil,
// a new slice is allocated for the backing slice. If len(data) == n, data is
// used as the backing slice, and changes to the elements of the returned VecDense
// will be reflected in data. If neither of these is true, NewVecDense will panic.
func NewVecDense(n int, data []float32) *VecDense {
	if n < 0 {
		panic("mat: negative dimension")
	}
	if len(data) != n && data != nil {
		panic(ErrShape)
	}
	if data == nil {
		data = make([]float32, n)
	}
	return &VecDense{
		mat: blas32.Vector{
			Inc:  1,
			Data: data,
		},
		n: n,
	}
}

// SliceVec returns a new Vector that shares backing data with the receiver.
// The returned matrix starts at i of the receiver and extends k-i elements.
// SliceVec panics with ErrIndexOutOfRange if the slice is outside the capacity
// of the receiver.
func (v *VecDense) SliceVec(i, k int) Vector {
	if i < 0 || k <= i || v.Cap() < k {
		panic(ErrIndexOutOfRange)
	}
	return &VecDense{
		n: k - i,
		mat: blas32.Vector{
			Inc:  v.mat.Inc,
			Data: v.mat.Data[i*v.mat.Inc : (k-1)*v.mat.Inc+1],
		},
	}
}

// Dims returns the number of rows and columns in the matrix. Columns is always 1
// for a non-Reset vector.
func (v *VecDense) Dims() (r, c int) {
	if v.IsZero() {
		return 0, 0
	}
	return v.n, 1
}

// Caps returns the number of rows and columns in the backing matrix. Columns is always 1
// for a non-Reset vector.
func (v *VecDense) Caps() (r, c int) {
	if v.IsZero() {
		return 0, 0
	}
	return v.Cap(), 1
}

// Len returns the length of the vector.
func (v *VecDense) Len() int {
	return v.n
}

// Cap returns the capacity of the vector.
func (v *VecDense) Cap() int {
	if v.IsZero() {
		return 0
	}
	return (cap(v.mat.Data)-1)/v.mat.Inc + 1
}

// T performs an implicit transpose by returning the receiver inside a Transpose.
func (v *VecDense) T() Matrix {
	return Transpose{v}
}

// TVec performs an implicit transpose by returning the receiver inside a TransposeVec.
func (v *VecDense) TVec() Vector {
	return TransposeVec{v}
}

// Reset zeros the length of the vector so that it can be reused as the
// receiver of a dimensionally restricted operation.
//
// See the Reseter interface for more information.
func (v *VecDense) Reset() {
	// No change of Inc or n to 0 may be
	// made unless both are set to 0.
	v.mat.Inc = 0
	v.n = 0
	v.mat.Data = v.mat.Data[:0]
}

// CloneVec makes a copy of a into the receiver, overwriting the previous value
// of the receiver.
func (v *VecDense) CloneVec(a Vector) {
	if v == a {
		return
	}
	v.n = a.Len()
	v.mat = blas32.Vector{
		Inc:  1,
		Data: use(v.mat.Data, v.n),
	}
	if r, ok := a.(RawVectorer); ok {
		blas32.Copy(v.n, r.RawVector(), v.mat)
		return
	}
	for i := 0; i < a.Len(); i++ {
		v.SetVec(i, a.AtVec(i))
	}
}

// VecDenseCopyOf returns a newly allocated copy of the elements of a.
func VecDenseCopyOf(a Vector) *VecDense {
	v := &VecDense{}
	v.CloneVec(a)
	return v
}

func (v *VecDense) RawVector() blas32.Vector {
	return v.mat
}

// CopyVec makes a copy of elements of a into the receiver. It is similar to the
// built-in copy; it copies as much as the overlap between the two vectors and
// returns the number of elements it copied.
func (v *VecDense) CopyVec(a Vector) int {
	n := min(v.Len(), a.Len())
	if v == a {
		return n
	}
	if r, ok := a.(RawVectorer); ok {
		blas32.Copy(n, r.RawVector(), v.mat)
		return n
	}
	for i := 0; i < n; i++ {
		v.setVec(i, a.AtVec(i))
	}
	return n
}

// ScaleVec scales the vector a by alpha, placing the result in the receiver.
func (v *VecDense) ScaleVec(alpha float32, a Vector) {
	n := a.Len()

	if v == a {
		if v.mat.Inc == 1 {
			f32.ScalUnitary(alpha, v.mat.Data)
			return
		}
		f32.ScalInc(alpha, v.mat.Data, uintptr(n), uintptr(v.mat.Inc))
		return
	}

	v.reuseAs(n)

	if rv, ok := a.(RawVectorer); ok {
		mat := rv.RawVector()
		v.checkOverlap(mat)
		if v.mat.Inc == 1 && mat.Inc == 1 {
			f32.ScalUnitaryTo(v.mat.Data, alpha, mat.Data)
			return
		}
		f32.ScalIncTo(v.mat.Data, uintptr(v.mat.Inc),
			alpha, mat.Data, uintptr(n), uintptr(mat.Inc))
		return
	}

	for i := 0; i < n; i++ {
		v.setVec(i, alpha*a.AtVec(i))
	}
}

// AddScaledVec adds the vectors a and alpha*b, placing the result in the receiver.
func (v *VecDense) AddScaledVec(a Vector, alpha float32, b Vector) {
	if alpha == 1 {
		v.AddVec(a, b)
		return
	}
	if alpha == -1 {
		v.SubVec(a, b)
		return
	}

	ar := a.Len()
	br := b.Len()

	if ar != br {
		panic(ErrShape)
	}

	var amat, bmat blas32.Vector
	fast := true
	aU, _ := untranspose(a)
	if rv, ok := aU.(RawVectorer); ok {
		amat = rv.RawVector()
		if v != a {
			v.checkOverlap(amat)
		}
	} else {
		fast = false
	}
	bU, _ := untranspose(b)
	if rv, ok := bU.(RawVectorer); ok {
		bmat = rv.RawVector()
		if v != b {
			v.checkOverlap(bmat)
		}
	} else {
		fast = false
	}

	v.reuseAs(ar)

	switch {
	case alpha == 0: // v <- a
		if v == a {
			return
		}
		v.CopyVec(a)
	case v == a && v == b: // v <- v + alpha * v = (alpha + 1) * v
		blas32.Scal(ar, alpha+1, v.mat)
	case !fast: // v <- a + alpha * b without blas32 support.
		for i := 0; i < ar; i++ {
			v.setVec(i, a.AtVec(i)+alpha*b.AtVec(i))
		}
	case v == a && v != b: // v <- v + alpha * b
		if v.mat.Inc == 1 && bmat.Inc == 1 {
			// Fast path for a common case.
			f32.AxpyUnitaryTo(v.mat.Data, alpha, bmat.Data, amat.Data)
		} else {
			f32.AxpyInc(alpha, bmat.Data, v.mat.Data,
				uintptr(ar), uintptr(bmat.Inc), uintptr(v.mat.Inc), 0, 0)
		}
	default: // v <- a + alpha * b or v <- a + alpha * v
		if v.mat.Inc == 1 && amat.Inc == 1 && bmat.Inc == 1 {
			// Fast path for a common case.
			f32.AxpyUnitaryTo(v.mat.Data, alpha, bmat.Data, amat.Data)
		} else {
			f32.AxpyIncTo(v.mat.Data, uintptr(v.mat.Inc), 0,
				alpha, bmat.Data, amat.Data,
				uintptr(ar), uintptr(bmat.Inc), uintptr(amat.Inc), 0, 0)
		}
	}
}

// AddVec adds the vectors a and b, placing the result in the receiver.
func (v *VecDense) AddVec(a, b Vector) {
	ar := a.Len()
	br := b.Len()

	if ar != br {
		panic(ErrShape)
	}

	v.reuseAs(ar)

	aU, _ := untranspose(a)
	bU, _ := untranspose(b)

	if arv, ok := aU.(RawVectorer); ok {
		if brv, ok := bU.(RawVectorer); ok {
			amat := arv.RawVector()
			bmat := brv.RawVector()

			if v != a {
				v.checkOverlap(amat)
			}
			if v != b {
				v.checkOverlap(bmat)
			}

			if v.mat.Inc == 1 && amat.Inc == 1 && bmat.Inc == 1 {
				// Fast path for a common case.
				f32.AxpyUnitaryTo(v.mat.Data, 1, bmat.Data, amat.Data)
				return
			}
			f32.AxpyIncTo(v.mat.Data, uintptr(v.mat.Inc), 0,
				1, bmat.Data, amat.Data,
				uintptr(ar), uintptr(bmat.Inc), uintptr(amat.Inc), 0, 0)
			return
		}
	}

	for i := 0; i < ar; i++ {
		v.setVec(i, a.AtVec(i)+b.AtVec(i))
	}
}

// SubVec subtracts the vector b from a, placing the result in the receiver.
func (v *VecDense) SubVec(a, b Vector) {
	ar := a.Len()
	br := b.Len()

	if ar != br {
		panic(ErrShape)
	}

	v.reuseAs(ar)

	aU, _ := untranspose(a)
	bU, _ := untranspose(b)

	if arv, ok := aU.(RawVectorer); ok {
		if brv, ok := bU.(RawVectorer); ok {
			amat := arv.RawVector()
			bmat := brv.RawVector()

			if v != a {
				v.checkOverlap(amat)
			}
			if v != b {
				v.checkOverlap(bmat)
			}

			if v.mat.Inc == 1 && amat.Inc == 1 && bmat.Inc == 1 {
				// Fast path for a common case.
				f32.AxpyUnitaryTo(v.mat.Data, -1, bmat.Data, amat.Data)
				return
			}
			f32.AxpyIncTo(v.mat.Data, uintptr(v.mat.Inc), 0,
				-1, bmat.Data, amat.Data,
				uintptr(ar), uintptr(bmat.Inc), uintptr(amat.Inc), 0, 0)
			return
		}
	}

	for i := 0; i < ar; i++ {
		v.setVec(i, a.AtVec(i)-b.AtVec(i))
	}
}

// MulElemVec performs element-wise multiplication of a and b, placing the result
// in the receiver.
func (v *VecDense) MulElemVec(a, b Vector) {
	ar := a.Len()
	br := b.Len()

	if ar != br {
		panic(ErrShape)
	}

	v.reuseAs(ar)

	aU, _ := untranspose(a)
	bU, _ := untranspose(b)

	if arv, ok := aU.(RawVectorer); ok {
		if brv, ok := bU.(RawVectorer); ok {
			amat := arv.RawVector()
			bmat := brv.RawVector()

			if v != a {
				v.checkOverlap(amat)
			}
			if v != b {
				v.checkOverlap(bmat)
			}

			if v.mat.Inc == 1 && amat.Inc == 1 && bmat.Inc == 1 {
				// Fast path for a common case.
				for i, a := range amat.Data {
					v.mat.Data[i] = a * bmat.Data[i]
				}
				return
			}
			var ia, ib int
			for i := 0; i < ar; i++ {
				v.setVec(i, amat.Data[ia]*bmat.Data[ib])
				ia += amat.Inc
				ib += bmat.Inc
			}
			return
		}
	}

	for i := 0; i < ar; i++ {
		v.setVec(i, a.AtVec(i)*b.AtVec(i))
	}
}

// DivElemVec performs element-wise division of a by b, placing the result
// in the receiver.
func (v *VecDense) DivElemVec(a, b Vector) {
	ar := a.Len()
	br := b.Len()

	if ar != br {
		panic(ErrShape)
	}

	v.reuseAs(ar)

	aU, _ := untranspose(a)
	bU, _ := untranspose(b)

	if arv, ok := aU.(RawVectorer); ok {
		if brv, ok := bU.(RawVectorer); ok {
			amat := arv.RawVector()
			bmat := brv.RawVector()

			if v != a {
				v.checkOverlap(amat)
			}
			if v != b {
				v.checkOverlap(bmat)
			}

			if v.mat.Inc == 1 && amat.Inc == 1 && bmat.Inc == 1 {
				// Fast path for a common case.
				for i, a := range amat.Data {
					v.setVec(i, a/bmat.Data[i])
				}
				return
			}
			var ia, ib int
			for i := 0; i < ar; i++ {
				v.setVec(i, amat.Data[ia]/bmat.Data[ib])
				ia += amat.Inc
				ib += bmat.Inc
			}
		}
	}

	for i := 0; i < ar; i++ {
		v.setVec(i, a.AtVec(i)/b.AtVec(i))
	}
}

// MulVec computes a * b. The result is stored into the receiver.
// MulVec panics if the number of columns in a does not equal the number of rows in b
// or if the number of columns in b does not equal 1.
func (v *VecDense) MulVec(a Matrix, b Vector) {
	r, c := a.Dims()
	br, bc := b.Dims()
	if c != br || bc != 1 {
		panic(ErrShape)
	}

	aU, trans := untranspose(a)
	var bmat blas32.Vector
	fast := true
	bU, _ := untranspose(b)
	if rv, ok := bU.(RawVectorer); ok {
		bmat = rv.RawVector()
		if v != b {
			v.checkOverlap(bmat)
		}
	} else {
		fast = false
	}

	v.reuseAs(r)
	var restore func()
	if v == aU {
		v, restore = v.isolatedWorkspace(aU.(*VecDense))
		defer restore()
	} else if v == b {
		v, restore = v.isolatedWorkspace(b)
		defer restore()
	}

	// TODO(kortschak): Improve the non-fast paths.
	switch aU := aU.(type) {
	case Vector:
		if b.Len() == 1 {
			// {n,1} x {1,1}
			v.ScaleVec(b.AtVec(0), aU)
			return
		}

		// {1,n} x {n,1}
		if fast {
			if rv, ok := aU.(RawVectorer); ok {
				amat := rv.RawVector()
				if v != aU {
					v.checkOverlap(amat)
				}

				if amat.Inc == 1 && bmat.Inc == 1 {
					// Fast path for a common case.
					v.setVec(0, f32.DotUnitary(amat.Data, bmat.Data))
					return
				}
				v.setVec(0, f32.DotInc(amat.Data, bmat.Data,
					uintptr(c), uintptr(amat.Inc), uintptr(bmat.Inc), 0, 0))
				return
			}
		}
		var sum float32
		for i := 0; i < c; i++ {
			sum += aU.AtVec(i) * b.AtVec(i)
		}
		v.setVec(0, sum)
		return
	case RawTriangular:
		v.CopyVec(b)
		amat := aU.RawTriangular()
		// We don't know that a is a *TriDense, so make
		// a temporary TriDense to check overlap.
		(&TriDense{mat: amat}).checkOverlap(v.asGeneral())
		ta := blas.NoTrans
		if trans {
			ta = blas.Trans
		}
		blas32.Trmv(ta, amat, v.mat)
	case RawMatrixer:
		if fast {
			amat := aU.RawMatrix()
			// We don't know that a is a *Dense, so make
			// a temporary Dense to check overlap.
			(&Dense{mat: amat}).checkOverlap(v.asGeneral())
			t := blas.NoTrans
			if trans {
				t = blas.Trans
			}
			blas32.Gemv(t, 1, amat, bmat, 0, v.mat)
			return
		}
	default:
		if fast {
			for i := 0; i < r; i++ {
				var f float32
				for j := 0; j < c; j++ {
					f += a.At(i, j) * bmat.Data[j*bmat.Inc]
				}
				v.setVec(i, f)
			}
			return
		}
	}

	for i := 0; i < r; i++ {
		var f float32
		for j := 0; j < c; j++ {
			f += a.At(i, j) * b.AtVec(j)
		}
		v.setVec(i, f)
	}
}

// reuseAs resizes an empty vector to a r×1 vector,
// or checks that a non-empty matrix is r×1.
func (v *VecDense) reuseAs(r int) {
	if r == 0 {
		panic(ErrZeroLength)
	}
	if v.IsZero() {
		v.mat = blas32.Vector{
			Inc:  1,
			Data: use(v.mat.Data, r),
		}
		v.n = r
		return
	}
	if r != v.n {
		panic(ErrShape)
	}
}

// IsZero returns whether the receiver is zero-sized. Zero-sized vectors can be the
// receiver for size-restricted operations. VecDenses can be zeroed using Reset.
func (v *VecDense) IsZero() bool {
	// It must be the case that v.Dims() returns
	// zeros in this case. See comment in Reset().
	return v.mat.Inc == 0
}

func (v *VecDense) isolatedWorkspace(a Vector) (n *VecDense, restore func()) {
	l := a.Len()
	if l == 0 {
		panic(ErrZeroLength)
	}
	n = getWorkspaceVec(l, false)
	return n, func() {
		v.CopyVec(n)
		putWorkspaceVec(n)
	}
}

// asDense returns a Dense representation of the receiver with the same
// underlying data.
func (v *VecDense) asDense() *Dense {
	return &Dense{
		mat:     v.asGeneral(),
		capRows: v.n,
		capCols: 1,
	}
}

// asGeneral returns a blas32.General representation of the receiver with the
// same underlying data.
func (v *VecDense) asGeneral() blas32.General {
	return blas32.General{
		Rows:   v.n,
		Cols:   1,
		Stride: v.mat.Inc,
		Data:   v.mat.Data,
	}
}

// ColViewOf reflects the column j of the RawMatrixer m, into the receiver
// backed by the same underlying data. The length of the receiver must either be
// zero or match the number of rows in m.
func (v *VecDense) ColViewOf(m RawMatrixer, j int) {
	rm := m.RawMatrix()

	if j >= rm.Cols || j < 0 {
		panic(ErrColAccess)
	}
	if !v.IsZero() && v.n != rm.Rows {
		panic(ErrShape)
	}

	v.mat.Inc = rm.Stride
	v.mat.Data = rm.Data[j : (rm.Rows-1)*rm.Stride+j+1]
	v.n = rm.Rows
}

// RowViewOf reflects the row i of the RawMatrixer m, into the receiver
// backed by the same underlying data. The length of the receiver must either be
// zero or match the number of columns in m.
func (v *VecDense) RowViewOf(m RawMatrixer, i int) {
	rm := m.RawMatrix()

	if i >= rm.Rows || i < 0 {
		panic(ErrRowAccess)
	}
	if !v.IsZero() && v.n != rm.Cols {
		panic(ErrShape)
	}

	v.mat.Inc = 1
	v.mat.Data = rm.Data[i*rm.Stride : i*rm.Stride+rm.Cols]
	v.n = rm.Cols
}
