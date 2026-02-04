//! Matrix operations with parallel computation support
//!
//! Optimized for large-scale matrix operations using:
//! - Cache-friendly memory layout (row-major)
//! - Parallel computation via rayon
//! - SIMD-friendly patterns

use crate::{EyeError, EyeResult, Scalar};
use rayon::prelude::*;

/// Dense matrix with row-major storage
#[derive(Debug, Clone)]
pub struct Matrix {
    data: Vec<Scalar>,
    rows: usize,
    cols: usize,
}

impl Matrix {
    /// Create a new matrix filled with zeros
    pub fn zeros(rows: usize, cols: usize) -> Self {
        Self {
            data: vec![0.0; rows * cols],
            rows,
            cols,
        }
    }

    /// Create a new matrix filled with ones
    pub fn ones(rows: usize, cols: usize) -> Self {
        Self {
            data: vec![1.0; rows * cols],
            rows,
            cols,
        }
    }

    /// Create an identity matrix
    pub fn identity(n: usize) -> Self {
        let mut m = Self::zeros(n, n);
        for i in 0..n {
            m.set(i, i, 1.0);
        }
        m
    }

    /// Create from row data
    pub fn from_rows(rows: &[&[Scalar]]) -> Self {
        let n_rows = rows.len();
        let n_cols = rows.first().map(|r| r.len()).unwrap_or(0);
        let mut data = Vec::with_capacity(n_rows * n_cols);
        for row in rows {
            data.extend_from_slice(row);
        }
        Self {
            data,
            rows: n_rows,
            cols: n_cols,
        }
    }

    /// Create from flat data
    pub fn from_data(rows: usize, cols: usize, data: Vec<Scalar>) -> EyeResult<Self> {
        if data.len() != rows * cols {
            return Err(EyeError::DimensionMismatch {
                expected: rows * cols,
                got: data.len(),
            });
        }
        Ok(Self { data, rows, cols })
    }

    #[inline]
    pub fn rows(&self) -> usize {
        self.rows
    }

    #[inline]
    pub fn cols(&self) -> usize {
        self.cols
    }

    #[inline]
    pub fn get(&self, row: usize, col: usize) -> Scalar {
        self.data[row * self.cols + col]
    }

    #[inline]
    pub fn set(&mut self, row: usize, col: usize, value: Scalar) {
        self.data[row * self.cols + col] = value;
    }

    /// Get raw data slice
    pub fn data(&self) -> &[Scalar] {
        &self.data
    }

    /// Get mutable raw data slice
    pub fn data_mut(&mut self) -> &mut [Scalar] {
        &mut self.data
    }

    /// Matrix addition
    pub fn add(&self, other: &Matrix) -> EyeResult<Matrix> {
        if self.rows != other.rows || self.cols != other.cols {
            return Err(EyeError::DimensionMismatch {
                expected: self.rows * self.cols,
                got: other.rows * other.cols,
            });
        }

        let data: Vec<Scalar> = self
            .data
            .par_iter()
            .zip(other.data.par_iter())
            .map(|(a, b)| a + b)
            .collect();

        Ok(Matrix {
            data,
            rows: self.rows,
            cols: self.cols,
        })
    }

    /// Matrix subtraction
    pub fn sub(&self, other: &Matrix) -> EyeResult<Matrix> {
        if self.rows != other.rows || self.cols != other.cols {
            return Err(EyeError::DimensionMismatch {
                expected: self.rows * self.cols,
                got: other.rows * other.cols,
            });
        }

        let data: Vec<Scalar> = self
            .data
            .par_iter()
            .zip(other.data.par_iter())
            .map(|(a, b)| a - b)
            .collect();

        Ok(Matrix {
            data,
            rows: self.rows,
            cols: self.cols,
        })
    }

    /// Scalar multiplication
    pub fn scale(&self, scalar: Scalar) -> Matrix {
        let data: Vec<Scalar> = self.data.par_iter().map(|a| a * scalar).collect();
        Matrix {
            data,
            rows: self.rows,
            cols: self.cols,
        }
    }

