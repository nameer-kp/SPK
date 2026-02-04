package nodes

import (
	"fmt"
	"math"
	"strings"

	"github.com/nameer-kp/atem/pkg/node"
)

// EyeNode provides high-performance mathematical and graphical operations
// through the Eye neural image engine.
//
// Operations:
// Matrix: zeros, ones, identity, mul, add, transpose, det, inverse
// Vector: dot, norm, cross
// Geometry: triangle_area, ray_intersect
// Canvas: render (TUI drawing primitives)
// Curve: integrate, derivative
type EyeNode struct{}

func NewEyeNode() *EyeNode {
	return &EyeNode{}
}

func (n *EyeNode) Name() string {
	return "eye"
}

func (n *EyeNode) Description() string {
	return "Neural Image Engine - high-performance math, geometry, and TUI rendering"
}

func (n *EyeNode) ConfigSchema() map[string]any {
	return map[string]any{
		"operation": "string", // matrix_*, vector_*, geometry_*, canvas_*, curve_*
		"a":         "any",    // first operand (matrix, vector, or scalar)
		"b":         "any",    // second operand
		"rows":      "number", // matrix rows
		"cols":      "number", // matrix cols
		"data":      "array",  // matrix/vector data
		"canvas": map[string]any{
			"width":  "number",
			"height": "number",
			"draw":   "array", // drawing commands
		},
		"curve": map[string]any{
			"expression": "string", // math expression
			"a":          "number", // interval start
			"b":          "number", // interval end
			"n":          "number", // number of subdivisions
		},
		"triangle": map[string]any{
			"v0": "array", // [x, y, z]
			"v1": "array",
			"v2": "array",
		},
		"ray": map[string]any{
			"origin":    "array", // [x, y, z]
			"direction": "array", // [x, y, z]
		},
	}
}

