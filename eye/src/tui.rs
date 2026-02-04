//! TUI Rendering Primitives
//!
//! "We don't render graphics... only dots"
//!
//! Building blocks for a terminal-based game engine.
//! Everything is ASCII/Unicode characters at integer coordinates.

use crate::{Scalar, Vec2};

/// Character-based pixel
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Pixel {
    pub ch: char,
    pub fg: Color,
    pub bg: Color,
    pub bold: bool,
}

impl Pixel {
    pub const EMPTY: Pixel = Pixel {
        ch: ' ',
        fg: Color::Default,
        bg: Color::Default,
        bold: false,
    };

    pub fn new(ch: char) -> Self {
        Pixel {
            ch,
            fg: Color::Default,
            bg: Color::Default,
            bold: false,
        }
    }

    pub fn with_fg(mut self, color: Color) -> Self {
        self.fg = color;
        self
    }

    pub fn with_bg(mut self, color: Color) -> Self {
        self.bg = color;
        self
    }

    pub fn bold(mut self) -> Self {
        self.bold = true;
        self
    }
}

/// Terminal color
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Color {
    Default,
    Black,
    Red,
    Green,
    Yellow,
    Blue,
    Magenta,
    Cyan,
    White,
    BrightBlack,
    BrightRed,
    BrightGreen,
    BrightYellow,
    BrightBlue,
    BrightMagenta,
    BrightCyan,
    BrightWhite,
    Rgb(u8, u8, u8),
    Indexed(u8),
}

impl Color {
    /// Convert to ANSI escape code (foreground)
    pub fn to_fg_ansi(&self) -> String {
        match self {
            Color::Default => "\x1b[39m".to_string(),
            Color::Black => "\x1b[30m".to_string(),
            Color::Red => "\x1b[31m".to_string(),
            Color::Green => "\x1b[32m".to_string(),
            Color::Yellow => "\x1b[33m".to_string(),
            Color::Blue => "\x1b[34m".to_string(),
            Color::Magenta => "\x1b[35m".to_string(),
            Color::Cyan => "\x1b[36m".to_string(),
            Color::White => "\x1b[37m".to_string(),
            Color::BrightBlack => "\x1b[90m".to_string(),
            Color::BrightRed => "\x1b[91m".to_string(),
            Color::BrightGreen => "\x1b[92m".to_string(),
            Color::BrightYellow => "\x1b[93m".to_string(),
            Color::BrightBlue => "\x1b[94m".to_string(),
            Color::BrightMagenta => "\x1b[95m".to_string(),
            Color::BrightCyan => "\x1b[96m".to_string(),
            Color::BrightWhite => "\x1b[97m".to_string(),
            Color::Rgb(r, g, b) => format!("\x1b[38;2;{};{};{}m", r, g, b),
            Color::Indexed(i) => format!("\x1b[38;5;{}m", i),
        }
    }

    /// Convert to ANSI escape code (background)
    pub fn to_bg_ansi(&self) -> String {
        match self {
            Color::Default => "\x1b[49m".to_string(),
            Color::Black => "\x1b[40m".to_string(),
            Color::Red => "\x1b[41m".to_string(),
            Color::Green => "\x1b[42m".to_string(),
            Color::Yellow => "\x1b[43m".to_string(),
            Color::Blue => "\x1b[44m".to_string(),
            Color::Magenta => "\x1b[45m".to_string(),
            Color::Cyan => "\x1b[46m".to_string(),
            Color::White => "\x1b[47m".to_string(),
            Color::BrightBlack => "\x1b[100m".to_string(),
            Color::BrightRed => "\x1b[101m".to_string(),
            Color::BrightGreen => "\x1b[102m".to_string(),
            Color::BrightYellow => "\x1b[103m".to_string(),
            Color::BrightBlue => "\x1b[104m".to_string(),
            Color::BrightMagenta => "\x1b[105m".to_string(),
            Color::BrightCyan => "\x1b[106m".to_string(),
            Color::BrightWhite => "\x1b[107m".to_string(),
            Color::Rgb(r, g, b) => format!("\x1b[48;2;{};{};{}m", r, g, b),
            Color::Indexed(i) => format!("\x1b[48;5;{}m", i),
        }
    }
}

