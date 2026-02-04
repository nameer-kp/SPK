//! FFI bindings for Go/C interop
//!
//! Provides a C-compatible interface for calling Eye functions from Go.

use crate::{curve, geometry, matrix::Matrix, tui, vector::Vector, Vec3};
use std::ffi::{CStr, CString};
use std::os::raw::{c_char, c_double, c_int, c_uint};
use std::ptr;
use std::slice;

/// Opaque handle for Matrix
pub struct EyeMatrix(Matrix);

/// Opaque handle for Vector
pub struct EyeVector(Vector);

/// Opaque handle for Canvas
pub struct EyeCanvas(tui::Canvas);

/// Result type for FFI
#[repr(C)]
pub struct EyeResult {
    pub success: c_int,
    pub error_msg: *mut c_char,
}

impl EyeResult {
    fn ok() -> Self {
        EyeResult {
            success: 1,
            error_msg: ptr::null_mut(),
        }
    }

    fn err(msg: &str) -> Self {
        let c_str = CString::new(msg).unwrap_or_default();
        EyeResult {
            success: 0,
            error_msg: c_str.into_raw(),
        }
    }
}

// ============ Matrix Functions ============

/// Create a new matrix filled with zeros
#[no_mangle]
pub extern "C" fn eye_matrix_zeros(rows: c_uint, cols: c_uint) -> *mut EyeMatrix {
    Box::into_raw(Box::new(EyeMatrix(Matrix::zeros(rows as usize, cols as usize))))
}

/// Create a new matrix filled with ones
#[no_mangle]
pub extern "C" fn eye_matrix_ones(rows: c_uint, cols: c_uint) -> *mut EyeMatrix {
    Box::into_raw(Box::new(EyeMatrix(Matrix::ones(rows as usize, cols as usize))))
}

/// Create an identity matrix
#[no_mangle]
pub extern "C" fn eye_matrix_identity(n: c_uint) -> *mut EyeMatrix {
    Box::into_raw(Box::new(EyeMatrix(Matrix::identity(n as usize))))
}

/// Create a matrix from flat data
#[no_mangle]
pub extern "C" fn eye_matrix_from_data(
    rows: c_uint,
    cols: c_uint,
    data: *const c_double,
    len: c_uint,
) -> *mut EyeMatrix {
    if data.is_null() || len != rows * cols {
        return ptr::null_mut();
    }

    let data_slice = unsafe { slice::from_raw_parts(data, len as usize) };
    match Matrix::from_data(rows as usize, cols as usize, data_slice.to_vec()) {
        Ok(m) => Box::into_raw(Box::new(EyeMatrix(m))),
        Err(_) => ptr::null_mut(),
    }
}

/// Free a matrix
#[no_mangle]
pub extern "C" fn eye_matrix_free(m: *mut EyeMatrix) {
    if !m.is_null() {
        unsafe {
            drop(Box::from_raw(m));
        }
    }
}

/// Get matrix dimensions
#[no_mangle]
pub extern "C" fn eye_matrix_rows(m: *const EyeMatrix) -> c_uint {
    if m.is_null() {
        return 0;
    }
    unsafe { (*m).0.rows() as c_uint }
}

#[no_mangle]
pub extern "C" fn eye_matrix_cols(m: *const EyeMatrix) -> c_uint {
    if m.is_null() {
        return 0;
    }
    unsafe { (*m).0.cols() as c_uint }
}

/// Get matrix element
#[no_mangle]
pub extern "C" fn eye_matrix_get(m: *const EyeMatrix, row: c_uint, col: c_uint) -> c_double {
    if m.is_null() {
        return 0.0;
    }
    unsafe { (*m).0.get(row as usize, col as usize) }
}

/// Set matrix element
#[no_mangle]
pub extern "C" fn eye_matrix_set(m: *mut EyeMatrix, row: c_uint, col: c_uint, value: c_double) {
    if m.is_null() {
        return;
    }
    unsafe {
        (*m).0.set(row as usize, col as usize, value);
    }
}

