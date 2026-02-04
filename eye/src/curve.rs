//! Curve operations and numerical integration
//!
//! "Find the real curve under the graph with god precision"
//!
//! Provides high-accuracy numerical integration, curve fitting,
//! and parametric curve operations.

use crate::{EyeError, EyeResult, Scalar, Vec2, Vec3};
use rayon::prelude::*;

/// Numerical integration using Simpson's rule
pub fn integrate<F>(f: F, a: Scalar, b: Scalar, n: usize) -> Scalar
where
    F: Fn(Scalar) -> Scalar + Sync,
{
    if n == 0 {
        return 0.0;
    }

    let n = if n % 2 == 1 { n + 1 } else { n }; // Simpson needs even intervals
    let h = (b - a) / n as Scalar;

    let sum: Scalar = (1..n)
        .into_par_iter()
        .map(|i| {
            let x = a + i as Scalar * h;
            let coef = if i % 2 == 0 { 2.0 } else { 4.0 };
            coef * f(x)
        })
        .sum();

    (h / 3.0) * (f(a) + sum + f(b))
}

/// Adaptive Simpson's integration for high precision
pub fn integrate_adaptive<F>(f: &F, a: Scalar, b: Scalar, tol: Scalar) -> Scalar
where
    F: Fn(Scalar) -> Scalar + Sync,
{
    adaptive_simpson(f, a, b, tol, simpson_estimate(f, a, b), 20)
}

fn simpson_estimate<F>(f: &F, a: Scalar, b: Scalar) -> Scalar
where
    F: Fn(Scalar) -> Scalar,
{
    let c = (a + b) / 2.0;
    let h = b - a;
    (h / 6.0) * (f(a) + 4.0 * f(c) + f(b))
}

fn adaptive_simpson<F>(
    f: &F,
    a: Scalar,
    b: Scalar,
    tol: Scalar,
    whole: Scalar,
    depth: usize,
) -> Scalar
where
    F: Fn(Scalar) -> Scalar + Sync,
{
    let c = (a + b) / 2.0;
    let left = simpson_estimate(f, a, c);
    let right = simpson_estimate(f, c, b);
    let delta = left + right - whole;

    if depth == 0 || delta.abs() <= 15.0 * tol {
        left + right + delta / 15.0
    } else {
        let half_tol = tol / 2.0;
        adaptive_simpson(f, a, c, half_tol, left, depth - 1)
            + adaptive_simpson(f, c, b, half_tol, right, depth - 1)
    }
}

/// Gaussian quadrature (high precision for smooth functions)
pub fn integrate_gauss<F>(f: F, a: Scalar, b: Scalar, order: usize) -> Scalar
where
    F: Fn(Scalar) -> Scalar,
{
    // Gauss-Legendre nodes and weights
    let (nodes, weights) = gauss_legendre_nodes(order);

    let mid = (a + b) / 2.0;
    let half = (b - a) / 2.0;

    let mut sum = 0.0;
    for i in 0..order {
        let x = mid + half * nodes[i];
        sum += weights[i] * f(x);
    }

    half * sum
}

fn gauss_legendre_nodes(order: usize) -> (Vec<Scalar>, Vec<Scalar>) {
    match order {
        1 => (vec![0.0], vec![2.0]),
        2 => (
            vec![-0.5773502691896257, 0.5773502691896257],
            vec![1.0, 1.0],
        ),
        3 => (
            vec![-0.7745966692414834, 0.0, 0.7745966692414834],
            vec![0.5555555555555556, 0.8888888888888888, 0.5555555555555556],
        ),
        4 => (
            vec![
                -0.8611363115940526,
                -0.3399810435848563,
                0.3399810435848563,
                0.8611363115940526,
            ],
            vec![
                0.3478548451374538,
                0.6521451548625461,
                0.6521451548625461,
                0.3478548451374538,
            ],
        ),
        5 => (
            vec![
                -0.9061798459386640,
                -0.5384693101056831,
                0.0,
                0.5384693101056831,
                0.9061798459386640,
            ],
            vec![
                0.2369268850561891,
                0.4786286704993665,
                0.5688888888888889,
                0.4786286704993665,
                0.2369268850561891,
            ],
        ),
        _ => {
            // Fall back to order 5 for unsupported orders
            gauss_legendre_nodes(5)
        }
    }
}

