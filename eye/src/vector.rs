//! Vector operations with SIMD-friendly patterns

use crate::{EyeError, EyeResult, Scalar};
use rayon::prelude::*;

/// N-dimensional vector
#[derive(Debug, Clone)]
pub struct Vector {
    data: Vec<Scalar>,
}

impl Vector {
    pub fn new(data: &[Scalar]) -> Self {
        Self {
            data: data.to_vec(),
        }
    }

    pub fn from_data(data: Vec<Scalar>) -> Self {
        Self { data }
    }

    pub fn zeros(n: usize) -> Self {
        Self { data: vec![0.0; n] }
    }

    pub fn ones(n: usize) -> Self {
        Self { data: vec![1.0; n] }
    }

    #[inline]
    pub fn len(&self) -> usize {
        self.data.len()
    }

    #[inline]
    pub fn is_empty(&self) -> bool {
        self.data.is_empty()
    }

    #[inline]
    pub fn get(&self, i: usize) -> Scalar {
        self.data[i]
    }

    #[inline]
    pub fn set(&mut self, i: usize, value: Scalar) {
        self.data[i] = value;
    }

    pub fn data(&self) -> &[Scalar] {
        &self.data
    }

    pub fn data_mut(&mut self) -> &mut [Scalar] {
        &mut self.data
    }

    /// Dot product
    pub fn dot(&self, other: &Vector) -> Scalar {
        self.data
            .par_iter()
            .zip(other.data.par_iter())
            .map(|(a, b)| a * b)
            .sum()
    }

    /// Euclidean norm (L2)
    pub fn norm(&self) -> Scalar {
        self.dot(self).sqrt()
    }

    /// L1 norm (Manhattan)
    pub fn norm_l1(&self) -> Scalar {
        self.data.par_iter().map(|x| x.abs()).sum()
    }

    /// L-infinity norm (max)
    pub fn norm_inf(&self) -> Scalar {
        self.data
            .par_iter()
            .map(|x| x.abs())
            .reduce(|| 0.0, |a, b| a.max(b))
    }

    /// Normalize to unit vector
    pub fn normalize(&self) -> EyeResult<Vector> {
        let n = self.norm();
        if n < 1e-15 {
            return Err(EyeError::InvalidInput("cannot normalize zero vector".into()));
        }
        Ok(self.scale(1.0 / n))
    }

    /// Vector addition
    pub fn add(&self, other: &Vector) -> EyeResult<Vector> {
        if self.len() != other.len() {
            return Err(EyeError::DimensionMismatch {
                expected: self.len(),
                got: other.len(),
            });
        }
        let data: Vec<Scalar> = self
            .data
            .par_iter()
            .zip(other.data.par_iter())
            .map(|(a, b)| a + b)
            .collect();
        Ok(Vector { data })
    }

    /// Vector subtraction
    pub fn sub(&self, other: &Vector) -> EyeResult<Vector> {
        if self.len() != other.len() {
            return Err(EyeError::DimensionMismatch {
                expected: self.len(),
                got: other.len(),
            });
        }
        let data: Vec<Scalar> = self
            .data
            .par_iter()
            .zip(other.data.par_iter())
            .map(|(a, b)| a - b)
            .collect();
        Ok(Vector { data })
    }

    /// Scalar multiplication
    pub fn scale(&self, scalar: Scalar) -> Vector {
        let data: Vec<Scalar> = self.data.par_iter().map(|a| a * scalar).collect();
        Vector { data }
    }

    /// Element-wise multiplication
    pub fn hadamard(&self, other: &Vector) -> EyeResult<Vector> {
        if self.len() != other.len() {
            return Err(EyeError::DimensionMismatch {
                expected: self.len(),
                got: other.len(),
            });
        }
        let data: Vec<Scalar> = self
            .data
            .par_iter()
            .zip(other.data.par_iter())
            .map(|(a, b)| a * b)
            .collect();
        Ok(Vector { data })
    }

    /// Apply function to each element
    pub fn map<F>(&self, f: F) -> Vector
    where
        F: Fn(Scalar) -> Scalar + Sync + Send,
    {
        let data: Vec<Scalar> = self.data.par_iter().map(|&x| f(x)).collect();
        Vector { data }
    }