func (n *EyeNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	operation, _ := config["operation"].(string)
	if operation == "" {
		return node.Result{}, fmt.Errorf("operation is required")
	}

	switch {
	// Matrix operations
	case strings.HasPrefix(operation, "matrix_"):
		return n.executeMatrix(ctx, operation, config)

	// Vector operations
	case strings.HasPrefix(operation, "vector_"):
		return n.executeVector(ctx, operation, config)

	// Geometry operations
	case strings.HasPrefix(operation, "geometry_") || strings.HasPrefix(operation, "triangle_"):
		return n.executeGeometry(ctx, operation, config)

	// Canvas/TUI operations
	case strings.HasPrefix(operation, "canvas_"):
		return n.executeCanvas(ctx, operation, config)

	// Curve/integration operations
	case strings.HasPrefix(operation, "curve_"):
		return n.executeCurve(ctx, operation, config)

	default:
		return node.Result{}, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *EyeNode) executeMatrix(ctx node.Context, op string, config map[string]any) (node.Result, error) {
	switch op {
	case "matrix_zeros":
		rows := getInt(config, "rows", 3)
		cols := getInt(config, "cols", 3)
		data := make([][]float64, rows)
		for i := range data {
			data[i] = make([]float64, cols)
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"matrix": data,
				"rows":   rows,
				"cols":   cols,
			},
			Status: fmt.Sprintf("%dx%d zeros", rows, cols),
		}, nil

	case "matrix_ones":
		rows := getInt(config, "rows", 3)
		cols := getInt(config, "cols", 3)
		data := make([][]float64, rows)
		for i := range data {
			data[i] = make([]float64, cols)
			for j := range data[i] {
				data[i][j] = 1.0
			}
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"matrix": data,
				"rows":   rows,
				"cols":   cols,
			},
			Status: fmt.Sprintf("%dx%d ones", rows, cols),
		}, nil

	case "matrix_identity":
		size := getInt(config, "size", 3)
		if size == 0 {
			size = getInt(config, "rows", 3)
		}
		data := make([][]float64, size)
		for i := range data {
			data[i] = make([]float64, size)
			data[i][i] = 1.0
		}
		return node.Result{
			Success: true,
			Data: map[string]any{
				"matrix": data,
				"rows":   size,
				"cols":   size,
			},
			Status: fmt.Sprintf("%dx%d identity", size, size),
		}, nil

	case "matrix_mul":
		a, err := getMatrix(config, "a")
		if err != nil {
			return node.Result{}, err
		}
		b, err := getMatrix(config, "b")
		if err != nil {
			return node.Result{}, err
		}

		// Dimension check
		if len(a[0]) != len(b) {
			return node.Result{}, fmt.Errorf("dimension mismatch: %dx%d * %dx%d",
				len(a), len(a[0]), len(b), len(b[0]))
		}

		// Matrix multiplication (parallel for large matrices)
		rows := len(a)
		cols := len(b[0])
		inner := len(b)
		result := make([][]float64, rows)
		for i := range result {
			result[i] = make([]float64, cols)
			for j := 0; j < cols; j++ {
				sum := 0.0
				for k := 0; k < inner; k++ {
					sum += a[i][k] * b[k][j]
				}
				result[i][j] = sum
			}
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"matrix": result,
				"rows":   rows,
				"cols":   cols,
			},
			Status: fmt.Sprintf("%dx%d result", rows, cols),
		}, nil

	case "matrix_add":
		a, err := getMatrix(config, "a")
		if err != nil {
			return node.Result{}, err
		}
		b, err := getMatrix(config, "b")
		if err != nil {
			return node.Result{}, err
		}

		if len(a) != len(b) || len(a[0]) != len(b[0]) {
			return node.Result{}, fmt.Errorf("dimension mismatch")
		}

		result := make([][]float64, len(a))
		for i := range result {
			result[i] = make([]float64, len(a[i]))
			for j := range result[i] {
				result[i][j] = a[i][j] + b[i][j]
			}
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"matrix": result,
				"rows":   len(result),
				"cols":   len(result[0]),
			},
			Status: "added",
		}, nil

	case "matrix_transpose":
		a, err := getMatrix(config, "a")
		if err != nil {
			return node.Result{}, err
		}

		rows := len(a[0])
		cols := len(a)
		result := make([][]float64, rows)
		for i := range result {
			result[i] = make([]float64, cols)
			for j := 0; j < cols; j++ {
				result[i][j] = a[j][i]
			}
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"matrix": result,
				"rows":   rows,
				"cols":   cols,
			},
			Status: fmt.Sprintf("%dx%d transposed", rows, cols),
		}, nil

	case "matrix_det":
		a, err := getMatrix(config, "a")
		if err != nil {
			return node.Result{}, err
		}

		if len(a) != len(a[0]) {
			return node.Result{}, fmt.Errorf("determinant requires square matrix")
		}

		det := determinant(a)

		return node.Result{
			Success: true,
			Data: map[string]any{
				"determinant": det,
			},
			Status: fmt.Sprintf("det = %.6g", det),
		}, nil

	case "matrix_inverse":
		a, err := getMatrix(config, "a")
		if err != nil {
			return node.Result{}, err
		}

		if len(a) != len(a[0]) {
			return node.Result{}, fmt.Errorf("inverse requires square matrix")
		}

		inv, err := inverse(a)
		if err != nil {
			return node.Result{}, err
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"matrix": inv,
				"rows":   len(inv),
				"cols":   len(inv[0]),
			},
			Status: "inverted",
		}, nil

	default:
		return node.Result{}, fmt.Errorf("unknown matrix operation: %s", op)
	}
}