/// Copy matrix data to buffer
#[no_mangle]
pub extern "C" fn eye_matrix_copy_data(m: *const EyeMatrix, buf: *mut c_double, len: c_uint) -> c_int {
    if m.is_null() || buf.is_null() {
        return 0;
    }

    let matrix = unsafe { &(*m).0 };
    let data = matrix.data();
    let copy_len = (len as usize).min(data.len());

    unsafe {
        ptr::copy_nonoverlapping(data.as_ptr(), buf, copy_len);
    }
    copy_len as c_int
}

/// Matrix multiplication
#[no_mangle]
pub extern "C" fn eye_matrix_mul(a: *const EyeMatrix, b: *const EyeMatrix) -> *mut EyeMatrix {
    if a.is_null() || b.is_null() {
        return ptr::null_mut();
    }

    let result = unsafe { (*a).0.mul(&(*b).0) };
    match result {
        Ok(m) => Box::into_raw(Box::new(EyeMatrix(m))),
        Err(_) => ptr::null_mut(),
    }
}

/// Strassen matrix multiplication
#[no_mangle]
pub extern "C" fn eye_matrix_mul_strassen(a: *const EyeMatrix, b: *const EyeMatrix) -> *mut EyeMatrix {
    if a.is_null() || b.is_null() {
        return ptr::null_mut();
    }

    let result = unsafe { (*a).0.mul_strassen(&(*b).0) };
    match result {
        Ok(m) => Box::into_raw(Box::new(EyeMatrix(m))),
        Err(_) => ptr::null_mut(),
    }
}

/// Matrix addition
#[no_mangle]
pub extern "C" fn eye_matrix_add(a: *const EyeMatrix, b: *const EyeMatrix) -> *mut EyeMatrix {
    if a.is_null() || b.is_null() {
        return ptr::null_mut();
    }

    let result = unsafe { (*a).0.add(&(*b).0) };
    match result {
        Ok(m) => Box::into_raw(Box::new(EyeMatrix(m))),
        Err(_) => ptr::null_mut(),
    }
}

/// Matrix transpose
#[no_mangle]
pub extern "C" fn eye_matrix_transpose(m: *const EyeMatrix) -> *mut EyeMatrix {
    if m.is_null() {
        return ptr::null_mut();
    }
    Box::into_raw(Box::new(EyeMatrix(unsafe { (*m).0.transpose() })))
}

/// Matrix determinant
#[no_mangle]
pub extern "C" fn eye_matrix_det(m: *const EyeMatrix, result: *mut c_double) -> c_int {
    if m.is_null() || result.is_null() {
        return 0;
    }

    match unsafe { (*m).0.det() } {
        Ok(det) => {
            unsafe { *result = det };
            1
        }
        Err(_) => 0,
    }
}

/// Matrix inverse
#[no_mangle]
pub extern "C" fn eye_matrix_inverse(m: *const EyeMatrix) -> *mut EyeMatrix {
    if m.is_null() {
        return ptr::null_mut();
    }

    match unsafe { (*m).0.inverse() } {
        Ok(inv) => Box::into_raw(Box::new(EyeMatrix(inv))),
        Err(_) => ptr::null_mut(),
    }
}

// ============ Vector Functions ============

/// Create a vector from data
#[no_mangle]
pub extern "C" fn eye_vector_from_data(data: *const c_double, len: c_uint) -> *mut EyeVector {
    if data.is_null() {
        return ptr::null_mut();
    }

    let data_slice = unsafe { slice::from_raw_parts(data, len as usize) };
    Box::into_raw(Box::new(EyeVector(Vector::new(data_slice))))
}

/// Create a zero vector
#[no_mangle]
pub extern "C" fn eye_vector_zeros(n: c_uint) -> *mut EyeVector {
    Box::into_raw(Box::new(EyeVector(Vector::zeros(n as usize))))
}