    /// Sum of all elements
    pub fn sum(&self) -> Scalar {
        self.data.par_iter().sum()
    }

    /// Mean of all elements
    pub fn mean(&self) -> Scalar {
        self.sum() / self.len() as Scalar
    }

    /// Variance
    pub fn variance(&self) -> Scalar {
        let mean = self.mean();
        self.data.par_iter().map(|x| (x - mean).powi(2)).sum::<Scalar>() / self.len() as Scalar
    }

    /// Standard deviation
    pub fn std_dev(&self) -> Scalar {
        self.variance().sqrt()
    }

    /// Cross product (3D only)
    pub fn cross(&self, other: &Vector) -> EyeResult<Vector> {
        if self.len() != 3 || other.len() != 3 {
            return Err(EyeError::InvalidInput("cross product requires 3D vectors".into()));
        }
        Ok(Vector::new(&[
            self.data[1] * other.data[2] - self.data[2] * other.data[1],
            self.data[2] * other.data[0] - self.data[0] * other.data[2],
            self.data[0] * other.data[1] - self.data[1] * other.data[0],
        ]))
    }

    /// Angle between vectors (radians)
    pub fn angle(&self, other: &Vector) -> Scalar {
        let dot = self.dot(other);
        let norms = self.norm() * other.norm();
        if norms < 1e-15 {
            return 0.0;
        }
        (dot / norms).clamp(-1.0, 1.0).acos()
    }

    /// Project self onto other
    pub fn project(&self, other: &Vector) -> EyeResult<Vector> {
        let dot = self.dot(other);
        let other_norm_sq = other.dot(other);
        if other_norm_sq < 1e-15 {
            return Err(EyeError::InvalidInput("cannot project onto zero vector".into()));
        }
        Ok(other.scale(dot / other_norm_sq))
    }

    /// Linear interpolation
    pub fn lerp(&self, other: &Vector, t: Scalar) -> EyeResult<Vector> {
        self.scale(1.0 - t).add(&other.scale(t))
    }
}

/// 3D Vector (specialized for geometry)
#[derive(Debug, Clone, Copy)]
pub struct Vec3 {
    pub x: Scalar,
    pub y: Scalar,
    pub z: Scalar,
}

impl Vec3 {
    pub const ZERO: Vec3 = Vec3 { x: 0.0, y: 0.0, z: 0.0 };
    pub const ONE: Vec3 = Vec3 { x: 1.0, y: 1.0, z: 1.0 };
    pub const X: Vec3 = Vec3 { x: 1.0, y: 0.0, z: 0.0 };
    pub const Y: Vec3 = Vec3 { x: 0.0, y: 1.0, z: 0.0 };
    pub const Z: Vec3 = Vec3 { x: 0.0, y: 0.0, z: 1.0 };

    #[inline]
    pub fn new(x: Scalar, y: Scalar, z: Scalar) -> Self {
        Self { x, y, z }
    }

    #[inline]
    pub fn dot(&self, other: Vec3) -> Scalar {
        self.x * other.x + self.y * other.y + self.z * other.z
    }

    #[inline]
    pub fn cross(&self, other: Vec3) -> Vec3 {
        Vec3 {
            x: self.y * other.z - self.z * other.y,
            y: self.z * other.x - self.x * other.z,
            z: self.x * other.y - self.y * other.x,
        }
    }

    #[inline]
    pub fn norm(&self) -> Scalar {
        (self.x * self.x + self.y * self.y + self.z * self.z).sqrt()
    }

    #[inline]
    pub fn norm_sq(&self) -> Scalar {
        self.x * self.x + self.y * self.y + self.z * self.z
    }

    pub fn normalize(&self) -> Vec3 {
        let n = self.norm();
        if n < 1e-15 {
            return Vec3::ZERO;
        }
        self.scale(1.0 / n)
    }

    #[inline]
    pub fn add(&self, other: Vec3) -> Vec3 {
        Vec3 {
            x: self.x + other.x,
            y: self.y + other.y,
            z: self.z + other.z,
        }
    }

    #[inline]
    pub fn sub(&self, other: Vec3) -> Vec3 {
        Vec3 {
            x: self.x - other.x,
            y: self.y - other.y,
            z: self.z - other.z,
        }
    }

