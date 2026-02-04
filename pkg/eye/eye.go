// Package eye provides Go bindings to the Eye neural image engine.
//
// Eye is a high-performance computational library for:
// - Matrix operations (parallel processing)
// - Vector mathematics
// - Geometric primitives (triangles, curves, surfaces)
// - Numerical integration
// - TUI rendering primitives
//
// The library is implemented in Rust for maximum performance and
// exposed to Go via CGO.
package eye

/*
#cgo LDFLAGS: -L${SRCDIR}/../../eye/target/release -leye -lm -ldl -lpthread
#cgo darwin LDFLAGS: -framework Security -framework CoreFoundation

#include <stdlib.h>
#include <stdint.h>

// Forward declarations for Eye FFI functions

// Matrix
typedef void* EyeMatrix;
EyeMatrix eye_matrix_zeros(unsigned int rows, unsigned int cols);
EyeMatrix eye_matrix_ones(unsigned int rows, unsigned int cols);
EyeMatrix eye_matrix_identity(unsigned int n);
EyeMatrix eye_matrix_from_data(unsigned int rows, unsigned int cols, double* data, unsigned int len);
void eye_matrix_free(EyeMatrix m);
unsigned int eye_matrix_rows(EyeMatrix m);
unsigned int eye_matrix_cols(EyeMatrix m);
double eye_matrix_get(EyeMatrix m, unsigned int row, unsigned int col);
void eye_matrix_set(EyeMatrix m, unsigned int row, unsigned int col, double value);
int eye_matrix_copy_data(EyeMatrix m, double* buf, unsigned int len);
EyeMatrix eye_matrix_mul(EyeMatrix a, EyeMatrix b);
EyeMatrix eye_matrix_mul_strassen(EyeMatrix a, EyeMatrix b);
EyeMatrix eye_matrix_add(EyeMatrix a, EyeMatrix b);
EyeMatrix eye_matrix_transpose(EyeMatrix m);
int eye_matrix_det(EyeMatrix m, double* result);
EyeMatrix eye_matrix_inverse(EyeMatrix m);

// Vector
typedef void* EyeVector;
EyeVector eye_vector_from_data(double* data, unsigned int len);
EyeVector eye_vector_zeros(unsigned int n);
void eye_vector_free(EyeVector v);
unsigned int eye_vector_len(EyeVector v);
double eye_vector_get(EyeVector v, unsigned int i);
void eye_vector_set(EyeVector v, unsigned int i, double value);
double eye_vector_dot(EyeVector a, EyeVector b);
double eye_vector_norm(EyeVector v);
EyeVector eye_vector_cross(EyeVector a, EyeVector b);

// Geometry
double eye_triangle_area(
    double v0x, double v0y, double v0z,
    double v1x, double v1y, double v1z,
    double v2x, double v2y, double v2z
);
int eye_triangle_intersect_ray(
    double v0x, double v0y, double v0z,
    double v1x, double v1y, double v1z,
    double v2x, double v2y, double v2z,
    double ox, double oy, double oz,
    double dx, double dy, double dz,
    double* t_out, double* u_out, double* v_out
);

// Canvas
typedef void* EyeCanvas;
EyeCanvas eye_canvas_new(unsigned int width, unsigned int height);
void eye_canvas_free(EyeCanvas c);
void eye_canvas_clear(EyeCanvas c);
void eye_canvas_put(EyeCanvas c, int x, int y, unsigned int ch);
void eye_canvas_line(EyeCanvas c, int x0, int y0, int x1, int y1, unsigned int ch);
void eye_canvas_rect(EyeCanvas c, int x, int y, int w, int h, unsigned int ch);
void eye_canvas_box(EyeCanvas c, int x, int y, int w, int h);
void eye_canvas_circle(EyeCanvas c, int cx, int cy, int r, unsigned int ch);
void eye_canvas_fill_circle(EyeCanvas c, int cx, int cy, int r, unsigned int ch);
void eye_canvas_triangle(EyeCanvas c, int x0, int y0, int x1, int y1, int x2, int y2, unsigned int ch);
void eye_canvas_fill_triangle(EyeCanvas c, int x0, int y0, int x1, int y1, int x2, int y2, unsigned int ch);
void eye_canvas_put_str(EyeCanvas c, int x, int y, char* s);
char* eye_canvas_render_plain(EyeCanvas c);
char* eye_canvas_render(EyeCanvas c);
void eye_string_free(char* s);
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

// Matrix represents a dense matrix with row-major storage.
type Matrix struct {
	ptr C.EyeMatrix
}

// NewMatrixZeros creates a new matrix filled with zeros.
func NewMatrixZeros(rows, cols int) *Matrix {
	m := &Matrix{ptr: C.eye_matrix_zeros(C.uint(rows), C.uint(cols))}
	runtime.SetFinalizer(m, (*Matrix).Free)
	return m
}

// NewMatrixOnes creates a new matrix filled with ones.
func NewMatrixOnes(rows, cols int) *Matrix {
	m := &Matrix{ptr: C.eye_matrix_ones(C.uint(rows), C.uint(cols))}
	runtime.SetFinalizer(m, (*Matrix).Free)
	return m
}

// NewMatrixIdentity creates an identity matrix.
func NewMatrixIdentity(n int) *Matrix {
	m := &Matrix{ptr: C.eye_matrix_identity(C.uint(n))}
	runtime.SetFinalizer(m, (*Matrix).Free)
	return m
}

// NewMatrixFromData creates a matrix from flat data.
func NewMatrixFromData(rows, cols int, data []float64) (*Matrix, error) {
	if len(data) != rows*cols {
		return nil, errors.New("data length does not match dimensions")
	}
	ptr := C.eye_matrix_from_data(
		C.uint(rows),
		C.uint(cols),
		(*C.double)(unsafe.Pointer(&data[0])),
		C.uint(len(data)),
	)
	if ptr == nil {
		return nil, errors.New("failed to create matrix")
	}
	m := &Matrix{ptr: ptr}
	runtime.SetFinalizer(m, (*Matrix).Free)
	return m, nil
}

// Free releases the matrix memory.
func (m *Matrix) Free() {
	if m.ptr != nil {
		C.eye_matrix_free(m.ptr)
		m.ptr = nil
	}
}

// Rows returns the number of rows.
func (m *Matrix) Rows() int {
	return int(C.eye_matrix_rows(m.ptr))
}

// Cols returns the number of columns.
func (m *Matrix) Cols() int {
	return int(C.eye_matrix_cols(m.ptr))
}

// Get returns the element at (row, col).
func (m *Matrix) Get(row, col int) float64 {
	return float64(C.eye_matrix_get(m.ptr, C.uint(row), C.uint(col)))
}

// Set sets the element at (row, col).
func (m *Matrix) Set(row, col int, value float64) {
	C.eye_matrix_set(m.ptr, C.uint(row), C.uint(col), C.double(value))
}

// Data returns the matrix data as a flat slice.
func (m *Matrix) Data() []float64 {
	size := m.Rows() * m.Cols()
	data := make([]float64, size)
	C.eye_matrix_copy_data(m.ptr, (*C.double)(unsafe.Pointer(&data[0])), C.uint(size))
	return data
}

// Mul performs matrix multiplication.
func (m *Matrix) Mul(other *Matrix) (*Matrix, error) {
	ptr := C.eye_matrix_mul(m.ptr, other.ptr)
	if ptr == nil {
		return nil, errors.New("matrix multiplication failed (dimension mismatch?)")
	}
	result := &Matrix{ptr: ptr}
	runtime.SetFinalizer(result, (*Matrix).Free)
	return result, nil
}

// MulStrassen performs matrix multiplication using Strassen's algorithm.
// More efficient for large matrices (O(n^2.807) vs O(n^3)).
func (m *Matrix) MulStrassen(other *Matrix) (*Matrix, error) {
	ptr := C.eye_matrix_mul_strassen(m.ptr, other.ptr)
	if ptr == nil {
		return nil, errors.New("Strassen multiplication failed")
	}
	result := &Matrix{ptr: ptr}
	runtime.SetFinalizer(result, (*Matrix).Free)
	return result, nil
}

// Add performs matrix addition.
func (m *Matrix) Add(other *Matrix) (*Matrix, error) {
	ptr := C.eye_matrix_add(m.ptr, other.ptr)
	if ptr == nil {
		return nil, errors.New("matrix addition failed")
	}
	result := &Matrix{ptr: ptr}
	runtime.SetFinalizer(result, (*Matrix).Free)
	return result, nil
}

// Transpose returns the transpose of the matrix.
func (m *Matrix) Transpose() *Matrix {
	result := &Matrix{ptr: C.eye_matrix_transpose(m.ptr)}
	runtime.SetFinalizer(result, (*Matrix).Free)
	return result
}

// Det computes the determinant.
func (m *Matrix) Det() (float64, error) {
	var result C.double
	if C.eye_matrix_det(m.ptr, &result) == 0 {
		return 0, errors.New("determinant failed (not square matrix?)")
	}
	return float64(result), nil
}

// Inverse computes the matrix inverse.
func (m *Matrix) Inverse() (*Matrix, error) {
	ptr := C.eye_matrix_inverse(m.ptr)
	if ptr == nil {
		return nil, errors.New("matrix inverse failed (singular matrix?)")
	}
	result := &Matrix{ptr: ptr}
	runtime.SetFinalizer(result, (*Matrix).Free)
	return result, nil
}

// Vector represents an N-dimensional vector.
type Vector struct {
	ptr C.EyeVector
}

// NewVector creates a vector from data.
func NewVector(data []float64) *Vector {
	v := &Vector{
		ptr: C.eye_vector_from_data(
			(*C.double)(unsafe.Pointer(&data[0])),
			C.uint(len(data)),
		),
	}
	runtime.SetFinalizer(v, (*Vector).Free)
	return v
}

// NewVectorZeros creates a zero vector.
func NewVectorZeros(n int) *Vector {
	v := &Vector{ptr: C.eye_vector_zeros(C.uint(n))}
	runtime.SetFinalizer(v, (*Vector).Free)
	return v
}

// Free releases the vector memory.
func (v *Vector) Free() {
	if v.ptr != nil {
		C.eye_vector_free(v.ptr)
		v.ptr = nil
	}
}

// Len returns the vector length.
func (v *Vector) Len() int {
	return int(C.eye_vector_len(v.ptr))
}

// Get returns the element at index i.
func (v *Vector) Get(i int) float64 {
	return float64(C.eye_vector_get(v.ptr, C.uint(i)))
}

// Set sets the element at index i.
func (v *Vector) Set(i int, value float64) {
	C.eye_vector_set(v.ptr, C.uint(i), C.double(value))
}

// Dot computes the dot product.
func (v *Vector) Dot(other *Vector) float64 {
	return float64(C.eye_vector_dot(v.ptr, other.ptr))
}

// Norm computes the Euclidean norm.
func (v *Vector) Norm() float64 {
	return float64(C.eye_vector_norm(v.ptr))
}

// Cross computes the cross product (3D only).
func (v *Vector) Cross(other *Vector) (*Vector, error) {
	ptr := C.eye_vector_cross(v.ptr, other.ptr)
	if ptr == nil {
		return nil, errors.New("cross product requires 3D vectors")
	}
	result := &Vector{ptr: ptr}
	runtime.SetFinalizer(result, (*Vector).Free)
	return result, nil
}

// Vec3 represents a 3D vector (lightweight, no FFI overhead).
type Vec3 struct {
	X, Y, Z float64
}

// Triangle represents a 3D triangle.
type Triangle struct {
	V0, V1, V2 Vec3
}

// Area computes the triangle area.
func (t Triangle) Area() float64 {
	return float64(C.eye_triangle_area(
		C.double(t.V0.X), C.double(t.V0.Y), C.double(t.V0.Z),
		C.double(t.V1.X), C.double(t.V1.Y), C.double(t.V1.Z),
		C.double(t.V2.X), C.double(t.V2.Y), C.double(t.V2.Z),
	))
}

// Ray represents a ray for intersection tests.
type Ray struct {
	Origin    Vec3
	Direction Vec3
}

// RayHit represents a ray-triangle intersection result.
type RayHit struct {
	T float64 // Distance along ray
	U float64 // Barycentric coordinate
	V float64 // Barycentric coordinate
}

// IntersectRay tests ray-triangle intersection.
func (t Triangle) IntersectRay(ray Ray) (RayHit, bool) {
	var tOut, uOut, vOut C.double
	hit := C.eye_triangle_intersect_ray(
		C.double(t.V0.X), C.double(t.V0.Y), C.double(t.V0.Z),
		C.double(t.V1.X), C.double(t.V1.Y), C.double(t.V1.Z),
		C.double(t.V2.X), C.double(t.V2.Y), C.double(t.V2.Z),
		C.double(ray.Origin.X), C.double(ray.Origin.Y), C.double(ray.Origin.Z),
		C.double(ray.Direction.X), C.double(ray.Direction.Y), C.double(ray.Direction.Z),
		&tOut, &uOut, &vOut,
	)
	if hit == 0 {
		return RayHit{}, false
	}
	return RayHit{T: float64(tOut), U: float64(uOut), V: float64(vOut)}, true
}

// Canvas represents a character-based canvas for TUI rendering.
type Canvas struct {
	ptr    C.EyeCanvas
	Width  int
	Height int
}

// NewCanvas creates a new canvas.
func NewCanvas(width, height int) *Canvas {
	c := &Canvas{
		ptr:    C.eye_canvas_new(C.uint(width), C.uint(height)),
		Width:  width,
		Height: height,
	}
	runtime.SetFinalizer(c, (*Canvas).Free)
	return c
}

// Free releases the canvas memory.
func (c *Canvas) Free() {
	if c.ptr != nil {
		C.eye_canvas_free(c.ptr)
		c.ptr = nil
	}
}

// Clear clears the canvas.
func (c *Canvas) Clear() {
	C.eye_canvas_clear(c.ptr)
}

// Put draws a character at (x, y).
func (c *Canvas) Put(x, y int, ch rune) {
	C.eye_canvas_put(c.ptr, C.int(x), C.int(y), C.uint(ch))
}

// Line draws a line from (x0, y0) to (x1, y1).
func (c *Canvas) Line(x0, y0, x1, y1 int, ch rune) {
	C.eye_canvas_line(c.ptr, C.int(x0), C.int(y0), C.int(x1), C.int(y1), C.uint(ch))
}

// Rect draws a rectangle outline.
func (c *Canvas) Rect(x, y, w, h int, ch rune) {
	C.eye_canvas_rect(c.ptr, C.int(x), C.int(y), C.int(w), C.int(h), C.uint(ch))
}

// Box draws a box with Unicode box-drawing characters.
func (c *Canvas) Box(x, y, w, h int) {
	C.eye_canvas_box(c.ptr, C.int(x), C.int(y), C.int(w), C.int(h))
}

// Circle draws a circle outline.
func (c *Canvas) Circle(cx, cy, r int, ch rune) {
	C.eye_canvas_circle(c.ptr, C.int(cx), C.int(cy), C.int(r), C.uint(ch))
}

// FillCircle draws a filled circle.
func (c *Canvas) FillCircle(cx, cy, r int, ch rune) {
	C.eye_canvas_fill_circle(c.ptr, C.int(cx), C.int(cy), C.int(r), C.uint(ch))
}

// Triangle draws a triangle outline.
func (c *Canvas) Triangle(x0, y0, x1, y1, x2, y2 int, ch rune) {
	C.eye_canvas_triangle(c.ptr, C.int(x0), C.int(y0), C.int(x1), C.int(y1), C.int(x2), C.int(y2), C.uint(ch))
}

// FillTriangle draws a filled triangle.
func (c *Canvas) FillTriangle(x0, y0, x1, y1, x2, y2 int, ch rune) {
	C.eye_canvas_fill_triangle(c.ptr, C.int(x0), C.int(y0), C.int(x1), C.int(y1), C.int(x2), C.int(y2), C.uint(ch))
}

// PutString draws a string starting at (x, y).
func (c *Canvas) PutString(x, y int, s string) {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))
	C.eye_canvas_put_str(c.ptr, C.int(x), C.int(y), cstr)
}

// Render returns the canvas as a string with ANSI color codes.
func (c *Canvas) Render() string {
	cstr := C.eye_canvas_render(c.ptr)
	defer C.eye_string_free(cstr)
	return C.GoString(cstr)
}

// RenderPlain returns the canvas as a plain string (no ANSI).
func (c *Canvas) RenderPlain() string {
	cstr := C.eye_canvas_render_plain(c.ptr)
	defer C.eye_string_free(cstr)
	return C.GoString(cstr)
}
