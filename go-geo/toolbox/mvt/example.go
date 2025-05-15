package mvt

//.*NO TEST
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/fredbi/geo/pkg/maptile"
	"github.com/fredbi/geo/utils/geojson"
)

func _() {
	//tile := maptile.New(17896, 24449, 16)
	tile := maptile.New(3142, 6687, 14)

	//file_name := fmt.Sprintf("pkg/mvt/testdata/%d-%d-%d.json", tile.Z, tile.X, tile.Y)
	filename := fmt.Sprintf("pkg/mvt/testdata/2018082412_0_mesh.geojson") //, tile.Z, tile.X, tile.Y)

	jsonData, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("err open files", err)
		return
	}

	// Start with a set of feature collections defining each layer in lon/lat (WGS84).
	collections := map[string]*geojson.FeatureCollection{}
	//fmt.Println("collections", collections)
	err = json.Unmarshal(jsonData, &collections)
	//err = collections.UnmarshalJson(json_data)

	if err != nil {
		fmt.Println("err when unmarshall jsondata", err)
	}
	//

	// Convert to a layers object and project to tile coordinates.
	layers := NewLayers(collections)
	layers.ProjectToTile(tile) // x, y, z

	// Simplify the geometry now that it's in the tile coordinate space.
	//layers.Simplify(simplify.DouglasPeucker(0.1))

	// Depending on use-case remove empty geometry, those two small to be
	// represented in this tile space.
	// In this case lines shorter than 1, and areas smaller than 1.
	//layers.RemoveEmpty(1.0, 1.0)

	// encoding using the Mapbox Vector Tile protobuf encoding.
	//data, err := mvt.Marshal(layers) // this data is NOT gzipped.

	// Sometimes MVT data is stored and transferred gzip compressed. In that case:
	data, err := MarshalGzipped(layers)
	if err != nil {
		log.Fatalf("MarshalGzipped error: %v", err)
	}
	f, err := os.Create("14-3142-6687.mvt")
	if err != nil {
		log.Fatalf("Create file error: %v", err)
	}
	defer f.Close()
	n2, err := f.Write(data)
	fmt.Printf("wrote %d bytes\n", n2)

	// error checking
	if err != nil {
		log.Fatalf("marshal error: %v", err)
	}

	_ = data
}