/// Character canvas for terminal rendering
#[derive(Debug, Clone)]
pub struct Canvas {
    pub width: usize,
    pub height: usize,
    pixels: Vec<Pixel>,
}

impl Canvas {
    pub fn new(width: usize, height: usize) -> Self {
        Self {
            width,
            height,
            pixels: vec![Pixel::EMPTY; width * height],
        }
    }

    pub fn clear(&mut self) {
        self.pixels.fill(Pixel::EMPTY);
    }

    pub fn fill(&mut self, pixel: Pixel) {
        self.pixels.fill(pixel);
    }

    #[inline]
    pub fn in_bounds(&self, x: i32, y: i32) -> bool {
        x >= 0 && y >= 0 && (x as usize) < self.width && (y as usize) < self.height
    }

    #[inline]
    pub fn get(&self, x: i32, y: i32) -> Option<Pixel> {
        if self.in_bounds(x, y) {
            Some(self.pixels[y as usize * self.width + x as usize])
        } else {
            None
        }
    }

    #[inline]
    pub fn set(&mut self, x: i32, y: i32, pixel: Pixel) {
        if self.in_bounds(x, y) {
            self.pixels[y as usize * self.width + x as usize] = pixel;
        }
    }

    #[inline]
    pub fn put(&mut self, x: i32, y: i32, ch: char) {
        self.set(x, y, Pixel::new(ch));
    }

    /// Draw a character with color
    pub fn put_colored(&mut self, x: i32, y: i32, ch: char, fg: Color, bg: Color) {
        self.set(x, y, Pixel::new(ch).with_fg(fg).with_bg(bg));
    }

    /// Draw a string starting at (x, y)
    pub fn put_str(&mut self, x: i32, y: i32, s: &str) {
        for (i, ch) in s.chars().enumerate() {
            self.put(x + i as i32, y, ch);
        }
    }

    /// Draw a point
    pub fn point(&mut self, x: i32, y: i32, ch: char) {
        self.put(x, y, ch);
    }

    /// Draw a line using Bresenham's algorithm
    pub fn line(&mut self, x0: i32, y0: i32, x1: i32, y1: i32, ch: char) {
        let dx = (x1 - x0).abs();
        let dy = -(y1 - y0).abs();
        let sx = if x0 < x1 { 1 } else { -1 };
        let sy = if y0 < y1 { 1 } else { -1 };
        let mut err = dx + dy;

        let mut x = x0;
        let mut y = y0;

        loop {
            self.put(x, y, ch);

            if x == x1 && y == y1 {
                break;
            }

            let e2 = 2 * err;
            if e2 >= dy {
                if x == x1 {
                    break;
                }
                err += dy;
                x += sx;
            }
            if e2 <= dx {
                if y == y1 {
                    break;
                }
                err += dx;
                y += sy;
            }
        }
    }

    /// Draw a line using better ASCII characters
    pub fn line_styled(&mut self, x0: i32, y0: i32, x1: i32, y1: i32) {
        let dx = x1 - x0;
        let dy = y1 - y0;

        // Choose character based on angle
        let ch = if dx == 0 {
            '│'
        } else if dy == 0 {
            '─'
        } else {
            let angle = (dy as Scalar).atan2(dx as Scalar);
            if angle.abs() < std::f64::consts::PI / 4.0 {
                '─'
            } else if angle.abs() > 3.0 * std::f64::consts::PI / 4.0 {
                '─'
            } else if angle > 0.0 {
                if (angle - std::f64::consts::PI / 4.0).abs() < std::f64::consts::PI / 8.0 {
                    '╲'
                } else {
                    '│'
                }
            } else if (angle + std::f64::consts::PI / 4.0).abs() < std::f64::consts::PI / 8.0 {
                '╱'
            } else {
                '│'
            }
        };

        self.line(x0, y0, x1, y1, ch);
    }

