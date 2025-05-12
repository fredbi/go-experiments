# A trivial tiling: geodesic coordinates

Let's figure out a cartoon tiling scheme.

Space: a (long,lat) portion of the polar plane, $`[-\pi,\pi[ \times [-\frac{\pi}{2},\frac{\pi}{2}[`$, as [WGS 84 system](https://en.wikipedia.org/wiki/World_Geodetic_System) coordinates.

Tile: a tile is the square on the surface of the Earth (e.g. a spherical square on a spherical model) centered around $`(X,Y)`$.

Level 0 tiles (largest possible): start with 8 "quarters" intersecting at the poles

Level _n_ tiles: when coordinates are considered with a precision $`\epsilon`$, a point belonging a tile centered on $`(X,Y)`$ is anywhere
within $`[X-\frac{\epsilon}{2},X+\frac{\epsilon}{2}[,[Y-\frac{\epsilon}{2},Y+\frac{\epsilon}{2}[`$.
This tile has an area in the polar plane of $`\epsilon^2`$ and an area of approx. $`R^2.\epsilon^2`$ on the surface of the Earth (spherical model approximation).

We may fill a level 0 tile with up to $`\frac{\pi^2 }{4 \epsilon^2}`$.

Let us consider the suite $`\epsilon_{n+1} = \frac{1}{2^n}\frac{\pi}{2}`$.
We define the level _n_ as tiles defined by coordinates with precision $`\epsilon_{n}`$.
We have each level-0 tile being covered by 4 level-1 tiles, each level-1 tile covered by 4 level-2 tiles, etc.

For $`n > 0`$, there are $`8 . 4^n` n-tiles covering the globe.

Self-similarity: all tiles are squares at any scale.

Scaling factor: The scaling (or zooming) factor defined above is 4.

Edges: coordinate lower bounds within the tile, and coordinate upper bounds are outside.

Prefix inclusion: $`(34.54,66.45) \subset (34.5,66.4)`$

Encodings:
* decimal (text) encoding. Example: (34.54,66.45) is a tile of precision $`1/100th`$
* binary IEEE float representation: need extra information to capture the precision?

Main issues with this trivial scheme:
* It is far from obvious to determine valid coordinates for a given level. Arbitrary numbers are generally not valid.
* Considering encoding issues, such as conveying the precision, an optimal encoding is difficult to find.
