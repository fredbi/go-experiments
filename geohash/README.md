# geohash: tilings of the Earth

TODO(fredbi): Push here last summer's reflections on a generic approach to geocoding using a geohashing schema.

This approach would naturally produce popular schemes such as GeoHash, OpenCode (fka PlusCodes), MapBox tiles as special cases.

S3 does not produce a geo hash properly, but kind of.

Programme for this investigation:

* Desirable properties of a geohashing schema: self-similarity, arbitrary precision, equal area tiling, compact encoding, simple tile shapes, ease of computation, etc
  * Analysis of the goals pursued by existing schemes, e.g., code readability over encoding compactness, ease of computation over consistent projected area etc
  * Shortcomings of popular geohashing schemes
  * The specifics of the S3 approach (hexagonal quasi-tiling, goal to preserve distances)
* Differences with the mathematical definition of a tiling: behavior at boundaries, does not necessarily emphasize regular polygons
* Spaces considered for tiling: surface of a sphere/ellipsoid/geoid, plane in polar coordinates
* Is a projection required? Not necessarily. Explain why.
* Problems solved by geohashing
  * localization
  * indexation
  * nearest neighbors (investigate findings from the MongoDB project, which attempted to follow that path)
  * collision/shape intersection problems (share algorithm to construct a minimal covering tile set for any polygon, resp. minimal inner tile set)
  * fast aggregation (share personal findings from past attempts - not very successful, not outright failures either... - to follow that path)
  * web rendering with fast zooming
* Polar plane tilings:
  * triangle, square (e.g., GeoHash), rectangle (e.g., OpenCode), hexagonal quasi-tiling (S3)
  * self-similar tilings: Gosper Islands fractal tiling
    * smoothing algorithm to make this of any practical use (your typical Gosper Island gets thousands of edges..)
* Spherical tilings:
  * spherical triangles and the tilings generated from there
