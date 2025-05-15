package postgis

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-spatial/geom/encoding/wkb"

	"github.com/fredbi/geo/pkg/tileprovider"

	"github.com/go-spatial/tegola/provider"

	"github.com/go-spatial/geom"
	"github.com/jackc/pgx"
)

// isSelectQuery is a regexp to check if a query starts with `SELECT`,
// case-insensitive and ignoring any preceeding whitespace and SQL comments.
var isSelectQuery = regexp.MustCompile(`(?i)^((\s*)(--.*\n)?)*select`)

func New(db *pgx.ConnPool) tileprovider.Tiler {
	return &Provider{conn: db}
}

type Provider struct {
	conn   *pgx.ConnPool
	config []layerConfig
}

type layerBuilder struct {
	provider *Provider

	// The Name of the layer
	name string
	// The SQL to use when querying PostGIS for this layer
	sql string
	// The ID field name, this will default to 'gid' if not set to something other then empty string.
	idField string
	// The Geometery field name, this will default to 'geom' if not set to something other then empty string.
	geomField string
	// the geometry type contained in this layer
	geomType string
	// The SRID that the data in the table is stored in. This will default to WebMercator
	srid uint64
}

func (l *layerBuilder) WithGeomType(geomType string) tileprovider.LayerBuilder {
	l.geomType = geomType
	return l
}
func (l *layerBuilder) WithSRID(srid uint64) tileprovider.LayerBuilder {
	l.srid = srid
	return l
}
func (l *layerBuilder) WithGeomField(fieldName string) tileprovider.LayerBuilder {
	l.geomField = fieldName
	return l
}
func (l *layerBuilder) WithQuery(sql string) tileprovider.LayerBuilder {
	l.sql = sql
	return l
}
func (l *layerBuilder) WithIDField(fieldName string) tileprovider.LayerBuilder {
	l.idField = fieldName
	return l
}

func (l *layerBuilder) Save() error {
	// validate
	if l.name == "" {
		return errors.New("name for layer  can't be empty")
	}

	if l.geomField == "" {
		return fmt.Errorf("for layer %q geom field name can't be empty", l.name)
	}

	if l.idField == "" {
		return fmt.Errorf("for layer %q id field name can't be empty", l.name)
	}

	if l.idField == l.geomField {
		return fmt.Errorf("for layer %q id field and geom field can't be the same", l.name)
	}

	if l.srid != tileprovider.WGS84 && l.srid != tileprovider.WebMercator {
		return fmt.Errorf("for layer %q the srid %d is unsupported, only %d and %d are supported", l.name, l.srid, tileprovider.WebMercator, tileprovider.WGS84)
	}

	if l.sql == "" {
		return fmt.Errorf("for layer %q the sql query is required", l.name)
	}

	if !isSelectQuery.MatchString(l.sql) {
		return fmt.Errorf("for layer %q only select queries are supported, got: %s", l.name, l.sql)
	}

	// convert !BOX! (MapServer) and !bbox! (Mapnik) to !BBOX! for compatibility
	sql := strings.Replace(strings.Replace(l.sql, "!BOX!", "!BBOX!", -1), "!bbox!", "!BBOX!", -1)
	// make sure that the sql has a !BBOX! token
	if !strings.Contains(sql, bboxToken) {
		return fmt.Errorf("SQL for layer %q is missing required token: %v", l.name, bboxToken)
	}
	if !strings.Contains(sql, "*") {
		if !strings.Contains(sql, l.geomField) {
			return fmt.Errorf("SQL for layer %q does not contain the geometry field: %v", l.name, l.geomField)
		}
		if !strings.Contains(sql, l.idField) {
			return fmt.Errorf("SQL for layer %q does not contain the id field for the geometry: %v", l.name, l.idField)
		}
	}

	var geom geom.Geometry
	if l.geomType != "" {
		v, err := geomNameToGeom(l.geomType)
		if err != nil {
			return fmt.Errorf("for layer %q %v", l.name, err)
		}
		geom = v
	} else {
		v, err := l.inferGeomType()
		if err != nil {
			return fmt.Errorf("for layer %q %v", l.name, err)
		}
		geom = v
	}

	l.provider.add(layerConfig{
		Name:      l.name,
		SQL:       sql,
		SRID:      l.srid,
		GeomField: l.geomField,
		GeomType:  geom,
		IDField:   l.idField,
	})
	return nil
}

func (l *layerBuilder) inferGeomType() (geom.Geometry, error) {
	var err error

	// we want to know the geom type instead of returning the geom data so we modify the SQL
	//
	// case insensitive search
	re := regexp.MustCompile(`(?i)ST_AsBinary`)
	sql := re.ReplaceAllString(l.sql, "ST_GeometryType")

	// we only need a single result set to sniff out the geometry type
	sql = fmt.Sprintf("%v LIMIT 1", sql)

	// if a !ZOOM! token exists, all features could be filtered out so we don't have a geometry to inspect it's type.
	// address this by replacing the !ZOOM! token with an ANY statement which includes all zooms
	sql = strings.Replace(sql, "!ZOOM!", "ANY('{0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24}')", 1)

	// we need a tile to run our sql through the replacer
	tile := tileprovider.NewTile(0, 0, 0)

	// normal replacer
	sql, err = replaceTokens(sql, l.srid, tile)
	if err != nil {
		return nil, err
	}

	rows, err := l.provider.conn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// fetch rows FieldDescriptions. this gives us the OID for the data types returned to aid in decoding
	fdescs := rows.FieldDescriptions()
	for rows.Next() {

		vals, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("error running SQL: %v ; %v", sql, err)
		}

		// iterate the values returned from our row, sniffing for the geomField or st_geometrytype field name
		for i, v := range vals {
			switch fdescs[i].Name {
			case l.geomField, "st_geometrytype":
				switch v {
				case "ST_Point":
					return geom.Point{}, nil
				case "ST_LineString":
					return geom.LineString{}, nil
				case "ST_Polygon":
					return geom.Polygon{}, nil
				case "ST_MultiPoint":
					return geom.MultiPoint{}, nil
				case "ST_MultiLineString":
					return geom.MultiLineString{}, nil
				case "ST_MultiPolygon":
					return geom.MultiPolygon{}, nil
				case "ST_GeometryCollection":
					return geom.Collection{}, nil
				default:
					return nil, fmt.Errorf("layer (%v) returned unsupported geometry type (%v)", l.name, fdescs[i].Name)
				}
			}
		}
	}

	return nil, rows.Err()
}

