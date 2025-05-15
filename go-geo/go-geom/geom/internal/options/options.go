package options

type (
	Layout interface {
		SRID() uint32
		set(*layout)
	}

	Equality interface {
		set(*equality)
	}

	Rounding interface {
		set(*rounding)
	}

	Topology interface {
		set(*topology)
	}

	Tesselator interface {
		set(*tesselator)
	}

	layout struct {
		srid uint32
	}

	equality struct{}

	topology struct{}

	rounding struct {
		precision uint32
	}

	tesselator struct{}
)

func (*layout) set(*layout)    {}
func (l *layout) SRID() uint32 { return l.srid }

func defaultLayout() *layout {
	return &layout{}
}

func (*equality) set(*equality) {}

func defaultEquality() *equality {
	return &equality{}
}

func (*tesselator) set(*tesselator) {}

func defaultTesselator() *tesselator {
	return &tesselator{}
}

func (*topology) set(*topology) {}

func defaultTopology() *topology {
	return &topology{}
}

func (*rounding) set(*rounding) {}

func defaultRounding() *rounding {
	return &rounding{precision: 6}
}

func WithSRID(srid uint32) func(Layout) {
	return func(cfg Layout) {
		cfg.set(&layout{srid: srid})
	}
}

func WithPrecision(precision uint32) func(Rounding) {
	return func(cfg Rounding) {
		cfg.set(&rounding{precision: precision})
	}
}

/*
type layoutSettings struct {
	srid      int
	precision int
}

// LayoutDefaults defines some defaults associated with layouts, such as the SRID or the rounding precision.
type LayoutDefaults map[Layout]layoutSettings

func (d LayoutDefaults) SRID(key Layout) int {
	v, ok := d[key]
	if !ok {
		return -1
	}
	return v.srid
}

func (d LayoutDefaults) SetSRID(key Layout, v int) {
	v, ok := d[key]
	if !ok {
		return
	}
	d[key].srid = v
}

func (d LayoutDefaults) SetPrecision(key Layout, v int) {
	v, ok := d[key]
	if !ok {
		return
	}
	d[key].precision = v
}

func (d LayoutDefaults) WithSRID(key Layout, v int) LayoutDefaults {
	d.SetSRID(key, v)
	return d
}

func (d LayoutDefaults) WithPrecision(key Layout, v int) LayoutDefaults {
	d.SetPrecision(key, v)
	return d
}

// NewLayoutDefaults returns a LayoutDefaults map to set up customized defaults for layouts
func NewLayoutDefaults() LayoutDefaults {
	return defaultSettings()
}

const defaultPrecision = 6

var (
	// defaultLayout defines the package level default layout used to create all geometries
	defautLayout    Layout
	m               sync.Mutex
	defaultSettings LayoutDefaults
)

func init() {
	defaultLayout = XY
	defaultSettings = makeDefaultSettings()
}

func makeDefaultSettings() map[Layout]layoutSettings {
	return map[Layout]layoutSettings{
		NoLayout: {srid: 0, precision: defaultPrecision},
		X:        {srid: 0, precision: defaultPrecision},
		// TODO
	}
}

func DefaultLayout() Layout {
	return defaultLayout
}

func SetDefaultLayout(layout Layout) {
	defaultLayout = layout
}
func SetDefaultPrecision(precision int) {
	m.Lock()
	defer m.Unlock()
	for k := range defaultSettings {
		v := defaultSettings[k]
		v.precision = precision
		defaultSettings[k] = v
	}
}
*/
