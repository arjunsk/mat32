package main

import (
	"github.com/arjunsk/mat32"
	"gonum.org/v1/gonum/mat"
	"math"
	"reflect"
	"unsafe"
)

type RealNumbers interface {
	float64 | float32
}

func L2Distance[T RealNumbers](v1, v2 []T) (float64, error) {
	var result0 []float32
	header0 := *(*reflect.SliceHeader)(unsafe.Pointer(&v1))
	result0 = *(*[]float32)(unsafe.Pointer(&header0))
	vec0 := mat32.NewVecDense(len(v1), result0)

	var result1 []float32
	header1 := *(*reflect.SliceHeader)(unsafe.Pointer(&v2))
	result1 = *(*[]float32)(unsafe.Pointer(&header1))
	vec1 := mat32.NewVecDense(len(v2), result1)

	diff := mat32.NewVecDense(vec0.Len(), nil)
	diff.SubVec(vec0, vec1)

	//vec0.SubVec(vec0, vec1)

	return math.Sqrt(float64(mat32.Dot(vec0, vec0))), nil
}

func L2Distance2[T RealNumbers](v1, v2 []float64) (float64, error) {
	vec0 := mat.NewVecDense(len(v1), v1)
	vec1 := mat.NewVecDense(len(v2), v2)

	diff := mat.NewVecDense(vec0.Len(), nil)
	diff.SubVec(vec0, vec1)
	return math.Sqrt(mat.Dot(vec0, vec0)), nil
}

//func NormalizeL2[T RealNumbers](v1 []T) ([]T, error) {
//	var result0 []float32
//	header0 := *(*reflect.SliceHeader)(unsafe.Pointer(&v1))
//	result0 = *(*[]float32)(unsafe.Pointer(&header0))
//	vec := mat32.NewVecDense(len(v1), result0)
//
//	norm := mat32.Norm(vec, 2)
//	if norm == 0 {
//		// NOTE: don't throw error here. If you throw error, then when a zero vector comes in the Vector Index
//		// Mapping Query, the query will fail. Instead, return the same zero vector.
//		// This is consistent with FAISS:https://github.com/facebookresearch/faiss/blob/0716bde2500edb2e18509bf05f5dfa37bd698082/faiss/utils/distances.cpp#L97
//		return v1, nil
//	}
//
//	vec.ScaleVec(1/norm, vec)
//
//	return ToMoArray[T](vec), nil
//}

func main() {

	vector2 := []float32{1, 2, 3}
	res2, _ := L2Distance[float32](vector2, vector2)
	println(res2)

	vector3 := []float64{1, 2, 3}
	res3, _ := L2Distance2[float64](vector3, vector3)
	println(res3)

	//println(res1, res2)
}