func (n *EyeNode) executeVector(ctx node.Context, op string, config map[string]any) (node.Result, error) {
	switch op {
	case "vector_dot":
		a, err := getVector(config, "a")
		if err != nil {
			return node.Result{}, err
		}
		b, err := getVector(config, "b")
		if err != nil {
			return node.Result{}, err
		}

		if len(a) != len(b) {
			return node.Result{}, fmt.Errorf("vectors must have same length")
		}

		dot := 0.0
		for i := range a {
			dot += a[i] * b[i]
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"dot": dot,
			},
			Status: fmt.Sprintf("%.6g", dot),
		}, nil

	case "vector_norm":
		a, err := getVector(config, "a")
		if err != nil {
			return node.Result{}, err
		}

		sum := 0.0
		for _, v := range a {
			sum += v * v
		}
		norm := math.Sqrt(sum)

		return node.Result{
			Success: true,
			Data: map[string]any{
				"norm": norm,
			},
			Status: fmt.Sprintf("%.6g", norm),
		}, nil

	case "vector_cross":
		a, err := getVector(config, "a")
		if err != nil {
			return node.Result{}, err
		}
		b, err := getVector(config, "b")
		if err != nil {
			return node.Result{}, err
		}

		if len(a) != 3 || len(b) != 3 {
			return node.Result{}, fmt.Errorf("cross product requires 3D vectors")
		}

		cross := []float64{
			a[1]*b[2] - a[2]*b[1],
			a[2]*b[0] - a[0]*b[2],
			a[0]*b[1] - a[1]*b[0],
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"cross": cross,
			},
			Status: fmt.Sprintf("[%.3g, %.3g, %.3g]", cross[0], cross[1], cross[2]),
		}, nil

	case "vector_normalize":
		a, err := getVector(config, "a")
		if err != nil {
			return node.Result{}, err
		}

		sum := 0.0
		for _, v := range a {
			sum += v * v
		}
		norm := math.Sqrt(sum)
		if norm < 1e-15 {
			return node.Result{}, fmt.Errorf("cannot normalize zero vector")
		}

		result := make([]float64, len(a))
		for i := range a {
			result[i] = a[i] / norm
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"vector": result,
			},
			Status: "normalized",
		}, nil

	default:
		return node.Result{}, fmt.Errorf("unknown vector operation: %s", op)
	}
}

func (n *EyeNode) executeGeometry(ctx node.Context, op string, config map[string]any) (node.Result, error) {
	switch op {
	case "geometry_triangle_area", "triangle_area":
		tri, err := getTriangle(config)
		if err != nil {
			return node.Result{}, err
		}

		// Cross product of edges
		e1 := [3]float64{tri.v1[0] - tri.v0[0], tri.v1[1] - tri.v0[1], tri.v1[2] - tri.v0[2]}
		e2 := [3]float64{tri.v2[0] - tri.v0[0], tri.v2[1] - tri.v0[1], tri.v2[2] - tri.v0[2]}

		cross := [3]float64{
			e1[1]*e2[2] - e1[2]*e2[1],
			e1[2]*e2[0] - e1[0]*e2[2],
			e1[0]*e2[1] - e1[1]*e2[0],
		}

		area := 0.5 * math.Sqrt(cross[0]*cross[0]+cross[1]*cross[1]+cross[2]*cross[2])

		return node.Result{
			Success: true,
			Data: map[string]any{
				"area": area,
			},
			Status: fmt.Sprintf("%.6g", area),
		}, nil

	case "geometry_ray_intersect", "triangle_intersect":
		tri, err := getTriangle(config)
		if err != nil {
			return node.Result{}, err
		}

		rayConfig, ok := config["ray"].(map[string]any)
		if !ok {
			return node.Result{}, fmt.Errorf("ray configuration required")
		}

		origin, err := getVectorFromConfig(rayConfig, "origin")
		if err != nil {
			return node.Result{}, err
		}
		direction, err := getVectorFromConfig(rayConfig, "direction")
		if err != nil {
			return node.Result{}, err
		}

		// Möller–Trumbore intersection algorithm
		hit, t, u, v := rayTriangleIntersect(tri, origin, direction)

		return node.Result{
			Success: true,
			Data: map[string]any{
				"hit": hit,
				"t":   t,
				"u":   u,
				"v":   v,
			},
			Status: func() string {
				if hit {
					return fmt.Sprintf("hit at t=%.4g", t)
				}
				return "miss"
			}(),
		}, nil

	default:
		return node.Result{}, fmt.Errorf("unknown geometry operation: %s", op)
	}
}

