# LaTeX Math Test File - Edge Cases

Test file for edge cases and potential rendering issues in KaTeX.

## Empty and Minimal Expressions

Empty braces: ${}$.

Single character: $x$.

Single number: $5$.

Single operator: $+$.

Just braces: $\{\ \}$.

## Escaped Characters

Dollar signs in text: The price is \$5.00.

Percent sign: $100\%$.

Ampersand: $A \& B$.

Backslash: $\backslash$.

## Deeply Nested Expressions

Triple nested fraction:
$$\frac{1}{\frac{1}{\frac{1}{x}}}$$

Deeply nested radicals:
$$\sqrt{\sqrt{\sqrt{\sqrt{x}}}}$$

Nested subscripts/superscripts: $x_{a_{b_{c}}}$, $y^{p^{q^{r}}}$.

Complex nesting:
$$\left(\frac{\sqrt{1+\left(\frac{a}{b}\right)^2}}{\left(\frac{c}{d}\right)^{n+1}}\right)^{1/3}$$

## Very Long Expressions

Long polynomial:
$$f(x) = a_0 + a_1 x + a_2 x^2 + a_3 x^3 + a_4 x^4 + a_5 x^5 + a_6 x^6 + a_7 x^7 + a_8 x^8 + a_9 x^9 + a_{10} x^{10}$$

Long continued fraction:
$$x = a_0 + \cfrac{1}{a_1 + \cfrac{1}{a_2 + \cfrac{1}{a_3 + \cfrac{1}{a_4}}}}$$

Very long sum:
$$\sum_{k=0}^{\infty} \sum_{j=0}^{k} \sum_{i=0}^{j} a_{ijk} x^i y^j z^k$$

## Large Matrices

5x5 matrix:
$$\begin{pmatrix}
a_{11} & a_{12} & a_{13} & a_{14} & a_{15} \\
a_{21} & a_{22} & a_{23} & a_{24} & a_{25} \\
a_{31} & a_{32} & a_{33} & a_{34} & a_{35} \\
a_{41} & a_{42} & a_{43} & a_{44} & a_{45} \\
a_{51} & a_{52} & a_{53} & a_{54} & a_{55}
\end{pmatrix}$$

Partitioned matrix:
$$\left(\begin{array}{c|c}
A & B \\
\hline
C & D
\end{array}\right)$$

## Mixed Inline and Display

Here's inline $x^2$ followed by display:
$$y = x^2$$
And then more inline $z = y + 1$ math.

Multiple inline in one line: We have $a = 1$, $b = 2$, $c = 3$, and therefore $a + b + c = 6$.

## Unicode in Math

Variables with accents: $\hat{x}$, $\tilde{y}$, $\bar{z}$.

Greek: $\alpha\beta\gamma\delta\epsilon\zeta\eta\theta\iota\kappa\lambda\mu\nu\xi\pi\rho\sigma\tau\upsilon\phi\chi\psi\omega$.

## Color (if supported)

Red text: $\color{red}{x^2}$.

Blue equation: $\color{blue}{\frac{a}{b}}$.

Mixed colors: $\color{red}{a} + \color{blue}{b} = \color{green}{c}$.

## Edge Case Spacing

No space: $xy$.

Thin space: $x\,y$.

Medium space: $x\;y$.

Thick space: $x\ y$.

Quad: $x\quad y$.

QQuad: $x\qquad y$.

Negative space: $x\!y$.

## Cancel and Strikethrough

Cancellation: $\cancel{x}$ (if cancel package loaded).

Crossing out: $\require{cancel}\cancel{abc}$.

## Commutative Diagrams (basic)

Simple diagram attempt:
$$A \xrightarrow{f} B \xrightarrow{g} C$$

With subscripts:
$$A \xrightarrow[\text{bottom}]{\text{top}} B$$

## Chemistry Notation

Chemical formula: $\ce{H2O}$ (if mhchem loaded, else $H_2O$).

Reaction: $\ce{2H2 + O2 -> 2H2O}$ or $2H_2 + O_2 \to 2H_2O$.

Equilibrium: $A + B \rightleftharpoons C + D$.

## Units (if supported)

With siunitx: $\SI{9.8}{m/s^2}$ or manually $9.8\,\text{m/s}^2$.

Temperature: $T = 300\,\text{K}$.

Energy: $E = 1.6 \times 10^{-19}\,\text{J}$.

## Error Cases (may not render)

These might cause rendering errors and test error handling:

Unbalanced braces (in code block to not break rendering):
```
$\frac{1}{2$
```

Missing argument:
```
$\sqrt{}$
```

Unknown command (may render as text):
$\unknowncommand$.

## Multiple Display Blocks

First block:
$$E = mc^2$$

Second block:
$$F = ma$$

Third block:
$$PV = nRT$$

## Alignment in Equations

Aligned equations:
$$\begin{aligned}
(x + y)^2 &= x^2 + 2xy + y^2 \\
(x - y)^2 &= x^2 - 2xy + y^2 \\
(x + y)(x - y) &= x^2 - y^2
\end{aligned}$$

## Equation Numbering (manual)

$$E = mc^2 \tag{1}$$

$$F = ma \tag{2}$$

## Mixed Content Stress Test

Given the function $f(x) = \int_0^x e^{-t^2}\, dt$, we can compute:

$$f'(x) = e^{-x^2}$$

And the Taylor series is:

$$f(x) = \sum_{n=0}^{\infty} \frac{(-1)^n x^{2n+1}}{n!(2n+1)}$$

For $x \in [-1, 1]$, we have $|f(x)| \leq 1$. The limit is:

$$\lim_{x \to \infty} f(x) = \frac{\sqrt{\pi}}{2}$$

This is related to the error function $\text{erf}(x) = \frac{2}{\sqrt{\pi}} f(x)$.
