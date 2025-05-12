# A trivial tiling: geodesic coordinates

Space: a (long,lat) portion of the polar plane, $`[-\pi,\pi[ \times [-\frac{\pi}{2},\frac{\pi}{2}[`$, as [WGS 84 system](https://en.wikipedia.org/wiki/World_Geodetic_System) coordinates.

Tile: a tile is the square on the surface of the Earth (e.g. a spherical square on a spherical model) centered around $`(X,Y)`$.

Level 0 tiles (largest possible): start with 8 quarters intersecting at the poles

Level _n_ tiles: when coordinates are considered with a precision $`\epsilon`$, a point belonging a tile centered on $`(X,Y)`$ is anywhere
within $`[X-\frac{\epsilon}{2},X+\frac{\epsilon}{2}[,[Y-\frac{\epsilon}{2},Y+\frac{\epsilon}{2}[`$.
This tile has an area in the polar plane of $`\epsilon^2`$ and an area of approx. $`R^2.\epsilon^2`$ on the surface of the Earth (spherical model approximation).

We define the level _n_ as tiles defined by coordinates with precision $`\epsilon = \frac{\pi}{2 \sqrt{n}}`$...

For $`n > 0`$, there are `xyz` tiles covering the globe.

Self-similarity: all tiles are squares at any scale.

Scaling factor: 

Edges: coordinate lower bounds within the tile, and coordinate upper bounds are outside.
Prefix inclusion: ?

Encodings:
* decimal (text) encoding. Example: (34.54,66.45) is a tile of precision $`1/100th`$
* binary IEEE float representation: need an extra information to capture the precision?


