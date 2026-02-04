//! Geometric primitives: triangles, rays, planes, meshes
//!
//! Building blocks for physics and graphics computations.

use crate::{Scalar, Vec2, Vec3};
use rayon::prelude::*;

/// Triangle primitive
#[derive(Debug, Clone, Copy)]
pub struct Triangle {
    pub v0: Vec3,
    pub v1: Vec3,
    pub v2: Vec3,
}

impl Triangle {
    pub fn new(v0: Vec3, v1: Vec3, v2: Vec3) -> Self {
        Self { v0, v1, v2 }
    }

    /// Compute triangle normal (not normalized)
    pub fn normal_unnormalized(&self) -> Vec3 {
        let e1 = self.v1.sub(self.v0);
        let e2 = self.v2.sub(self.v0);
        e1.cross(e2)
    }

    /// Compute triangle normal (unit vector)
    pub fn normal(&self) -> Vec3 {
        self.normal_unnormalized().normalize()
    }

    /// Compute triangle area
    pub fn area(&self) -> Scalar {
        self.normal_unnormalized().norm() * 0.5
    }

    /// Compute centroid
    pub fn centroid(&self) -> Vec3 {
        Vec3::new(
            (self.v0.x + self.v1.x + self.v2.x) / 3.0,
            (self.v0.y + self.v1.y + self.v2.y) / 3.0,
            (self.v0.z + self.v1.z + self.v2.z) / 3.0,
        )
    }

    /// Barycentric coordinates of a point
    pub fn barycentric(&self, p: Vec3) -> (Scalar, Scalar, Scalar) {
        let v0v1 = self.v1.sub(self.v0);
        let v0v2 = self.v2.sub(self.v0);
        let v0p = p.sub(self.v0);

        let d00 = v0v1.dot(v0v1);
        let d01 = v0v1.dot(v0v2);
        let d11 = v0v2.dot(v0v2);
        let d20 = v0p.dot(v0v1);
        let d21 = v0p.dot(v0v2);

        let denom = d00 * d11 - d01 * d01;
        if denom.abs() < 1e-15 {
            return (1.0 / 3.0, 1.0 / 3.0, 1.0 / 3.0);
        }

        let v = (d11 * d20 - d01 * d21) / denom;
        let w = (d00 * d21 - d01 * d20) / denom;
        let u = 1.0 - v - w;

        (u, v, w)
    }

    /// Check if point is inside triangle (using barycentric)
    pub fn contains(&self, p: Vec3) -> bool {
        let (u, v, w) = self.barycentric(p);
        u >= 0.0 && v >= 0.0 && w >= 0.0
    }

    /// Ray-triangle intersection (Möller–Trumbore algorithm)
    pub fn intersect_ray(&self, origin: Vec3, dir: Vec3) -> Option<(Scalar, Scalar, Scalar)> {
        let epsilon = 1e-10;

        let e1 = self.v1.sub(self.v0);
        let e2 = self.v2.sub(self.v0);

        let h = dir.cross(e2);
        let a = e1.dot(h);

        if a > -epsilon && a < epsilon {
            return None; // Ray parallel to triangle
        }

        let f = 1.0 / a;
        let s = origin.sub(self.v0);
        let u = f * s.dot(h);

        if u < 0.0 || u > 1.0 {
            return None;
        }

        let q = s.cross(e1);
        let v = f * dir.dot(q);

        if v < 0.0 || u + v > 1.0 {
            return None;
        }

        let t = f * e2.dot(q);

        if t > epsilon {
            Some((t, u, v))
        } else {
            None
        }
    }

    /// Signed volume of tetrahedron formed with origin
    pub fn signed_volume(&self) -> Scalar {
        self.v0.dot(self.v1.cross(self.v2)) / 6.0
    }
}

/// Ray for intersection tests
#[derive(Debug, Clone, Copy)]
pub struct Ray {
    pub origin: Vec3,
    pub direction: Vec3,
}

impl Ray {
    pub fn new(origin: Vec3, direction: Vec3) -> Self {
        Self {
            origin,
            direction: direction.normalize(),
        }
    }

    pub fn at(&self, t: Scalar) -> Vec3 {
        self.origin.add(self.direction.scale(t))
    }
}

/// Infinite plane
#[derive(Debug, Clone, Copy)]
pub struct Plane {
    pub normal: Vec3,
    pub d: Scalar, // distance from origin
}

