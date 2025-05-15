package postgis

import (
	"context"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/tegola/provider"
	"github.com/jackc/pgx"
	"github.com/fredbi/geo/pkg/tileprovider"
	"github.com/stretchr/testify/require"
)

func TestInferLayerGeomType(t *testing.T) {
	cfg, err := loadConfig()
	require.NoError(t, err)
	dsn := cfg.GetString("postgres.tegola.url")

	if !sqlTestsEnabled(dsn) {
		t.Skipf("Skipping postgis tile provider test, not in Circle CI")
	}

	type tlayer struct {
		sql      string
		name     string
		geomType string
	}
	type tcase struct {
		layers       []tlayer
		expectedName string
		geom         geom.Geometry
		err          string
	}

	cc, err := pgx.ParseURI(dsn)
	require.NoError(t, err)

	cp := pgx.ConnPoolConfig{
		ConnConfig: cc,
	}
	conn, err := pgx.NewConnPool(cp)
	require.NoError(t, err)
	defer conn.Close()
	fn := func(t *testing.T, tc tcase) {

		provider := New(conn)
		for _, l := range tc.layers {
			lyr := provider.NewLayer(l.name).
				WithIDField("gid").
				WithQuery(l.sql).
				WithGeomType(l.geomType)

			err := lyr.Save()
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
				return
			}
			require.NoError(t, err)
		}

		p := provider.(*Provider)
		layer, _ := p.layer(tc.expectedName)
		require.EqualValues(t, tc.geom, layer.GeomType)
	}

	tests := map[string]tcase{
		"1": {
			layers: []tlayer{
				{
					name: "land",
					sql:  "SELECT gid, ST_AsBinary(geom) FROM ne_10m_land_scale_rank WHERE geom && !BBOX!",
				},
			},
			expectedName: "land",
			geom:         geom.MultiPolygon{},
		},
		"zoom token replacement": {
			layers: []tlayer{
				{
					name: "land",
					sql:  "SELECT gid, ST_AsBinary(geom) FROM ne_10m_land_scale_rank WHERE gid = !ZOOM! AND geom && !BBOX!",
				},
			},
			expectedName: "land",
			geom:         geom.MultiPolygon{},
		},
		"configured geometry_type": {
			layers: []tlayer{
				{
					name:     "land",
					geomType: "multipolygon",
					sql:      "SELECT gid, ST_AsBinary(geom) FROM invalid_table_to_check_query_table_was_not_inspected WHERE geom && !BBOX!",
				},
			},
			expectedName: "land",
			geom:         geom.MultiPolygon{},
		},
		"configured geometry_type (case insensitive)": {
			layers: []tlayer{
				{
					name:     "land",
					geomType: "MultiPolyGOn",
					sql:      "SELECT gid, ST_AsBinary(geom) FROM invalid_table_to_check_query_table_was_not_inspected WHERE geom && !BBOX!",
				},
			},
			expectedName: "land",
			geom:         geom.MultiPolygon{},
		},
		"invalid configured geometry_type": {
			layers: []tlayer{
				{
					name:     "land",
					geomType: "invalid",
					sql:      "SELECT gid, ST_AsBinary(geom) FROM invalid_table_to_check_query_table_was_not_inspected WHERE geom && !BBOX!",
				},
			},
			expectedName: "land",
			err:          "for layer \"land\" unsupported geometry type \"invalid\"",
			geom:         geom.MultiPolygon{},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestTileFeatures(t *testing.T) {
	cfg, err := loadConfig()
	require.NoError(t, err)
	dsn := cfg.GetString("postgres.tegola.url")

	if !sqlTestsEnabled(dsn) {
		t.Skipf("Skipping postgis tile provider test, not in Circle CI")
	}

	type tlayer struct {
		sql       string
		name      string
		geomField string
		idField   string
		geomType  string
	}

	type tcase struct {
		layer                tlayer
		tile                 *tileprovider.Tile
		expectedFeatureCount int
		expectedTags         []string
	}

	cc, err := pgx.ParseURI(dsn)
	require.NoError(t, err)

	cp := pgx.ConnPoolConfig{
		ConnConfig: cc,
	}
	conn, err := pgx.NewConnPool(cp)
	require.NoError(t, err)
	defer conn.Close()

	fn := func(t *testing.T, tc tcase) {
		p := New(conn)
		l := tc.layer
		lyr := p.NewLayer(l.name).
			WithIDField("gid").
			WithQuery(l.sql).
			WithGeomType(l.geomType).
			WithSRID(tileprovider.WebMercator)

		if l.geomField != "" {
			lyr.WithGeomField(l.geomField)
		}
		if l.idField != "" {
			lyr.WithIDField(l.idField)
		}
		err := lyr.Save()
		require.NoError(t, err)

		var featureCount int
		err = p.TileFeatures(context.Background(), l.name, tc.tile, func(f *provider.Feature) error {
			// only verify tags on first feature
			if featureCount == 0 {
				for _, tag := range tc.expectedTags {
					require.Contains(t, f.Tags, tag)
				}
			}

			featureCount++

			return nil
		})
		require.NoError(t, err)
		require.EqualValues(t, tc.expectedFeatureCount, featureCount)
	}

	tests := map[string]tcase{
		"SQL with !ZOOM!": {
			layer: tlayer{
				name: "land",
				sql:  "SELECT gid, ST_AsBinary(geom) AS geom FROM ne_10m_land_scale_rank WHERE scalerank=!ZOOM! AND geom && !BBOX!",
			},
			tile:                 tileprovider.NewTile(1, 1, 1),
			expectedFeatureCount: 98,
		},
		"SQL with comments": {
			layer: tlayer{
				name: "land",
				sql:  " -- this is a comment\n -- accross multiple lines \n \tSELECT gid, -- gid \nST_AsBinary(geom) AS geom -- geom \n FROM ne_10m_land_scale_rank WHERE scalerank=!ZOOM! AND geom && !BBOX! -- comment at the end",
			},
			tile:                 tileprovider.NewTile(1, 1, 1),
			expectedFeatureCount: 98,
		},
		"decode numeric(x,x) types": {
			layer: tlayer{
				name:      "buildings",
				idField:   "osm_id",
				geomField: "geometry",
				sql:       "SELECT ST_AsBinary(geometry) AS geometry, osm_id, name, nullif(as_numeric(height),-1) AS height, type FROM osm_buildings_test WHERE geometry && !BBOX!",
			},
			tile:                 tileprovider.NewTile(16, 11241, 26168),
			expectedFeatureCount: 101,
			expectedTags:         []string{"name", "type"}, // height can be null and therefore missing from the tags
		},
		"gracefully handle 3d point": {
			layer: tlayer{
				name:      "three_d_points",
				idField:   "id",
				geomField: "geom",
				sql:       "SELECT ST_AsBinary(geom) AS geom, id FROM three_d_test WHERE geom && !BBOX!",
			},
			tile:                 tileprovider.NewTile(0, 0, 0),
			expectedFeatureCount: 0,
		},
		"gracefully handle null geometry": {
			layer: tlayer{
				name:      "null_geom",
				idField:   "id",
				geomField: "geometry",
				sql:       "SELECT id, ST_AsBinary(geometry) AS geometry, !BBOX! as bbox FROM null_geom_test",
			},
			tile:                 tileprovider.NewTile(16, 11241, 26168),
			expectedFeatureCount: 1,
			expectedTags:         []string{"bbox"},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}
