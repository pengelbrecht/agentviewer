# LaTeX Math Test File - Basic Formulas

Test file for verifying KaTeX rendering of basic mathematical expressions.

## Inline Math

Basic inline math: $x = 5$ and $y = 10$.

Variables and operations: $a + b = c$, $x - y$, $m \times n$, $p \div q$.

Greek letters: $\alpha$, $\beta$, $\gamma$, $\delta$, $\epsilon$, $\theta$, $\lambda$, $\mu$, $\pi$, $\sigma$, $\omega$.

Uppercase Greek: $\Gamma$, $\Delta$, $\Theta$, $\Lambda$, $\Pi$, $\Sigma$, $\Omega$.

Subscripts: $x_1$, $x_2$, $a_i$, $b_{ij}$, $c_{n+1}$.

Superscripts: $x^2$, $y^3$, $e^x$, $2^{10}$, $a^{b+c}$.

Combined: $x_i^2$, $a_{ij}^{kl}$, $\sum_i^n$.

## Display Math (Block Equations)

The quadratic formula:
$$x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}$$

Euler's identity:
$$e^{i\pi} + 1 = 0$$

Pythagorean theorem:
$$a^2 + b^2 = c^2$$

Definition of e:
$$e = \lim_{n \to \infty} \left(1 + \frac{1}{n}\right)^n$$

## Fractions

Simple: $\frac{1}{2}$, $\frac{a}{b}$, $\frac{x+1}{x-1}$.

Nested: $\frac{1}{1 + \frac{1}{x}}$.

Complex:
$$\frac{\frac{a}{b}}{\frac{c}{d}} = \frac{ad}{bc}$$

## Square Roots

Basic: $\sqrt{2}$, $\sqrt{x}$, $\sqrt{a+b}$.

Higher roots: $\sqrt[3]{8}$, $\sqrt[n]{x}$, $\sqrt[4]{16}$.

Nested: $\sqrt{1 + \sqrt{2 + \sqrt{3}}}$.

## Sums and Products

Sum notation: $\sum_{i=1}^{n} i = \frac{n(n+1)}{2}$.

Product notation: $\prod_{i=1}^{n} i = n!$.

Display sum:
$$\sum_{k=0}^{\infty} \frac{x^k}{k!} = e^x$$

Display product:
$$\prod_{p \text{ prime}} \frac{1}{1-p^{-s}} = \sum_{n=1}^{\infty} \frac{1}{n^s}$$

## Integrals

Indefinite: $\int x^2 \, dx = \frac{x^3}{3} + C$.

Definite: $\int_0^1 x \, dx = \frac{1}{2}$.

Multiple:
$$\iint_D f(x,y) \, dA$$

$$\iiint_V f(x,y,z) \, dV$$

Contour integral: $\oint_C f(z) \, dz$.

## Limits

Basic: $\lim_{x \to 0} \frac{\sin x}{x} = 1$.

Infinity: $\lim_{n \to \infty} \frac{1}{n} = 0$.

Display:
$$\lim_{h \to 0} \frac{f(x+h) - f(x)}{h} = f'(x)$$

## Derivatives

Prime notation: $f'(x)$, $f''(x)$, $f'''(x)$.

Leibniz notation: $\frac{dy}{dx}$, $\frac{d^2y}{dx^2}$.

Partial derivatives: $\frac{\partial f}{\partial x}$, $\frac{\partial^2 f}{\partial x \partial y}$.

Nabla: $\nabla f = \left(\frac{\partial f}{\partial x}, \frac{\partial f}{\partial y}, \frac{\partial f}{\partial z}\right)$.

## Basic Sets and Logic

Set notation: $\{1, 2, 3\}$, $\{x \mid x > 0\}$.

Membership: $x \in A$, $y \notin B$.

Subset: $A \subset B$, $A \subseteq B$.

Union/intersection: $A \cup B$, $A \cap B$.

Set operations: $A \setminus B$, $\overline{A}$.

Logic: $p \land q$, $p \lor q$, $\neg p$, $p \implies q$, $p \iff q$.

Quantifiers: $\forall x$, $\exists y$, $\nexists z$.