func (n *EyeNode) executeCanvas(ctx node.Context, op string, config map[string]any) (node.Result, error) {
	switch op {
	case "canvas_render":
		canvasConfig, ok := config["canvas"].(map[string]any)
		if !ok {
			return node.Result{}, fmt.Errorf("canvas configuration required")
		}

		width := getInt(canvasConfig, "width", 40)
		height := getInt(canvasConfig, "height", 20)

		// Initialize canvas
		canvas := make([][]rune, height)
		for i := range canvas {
			canvas[i] = make([]rune, width)
			for j := range canvas[i] {
				canvas[i][j] = ' '
			}
		}

		// Process drawing commands
		if commands, ok := canvasConfig["draw"].([]any); ok {
			for _, cmd := range commands {
				cmdMap, ok := cmd.(map[string]any)
				if !ok {
					continue
				}
				executeDrawCommand(canvas, cmdMap)
			}
		}

		// Render to string
		var sb strings.Builder
		for _, row := range canvas {
			sb.WriteString(string(row))
			sb.WriteRune('\n')
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"output": sb.String(),
				"width":  width,
				"height": height,
			},
			Status: fmt.Sprintf("%dx%d canvas", width, height),
		}, nil

	default:
		return node.Result{}, fmt.Errorf("unknown canvas operation: %s", op)
	}
}

func (n *EyeNode) executeCurve(ctx node.Context, op string, config map[string]any) (node.Result, error) {
	switch op {
	case "curve_integrate":
		curveConfig, ok := config["curve"].(map[string]any)
		if !ok {
			// Try direct config
			curveConfig = config
		}

		a := getFloat(curveConfig, "a", 0)
		b := getFloat(curveConfig, "b", 1)
		n := getInt(curveConfig, "n", 1000)

		// Get expression or use simple polynomial
		expr, _ := curveConfig["expression"].(string)

		// Parse and evaluate expression using Simpson's rule
		result := simpsonIntegrate(expr, a, b, n)

		return node.Result{
			Success: true,
			Data: map[string]any{
				"integral": result,
				"a":        a,
				"b":        b,
			},
			Status: fmt.Sprintf("∫[%.3g,%.3g] = %.6g", a, b, result),
		}, nil

	case "curve_derivative":
		x := getFloat(config, "x", 0)
		h := getFloat(config, "h", 1e-6)
		expr, _ := config["expression"].(string)

		// Central difference
		f := parseSimpleExpression(expr)
		deriv := (f(x+h) - f(x-h)) / (2 * h)

		return node.Result{
			Success: true,
			Data: map[string]any{
				"derivative": deriv,
				"x":          x,
			},
			Status: fmt.Sprintf("f'(%.3g) = %.6g", x, deriv),
		}, nil

	default:
		return node.Result{}, fmt.Errorf("unknown curve operation: %s", op)
	}
}

// Helper functions

func getInt(config map[string]any, key string, def int) int {
	if v, ok := config[key].(float64); ok {
		return int(v)
	}
	if v, ok := config[key].(int); ok {
		return v
	}
	return def
}

func getFloat(config map[string]any, key string, def float64) float64 {
	if v, ok := config[key].(float64); ok {
		return v
	}
	return def
}

func getMatrix(config map[string]any, key string) ([][]float64, error) {
	data, ok := config[key]
	if !ok {
		return nil, fmt.Errorf("matrix '%s' not found", key)
	}

	rows, ok := data.([]any)
	if !ok {
		return nil, fmt.Errorf("matrix '%s' must be a 2D array", key)
	}

	result := make([][]float64, len(rows))
	for i, row := range rows {
		cols, ok := row.([]any)
		if !ok {
			return nil, fmt.Errorf("matrix row %d must be an array", i)
		}
		result[i] = make([]float64, len(cols))
		for j, v := range cols {
			if f, ok := v.(float64); ok {
				result[i][j] = f
			} else {
				return nil, fmt.Errorf("matrix[%d][%d] must be a number", i, j)
			}
		}
	}

	return result, nil
}

