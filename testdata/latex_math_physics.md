# LaTeX Math Test File - Physics and Applied Math

Test file for physics formulas and applied mathematics expressions.

## Classical Mechanics

Newton's second law: $\vec{F} = m\vec{a}$.

Kinetic energy: $K = \frac{1}{2}mv^2$.

Potential energy (gravity): $U = mgh$.

Work-energy theorem: $W = \Delta K = K_f - K_i$.

Momentum: $\vec{p} = m\vec{v}$.

Angular momentum: $\vec{L} = \vec{r} \times \vec{p}$.

Torque: $\vec{\tau} = \vec{r} \times \vec{F}$.

Lagrangian:
$$L = T - V = \frac{1}{2}m\dot{q}^2 - V(q)$$

Euler-Lagrange equation:
$$\frac{d}{dt}\left(\frac{\partial L}{\partial \dot{q}}\right) - \frac{\partial L}{\partial q} = 0$$

Hamiltonian:
$$H = \sum_i p_i \dot{q}_i - L$$

## Electromagnetism

Coulomb's law: $\vec{F} = k_e \frac{q_1 q_2}{r^2} \hat{r}$.

Electric field: $\vec{E} = \frac{\vec{F}}{q}$.

Gauss's law:
$$\oint_S \vec{E} \cdot d\vec{A} = \frac{Q_{\text{enc}}}{\epsilon_0}$$

Magnetic force: $\vec{F} = q\vec{v} \times \vec{B}$.

Faraday's law:
$$\oint_C \vec{E} \cdot d\vec{l} = -\frac{d\Phi_B}{dt}$$

Maxwell's equations:
$$\begin{aligned}
\nabla \cdot \vec{E} &= \frac{\rho}{\epsilon_0} \\
\nabla \cdot \vec{B} &= 0 \\
\nabla \times \vec{E} &= -\frac{\partial \vec{B}}{\partial t} \\
\nabla \times \vec{B} &= \mu_0 \vec{J} + \mu_0 \epsilon_0 \frac{\partial \vec{E}}{\partial t}
\end{aligned}$$

Wave equation:
$$\nabla^2 \vec{E} = \mu_0 \epsilon_0 \frac{\partial^2 \vec{E}}{\partial t^2}$$

## Thermodynamics

Ideal gas law: $PV = nRT$.

First law: $\Delta U = Q - W$.

Entropy change: $\Delta S = \int \frac{dQ}{T}$.

Boltzmann entropy: $S = k_B \ln \Omega$.

Helmholtz free energy: $F = U - TS$.

Gibbs free energy: $G = H - TS$.

Maxwell relation:
$$\left(\frac{\partial T}{\partial V}\right)_S = -\left(\frac{\partial P}{\partial S}\right)_V$$

## Quantum Mechanics

Schrödinger equation:
$$i\hbar \frac{\partial}{\partial t}|\Psi\rangle = \hat{H}|\Psi\rangle$$

Time-independent form:
$$\hat{H}\psi = E\psi$$

Momentum operator: $\hat{p} = -i\hbar \nabla$.

Position-momentum commutator: $[\hat{x}, \hat{p}] = i\hbar$.

Uncertainty principle:
$$\Delta x \Delta p \geq \frac{\hbar}{2}$$

Bra-ket notation: $\langle \phi | \psi \rangle$, $|\psi\rangle\langle\phi|$.

Expectation value: $\langle A \rangle = \langle \psi | \hat{A} | \psi \rangle$.

Harmonic oscillator:
$$E_n = \hbar\omega\left(n + \frac{1}{2}\right)$$

## Special Relativity

Lorentz factor: $\gamma = \frac{1}{\sqrt{1 - v^2/c^2}}$.

Time dilation: $\Delta t' = \gamma \Delta t$.

Length contraction: $L' = \frac{L}{\gamma}$.

Energy-momentum relation:
$$E^2 = (pc)^2 + (mc^2)^2$$

Rest energy: $E_0 = mc^2$.

Relativistic momentum: $\vec{p} = \gamma m \vec{v}$.

## Statistical Mechanics

Partition function:
$$Z = \sum_i e^{-\beta E_i}$$

where $\beta = \frac{1}{k_B T}$.

Boltzmann distribution:
$$P_i = \frac{e^{-\beta E_i}}{Z}$$

Average energy:
$$\langle E \rangle = -\frac{\partial \ln Z}{\partial \beta}$$

Fermi-Dirac distribution:
$$f(E) = \frac{1}{e^{(E-\mu)/k_B T} + 1}$$

Bose-Einstein distribution:
$$n(E) = \frac{1}{e^{(E-\mu)/k_B T} - 1}$$

## Differential Equations

Simple harmonic motion:
$$\frac{d^2x}{dt^2} + \omega^2 x = 0$$

General solution: $x(t) = A\cos(\omega t) + B\sin(\omega t)$.

Damped oscillator:
$$\frac{d^2x}{dt^2} + 2\gamma\frac{dx}{dt} + \omega_0^2 x = 0$$

Wave equation (1D):
$$\frac{\partial^2 u}{\partial t^2} = c^2 \frac{\partial^2 u}{\partial x^2}$$

Heat equation:
$$\frac{\partial u}{\partial t} = \alpha \nabla^2 u$$

Laplace's equation:
$$\nabla^2 \phi = 0$$

## Fourier Analysis

Fourier series:
$$f(x) = \frac{a_0}{2} + \sum_{n=1}^{\infty} \left(a_n \cos\frac{2\pi nx}{L} + b_n \sin\frac{2\pi nx}{L}\right)$$

Fourier coefficients:
$$a_n = \frac{2}{L}\int_0^L f(x)\cos\frac{2\pi nx}{L}\,dx$$

Fourier transform:
$$\hat{f}(\omega) = \int_{-\infty}^{\infty} f(t) e^{-i\omega t}\, dt$$

Inverse transform:
$$f(t) = \frac{1}{2\pi}\int_{-\infty}^{\infty} \hat{f}(\omega) e^{i\omega t}\, d\omega$$

Convolution theorem: $\widehat{f * g} = \hat{f} \cdot \hat{g}$.

Parseval's theorem:
$$\int_{-\infty}^{\infty} |f(t)|^2\, dt = \frac{1}{2\pi}\int_{-\infty}^{\infty} |\hat{f}(\omega)|^2\, d\omega$$
