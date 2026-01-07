# LaTeX Math Test File - Advanced Formulas

Test file for verifying KaTeX rendering of advanced mathematical expressions.

## Matrices and Vectors

Row vector: $\vec{v} = (v_1, v_2, v_3)$.

Column vector:
$$\mathbf{v} = \begin{pmatrix} v_1 \\ v_2 \\ v_3 \end{pmatrix}$$

Matrix with parentheses:
$$A = \begin{pmatrix} a & b \\ c & d \end{pmatrix}$$

Matrix with brackets:
$$B = \begin{bmatrix} 1 & 2 & 3 \\ 4 & 5 & 6 \end{bmatrix}$$

Matrix with vertical bars (determinant):
$$\det(A) = \begin{vmatrix} a & b \\ c & d \end{vmatrix} = ad - bc$$

Identity matrix:
$$I_3 = \begin{bmatrix} 1 & 0 & 0 \\ 0 & 1 & 0 \\ 0 & 0 & 1 \end{bmatrix}$$

Augmented matrix:
$$\left[\begin{array}{cc|c} 1 & 2 & 5 \\ 3 & 4 & 6 \end{array}\right]$$

## Piecewise Functions

Absolute value:
$$|x| = \begin{cases} x & \text{if } x \geq 0 \\ -x & \text{if } x < 0 \end{cases}$$

Sign function:
$$\text{sgn}(x) = \begin{cases} -1 & x < 0 \\ 0 & x = 0 \\ 1 & x > 0 \end{cases}$$

Heaviside step:
$$H(x) = \begin{cases} 0 & x < 0 \\ \frac{1}{2} & x = 0 \\ 1 & x > 0 \end{cases}$$

## Systems of Equations

Linear system:
$$\begin{cases} 2x + 3y = 7 \\ x - y = 1 \end{cases}$$

Aligned system:
$$\begin{aligned} x + y + z &= 6 \\ 2x - y + z &= 3 \\ x + 2y - z &= 0 \end{aligned}$$

## Brackets and Delimiters

Auto-sizing: $\left(\frac{a}{b}\right)$, $\left[\frac{x^2}{y}\right]$, $\left\{\frac{1}{n}\right\}$.

Angle brackets: $\langle x, y \rangle$, $\left\langle\frac{a}{b}\right\rangle$.

Floor and ceiling: $\lfloor x \rfloor$, $\lceil y \rceil$, $\left\lfloor\frac{n}{2}\right\rfloor$.

Mixed: $\left(\frac{a}{b}\right]$, $\left[0, 1\right)$.

Nested:
$$\left[\left(\frac{a^2 + b^2}{c^2}\right)^{1/2}\right]^n$$

## Arrows and Relations

Arrows: $\to$, $\leftarrow$, $\leftrightarrow$, $\Rightarrow$, $\Leftarrow$, $\Leftrightarrow$.

Long arrows: $\longrightarrow$, $\longleftrightarrow$.

Maps to: $x \mapsto f(x)$.

Relations: $<$, $>$, $\leq$, $\geq$, $\neq$, $\approx$, $\equiv$, $\cong$, $\sim$, $\propto$.

Ordering: $\ll$, $\gg$, $\prec$, $\succ$.

## Operators and Functions

Trigonometric: $\sin\theta$, $\cos\theta$, $\tan\theta$, $\sec\theta$, $\csc\theta$, $\cot\theta$.

Inverse trig: $\arcsin x$, $\arccos x$, $\arctan x$.

Hyperbolic: $\sinh x$, $\cosh x$, $\tanh x$.

Logarithms: $\log x$, $\ln x$, $\log_2 x$, $\log_{10} x$.

Exponential: $\exp(x)$, $e^x$.

Other: $\min$, $\max$, $\sup$, $\inf$, $\lim$, $\gcd$, $\det$, $\dim$, $\ker$, $\arg$.

## Accents and Decorations

Hats: $\hat{x}$, $\widehat{ABC}$.

Bars: $\bar{x}$, $\overline{AB}$.

Dots: $\dot{x}$, $\ddot{x}$, $\dddot{x}$.

Tildes: $\tilde{x}$, $\widetilde{ABC}$.

Vectors: $\vec{v}$, $\overrightarrow{AB}$.

Underbrace/overbrace:
$$\overbrace{1 + 2 + \cdots + n}^{n \text{ terms}} = \frac{n(n+1)}{2}$$

$$\underbrace{a + a + \cdots + a}_{n \text{ times}} = na$$

## Spacing and Formatting

Different spaces: $a\,b$, $a\;b$, $a\quad b$, $a\qquad b$.

Text in math: $x = 5 \text{ when } y > 0$.

Bold math: $\mathbf{v}$, $\mathbf{A}$, $\boldsymbol{\alpha}$.

Script: $\mathcal{L}$, $\mathcal{F}$, $\mathcal{H}$.

Blackboard bold: $\mathbb{N}$, $\mathbb{Z}$, $\mathbb{Q}$, $\mathbb{R}$, $\mathbb{C}$.

Fraktur: $\mathfrak{g}$, $\mathfrak{A}$.

## Big Operators

Large sum:
$$\sum_{i=1}^{n} a_i = a_1 + a_2 + \cdots + a_n$$

Large product:
$$\prod_{i=1}^{n} a_i = a_1 \cdot a_2 \cdots a_n$$

Large union:
$$\bigcup_{i=1}^{n} A_i = A_1 \cup A_2 \cup \cdots \cup A_n$$

Large intersection:
$$\bigcap_{i=1}^{n} A_i = A_1 \cap A_2 \cap \cdots \cap A_n$$

Coprod:
$$\coprod_{i \in I} X_i$$

## Dots and Ellipsis

Horizontal dots: $1, 2, \ldots, n$ or $1 + 2 + \cdots + n$.

Vertical dots:
$$\begin{pmatrix} a_1 \\ a_2 \\ \vdots \\ a_n \end{pmatrix}$$

Diagonal dots:
$$\begin{pmatrix} a_{11} & \cdots & a_{1n} \\ \vdots & \ddots & \vdots \\ a_{m1} & \cdots & a_{mn} \end{pmatrix}$$
