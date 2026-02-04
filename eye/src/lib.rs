//! Eye - Neural Image Engine
//!
//! High-performance computational primitives for:
//! - Matrix operations (planet-scale parallel processing)
//! - Vector mathematics
//! - Geometric primitives (triangles, curves, surfaces)
//! - Numerical integration (area under curves)
//!
//! Designed for FFI with Go/Atem workflow engine.

pub mod matrix;
pub mod vector;
pub mod geometry;
pub mod curve;
pub mod ffi;
pub mod tui;

pub use matrix::*;
pub use vector::*;
pub use geometry::*;
pub use curve::*;

/// Precision type - f64 for "god precision"
pub type Scalar = f64;

/// Result type for Eye operations
pub type EyeResult<T> = Result<T, EyeError>;

/// Error types for Eye operations
#[derive(Debug, Clone)]
pub enum EyeError {
    DimensionMismatch { expected: usize, got: usize },
    SingularMatrix,
    InvalidInput(String),
    ComputationOverflow,
    NotConverged { iterations: usize, residual: f64 },
}

impl std::fmt::Display for EyeError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            EyeError::DimensionMismatch { expected, got } => {
                write!(f, "dimension mismatch: expected {}, got {}", expected, got)
            }
            EyeError::SingularMatrix => write!(f, "singular matrix"),
            EyeError::InvalidInput(msg) => write!(f, "invalid input: {}", msg),
            EyeError::ComputationOverflow => write!(f, "computation overflow"),
            EyeError::NotConverged { iterations, residual } => {
                write!(f, "not converged after {} iterations, residual: {}", iterations, residual)
            }
        }
    }
}

impl std::error::Error for EyeError {}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_matrix_multiply() {
        let a = Matrix::from_rows(&[
            &[1.0, 2.0],
            &[3.0, 4.0],
        ]);
        let b = Matrix::from_rows(&[
            &[5.0, 6.0],
            &[7.0, 8.0],
        ]);
        let c = a.mul(&b).unwrap();
        assert_eq!(c.get(0, 0), 19.0);
        assert_eq!(c.get(0, 1), 22.0);
        assert_eq!(c.get(1, 0), 43.0);
        assert_eq!(c.get(1, 1), 50.0);
    }

    #[test]
    fn test_vector_dot() {
        let a = Vector::new(&[1.0, 2.0, 3.0]);
        let b = Vector::new(&[4.0, 5.0, 6.0]);
        assert_eq!(a.dot(&b), 32.0);
    }

    #[test]
    fn test_triangle_area() {
        let t = Triangle::new(
            Vec3::new(0.0, 0.0, 0.0),
            Vec3::new(1.0, 0.0, 0.0),
            Vec3::new(0.0, 1.0, 0.0),
        );
        assert!((t.area() - 0.5).abs() < 1e-10);
    }

    #[test]
    fn test_curve_integration() {
        // Integrate x^2 from 0 to 1 = 1/3
        let result = curve::integrate(|x| x * x, 0.0, 1.0, 1000);
        assert!((result - 1.0/3.0).abs() < 1e-6);
    }
}