/// Romberg integration (extrapolated trapezoidal rule)
pub fn integrate_romberg<F>(f: F, a: Scalar, b: Scalar, max_iter: usize, tol: Scalar) -> Scalar
where
    F: Fn(Scalar) -> Scalar,
{
    let mut r = vec![vec![0.0; max_iter]; max_iter];

    // Initial trapezoidal estimate
    let mut h = b - a;
    r[0][0] = h * (f(a) + f(b)) / 2.0;

    for i in 1..max_iter {
        h /= 2.0;

        // Trapezoidal rule with new points
        let sum: Scalar = (1..=(1 << (i - 1)))
            .map(|k| f(a + (2 * k - 1) as Scalar * h))
            .sum();
        r[i][0] = r[i - 1][0] / 2.0 + h * sum;

        // Richardson extrapolation
        for j in 1..=i {
            let factor = (4_u64.pow(j as u32)) as Scalar;
            r[i][j] = (factor * r[i][j - 1] - r[i - 1][j - 1]) / (factor - 1.0);
        }

        // Check convergence
        if i > 0 && (r[i][i] - r[i - 1][i - 1]).abs() < tol {
            return r[i][i];
        }
    }

    r[max_iter - 1][max_iter - 1]
}

/// Monte Carlo integration (for high-dimensional integrals)
pub fn integrate_monte_carlo<F>(f: F, bounds: &[(Scalar, Scalar)], samples: usize) -> (Scalar, Scalar)
where
    F: Fn(&[Scalar]) -> Scalar + Sync + Send,
{
    use std::sync::atomic::{AtomicU64, Ordering};

    let dim = bounds.len();
    let volume: Scalar = bounds.iter().map(|(a, b)| b - a).product();

    // Use simple LCG for reproducible parallel random
    let seed = AtomicU64::new(42);

    let results: Vec<Scalar> = (0..samples)
        .into_par_iter()
        .map(|_| {
            let s = seed.fetch_add(1, Ordering::Relaxed);
            let mut rng_state = s;
            let mut point = Vec::with_capacity(dim);
            for (a, b) in bounds {
                // LCG: x' = (a*x + c) mod m
                rng_state = rng_state.wrapping_mul(6364136223846793005).wrapping_add(1442695040888963407);
                let u = (rng_state as Scalar) / (u64::MAX as Scalar);
                point.push(a + u * (b - a));
            }
            f(&point)
        })
        .collect();

    let n = samples as Scalar;
    let mean = results.iter().sum::<Scalar>() / n;
    let variance = results.iter().map(|x| (x - mean).powi(2)).sum::<Scalar>() / (n - 1.0);
    let std_error = (variance / n).sqrt();

    (volume * mean, volume * std_error)
}

/// Arc length of parametric curve
pub fn arc_length<F>(curve: F, t0: Scalar, t1: Scalar, n: usize) -> Scalar
where
    F: Fn(Scalar) -> Vec3 + Sync,
{
    let dt = (t1 - t0) / n as Scalar;

    (0..n)
        .into_par_iter()
        .map(|i| {
            let t = t0 + i as Scalar * dt;
            let p1 = curve(t);
            let p2 = curve(t + dt);
            p2.sub(p1).norm()
        })
        .sum()
}

/// Arc length with derivative (more accurate)
pub fn arc_length_derivative<F, G>(
    _curve: F,
    derivative: G,
    t0: Scalar,
    t1: Scalar,
    n: usize,
) -> Scalar
where
    F: Fn(Scalar) -> Vec3,
    G: Fn(Scalar) -> Vec3 + Sync,
{
    integrate(|t| derivative(t).norm(), t0, t1, n)
}

/// Bezier curve evaluation
#[derive(Debug, Clone)]
pub struct BezierCurve {
    pub control_points: Vec<Vec3>,
}

impl BezierCurve {
    pub fn new(control_points: Vec<Vec3>) -> Self {
        Self { control_points }
    }