func geomNameToGeom(name string) (geom.Geometry, error) {
	switch strings.ToLower(name) {
	case "point":
		return geom.Point{}, nil
	case "linestring":
		return geom.LineString{}, nil
	case "polygon":
		return geom.Polygon{}, nil
	case "multipoint":
		return geom.MultiPoint{}, nil
	case "multilinestring":
		return geom.MultiLineString{}, nil
	case "multipolygon":
		return geom.MultiPolygon{}, nil
	case "geometrycollection":
		return geom.Collection{}, nil
	default:
		return nil, fmt.Errorf("unsupported geometry type %q", name)
	}
}

func (p *Provider) NewLayer(name string) tileprovider.LayerBuilder {
	return &layerBuilder{
		name:      name,
		srid:      tileprovider.WGS84,
		geomField: "geom",
		idField:   "id",
		provider:  p,
	}
}

func (p *Provider) add(opt layerConfig) {
	p.config = append(p.config, opt)
}

type layerConfig struct {
	// The Name of the layer
	Name string
	// The SQL to use when querying PostGIS for this layer
	SQL string
	// The ID field name, this will default to 'gid' if not set to something other then empty string.
	IDField string
	// The Geometery field name, this will default to 'geom' if not set to something other then empty string.
	GeomField string
	// GeomType is the the type of geometry returned from the SQL
	GeomType geom.Geometry
	// The SRID that the data in the table is stored in. This will default to WebMercator
	SRID uint64
}

// TileFeature will stream decoded features to the callback function fn
// if fn returns ErrCanceled, the TileFeatures method should stop processing
func (p *Provider) TileFeatures(ctx context.Context, layer string, t provider.Tile, fn func(f *provider.Feature) error) error {
	// fetch the provider layer
	plyr, ok := p.layer(layer)
	if !ok {
		return fmt.Errorf("layer %q can't be found", layer)
	}

	sql, err := replaceTokens(plyr.SQL, plyr.SRID, t)
	if err != nil {
		return fmt.Errorf("error replacing layer tokens for layer (%v) SQL (%v): %v", layer, sql, err)
	}

	// context check
	if err = ctx.Err(); err != nil {
		return err
	}

	rows, err := p.conn.Query(sql)
	if err != nil {
		return fmt.Errorf("error running layer (%v) SQL (%v): %v", layer, sql, err)
	}
	defer rows.Close()

	// fetch rows FieldDescriptions. this gives us the OID for the data types returned to aid in decoding
	fdescs := rows.FieldDescriptions()

	for rows.Next() {
		// context check
		if err := ctx.Err(); err != nil {
			return err
		}

		// fetch row values
		vals, err := rows.Values()
		if err != nil {
			return fmt.Errorf("error running layer (%v) SQL (%v): %v", layer, sql, err)
		}

		gid, geobytes, tags, err := decipherFields(ctx, plyr.GeomField, plyr.IDField, fdescs, vals)
		if err != nil {
			switch err {
			case context.Canceled:
				return err
			default:
				return fmt.Errorf("for layer (%v) %v", plyr.Name, err)
			}
		}

		// check that we have geometry data. if not, skip the feature
		if len(geobytes) == 0 {
			continue
		}

		// decode our WKB
		geom, err := wkb.DecodeBytes(geobytes)
		if err != nil {
			switch err.(type) {
			case wkb.ErrUnknownGeometryType:
				continue
			default:
				return fmt.Errorf("unable to decode layer (%v) geometry field (%v) into wkb where (%v = %v): %v", layer, plyr.GeomField, plyr.IDField, gid, err)
			}
		}

		feature := provider.Feature{
			ID:       gid,
			Geometry: geom,
			SRID:     plyr.SRID,
			Tags:     tags,
		}

		// pass the feature to the provided callback
		if err = fn(&feature); err != nil {
			return err
		}
	}

	return rows.Err()
}

// Layer fetches an individual layer from the provider, if it's configured
// if no name is provider, the first layer is returned
func (p *Provider) layer(name string) (layerConfig, bool) {
	if len(p.config) == 0 {
		return layerConfig{}, false
	}
	if name == "" {
		return p.config[0], true
	}

	for _, l := range p.config {
		if l.Name == name {
			return l, true
		}
	}
	return layerConfig{}, false
}

// Layers returns information about the various layers the provider supports
func (p *Provider) Layers() ([]provider.LayerInfo, error) {
	result := make([]provider.LayerInfo, len(p.config))
	for i, v := range p.config {
		result[i] = linfo{l: v}
	}
	return result, nil
}

type linfo struct {
	l layerConfig
}

func (l linfo) Name() string {
	return l.l.Name
}
func (l linfo) GeomType() geom.Geometry {
	return l.l.GeomType
}
func (l linfo) SRID() uint64 {
	return l.l.SRID
}