    /// Matrix multiplication (parallel)
    pub fn mul(&self, other: &Matrix) -> EyeResult<Matrix> {
        if self.cols != other.rows {
            return Err(EyeError::DimensionMismatch {
                expected: self.cols,
                got: other.rows,
            });
        }

        let mut result = Matrix::zeros(self.rows, other.cols);

        // Parallel over rows
        result
            .data
            .par_chunks_mut(other.cols)
            .enumerate()
            .for_each(|(i, row_chunk)| {
                for k in 0..self.cols {
                    let a_ik = self.get(i, k);
                    for j in 0..other.cols {
                        row_chunk[j] += a_ik * other.get(k, j);
                    }
                }
            });

        Ok(result)
    }

    /// Transpose
    pub fn transpose(&self) -> Matrix {
        let mut result = Matrix::zeros(self.cols, self.rows);
        for i in 0..self.rows {
            for j in 0..self.cols {
                result.set(j, i, self.get(i, j));
            }
        }
        result
    }

    /// Trace (sum of diagonal elements)
    pub fn trace(&self) -> Scalar {
        let n = self.rows.min(self.cols);
        (0..n).map(|i| self.get(i, i)).sum()
    }

    /// Frobenius norm
    pub fn frobenius_norm(&self) -> Scalar {
        self.data.par_iter().map(|x| x * x).sum::<Scalar>().sqrt()
    }

    /// Element-wise multiplication (Hadamard product)
    pub fn hadamard(&self, other: &Matrix) -> EyeResult<Matrix> {
        if self.rows != other.rows || self.cols != other.cols {
            return Err(EyeError::DimensionMismatch {
                expected: self.rows * self.cols,
                got: other.rows * other.cols,
            });
        }

        let data: Vec<Scalar> = self
            .data
            .par_iter()
            .zip(other.data.par_iter())
            .map(|(a, b)| a * b)
            .collect();

        Ok(Matrix {
            data,
            rows: self.rows,
            cols: self.cols,
        })
    }

    /// Determinant (for small matrices)
    pub fn det(&self) -> EyeResult<Scalar> {
        if self.rows != self.cols {
            return Err(EyeError::InvalidInput("determinant requires square matrix".into()));
        }

        match self.rows {
            0 => Ok(1.0),
            1 => Ok(self.get(0, 0)),
            2 => Ok(self.get(0, 0) * self.get(1, 1) - self.get(0, 1) * self.get(1, 0)),
            3 => {
                let a = self.get(0, 0);
                let b = self.get(0, 1);
                let c = self.get(0, 2);
                let d = self.get(1, 0);
                let e = self.get(1, 1);
                let f = self.get(1, 2);
                let g = self.get(2, 0);
                let h = self.get(2, 1);
                let i = self.get(2, 2);
                Ok(a * (e * i - f * h) - b * (d * i - f * g) + c * (d * h - e * g))
            }
            _ => self.det_lu(),
        }
    }

    /// Determinant via LU decomposition for larger matrices
    fn det_lu(&self) -> EyeResult<Scalar> {
        let n = self.rows;
        let mut lu = self.clone();
        let mut det = 1.0;

        for k in 0..n {
            // Find pivot
            let mut max_val = lu.get(k, k).abs();
            let mut max_row = k;
            for i in (k + 1)..n {
                let val = lu.get(i, k).abs();
                if val > max_val {
                    max_val = val;
                    max_row = i;
                }
            }

            if max_val < 1e-15 {
                return Ok(0.0); // Singular matrix
            }

            // Swap rows if needed
            if max_row != k {
                for j in 0..n {
                    let tmp = lu.get(k, j);
                    lu.set(k, j, lu.get(max_row, j));
                    lu.set(max_row, j, tmp);
                }
                det = -det;
            }

            det *= lu.get(k, k);

            // Eliminate
            for i in (k + 1)..n {
                let factor = lu.get(i, k) / lu.get(k, k);
                for j in k..n {
                    let val = lu.get(i, j) - factor * lu.get(k, j);
                    lu.set(i, j, val);
                }
            }
        }

        Ok(det)
    }