    /// Draw a rectangle outline
    pub fn rect(&mut self, x: i32, y: i32, w: i32, h: i32, ch: char) {
        self.line(x, y, x + w - 1, y, ch);
        self.line(x, y + h - 1, x + w - 1, y + h - 1, ch);
        self.line(x, y, x, y + h - 1, ch);
        self.line(x + w - 1, y, x + w - 1, y + h - 1, ch);
    }

    /// Draw a box with box-drawing characters
    pub fn box_draw(&mut self, x: i32, y: i32, w: i32, h: i32) {
        // Corners
        self.put(x, y, '┌');
        self.put(x + w - 1, y, '┐');
        self.put(x, y + h - 1, '└');
        self.put(x + w - 1, y + h - 1, '┘');

        // Horizontal lines
        for i in 1..(w - 1) {
            self.put(x + i, y, '─');
            self.put(x + i, y + h - 1, '─');
        }

        // Vertical lines
        for i in 1..(h - 1) {
            self.put(x, y + i, '│');
            self.put(x + w - 1, y + i, '│');
        }
    }

    /// Fill a rectangle
    pub fn fill_rect(&mut self, x: i32, y: i32, w: i32, h: i32, ch: char) {
        for dy in 0..h {
            for dx in 0..w {
                self.put(x + dx, y + dy, ch);
            }
        }
    }

    /// Draw a circle using midpoint algorithm
    pub fn circle(&mut self, cx: i32, cy: i32, r: i32, ch: char) {
        let mut x = r;
        let mut y = 0;
        let mut err = 0;

        while x >= y {
            self.put(cx + x, cy + y, ch);
            self.put(cx + y, cy + x, ch);
            self.put(cx - y, cy + x, ch);
            self.put(cx - x, cy + y, ch);
            self.put(cx - x, cy - y, ch);
            self.put(cx - y, cy - x, ch);
            self.put(cx + y, cy - x, ch);
            self.put(cx + x, cy - y, ch);

            y += 1;
            err += 1 + 2 * y;
            if 2 * (err - x) + 1 > 0 {
                x -= 1;
                err += 1 - 2 * x;
            }
        }
    }

    /// Fill a circle
    pub fn fill_circle(&mut self, cx: i32, cy: i32, r: i32, ch: char) {
        for y in (cy - r)..=(cy + r) {
            for x in (cx - r)..=(cx + r) {
                let dx = x - cx;
                let dy = y - cy;
                if dx * dx + dy * dy <= r * r {
                    self.put(x, y, ch);
                }
            }
        }
    }

    /// Draw an ellipse
    pub fn ellipse(&mut self, cx: i32, cy: i32, rx: i32, ry: i32, ch: char) {
        let rx2 = rx * rx;
        let ry2 = ry * ry;
        let two_rx2 = 2 * rx2;
        let two_ry2 = 2 * ry2;

        let mut x = 0;
        let mut y = ry;
        let mut px = 0;
        let mut py = two_rx2 * y;

        // Region 1
        let mut p = ry2 - rx2 * ry + rx2 / 4;
        while px < py {
            self.put(cx + x, cy + y, ch);
            self.put(cx - x, cy + y, ch);
            self.put(cx + x, cy - y, ch);
            self.put(cx - x, cy - y, ch);

            x += 1;
            px += two_ry2;
            if p < 0 {
                p += ry2 + px;
            } else {
                y -= 1;
                py -= two_rx2;
                p += ry2 + px - py;
            }
        }

        // Region 2
        p = ry2 * (x * x + x) + rx2 * (y - 1) * (y - 1) - rx2 * ry2;
        while y >= 0 {
            self.put(cx + x, cy + y, ch);
            self.put(cx - x, cy + y, ch);
            self.put(cx + x, cy - y, ch);
            self.put(cx - x, cy - y, ch);

            y -= 1;
            py -= two_rx2;
            if p > 0 {
                p += rx2 - py;
            } else {
                x += 1;
                px += two_ry2;
                p += rx2 - py + px;
            }
        }
    }