/// Free a vector
#[no_mangle]
pub extern "C" fn eye_vector_free(v: *mut EyeVector) {
    if !v.is_null() {
        unsafe {
            drop(Box::from_raw(v));
        }
    }
}

/// Vector length
#[no_mangle]
pub extern "C" fn eye_vector_len(v: *const EyeVector) -> c_uint {
    if v.is_null() {
        return 0;
    }
    unsafe { (*v).0.len() as c_uint }
}

/// Vector get element
#[no_mangle]
pub extern "C" fn eye_vector_get(v: *const EyeVector, i: c_uint) -> c_double {
    if v.is_null() {
        return 0.0;
    }
    unsafe { (*v).0.get(i as usize) }
}

/// Vector set element
#[no_mangle]
pub extern "C" fn eye_vector_set(v: *mut EyeVector, i: c_uint, value: c_double) {
    if v.is_null() {
        return;
    }
    unsafe {
        (*v).0.set(i as usize, value);
    }
}

/// Dot product
#[no_mangle]
pub extern "C" fn eye_vector_dot(a: *const EyeVector, b: *const EyeVector) -> c_double {
    if a.is_null() || b.is_null() {
        return 0.0;
    }
    unsafe { (*a).0.dot(&(*b).0) }
}

/// Vector norm
#[no_mangle]
pub extern "C" fn eye_vector_norm(v: *const EyeVector) -> c_double {
    if v.is_null() {
        return 0.0;
    }
    unsafe { (*v).0.norm() }
}

/// Cross product (3D)
#[no_mangle]
pub extern "C" fn eye_vector_cross(a: *const EyeVector, b: *const EyeVector) -> *mut EyeVector {
    if a.is_null() || b.is_null() {
        return ptr::null_mut();
    }

    match unsafe { (*a).0.cross(&(*b).0) } {
        Ok(v) => Box::into_raw(Box::new(EyeVector(v))),
        Err(_) => ptr::null_mut(),
    }
}

// ============ Geometry Functions ============

/// Triangle area
#[no_mangle]
pub extern "C" fn eye_triangle_area(
    v0x: c_double, v0y: c_double, v0z: c_double,
    v1x: c_double, v1y: c_double, v1z: c_double,
    v2x: c_double, v2y: c_double, v2z: c_double,
) -> c_double {
    let tri = geometry::Triangle::new(
        Vec3::new(v0x, v0y, v0z),
        Vec3::new(v1x, v1y, v1z),
        Vec3::new(v2x, v2y, v2z),
    );
    tri.area()
}

/// Ray-triangle intersection
#[no_mangle]
pub extern "C" fn eye_triangle_intersect_ray(
    v0x: c_double, v0y: c_double, v0z: c_double,
    v1x: c_double, v1y: c_double, v1z: c_double,
    v2x: c_double, v2y: c_double, v2z: c_double,
    ox: c_double, oy: c_double, oz: c_double,
    dx: c_double, dy: c_double, dz: c_double,
    t_out: *mut c_double,
    u_out: *mut c_double,
    v_out: *mut c_double,
) -> c_int {
    let tri = geometry::Triangle::new(
        Vec3::new(v0x, v0y, v0z),
        Vec3::new(v1x, v1y, v1z),
        Vec3::new(v2x, v2y, v2z),
    );

    match tri.intersect_ray(Vec3::new(ox, oy, oz), Vec3::new(dx, dy, dz)) {
        Some((t, u, v)) => {
            if !t_out.is_null() {
                unsafe { *t_out = t };
            }
            if !u_out.is_null() {
                unsafe { *u_out = u };
            }
            if !v_out.is_null() {
                unsafe { *v_out = v };
            }
            1
        }
        None => 0,
    }
}

// ============ Curve Functions ============

