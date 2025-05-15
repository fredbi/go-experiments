package postgis

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"testing"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/jackc/pgx"
	"github.com/fredbi/geo/pkg/tileprovider"
	"github.com/fredbi/geo/utils/cli/envk"
	"github.com/fredbi/geo/utils/config"
	"github.com/spf13/viper"
)

func TestReplaceTokens(t *testing.T) {
	type tcase struct {
		sql      string
		srid     uint64
		tile     *tileprovider.Tile
		expected string
	}

	fn := func(t *testing.T, tc tcase) {
		sql, err := replaceTokens(tc.sql, tc.srid, tc.tile)
		if err != nil {
			t.Errorf("unexpected error, Expected nil Got %v", err)
			return
		}

		if sql != tc.expected {
			t.Errorf("incorrect sql,\n Expected \n \t%v\n Got \n \t%v", tc.expected, sql)
			return
		}
	}

	tests := map[string]tcase{
		"replace BBOX": {
			sql:      "SELECT * FROM foo WHERE geom && !BBOX!",
			srid:     tileprovider.WebMercator,
			tile:     tileprovider.NewTile(2, 1, 1),
			expected: "SELECT * FROM foo WHERE geom && ST_MakeEnvelope(-1.017529720390625e+07,-156543.03390625,156543.03390624933,1.017529720390625e+07,3857)",
		},
		"replace BBOX with != in query": {
			sql:      "SELECT * FROM foo WHERE geom && !BBOX! AND bar != 42",
			srid:     tileprovider.WebMercator,
			tile:     tileprovider.NewTile(2, 1, 1),
			expected: "SELECT * FROM foo WHERE geom && ST_MakeEnvelope(-1.017529720390625e+07,-156543.03390625,156543.03390624933,1.017529720390625e+07,3857) AND bar != 42",
		},
		"replace BBOX and ZOOM 1": {
			sql:      "SELECT id, scalerank=!ZOOM! FROM foo WHERE geom && !BBOX!",
			srid:     tileprovider.WebMercator,
			tile:     tileprovider.NewTile(2, 1, 1),
			expected: "SELECT id, scalerank=2 FROM foo WHERE geom && ST_MakeEnvelope(-1.017529720390625e+07,-156543.03390625,156543.03390624933,1.017529720390625e+07,3857)",
		},
		"replace BBOX and ZOOM 2": {
			sql:      "SELECT id, scalerank=!ZOOM! FROM foo WHERE geom && !BBOX!",
			srid:     tileprovider.WebMercator,
			tile:     tileprovider.NewTile(16, 11241, 26168),
			expected: "SELECT id, scalerank=16 FROM foo WHERE geom && ST_MakeEnvelope(-1.3163688815956049e+07,4.0352540420407774e+06,-1.3163058210472783e+07,4.035884647524042e+06,3857)",
		},
		"replace pixel_width/height and scale_denominator": {
			sql:      "SELECT id, !pixel_width! as width, !pixel_height! as height, !scale_denominator! as scale_denom FROM foo WHERE geom && !BBOX!",
			srid:     tileprovider.WebMercator,
			tile:     tileprovider.NewTile(11, 1070, 676),
			expected: "SELECT id, 76.43702827453626 as width, 76.43702827453671 as height, 272989.3866947724 as scale_denom FROM foo WHERE geom && ST_MakeEnvelope(899816.6968478388,6.789748347570495e+06,919996.0723123164,6.809927723034973e+06,3857)",
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestUppercaseTokens(t *testing.T) {
	type tcase struct {
		str      string
		expected string
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			out := uppercaseTokens(tc.str)

			if out != tc.expected {
				t.Errorf("expected \n \t%v\n out \n \t%v", tc.expected, out)
				return
			}
		}
	}

	tests := map[string]tcase{
		"uppercase tokens": {
			str:      "this !lower! case !STrInG! should uppercase !TOKENS!",
			expected: "this !LOWER! case !STRING! should uppercase !TOKENS!",
		},
		"no tokens": {
			str:      "no token",
			expected: "no token",
		},
		"empty string": {
			str:      "",
			expected: "",
		},
		"unclosed token": {
			str:      "unclosed !token",
			expected: "unclosed !token",
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func loadConfig() (*viper.Viper, error) {
	return config.DefaultLoader.WithBasePath("../../..").LoadFor(envk.StringOrDefault("APP_ENV", "test"))
}

func sqlTestsEnabled(dsn string) bool {
	if dsn == "" {
		return false
	}
	u, err := url.Parse(dsn)
	if err != nil {
		log.Println(err)
		return false
	}
	if u == nil || len(u.Path) == 0 {
		return false
	}

	dbName := u.Path[1:]
	u.Path = ""

	db, err := sqlx.Open("pgx", u.String())
	if err != nil {
		log.Println(err)
		return false
	}
	defer db.Close()

	var ignored sql.NullString
	/* #nosec */
	err = db.QueryRow("SELECT datname FROM pg_database WHERE datname = '" + dbName + "';").Scan(&ignored)
	log.Println(err)
	return err == nil
}

func TestDecipherFields(t *testing.T) {
	cfg, err := loadConfig()
	require.NoError(t, err)
	dsn := cfg.GetString("postgres.tegola.url")

	if !sqlTestsEnabled(dsn) {
		t.Skipf("Skipping postgis tile provider test, not in Circle CI")
	}

	type tcase struct {
		sql              string
		expectedRowCount int
		expectedTags     map[string]interface{}
	}

	cc, err := pgx.ParseURI(dsn)
	require.NoError(t, err)

	conn, err := pgx.Connect(cc)
	require.NoError(t, err)
	defer conn.Close()

	fn := func(t *testing.T, tc tcase) {
		rows, err := conn.Query(tc.sql)
		require.NoError(t, err)
		defer rows.Close()

		var rowCount int
		for rows.Next() {
			geoFieldname := "geom"
			idFieldname := "id"
			descriptions := rows.FieldDescriptions()

			vals, err := rows.Values()
			require.NoError(t, err)

			_, _, tags, err := decipherFields(context.TODO(), geoFieldname, idFieldname, descriptions, vals)
			require.NoError(t, err)

			require.EqualValues(t, tc.expectedTags, tags)
			rowCount++
		}
		require.NoError(t, rows.Err())
		require.Equal(t, tc.expectedRowCount, rowCount)
	}

	tests := map[string]tcase{
		"hstore 1": {
			sql:              "SELECT id, tags, int8_test FROM hstore_test WHERE id = 1;",
			expectedRowCount: 1,
			expectedTags: map[string]interface{}{
				"height":    "9",
				"int8_test": int64(1000888),
			},
		},
		"hstore 2": {
			sql:              "SELECT id, tags, int8_test FROM hstore_test WHERE id = 2;",
			expectedRowCount: 1,
			expectedTags: map[string]interface{}{
				"hello":     "there",
				"good":      "day",
				"int8_test": int64(8880001),
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}