impl Plane {
    pub fn new(normal: Vec3, point: Vec3) -> Self {
        let n = normal.normalize();
        Self {
            normal: n,
            d: -n.dot(point),
        }
    }

    pub fn from_points(p1: Vec3, p2: Vec3, p3: Vec3) -> Self {
        let normal = p2.sub(p1).cross(p3.sub(p1)).normalize();
        Self {
            normal,
            d: -normal.dot(p1),
        }
    }

    /// Signed distance from point to plane
    pub fn distance(&self, p: Vec3) -> Scalar {
        self.normal.dot(p) + self.d
    }

    /// Project point onto plane
    pub fn project(&self, p: Vec3) -> Vec3 {
        p.sub(self.normal.scale(self.distance(p)))
    }

    /// Ray-plane intersection
    pub fn intersect_ray(&self, ray: &Ray) -> Option<Scalar> {
        let denom = self.normal.dot(ray.direction);
        if denom.abs() < 1e-10 {
            return None; // Ray parallel to plane
        }
        let t = -(self.normal.dot(ray.origin) + self.d) / denom;
        if t >= 0.0 {
            Some(t)
        } else {
            None
        }
    }
}

/// Axis-aligned bounding box
#[derive(Debug, Clone, Copy)]
pub struct AABB {
    pub min: Vec3,
    pub max: Vec3,
}

impl AABB {
    pub fn new(min: Vec3, max: Vec3) -> Self {
        Self { min, max }
    }

    pub fn from_points(points: &[Vec3]) -> Option<Self> {
        if points.is_empty() {
            return None;
        }
        let mut min = points[0];
        let mut max = points[0];
        for p in &points[1..] {
            min.x = min.x.min(p.x);
            min.y = min.y.min(p.y);
            min.z = min.z.min(p.z);
            max.x = max.x.max(p.x);
            max.y = max.y.max(p.y);
            max.z = max.z.max(p.z);
        }
        Some(Self { min, max })
    }

    pub fn center(&self) -> Vec3 {
        self.min.add(self.max).scale(0.5)
    }

    pub fn size(&self) -> Vec3 {
        self.max.sub(self.min)
    }

    pub fn contains(&self, p: Vec3) -> bool {
        p.x >= self.min.x
            && p.x <= self.max.x
            && p.y >= self.min.y
            && p.y <= self.max.y
            && p.z >= self.min.z
            && p.z <= self.max.z
    }

    pub fn intersects(&self, other: &AABB) -> bool {
        self.min.x <= other.max.x
            && self.max.x >= other.min.x
            && self.min.y <= other.max.y
            && self.max.y >= other.min.y
            && self.min.z <= other.max.z
            && self.max.z >= other.min.z
    }

    /// Ray-AABB intersection (returns entry and exit t values)
    pub fn intersect_ray(&self, ray: &Ray) -> Option<(Scalar, Scalar)> {
        let inv_dir = Vec3::new(
            1.0 / ray.direction.x,
            1.0 / ray.direction.y,
            1.0 / ray.direction.z,
        );

        let t1 = (self.min.x - ray.origin.x) * inv_dir.x;
        let t2 = (self.max.x - ray.origin.x) * inv_dir.x;
        let t3 = (self.min.y - ray.origin.y) * inv_dir.y;
        let t4 = (self.max.y - ray.origin.y) * inv_dir.y;
        let t5 = (self.min.z - ray.origin.z) * inv_dir.z;
        let t6 = (self.max.z - ray.origin.z) * inv_dir.z;

        let tmin = t1.min(t2).max(t3.min(t4)).max(t5.min(t6));
        let tmax = t1.max(t2).min(t3.max(t4)).min(t5.max(t6));

        if tmax < 0.0 || tmin > tmax {
            None
        } else {
            Some((tmin.max(0.0), tmax))
        }
    }

    pub fn union(&self, other: &AABB) -> AABB {
        AABB {
            min: Vec3::new(
                self.min.x.min(other.min.x),
                self.min.y.min(other.min.y),
                self.min.z.min(other.min.z),
            ),
            max: Vec3::new(
                self.max.x.max(other.max.x),
                self.max.y.max(other.max.y),
                self.max.z.max(other.max.z),
            ),
        }
    }

    pub fn expand(&self, amount: Scalar) -> AABB {
        AABB {
            min: Vec3::new(self.min.x - amount, self.min.y - amount, self.min.z - amount),
            max: Vec3::new(self.max.x + amount, self.max.y + amount, self.max.z + amount),
        }
    }
}