    #[inline]
    pub fn scale(&self, s: Scalar) -> Vec3 {
        Vec3 {
            x: self.x * s,
            y: self.y * s,
            z: self.z * s,
        }
    }

    #[inline]
    pub fn neg(&self) -> Vec3 {
        Vec3 {
            x: -self.x,
            y: -self.y,
            z: -self.z,
        }
    }

    pub fn lerp(&self, other: Vec3, t: Scalar) -> Vec3 {
        Vec3 {
            x: self.x + (other.x - self.x) * t,
            y: self.y + (other.y - self.y) * t,
            z: self.z + (other.z - self.z) * t,
        }
    }

    pub fn reflect(&self, normal: Vec3) -> Vec3 {
        self.sub(normal.scale(2.0 * self.dot(normal)))
    }

    pub fn to_array(&self) -> [Scalar; 3] {
        [self.x, self.y, self.z]
    }

    pub fn from_array(arr: [Scalar; 3]) -> Self {
        Self { x: arr[0], y: arr[1], z: arr[2] }
    }
}

/// 2D Vector
#[derive(Debug, Clone, Copy)]
pub struct Vec2 {
    pub x: Scalar,
    pub y: Scalar,
}

impl Vec2 {
    pub const ZERO: Vec2 = Vec2 { x: 0.0, y: 0.0 };
    pub const ONE: Vec2 = Vec2 { x: 1.0, y: 1.0 };

    #[inline]
    pub fn new(x: Scalar, y: Scalar) -> Self {
        Self { x, y }
    }

    #[inline]
    pub fn dot(&self, other: Vec2) -> Scalar {
        self.x * other.x + self.y * other.y
    }

    /// 2D cross product (returns scalar - the z component of 3D cross)
    #[inline]
    pub fn cross(&self, other: Vec2) -> Scalar {
        self.x * other.y - self.y * other.x
    }

    #[inline]
    pub fn norm(&self) -> Scalar {
        (self.x * self.x + self.y * self.y).sqrt()
    }

    pub fn normalize(&self) -> Vec2 {
        let n = self.norm();
        if n < 1e-15 {
            return Vec2::ZERO;
        }
        Vec2 { x: self.x / n, y: self.y / n }
    }

    #[inline]
    pub fn add(&self, other: Vec2) -> Vec2 {
        Vec2 { x: self.x + other.x, y: self.y + other.y }
    }

    #[inline]
    pub fn sub(&self, other: Vec2) -> Vec2 {
        Vec2 { x: self.x - other.x, y: self.y - other.y }
    }

    #[inline]
    pub fn scale(&self, s: Scalar) -> Vec2 {
        Vec2 { x: self.x * s, y: self.y * s }
    }

    /// Perpendicular vector (rotate 90 degrees)
    pub fn perp(&self) -> Vec2 {
        Vec2 { x: -self.y, y: self.x }
    }

    pub fn lerp(&self, other: Vec2, t: Scalar) -> Vec2 {
        Vec2 {
            x: self.x + (other.x - self.x) * t,
            y: self.y + (other.y - self.y) * t,
        }
    }
}

/// 4D Vector (for homogeneous coordinates and quaternions)
#[derive(Debug, Clone, Copy)]
pub struct Vec4 {
    pub x: Scalar,
    pub y: Scalar,
    pub z: Scalar,
    pub w: Scalar,
}

impl Vec4 {
    #[inline]
    pub fn new(x: Scalar, y: Scalar, z: Scalar, w: Scalar) -> Self {
        Self { x, y, z, w }
    }

    #[inline]
    pub fn from_vec3(v: Vec3, w: Scalar) -> Self {
        Self { x: v.x, y: v.y, z: v.z, w }
    }

    pub fn to_vec3(&self) -> Vec3 {
        if self.w.abs() < 1e-15 {
            Vec3::new(self.x, self.y, self.z)
        } else {
            Vec3::new(self.x / self.w, self.y / self.w, self.z / self.w)
        }
    }

    #[inline]
    pub fn dot(&self, other: Vec4) -> Scalar {
        self.x * other.x + self.y * other.y + self.z * other.z + self.w * other.w
    }
}