    /// Draw a triangle outline
    pub fn triangle(&mut self, x0: i32, y0: i32, x1: i32, y1: i32, x2: i32, y2: i32, ch: char) {
        self.line(x0, y0, x1, y1, ch);
        self.line(x1, y1, x2, y2, ch);
        self.line(x2, y2, x0, y0, ch);
    }

    /// Fill a triangle (scanline algorithm)
    pub fn fill_triangle(&mut self, x0: i32, y0: i32, x1: i32, y1: i32, x2: i32, y2: i32, ch: char) {
        // Sort vertices by y
        let mut v = [(x0, y0), (x1, y1), (x2, y2)];
        v.sort_by_key(|p| p.1);
        let [(x0, y0), (x1, y1), (x2, y2)] = v;

        fn interpolate(y: i32, y0: i32, x0: i32, y1: i32, x1: i32) -> i32 {
            if y1 == y0 {
                x0
            } else {
                x0 + (x1 - x0) * (y - y0) / (y1 - y0)
            }
        }

        for y in y0..=y2 {
            let xa = interpolate(y, y0, x0, y2, x2);
            let xb = if y < y1 {
                interpolate(y, y0, x0, y1, x1)
            } else {
                interpolate(y, y1, x1, y2, x2)
            };

            let (x_start, x_end) = if xa < xb { (xa, xb) } else { (xb, xa) };
            for x in x_start..=x_end {
                self.put(x, y, ch);
            }
        }
    }

    /// Draw a Bezier curve
    pub fn bezier(&mut self, points: &[Vec2], ch: char, steps: usize) {
        if points.len() < 2 {
            return;
        }

        let bezier = crate::curve::BezierCurve::new(
            points.iter().map(|p| crate::Vec3::new(p.x, p.y, 0.0)).collect()
        );

        let mut prev: Option<(i32, i32)> = None;
        for i in 0..=steps {
            let t = i as Scalar / steps as Scalar;
            let p = bezier.evaluate(t);
            let x = p.x.round() as i32;
            let y = p.y.round() as i32;

            if let Some((px, py)) = prev {
                self.line(px, py, x, y, ch);
            }
            prev = Some((x, y));
        }
    }

    /// Flood fill
    pub fn flood_fill(&mut self, x: i32, y: i32, new_ch: char) {
        if !self.in_bounds(x, y) {
            return;
        }

        let old = self.get(x, y).unwrap();
        if old.ch == new_ch {
            return;
        }

        let old_ch = old.ch;
        let mut stack = vec![(x, y)];

        while let Some((cx, cy)) = stack.pop() {
            if !self.in_bounds(cx, cy) {
                continue;
            }
            if let Some(p) = self.get(cx, cy) {
                if p.ch == old_ch {
                    self.put(cx, cy, new_ch);
                    stack.push((cx + 1, cy));
                    stack.push((cx - 1, cy));
                    stack.push((cx, cy + 1));
                    stack.push((cx, cy - 1));
                }
            }
        }
    }

    /// Render canvas to string with ANSI codes
    pub fn render(&self) -> String {
        let mut out = String::with_capacity(self.width * self.height * 2);
        let mut prev_fg = Color::Default;
        let mut prev_bg = Color::Default;

        for y in 0..self.height {
            for x in 0..self.width {
                let p = self.pixels[y * self.width + x];

                if p.fg != prev_fg {
                    out.push_str(&p.fg.to_fg_ansi());
                    prev_fg = p.fg;
                }
                if p.bg != prev_bg {
                    out.push_str(&p.bg.to_bg_ansi());
                    prev_bg = p.bg;
                }
                if p.bold {
                    out.push_str("\x1b[1m");
                }
                out.push(p.ch);
                if p.bold {
                    out.push_str("\x1b[22m");
                }
            }
            out.push('\n');
        }

        // Reset colors at end
        out.push_str("\x1b[0m");
        out
    }