/// Sphere
#[derive(Debug, Clone, Copy)]
pub struct Sphere {
    pub center: Vec3,
    pub radius: Scalar,
}

impl Sphere {
    pub fn new(center: Vec3, radius: Scalar) -> Self {
        Self { center, radius }
    }

    pub fn contains(&self, p: Vec3) -> bool {
        p.sub(self.center).norm_sq() <= self.radius * self.radius
    }

    pub fn intersects_sphere(&self, other: &Sphere) -> bool {
        let d = self.center.sub(other.center).norm();
        d <= self.radius + other.radius
    }

    pub fn intersect_ray(&self, ray: &Ray) -> Option<(Scalar, Scalar)> {
        let oc = ray.origin.sub(self.center);
        let a = ray.direction.dot(ray.direction);
        let b = 2.0 * oc.dot(ray.direction);
        let c = oc.dot(oc) - self.radius * self.radius;
        let discriminant = b * b - 4.0 * a * c;

        if discriminant < 0.0 {
            None
        } else {
            let sqrt_d = discriminant.sqrt();
            let t1 = (-b - sqrt_d) / (2.0 * a);
            let t2 = (-b + sqrt_d) / (2.0 * a);
            Some((t1, t2))
        }
    }

    /// Surface area
    pub fn surface_area(&self) -> Scalar {
        4.0 * std::f64::consts::PI * self.radius * self.radius
    }

    /// Volume
    pub fn volume(&self) -> Scalar {
        (4.0 / 3.0) * std::f64::consts::PI * self.radius * self.radius * self.radius
    }
}

/// Triangle mesh with parallel operations
#[derive(Debug, Clone)]
pub struct Mesh {
    pub vertices: Vec<Vec3>,
    pub indices: Vec<[usize; 3]>, // triangle indices
    pub normals: Option<Vec<Vec3>>,
}

impl Mesh {
    pub fn new(vertices: Vec<Vec3>, indices: Vec<[usize; 3]>) -> Self {
        Self {
            vertices,
            indices,
            normals: None,
        }
    }

    pub fn triangle(&self, i: usize) -> Triangle {
        let idx = self.indices[i];
        Triangle::new(
            self.vertices[idx[0]],
            self.vertices[idx[1]],
            self.vertices[idx[2]],
        )
    }

    pub fn triangle_count(&self) -> usize {
        self.indices.len()
    }

    /// Compute vertex normals (parallel)
    pub fn compute_normals(&mut self) {
        let mut normals = vec![Vec3::ZERO; self.vertices.len()];

        // Accumulate face normals to vertices
        for idx in &self.indices {
            let tri = Triangle::new(
                self.vertices[idx[0]],
                self.vertices[idx[1]],
                self.vertices[idx[2]],
            );
            let n = tri.normal_unnormalized();
            for &i in idx {
                normals[i] = normals[i].add(n);
            }
        }

        // Normalize
        for n in &mut normals {
            *n = n.normalize();
        }

        self.normals = Some(normals);
    }

    /// Compute bounding box
    pub fn bounding_box(&self) -> Option<AABB> {
        AABB::from_points(&self.vertices)
    }

    /// Total surface area (parallel)
    pub fn surface_area(&self) -> Scalar {
        self.indices
            .par_iter()
            .map(|idx| {
                Triangle::new(
                    self.vertices[idx[0]],
                    self.vertices[idx[1]],
                    self.vertices[idx[2]],
                )
                .area()
            })
            .sum()
    }

    /// Total volume (assuming closed mesh, parallel)
    pub fn volume(&self) -> Scalar {
        self.indices
            .par_iter()
            .map(|idx| {
                Triangle::new(
                    self.vertices[idx[0]],
                    self.vertices[idx[1]],
                    self.vertices[idx[2]],
                )
                .signed_volume()
            })
            .sum::<Scalar>()
            .abs()
    }

    /// Centroid (parallel)
    pub fn centroid(&self) -> Vec3 {
        if self.vertices.is_empty() {
            return Vec3::ZERO;
        }

        let sum: (Scalar, Scalar, Scalar) = self
            .vertices
            .par_iter()
            .map(|v| (v.x, v.y, v.z))
            .reduce(
                || (0.0, 0.0, 0.0),
                |a, b| (a.0 + b.0, a.1 + b.1, a.2 + b.2),
            );

        let n = self.vertices.len() as Scalar;
        Vec3::new(sum.0 / n, sum.1 / n, sum.2 / n)
    }