    /// Inverse (Gauss-Jordan)
    pub fn inverse(&self) -> EyeResult<Matrix> {
        if self.rows != self.cols {
            return Err(EyeError::InvalidInput("inverse requires square matrix".into()));
        }

        let n = self.rows;
        let mut aug = Matrix::zeros(n, 2 * n);

        // Build augmented matrix [A | I]
        for i in 0..n {
            for j in 0..n {
                aug.set(i, j, self.get(i, j));
            }
            aug.set(i, n + i, 1.0);
        }

        // Gauss-Jordan elimination
        for k in 0..n {
            // Find pivot
            let mut max_val = aug.get(k, k).abs();
            let mut max_row = k;
            for i in (k + 1)..n {
                let val = aug.get(i, k).abs();
                if val > max_val {
                    max_val = val;
                    max_row = i;
                }
            }

            if max_val < 1e-15 {
                return Err(EyeError::SingularMatrix);
            }

            // Swap rows
            if max_row != k {
                for j in 0..(2 * n) {
                    let tmp = aug.get(k, j);
                    aug.set(k, j, aug.get(max_row, j));
                    aug.set(max_row, j, tmp);
                }
            }

            // Scale pivot row
            let pivot = aug.get(k, k);
            for j in 0..(2 * n) {
                aug.set(k, j, aug.get(k, j) / pivot);
            }

            // Eliminate column
            for i in 0..n {
                if i != k {
                    let factor = aug.get(i, k);
                    for j in 0..(2 * n) {
                        let val = aug.get(i, j) - factor * aug.get(k, j);
                        aug.set(i, j, val);
                    }
                }
            }
        }

        // Extract inverse
        let mut inv = Matrix::zeros(n, n);
        for i in 0..n {
            for j in 0..n {
                inv.set(i, j, aug.get(i, n + j));
            }
        }

        Ok(inv)
    }

    /// Solve linear system Ax = b
    pub fn solve(&self, b: &crate::Vector) -> EyeResult<crate::Vector> {
        let inv = self.inverse()?;
        let b_mat = Matrix::from_data(b.len(), 1, b.data().to_vec())?;
        let x_mat = inv.mul(&b_mat)?;
        Ok(crate::Vector::from_data(x_mat.data().to_vec()))
    }

    /// Apply element-wise function
    pub fn map<F>(&self, f: F) -> Matrix
    where
        F: Fn(Scalar) -> Scalar + Sync + Send,
    {
        let data: Vec<Scalar> = self.data.par_iter().map(|&x| f(x)).collect();
        Matrix {
            data,
            rows: self.rows,
            cols: self.cols,
        }
    }

    /// Pad matrix with zeros
    pub fn pad(&self, new_rows: usize, new_cols: usize) -> Matrix {
        let mut result = Matrix::zeros(new_rows, new_cols);
        for i in 0..self.rows {
            for j in 0..self.cols {
                result.set(i, j, self.get(i, j));
            }
        }
        result
    }

    /// Extract submatrix
    pub fn submatrix(&self, row_start: usize, col_start: usize, rows: usize, cols: usize) -> Matrix {
        let mut result = Matrix::zeros(rows, cols);
        for i in 0..rows {
            for j in 0..cols {
                result.set(i, j, self.get(row_start + i, col_start + j));
            }
        }
        result
    }