    /// Render to plain string (no ANSI)
    pub fn render_plain(&self) -> String {
        let mut out = String::with_capacity(self.width * self.height + self.height);
        for y in 0..self.height {
            for x in 0..self.width {
                out.push(self.pixels[y * self.width + x].ch);
            }
            out.push('\n');
        }
        out
    }
}

/// Sprite - a small canvas that can be drawn onto another canvas
#[derive(Debug, Clone)]
pub struct Sprite {
    pub canvas: Canvas,
    pub x: i32,
    pub y: i32,
    pub visible: bool,
}

impl Sprite {
    pub fn new(width: usize, height: usize) -> Self {
        Self {
            canvas: Canvas::new(width, height),
            x: 0,
            y: 0,
            visible: true,
        }
    }

    pub fn from_str(s: &str) -> Self {
        let lines: Vec<&str> = s.lines().collect();
        let height = lines.len();
        let width = lines.iter().map(|l| l.chars().count()).max().unwrap_or(0);

        let mut sprite = Self::new(width, height);
        for (y, line) in lines.iter().enumerate() {
            for (x, ch) in line.chars().enumerate() {
                sprite.canvas.put(x as i32, y as i32, ch);
            }
        }
        sprite
    }

    /// Draw sprite onto canvas
    pub fn draw_onto(&self, canvas: &mut Canvas) {
        if !self.visible {
            return;
        }

        for y in 0..self.canvas.height {
            for x in 0..self.canvas.width {
                let p = self.canvas.pixels[y * self.canvas.width + x];
                if p.ch != ' ' {
                    canvas.set(self.x + x as i32, self.y + y as i32, p);
                }
            }
        }
    }
}

/// Block characters for higher resolution rendering
pub mod blocks {
    /// Unicode block characters for sub-character rendering
    pub const FULL: char = '█';
    pub const UPPER_HALF: char = '▀';
    pub const LOWER_HALF: char = '▄';
    pub const LEFT_HALF: char = '▌';
    pub const RIGHT_HALF: char = '▐';

    /// Shade characters
    pub const LIGHT_SHADE: char = '░';
    pub const MEDIUM_SHADE: char = '▒';
    pub const DARK_SHADE: char = '▓';

    /// Quarter blocks
    pub const LOWER_LEFT: char = '▖';
    pub const LOWER_RIGHT: char = '▗';
    pub const UPPER_LEFT: char = '▘';
    pub const UPPER_RIGHT: char = '▝';
    pub const DIAGONAL_UPPER_LEFT: char = '▛';
    pub const DIAGONAL_UPPER_RIGHT: char = '▜';
    pub const DIAGONAL_LOWER_LEFT: char = '▙';
    pub const DIAGONAL_LOWER_RIGHT: char = '▟';

    /// Braille patterns for 2x4 sub-pixel rendering
    /// Each character represents an 2x4 grid of dots
    pub fn braille(dots: [[bool; 2]; 4]) -> char {
        let mut code: u32 = 0x2800; // Braille base
        if dots[0][0] { code |= 0x01; }
        if dots[1][0] { code |= 0x02; }
        if dots[2][0] { code |= 0x04; }
        if dots[0][1] { code |= 0x08; }
        if dots[1][1] { code |= 0x10; }
        if dots[2][1] { code |= 0x20; }
        if dots[3][0] { code |= 0x40; }
        if dots[3][1] { code |= 0x80; }
        char::from_u32(code).unwrap_or(' ')
    }
}