    /// Find closest intersection with ray (parallel)
    pub fn intersect_ray(&self, ray: &Ray) -> Option<(usize, Scalar, Scalar, Scalar)> {
        self.indices
            .par_iter()
            .enumerate()
            .filter_map(|(i, idx)| {
                let tri = Triangle::new(
                    self.vertices[idx[0]],
                    self.vertices[idx[1]],
                    self.vertices[idx[2]],
                );
                tri.intersect_ray(ray.origin, ray.direction)
                    .map(|(t, u, v)| (i, t, u, v))
            })
            .min_by(|a, b| a.1.partial_cmp(&b.1).unwrap())
    }
}

/// 2D Polygon
#[derive(Debug, Clone)]
pub struct Polygon2D {
    pub vertices: Vec<Vec2>,
}

impl Polygon2D {
    pub fn new(vertices: Vec<Vec2>) -> Self {
        Self { vertices }
    }

    /// Signed area (positive = counter-clockwise)
    pub fn signed_area(&self) -> Scalar {
        let n = self.vertices.len();
        if n < 3 {
            return 0.0;
        }
        let mut area = 0.0;
        for i in 0..n {
            let j = (i + 1) % n;
            area += self.vertices[i].x * self.vertices[j].y;
            area -= self.vertices[j].x * self.vertices[i].y;
        }
        area * 0.5
    }

    /// Absolute area
    pub fn area(&self) -> Scalar {
        self.signed_area().abs()
    }

    /// Centroid
    pub fn centroid(&self) -> Vec2 {
        let n = self.vertices.len();
        if n == 0 {
            return Vec2::ZERO;
        }
        let mut cx = 0.0;
        let mut cy = 0.0;
        let mut a = 0.0;

        for i in 0..n {
            let j = (i + 1) % n;
            let cross = self.vertices[i].x * self.vertices[j].y
                - self.vertices[j].x * self.vertices[i].y;
            cx += (self.vertices[i].x + self.vertices[j].x) * cross;
            cy += (self.vertices[i].y + self.vertices[j].y) * cross;
            a += cross;
        }

        if a.abs() < 1e-15 {
            // Degenerate polygon, return average
            let sum: (Scalar, Scalar) = self.vertices.iter().map(|v| (v.x, v.y)).fold(
                (0.0, 0.0),
                |acc, v| (acc.0 + v.0, acc.1 + v.1),
            );
            return Vec2::new(sum.0 / n as Scalar, sum.1 / n as Scalar);
        }

        a *= 3.0;
        Vec2::new(cx / a, cy / a)
    }

    /// Check if point is inside polygon (ray casting)
    pub fn contains(&self, p: Vec2) -> bool {
        let n = self.vertices.len();
        if n < 3 {
            return false;
        }

        let mut inside = false;
        let mut j = n - 1;

        for i in 0..n {
            let vi = &self.vertices[i];
            let vj = &self.vertices[j];

            if ((vi.y > p.y) != (vj.y > p.y))
                && (p.x < (vj.x - vi.x) * (p.y - vi.y) / (vj.y - vi.y) + vi.x)
            {
                inside = !inside;
            }
            j = i;
        }

        inside
    }

    /// Perimeter
    pub fn perimeter(&self) -> Scalar {
        let n = self.vertices.len();
        if n < 2 {
            return 0.0;
        }
        let mut p = 0.0;
        for i in 0..n {
            let j = (i + 1) % n;
            p += self.vertices[i].sub(self.vertices[j]).norm();
        }
        p
    }

    /// Is convex?
    pub fn is_convex(&self) -> bool {
        let n = self.vertices.len();
        if n < 3 {
            return true;
        }

        let mut sign = 0;
        for i in 0..n {
            let d1 = self.vertices[(i + 1) % n].sub(self.vertices[i]);
            let d2 = self.vertices[(i + 2) % n].sub(self.vertices[(i + 1) % n]);
            let cross = d1.cross(d2);

            if cross.abs() > 1e-10 {
                let current_sign = if cross > 0.0 { 1 } else { -1 };
                if sign == 0 {
                    sign = current_sign;
                } else if sign != current_sign {
                    return false;
                }
            }
        }
        true
    }
}