/// Numerical integration (Simpson's rule)
#[no_mangle]
pub extern "C" fn eye_integrate_simpson(
    func: extern "C" fn(c_double) -> c_double,
    a: c_double,
    b: c_double,
    n: c_uint,
) -> c_double {
    curve::integrate(|x| func(x), a, b, n as usize)
}

/// Gaussian quadrature integration
#[no_mangle]
pub extern "C" fn eye_integrate_gauss(
    func: extern "C" fn(c_double) -> c_double,
    a: c_double,
    b: c_double,
    order: c_uint,
) -> c_double {
    curve::integrate_gauss(|x| func(x), a, b, order as usize)
}

/// Romberg integration
#[no_mangle]
pub extern "C" fn eye_integrate_romberg(
    func: extern "C" fn(c_double) -> c_double,
    a: c_double,
    b: c_double,
    max_iter: c_uint,
    tol: c_double,
) -> c_double {
    curve::integrate_romberg(|x| func(x), a, b, max_iter as usize, tol)
}

/// Numerical derivative
#[no_mangle]
pub extern "C" fn eye_derivative(
    func: extern "C" fn(c_double) -> c_double,
    x: c_double,
    h: c_double,
) -> c_double {
    curve::derivative(|t| func(t), x, h)
}

/// Find root (Newton-Raphson)
#[no_mangle]
pub extern "C" fn eye_find_root(
    f: extern "C" fn(c_double) -> c_double,
    df: extern "C" fn(c_double) -> c_double,
    x0: c_double,
    tol: c_double,
    max_iter: c_uint,
    result: *mut c_double,
) -> c_int {
    match curve::find_root(|x| f(x), |x| df(x), x0, tol, max_iter as usize) {
        Ok(root) => {
            if !result.is_null() {
                unsafe { *result = root };
            }
            1
        }
        Err(_) => 0,
    }
}

// ============ TUI Canvas Functions ============

/// Create a canvas
#[no_mangle]
pub extern "C" fn eye_canvas_new(width: c_uint, height: c_uint) -> *mut EyeCanvas {
    Box::into_raw(Box::new(EyeCanvas(tui::Canvas::new(width as usize, height as usize))))
}

/// Free a canvas
#[no_mangle]
pub extern "C" fn eye_canvas_free(c: *mut EyeCanvas) {
    if !c.is_null() {
        unsafe {
            drop(Box::from_raw(c));
        }
    }
}

/// Clear canvas
#[no_mangle]
pub extern "C" fn eye_canvas_clear(c: *mut EyeCanvas) {
    if c.is_null() {
        return;
    }
    unsafe {
        (*c).0.clear();
    }
}

/// Put character
#[no_mangle]
pub extern "C" fn eye_canvas_put(c: *mut EyeCanvas, x: c_int, y: c_int, ch: c_uint) {
    if c.is_null() {
        return;
    }
    if let Some(chr) = char::from_u32(ch) {
        unsafe {
            (*c).0.put(x, y, chr);
        }
    }
}

/// Draw line
#[no_mangle]
pub extern "C" fn eye_canvas_line(
    c: *mut EyeCanvas,
    x0: c_int, y0: c_int,
    x1: c_int, y1: c_int,
    ch: c_uint,
) {
    if c.is_null() {
        return;
    }
    if let Some(chr) = char::from_u32(ch) {
        unsafe {
            (*c).0.line(x0, y0, x1, y1, chr);
        }
    }
}

/// Draw rectangle
#[no_mangle]
pub extern "C" fn eye_canvas_rect(
    c: *mut EyeCanvas,
    x: c_int, y: c_int,
    w: c_int, h: c_int,
    ch: c_uint,
) {
    if c.is_null() {
        return;
    }
    if let Some(chr) = char::from_u32(ch) {
        unsafe {
            (*c).0.rect(x, y, w, h, chr);
        }
    }
}