/// High-resolution canvas using Braille characters
/// Provides 2x4 sub-pixel resolution per character cell
#[derive(Debug, Clone)]
pub struct BrailleCanvas {
    width: usize,  // Character width
    height: usize, // Character height
    pixels: Vec<bool>, // Raw pixel data (width*2 x height*4)
}

impl BrailleCanvas {
    pub fn new(char_width: usize, char_height: usize) -> Self {
        Self {
            width: char_width,
            height: char_height,
            pixels: vec![false; char_width * 2 * char_height * 4],
        }
    }

    /// Pixel width (2x character width)
    pub fn pixel_width(&self) -> usize {
        self.width * 2
    }

    /// Pixel height (4x character height)
    pub fn pixel_height(&self) -> usize {
        self.height * 4
    }

    pub fn clear(&mut self) {
        self.pixels.fill(false);
    }

    #[inline]
    fn in_bounds(&self, x: i32, y: i32) -> bool {
        x >= 0 && y >= 0 && (x as usize) < self.pixel_width() && (y as usize) < self.pixel_height()
    }

    #[inline]
    pub fn set(&mut self, x: i32, y: i32, on: bool) {
        let pw = self.width * 2;
        let ph = self.height * 4;
        if x >= 0 && y >= 0 && (x as usize) < pw && (y as usize) < ph {
            self.pixels[y as usize * pw + x as usize] = on;
        }
    }

    #[inline]
    pub fn get(&self, x: i32, y: i32) -> bool {
        let pw = self.width * 2;
        let ph = self.height * 4;
        if x >= 0 && y >= 0 && (x as usize) < pw && (y as usize) < ph {
            self.pixels[y as usize * pw + x as usize]
        } else {
            false
        }
    }

    pub fn point(&mut self, x: i32, y: i32) {
        self.set(x, y, true);
    }

    pub fn line(&mut self, x0: i32, y0: i32, x1: i32, y1: i32) {
        let dx = (x1 - x0).abs();
        let dy = -(y1 - y0).abs();
        let sx = if x0 < x1 { 1 } else { -1 };
        let sy = if y0 < y1 { 1 } else { -1 };
        let mut err = dx + dy;
        let mut x = x0;
        let mut y = y0;

        loop {
            self.point(x, y);
            if x == x1 && y == y1 {
                break;
            }
            let e2 = 2 * err;
            if e2 >= dy {
                if x == x1 { break; }
                err += dy;
                x += sx;
            }
            if e2 <= dx {
                if y == y1 { break; }
                err += dx;
                y += sy;
            }
        }
    }

    pub fn circle(&mut self, cx: i32, cy: i32, r: i32) {
        let mut x = r;
        let mut y = 0;
        let mut err = 0;

        while x >= y {
            self.point(cx + x, cy + y);
            self.point(cx + y, cy + x);
            self.point(cx - y, cy + x);
            self.point(cx - x, cy + y);
            self.point(cx - x, cy - y);
            self.point(cx - y, cy - x);
            self.point(cx + y, cy - x);
            self.point(cx + x, cy - y);

            y += 1;
            err += 1 + 2 * y;
            if 2 * (err - x) + 1 > 0 {
                x -= 1;
                err += 1 - 2 * x;
            }
        }
    }

    /// Render to Canvas
    pub fn to_canvas(&self) -> Canvas {
        let mut canvas = Canvas::new(self.width, self.height);

        for cy in 0..self.height {
            for cx in 0..self.width {
                let mut dots = [[false; 2]; 4];
                for dy in 0..4 {
                    for dx in 0..2 {
                        let px = cx * 2 + dx;
                        let py = cy * 4 + dy;
                        dots[dy][dx] = self.pixels[py * self.pixel_width() + px];
                    }
                }
                canvas.put(cx as i32, cy as i32, blocks::braille(dots));
            }
        }

        canvas
    }

    /// Render to string
    pub fn render(&self) -> String {
        self.to_canvas().render_plain()
    }
}