    /// Strassen's algorithm for large matrix multiplication
    /// O(n^2.807) instead of O(n^3)
    pub fn mul_strassen(&self, other: &Matrix) -> EyeResult<Matrix> {
        if self.cols != other.rows {
            return Err(EyeError::DimensionMismatch {
                expected: self.cols,
                got: other.rows,
            });
        }

        // For small matrices, use standard multiplication
        if self.rows <= 64 || self.cols <= 64 || other.cols <= 64 {
            return self.mul(other);
        }

        // Pad to power of 2
        let n = self.rows.max(self.cols).max(other.cols).next_power_of_two();
        let a = self.pad(n, n);
        let b = other.pad(n, n);

        let c = strassen_recursive(&a, &b, 64);

        // Extract result
        Ok(c.submatrix(0, 0, self.rows, other.cols))
    }

    /// Blocked matrix multiplication for cache efficiency
    pub fn mul_blocked(&self, other: &Matrix, block_size: usize) -> EyeResult<Matrix> {
        if self.cols != other.rows {
            return Err(EyeError::DimensionMismatch {
                expected: self.cols,
                got: other.rows,
            });
        }

        let mut result = Matrix::zeros(self.rows, other.cols);

        // Process in blocks
        for ii in (0..self.rows).step_by(block_size) {
            for jj in (0..other.cols).step_by(block_size) {
                for kk in (0..self.cols).step_by(block_size) {
                    let i_end = (ii + block_size).min(self.rows);
                    let j_end = (jj + block_size).min(other.cols);
                    let k_end = (kk + block_size).min(self.cols);

                    for i in ii..i_end {
                        for k in kk..k_end {
                            let a_ik = self.get(i, k);
                            for j in jj..j_end {
                                let val = result.get(i, j) + a_ik * other.get(k, j);
                                result.set(i, j, val);
                            }
                        }
                    }
                }
            }
        }

        Ok(result)
    }
}

fn strassen_recursive(a: &Matrix, b: &Matrix, threshold: usize) -> Matrix {
    let n = a.rows();

    if n <= threshold {
        return a.mul(b).unwrap();
    }

    let half = n / 2;

    // Split into quadrants
    let a11 = a.submatrix(0, 0, half, half);
    let a12 = a.submatrix(0, half, half, half);
    let a21 = a.submatrix(half, 0, half, half);
    let a22 = a.submatrix(half, half, half, half);

    let b11 = b.submatrix(0, 0, half, half);
    let b12 = b.submatrix(0, half, half, half);
    let b21 = b.submatrix(half, 0, half, half);
    let b22 = b.submatrix(half, half, half, half);

    // Compute the 7 Strassen products (in parallel)
    let (m1, m2) = rayon::join(
        || strassen_recursive(&a11.add(&a22).unwrap(), &b11.add(&b22).unwrap(), threshold),
        || strassen_recursive(&a21.add(&a22).unwrap(), &b11, threshold),
    );

    let (m3, m4) = rayon::join(
        || strassen_recursive(&a11, &b12.sub(&b22).unwrap(), threshold),
        || strassen_recursive(&a22, &b21.sub(&b11).unwrap(), threshold),
    );

    let (m5, (m6, m7)) = rayon::join(
        || strassen_recursive(&a11.add(&a12).unwrap(), &b22, threshold),
        || rayon::join(
            || strassen_recursive(&a21.sub(&a11).unwrap(), &b11.add(&b12).unwrap(), threshold),
            || strassen_recursive(&a12.sub(&a22).unwrap(), &b21.add(&b22).unwrap(), threshold),
        ),
    );

    // Combine results
    let c11 = m1.add(&m4).unwrap().sub(&m5).unwrap().add(&m7).unwrap();
    let c12 = m3.add(&m5).unwrap();
    let c21 = m2.add(&m4).unwrap();
    let c22 = m1.sub(&m2).unwrap().add(&m3).unwrap().add(&m6).unwrap();

    // Assemble result
    let mut result = Matrix::zeros(n, n);
    for i in 0..half {
        for j in 0..half {
            result.set(i, j, c11.get(i, j));
            result.set(i, half + j, c12.get(i, j));
            result.set(half + i, j, c21.get(i, j));
            result.set(half + i, half + j, c22.get(i, j));
        }
    }

    result
}