/// Draw box
#[no_mangle]
pub extern "C" fn eye_canvas_box(c: *mut EyeCanvas, x: c_int, y: c_int, w: c_int, h: c_int) {
    if c.is_null() {
        return;
    }
    unsafe {
        (*c).0.box_draw(x, y, w, h);
    }
}

/// Draw circle
#[no_mangle]
pub extern "C" fn eye_canvas_circle(c: *mut EyeCanvas, cx: c_int, cy: c_int, r: c_int, ch: c_uint) {
    if c.is_null() {
        return;
    }
    if let Some(chr) = char::from_u32(ch) {
        unsafe {
            (*c).0.circle(cx, cy, r, chr);
        }
    }
}

/// Fill circle
#[no_mangle]
pub extern "C" fn eye_canvas_fill_circle(c: *mut EyeCanvas, cx: c_int, cy: c_int, r: c_int, ch: c_uint) {
    if c.is_null() {
        return;
    }
    if let Some(chr) = char::from_u32(ch) {
        unsafe {
            (*c).0.fill_circle(cx, cy, r, chr);
        }
    }
}

/// Draw triangle
#[no_mangle]
pub extern "C" fn eye_canvas_triangle(
    c: *mut EyeCanvas,
    x0: c_int, y0: c_int,
    x1: c_int, y1: c_int,
    x2: c_int, y2: c_int,
    ch: c_uint,
) {
    if c.is_null() {
        return;
    }
    if let Some(chr) = char::from_u32(ch) {
        unsafe {
            (*c).0.triangle(x0, y0, x1, y1, x2, y2, chr);
        }
    }
}

/// Fill triangle
#[no_mangle]
pub extern "C" fn eye_canvas_fill_triangle(
    c: *mut EyeCanvas,
    x0: c_int, y0: c_int,
    x1: c_int, y1: c_int,
    x2: c_int, y2: c_int,
    ch: c_uint,
) {
    if c.is_null() {
        return;
    }
    if let Some(chr) = char::from_u32(ch) {
        unsafe {
            (*c).0.fill_triangle(x0, y0, x1, y1, x2, y2, chr);
        }
    }
}

/// Put string
#[no_mangle]
pub extern "C" fn eye_canvas_put_str(c: *mut EyeCanvas, x: c_int, y: c_int, s: *const c_char) {
    if c.is_null() || s.is_null() {
        return;
    }
    let cstr = unsafe { CStr::from_ptr(s) };
    if let Ok(str_val) = cstr.to_str() {
        unsafe {
            (*c).0.put_str(x, y, str_val);
        }
    }
}

/// Render canvas to string (plain, no ANSI)
#[no_mangle]
pub extern "C" fn eye_canvas_render_plain(c: *const EyeCanvas) -> *mut c_char {
    if c.is_null() {
        return ptr::null_mut();
    }
    let rendered = unsafe { (*c).0.render_plain() };
    CString::new(rendered).map(|s| s.into_raw()).unwrap_or(ptr::null_mut())
}

/// Render canvas to string (with ANSI)
#[no_mangle]
pub extern "C" fn eye_canvas_render(c: *const EyeCanvas) -> *mut c_char {
    if c.is_null() {
        return ptr::null_mut();
    }
    let rendered = unsafe { (*c).0.render() };
    CString::new(rendered).map(|s| s.into_raw()).unwrap_or(ptr::null_mut())
}

/// Free a string returned by Eye
#[no_mangle]
pub extern "C" fn eye_string_free(s: *mut c_char) {
    if !s.is_null() {
        unsafe {
            drop(CString::from_raw(s));
        }
    }
}

/// Free result error message
#[no_mangle]
pub extern "C" fn eye_result_free(r: *mut EyeResult) {
    if r.is_null() {
        return;
    }
    unsafe {
        if !(*r).error_msg.is_null() {
            drop(CString::from_raw((*r).error_msg));
        }
    }
}