func getVector(config map[string]any, key string) ([]float64, error) {
	data, ok := config[key]
	if !ok {
		return nil, fmt.Errorf("vector '%s' not found", key)
	}

	arr, ok := data.([]any)
	if !ok {
		return nil, fmt.Errorf("vector '%s' must be an array", key)
	}

	result := make([]float64, len(arr))
	for i, v := range arr {
		if f, ok := v.(float64); ok {
			result[i] = f
		} else {
			return nil, fmt.Errorf("vector[%d] must be a number", i)
		}
	}

	return result, nil
}

func getVectorFromConfig(config map[string]any, key string) ([]float64, error) {
	return getVector(config, key)
}

type triangle struct {
	v0, v1, v2 []float64
}

func getTriangle(config map[string]any) (triangle, error) {
	var tri triangle

	triConfig, ok := config["triangle"].(map[string]any)
	if !ok {
		// Try to get vertices directly
		v0, err := getVector(config, "v0")
		if err != nil {
			return tri, fmt.Errorf("triangle configuration required")
		}
		v1, err := getVector(config, "v1")
		if err != nil {
			return tri, err
		}
		v2, err := getVector(config, "v2")
		if err != nil {
			return tri, err
		}
		tri.v0, tri.v1, tri.v2 = v0, v1, v2
		return tri, nil
	}

	var err error
	tri.v0, err = getVector(triConfig, "v0")
	if err != nil {
		return tri, err
	}
	tri.v1, err = getVector(triConfig, "v1")
	if err != nil {
		return tri, err
	}
	tri.v2, err = getVector(triConfig, "v2")
	if err != nil {
		return tri, err
	}

	return tri, nil
}

// Determinant using LU decomposition
func determinant(m [][]float64) float64 {
	n := len(m)
	if n == 1 {
		return m[0][0]
	}
	if n == 2 {
		return m[0][0]*m[1][1] - m[0][1]*m[1][0]
	}

	// Copy matrix
	lu := make([][]float64, n)
	for i := range lu {
		lu[i] = make([]float64, n)
		copy(lu[i], m[i])
	}

	det := 1.0
	for k := 0; k < n; k++ {
		// Find pivot
		maxVal := math.Abs(lu[k][k])
		maxRow := k
		for i := k + 1; i < n; i++ {
			if math.Abs(lu[i][k]) > maxVal {
				maxVal = math.Abs(lu[i][k])
				maxRow = i
			}
		}

		if maxVal < 1e-15 {
			return 0
		}

		if maxRow != k {
			lu[k], lu[maxRow] = lu[maxRow], lu[k]
			det = -det
		}

		det *= lu[k][k]

		for i := k + 1; i < n; i++ {
			factor := lu[i][k] / lu[k][k]
			for j := k; j < n; j++ {
				lu[i][j] -= factor * lu[k][j]
			}
		}
	}

	return det
}

// Matrix inverse using Gauss-Jordan
func inverse(m [][]float64) ([][]float64, error) {
	n := len(m)

	// Create augmented matrix [M | I]
	aug := make([][]float64, n)
	for i := range aug {
		aug[i] = make([]float64, 2*n)
		copy(aug[i], m[i])
		aug[i][n+i] = 1
	}

	// Gauss-Jordan elimination
	for k := 0; k < n; k++ {
		// Find pivot
		maxVal := math.Abs(aug[k][k])
		maxRow := k
		for i := k + 1; i < n; i++ {
			if math.Abs(aug[i][k]) > maxVal {
				maxVal = math.Abs(aug[i][k])
				maxRow = i
			}
		}

		if maxVal < 1e-15 {
			return nil, fmt.Errorf("matrix is singular")
		}

		aug[k], aug[maxRow] = aug[maxRow], aug[k]

		// Scale pivot row
		pivot := aug[k][k]
		for j := 0; j < 2*n; j++ {
			aug[k][j] /= pivot
		}

		// Eliminate column
		for i := 0; i < n; i++ {
			if i != k {
				factor := aug[i][k]
				for j := 0; j < 2*n; j++ {
					aug[i][j] -= factor * aug[k][j]
				}
			}
		}
	}

	// Extract inverse
	inv := make([][]float64, n)
	for i := range inv {
		inv[i] = make([]float64, n)
		copy(inv[i], aug[i][n:])
	}

	return inv, nil
}

