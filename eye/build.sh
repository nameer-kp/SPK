#!/bin/bash
# Build script for Eye neural image engine

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Building Eye library..."

# Build release version
cargo build --release

echo "Eye library built successfully!"
echo "Static library: target/release/libeye.a"
echo "Dynamic library: target/release/libeye.dylib (macOS) or libeye.so (Linux)"

# Create header file for FFI
echo "Generating C header..."
cat > include/eye.h << 'EOF'
// Eye Neural Image Engine - C Header
// Auto-generated - do not edit

#ifndef EYE_H
#define EYE_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// Opaque types
typedef void* EyeMatrix;
typedef void* EyeVector;
typedef void* EyeCanvas;

// Matrix functions
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

// Vector functions
EyeVector eye_vector_from_data(double* data, unsigned int len);
EyeVector eye_vector_zeros(unsigned int n);
void eye_vector_free(EyeVector v);
unsigned int eye_vector_len(EyeVector v);
double eye_vector_get(EyeVector v, unsigned int i);
void eye_vector_set(EyeVector v, unsigned int i, double value);
double eye_vector_dot(EyeVector a, EyeVector b);
double eye_vector_norm(EyeVector v);
EyeVector eye_vector_cross(EyeVector a, EyeVector b);

// Geometry functions
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

// Curve functions (using function pointers)
typedef double (*EyeFunc)(double);
double eye_integrate_simpson(EyeFunc func, double a, double b, unsigned int n);
double eye_integrate_gauss(EyeFunc func, double a, double b, unsigned int order);
double eye_integrate_romberg(EyeFunc func, double a, double b, unsigned int max_iter, double tol);
double eye_derivative(EyeFunc func, double x, double h);
int eye_find_root(EyeFunc f, EyeFunc df, double x0, double tol, unsigned int max_iter, double* result);

// Canvas functions
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

// Memory management
void eye_string_free(char* s);

#ifdef __cplusplus
}
#endif

#endif // EYE_H
EOF

mkdir -p include

echo "Header generated: include/eye.h"
echo ""
echo "To use with Go, ensure the library is built and linked correctly."
echo "The Go package is at: pkg/eye/eye.go"