    /// Evaluate curve at parameter t ∈ [0, 1] using De Casteljau's algorithm
    pub fn evaluate(&self, t: Scalar) -> Vec3 {
        let n = self.control_points.len();
        if n == 0 {
            return Vec3::ZERO;
        }
        if n == 1 {
            return self.control_points[0];
        }

        let mut points = self.control_points.clone();
        for r in 1..n {
            for i in 0..(n - r) {
                points[i] = points[i].lerp(points[i + 1], t);
            }
        }
        points[0]
    }

    /// Evaluate derivative at parameter t
    pub fn derivative(&self, t: Scalar) -> Vec3 {
        let n = self.control_points.len();
        if n < 2 {
            return Vec3::ZERO;
        }

        // Derivative control points
        let d_points: Vec<Vec3> = (0..n - 1)
            .map(|i| {
                self.control_points[i + 1]
                    .sub(self.control_points[i])
                    .scale((n - 1) as Scalar)
            })
            .collect();

        BezierCurve::new(d_points).evaluate(t)
    }

    /// Arc length
    pub fn arc_length(&self, n: usize) -> Scalar {
        arc_length(|t| self.evaluate(t), 0.0, 1.0, n)
    }

    /// Split curve at parameter t
    pub fn split(&self, t: Scalar) -> (BezierCurve, BezierCurve) {
        let n = self.control_points.len();
        let mut left = Vec::with_capacity(n);
        let mut right = Vec::with_capacity(n);

        let mut points = self.control_points.clone();
        left.push(points[0]);
        right.push(points[n - 1]);

        for r in 1..n {
            for i in 0..(n - r) {
                points[i] = points[i].lerp(points[i + 1], t);
            }
            left.push(points[0]);
            right.push(points[n - 1 - r]);
        }

        right.reverse();
        (BezierCurve::new(left), BezierCurve::new(right))
    }
}

/// B-spline curve
#[derive(Debug, Clone)]
pub struct BSpline {
    pub control_points: Vec<Vec3>,
    pub knots: Vec<Scalar>,
    pub degree: usize,
}

impl BSpline {
    pub fn new(control_points: Vec<Vec3>, degree: usize) -> Self {
        let n = control_points.len();
        let k = degree + 1;
        let num_knots = n + k;

        // Generate uniform knots
        let mut knots = Vec::with_capacity(num_knots);
        for i in 0..num_knots {
            if i < k {
                knots.push(0.0);
            } else if i >= n {
                knots.push(1.0);
            } else {
                knots.push((i - degree) as Scalar / (n - degree) as Scalar);
            }
        }

        Self {
            control_points,
            knots,
            degree,
        }
    }

    /// Evaluate using De Boor's algorithm
    pub fn evaluate(&self, t: Scalar) -> Vec3 {
        let t = t.clamp(0.0, 1.0);
        let n = self.control_points.len();
        let p = self.degree;

        // Find knot span
        let mut k = p;
        for i in p..(n) {
            if t >= self.knots[i] && t < self.knots[i + 1] {
                k = i;
                break;
            }
        }
        if t >= self.knots[n] {
            k = n - 1;
        }

        // De Boor's algorithm
        let mut d: Vec<Vec3> = (0..=p)
            .map(|j| self.control_points[(k - p + j).min(n - 1)])
            .collect();

        for r in 1..=p {
            for j in (r..=p).rev() {
                let i = k - p + j;
                let alpha = if (self.knots[i + p + 1 - r] - self.knots[i]).abs() < 1e-15 {
                    0.0
                } else {
                    (t - self.knots[i]) / (self.knots[i + p + 1 - r] - self.knots[i])
                };
                d[j] = d[j - 1].lerp(d[j], alpha);
            }
        }

        d[p]
    }
}

/// Curve fitting via least squares
pub fn fit_polynomial(x: &[Scalar], y: &[Scalar], degree: usize) -> EyeResult<Vec<Scalar>> {
    if x.len() != y.len() {
        return Err(EyeError::DimensionMismatch {
            expected: x.len(),
            got: y.len(),
        });
    }

    let n = x.len();
    let m = degree + 1;

    if n < m {
        return Err(EyeError::InvalidInput("not enough points for polynomial degree".into()));
    }

    // Build Vandermonde matrix
    let mut a = crate::Matrix::zeros(n, m);
    for i in 0..n {
        let mut xi = 1.0;
        for j in 0..m {
            a.set(i, j, xi);
            xi *= x[i];
        }
    }

    // Solve normal equations: A^T A c = A^T y
    let at = a.transpose();
    let ata = at.mul(&a)?;
    let aty_data: Vec<Scalar> = (0..m)
        .map(|j| (0..n).map(|i| at.get(j, i) * y[i]).sum())
        .collect();
    let aty = crate::Vector::from_data(aty_data);

    let coeffs = ata.solve(&aty)?;
    Ok(coeffs.data().to_vec())
}