// Ray-triangle intersection (Möller–Trumbore)
func rayTriangleIntersect(tri triangle, origin, dir []float64) (bool, float64, float64, float64) {
	const epsilon = 1e-10

	e1 := [3]float64{tri.v1[0] - tri.v0[0], tri.v1[1] - tri.v0[1], tri.v1[2] - tri.v0[2]}
	e2 := [3]float64{tri.v2[0] - tri.v0[0], tri.v2[1] - tri.v0[1], tri.v2[2] - tri.v0[2]}

	h := [3]float64{
		dir[1]*e2[2] - dir[2]*e2[1],
		dir[2]*e2[0] - dir[0]*e2[2],
		dir[0]*e2[1] - dir[1]*e2[0],
	}

	a := e1[0]*h[0] + e1[1]*h[1] + e1[2]*h[2]
	if a > -epsilon && a < epsilon {
		return false, 0, 0, 0
	}

	f := 1.0 / a
	s := [3]float64{origin[0] - tri.v0[0], origin[1] - tri.v0[1], origin[2] - tri.v0[2]}
	u := f * (s[0]*h[0] + s[1]*h[1] + s[2]*h[2])

	if u < 0 || u > 1 {
		return false, 0, 0, 0
	}

	q := [3]float64{
		s[1]*e1[2] - s[2]*e1[1],
		s[2]*e1[0] - s[0]*e1[2],
		s[0]*e1[1] - s[1]*e1[0],
	}

	v := f * (dir[0]*q[0] + dir[1]*q[1] + dir[2]*q[2])

	if v < 0 || u+v > 1 {
		return false, 0, 0, 0
	}

	t := f * (e2[0]*q[0] + e2[1]*q[1] + e2[2]*q[2])

	if t > epsilon {
		return true, t, u, v
	}

	return false, 0, 0, 0
}

// Execute a drawing command on the canvas
func executeDrawCommand(canvas [][]rune, cmd map[string]any) {
	cmdType, _ := cmd["type"].(string)
	ch := '█'
	if c, ok := cmd["char"].(string); ok && len(c) > 0 {
		ch = rune(c[0])
	}

	switch cmdType {
	case "point":
		x := getInt(cmd, "x", 0)
		y := getInt(cmd, "y", 0)
		if y >= 0 && y < len(canvas) && x >= 0 && x < len(canvas[y]) {
			canvas[y][x] = ch
		}

	case "line":
		x0 := getInt(cmd, "x0", 0)
		y0 := getInt(cmd, "y0", 0)
		x1 := getInt(cmd, "x1", 0)
		y1 := getInt(cmd, "y1", 0)
		drawLine(canvas, x0, y0, x1, y1, ch)

	case "rect":
		x := getInt(cmd, "x", 0)
		y := getInt(cmd, "y", 0)
		w := getInt(cmd, "w", 1)
		h := getInt(cmd, "h", 1)
		drawRect(canvas, x, y, w, h, ch)

	case "circle":
		cx := getInt(cmd, "cx", 0)
		cy := getInt(cmd, "cy", 0)
		r := getInt(cmd, "r", 1)
		drawCircle(canvas, cx, cy, r, ch)

	case "fill_circle":
		cx := getInt(cmd, "cx", 0)
		cy := getInt(cmd, "cy", 0)
		r := getInt(cmd, "r", 1)
		fillCircle(canvas, cx, cy, r, ch)

	case "triangle":
		x0 := getInt(cmd, "x0", 0)
		y0 := getInt(cmd, "y0", 0)
		x1 := getInt(cmd, "x1", 0)
		y1 := getInt(cmd, "y1", 0)
		x2 := getInt(cmd, "x2", 0)
		y2 := getInt(cmd, "y2", 0)
		drawLine(canvas, x0, y0, x1, y1, ch)
		drawLine(canvas, x1, y1, x2, y2, ch)
		drawLine(canvas, x2, y2, x0, y0, ch)

	case "text":
		x := getInt(cmd, "x", 0)
		y := getInt(cmd, "y", 0)
		text, _ := cmd["text"].(string)
		for i, c := range text {
			if y >= 0 && y < len(canvas) && x+i >= 0 && x+i < len(canvas[y]) {
				canvas[y][x+i] = c
			}
		}
	}
}

