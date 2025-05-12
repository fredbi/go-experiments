# A trivial tiling: geodesic coordinates

Space: a (long,lat) portion of the polar plane, -$`(\pi,\pi( x (-\pi/2,\pi/2(`$, as [WGS 84 system](https://en.wikipedia.org/wiki/World_Geodetic_System) coordinates.

Tile: a tile is the square on the surface of the Earth (e.g. a spherical square on a spherical model) centered around (X,Y).

Level 0 tiles (largest possible): start with 8 quarters intersecting at the poles

Level _n_ tiles: when coordinates are considered with a precision $`\epsilon/2`$, a point belonging a tile centered on (X,Y) is anywhere
within $`(-X-\epsilon/2,X+\epsilon/2),(-Y-\epsilon/2,Y+\epsilon/2)`$. This tile has an area in the polar plane of $`\epsilon^2`$ and an area of approx.
$`R^2.\epsilon^2`$ on the surface of the Earth (spherical model approximation).

We define the level _n_ as tiles defined by coordinates with precision $`\epsilon = \frac{\pi}{2 \sqrt{n}}`$.
