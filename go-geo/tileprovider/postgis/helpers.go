package postgis

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-spatial/tegola/basic"
	"github.com/go-spatial/tegola/provider"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
)

const (
	bboxToken = "!BBOX!"
	zoomToken = "!ZOOM!"
	//#nosec
	scaleDenominatorToken = "!SCALE_DENOMINATOR!"
	//#nosec
	pixelWidthToken = "!PIXEL_WIDTH!"
	//#nosec
	pixelHeightToken = "!PIXEL_HEIGHT!"
)

// replaceTokens replaces tokens in the provided SQL string
//
// !BBOX! - the bounding box of the tile
// !ZOOM! - the tile Z value
// !SCALE_DENOMINATOR! - scale denominator, assuming 90.7 DPI (i.e. 0.28mm pixel size)
// !PIXEL_WIDTH! - the pixel width in meters, assuming 256x256 tiles
// !PIXEL_HEIGHT! - the pixel height in meters, assuming 256x256 tiles
func replaceTokens(sql string, srid uint64, tile provider.Tile) (string, error) {

	bufferedExtent, _ := tile.BufferedExtent()

	// TODO: leverage helper functions for minx / miny to make this easier to follow
	// TODO: it's currently assumed the tile will always be in WebMercator. Need to support different projections
	minGeo, err := basic.FromWebMercator(srid, basic.Point{bufferedExtent.MinX(), bufferedExtent.MinY()})
	if err != nil {
		return "", fmt.Errorf("convert tile point: %v", err)
	}

	maxGeo, err := basic.FromWebMercator(srid, basic.Point{bufferedExtent.MaxX(), bufferedExtent.MaxY()})
	if err != nil {
		return "", fmt.Errorf("convert tile point: %v", err)
	}

	minPt, maxPt := minGeo.AsPoint(), maxGeo.AsPoint()

	bbox := fmt.Sprintf("ST_MakeEnvelope(%g,%g,%g,%g,%d)", minPt.X(), minPt.Y(), maxPt.X(), maxPt.Y(), srid)

	extent, _ := tile.Extent()
	// TODO: Always convert to meter if we support different projections
	pixelWidth := (extent.MaxX() - extent.MinX()) / 256
	pixelHeight := (extent.MaxY() - extent.MinY()) / 256
	scaleDenominator := pixelWidth / 0.00028 /* px size in m */

	// replace query string tokens
	z, _, _ := tile.ZXY()
	tokenReplacer := strings.NewReplacer(
		bboxToken, bbox,
		zoomToken, strconv.FormatUint(uint64(z), 10),
		scaleDenominatorToken, strconv.FormatFloat(scaleDenominator, 'f', -1, 64),
		pixelWidthToken, strconv.FormatFloat(pixelWidth, 'f', -1, 64),
		pixelHeightToken, strconv.FormatFloat(pixelHeight, 'f', -1, 64),
	)

	uppercaseTokenSQL := uppercaseTokens(sql)

	return tokenReplacer.Replace(uppercaseTokenSQL), nil
}

var tokenRe = regexp.MustCompile("![a-zA-Z0-9_-]+!")

//	uppercaseTokens converts all !tokens! to uppercase !TOKENS!. Tokens can
//	contain alphanumerics, dash and underline chars.
func uppercaseTokens(str string) string {
	return tokenRe.ReplaceAllStringFunc(str, strings.ToUpper)
}

func transformVal(valType pgtype.OID, val interface{}) (interface{}, error) {
	switch valType {
	default:
		switch vt := val.(type) {
		default:
			log.Printf("%v type is not supported. (Expected it to be a stringer type)", valType)
			return nil, fmt.Errorf("%v type is not supported. (Expected it to be a stringer type)", valType)
		case fmt.Stringer:
			return vt.String(), nil
		case string:
			return vt, nil
		}
	case pgtype.BoolOID, pgtype.ByteaOID, pgtype.TextOID, pgtype.OIDOID, pgtype.VarcharOID, pgtype.JSONBOID:
		return val, nil
	case pgtype.Int8OID, pgtype.Int2OID, pgtype.Int4OID, pgtype.Float4OID, pgtype.Float8OID:
		switch vt := val.(type) {
		case int8:
			return int64(vt), nil
		case int16:
			return int64(vt), nil
		case int32:
			return int64(vt), nil
		case int64, uint64:
			return vt, nil
		case uint8:
			return int64(vt), nil
		case uint16:
			return int64(vt), nil
		case uint32:
			return int64(vt), nil
		case float32:
			return float64(vt), nil
		case float64:
			return vt, nil
		default: // should never happen.
			return nil, fmt.Errorf("%v type is not supported. (should never happen)", valType)
		}
	case pgtype.DateOID, pgtype.TimestampOID, pgtype.TimestamptzOID:
		return fmt.Sprintf("%v", val), nil
	}
}

func decipherFields(ctx context.Context, geoFieldname, idFieldname string, descriptions []pgx.FieldDescription, values []interface{}) (gid uint64, geom []byte, tags map[string]interface{}, err error) {
	var ok bool
	tags = make(map[string]interface{})

	for i := range values {
		// do a quick check
		if err = ctx.Err(); err != nil {
			return 0, nil, nil, err
		}

		// skip nil values.
		if values[i] == nil {
			continue
		}

		desc := descriptions[i]

		switch desc.Name {
		case geoFieldname:
			if geom, ok = values[i].([]byte); !ok {
				return 0, nil, nil, fmt.Errorf("unable to convert geometry field (%v) into bytes", geoFieldname)
			}
		case idFieldname:
			gid, err = gID(values[i])
		default:
			switch vex := values[i].(type) {
			case map[string]pgtype.Text:
				for k, v := range vex {
					// we need to check if the key already exists. if it does, then don't overwrite it
					if _, ok := tags[k]; !ok {
						tags[k] = v.String
					}
				}
			case *pgtype.Numeric:
				var num float64
				if err = vex.AssignTo(&num); err != nil {
					return gid, geom, tags, fmt.Errorf("unable to convert field [%v] (%v) of type (%v - %v) to a suitable value: %+v", i, desc.Name, desc.DataType, desc.DataTypeName, values[i])
				}

				tags[desc.Name] = num
			default:
				var value interface{}
				value, err = transformVal(desc.DataType, values[i])
				if err != nil {
					return gid, geom, tags, fmt.Errorf("unable to convert field [%v] (%v) of type (%v - %v) to a suitable value: %+v", i, desc.Name, desc.DataType, desc.DataTypeName, values[i])
				}

				tags[desc.Name] = value
			}
		}
	}

	return gid, geom, tags, err
}

func gID(v interface{}) (uint64, error) {
	switch aval := v.(type) {
	case float64:
		return uint64(aval), nil
	case int64:
		return uint64(aval), nil
	case uint64:
		return aval, nil
	case uint:
		return uint64(aval), nil
	case int8:
		return uint64(aval), nil
	case uint8:
		return uint64(aval), nil
	case uint16:
		return uint64(aval), nil
	case int32:
		return uint64(aval), nil
	case uint32:
		return uint64(aval), nil
	case string:
		return strconv.ParseUint(aval, 10, 64)
	default:
		return 0, fmt.Errorf("unable to convert field into a uint64")
	}
}