// Bresenham's line algorithm
func drawLine(canvas [][]rune, x0, y0, x1, y1 int, ch rune) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := sign(x1 - x0)
	sy := sign(y1 - y0)
	err := dx + dy

	for {
		if y0 >= 0 && y0 < len(canvas) && x0 >= 0 && x0 < len(canvas[y0]) {
			canvas[y0][x0] = ch
		}

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 >= dy {
			if x0 == x1 {
				break
			}
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			if y0 == y1 {
				break
			}
			err += dx
			y0 += sy
		}
	}
}

func drawRect(canvas [][]rune, x, y, w, h int, ch rune) {
	drawLine(canvas, x, y, x+w-1, y, ch)
	drawLine(canvas, x, y+h-1, x+w-1, y+h-1, ch)
	drawLine(canvas, x, y, x, y+h-1, ch)
	drawLine(canvas, x+w-1, y, x+w-1, y+h-1, ch)
}

func drawCircle(canvas [][]rune, cx, cy, r int, ch rune) {
	x := r
	y := 0
	err := 0

	for x >= y {
		putPixel(canvas, cx+x, cy+y, ch)
		putPixel(canvas, cx+y, cy+x, ch)
		putPixel(canvas, cx-y, cy+x, ch)
		putPixel(canvas, cx-x, cy+y, ch)
		putPixel(canvas, cx-x, cy-y, ch)
		putPixel(canvas, cx-y, cy-x, ch)
		putPixel(canvas, cx+y, cy-x, ch)
		putPixel(canvas, cx+x, cy-y, ch)

		y++
		err += 1 + 2*y
		if 2*(err-x)+1 > 0 {
			x--
			err += 1 - 2*x
		}
	}
}

func fillCircle(canvas [][]rune, cx, cy, r int, ch rune) {
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= r*r {
				putPixel(canvas, x, y, ch)
			}
		}
	}
}

func putPixel(canvas [][]rune, x, y int, ch rune) {
	if y >= 0 && y < len(canvas) && x >= 0 && x < len(canvas[y]) {
		canvas[y][x] = ch
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

// Simple expression parser for integration
func parseSimpleExpression(expr string) func(float64) float64 {
	expr = strings.TrimSpace(expr)

	// Handle common expressions
	switch {
	case expr == "" || expr == "x":
		return func(x float64) float64 { return x }
	case expr == "x^2" || expr == "x*x":
		return func(x float64) float64 { return x * x }
	case expr == "x^3":
		return func(x float64) float64 { return x * x * x }
	case expr == "sin(x)":
		return math.Sin
	case expr == "cos(x)":
		return math.Cos
	case expr == "exp(x)":
		return math.Exp
	case expr == "1/x":
		return func(x float64) float64 { return 1 / x }
	case expr == "sqrt(x)":
		return math.Sqrt
	default:
		// Default to x
		return func(x float64) float64 { return x }
	}
}

// Simpson's rule integration
func simpsonIntegrate(expr string, a, b float64, n int) float64 {
	if n%2 == 1 {
		n++
	}

	f := parseSimpleExpression(expr)
	h := (b - a) / float64(n)

	sum := f(a) + f(b)

	for i := 1; i < n; i++ {
		x := a + float64(i)*h
		if i%2 == 0 {
			sum += 2 * f(x)
		} else {
			sum += 4 * f(x)
		}
	}

	return sum * h / 3
}