/// Evaluate polynomial at x
pub fn eval_polynomial(coeffs: &[Scalar], x: Scalar) -> Scalar {
    let mut result = 0.0;
    let mut xi = 1.0;
    for &c in coeffs {
        result += c * xi;
        xi *= x;
    }
    result
}

/// Numerical derivative
pub fn derivative<F>(f: F, x: Scalar, h: Scalar) -> Scalar
where
    F: Fn(Scalar) -> Scalar,
{
    // Central difference (O(h^2) accuracy)
    (f(x + h) - f(x - h)) / (2.0 * h)
}

/// Second derivative
pub fn second_derivative<F>(f: F, x: Scalar, h: Scalar) -> Scalar
where
    F: Fn(Scalar) -> Scalar,
{
    (f(x + h) - 2.0 * f(x) + f(x - h)) / (h * h)
}

/// Find root using Newton-Raphson method
pub fn find_root<F, G>(
    f: F,
    df: G,
    x0: Scalar,
    tol: Scalar,
    max_iter: usize,
) -> EyeResult<Scalar>
where
    F: Fn(Scalar) -> Scalar,
    G: Fn(Scalar) -> Scalar,
{
    let mut x = x0;

    for i in 0..max_iter {
        let fx = f(x);
        if fx.abs() < tol {
            return Ok(x);
        }

        let dfx = df(x);
        if dfx.abs() < 1e-15 {
            return Err(EyeError::NotConverged {
                iterations: i,
                residual: fx.abs(),
            });
        }

        x -= fx / dfx;
    }

    Err(EyeError::NotConverged {
        iterations: max_iter,
        residual: f(x).abs(),
    })
}

/// Find root using bisection (guaranteed convergence for continuous functions)
pub fn find_root_bisection<F>(
    f: F,
    mut a: Scalar,
    mut b: Scalar,
    tol: Scalar,
    max_iter: usize,
) -> EyeResult<Scalar>
where
    F: Fn(Scalar) -> Scalar,
{
    let mut fa = f(a);
    let fb = f(b);

    if fa * fb > 0.0 {
        return Err(EyeError::InvalidInput(
            "function must have opposite signs at interval endpoints".into(),
        ));
    }

    for _ in 0..max_iter {
        let c = (a + b) / 2.0;
        let fc = f(c);

        if fc.abs() < tol || (b - a) / 2.0 < tol {
            return Ok(c);
        }

        if fa * fc < 0.0 {
            b = c;
        } else {
            a = c;
            fa = fc;
        }
    }

    Ok((a + b) / 2.0)
}

/// Curvature of a parametric curve at t
pub fn curvature<F, G>(
    derivative: F,
    second_derivative: G,
    t: Scalar,
) -> Scalar
where
    F: Fn(Scalar) -> Vec3,
    G: Fn(Scalar) -> Vec3,
{
    let d1 = derivative(t);
    let d2 = second_derivative(t);
    let cross = d1.cross(d2);
    let d1_norm = d1.norm();

    if d1_norm < 1e-15 {
        return 0.0;
    }

    cross.norm() / d1_norm.powi(3)
}

/// 2D curvature
pub fn curvature_2d<F, G>(
    derivative: F,
    second_derivative: G,
    t: Scalar,
) -> Scalar
where
    F: Fn(Scalar) -> Vec2,
    G: Fn(Scalar) -> Vec2,
{
    let d1 = derivative(t);
    let d2 = second_derivative(t);

    let cross = d1.x * d2.y - d1.y * d2.x;
    let d1_norm = d1.norm();

    if d1_norm < 1e-15 {
        return 0.0;
    }

    cross / d1_norm.powi(3)
}
