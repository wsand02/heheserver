package cache

import (
	"image"
	"log"

	"github.com/dgraph-io/ristretto/v2"
)

type IgnoreCache struct {
	*ristretto.Cache[string, []string]
}

var ignoreCache *IgnoreCache

func sizeToNCMB(size int64) (int64, int64) {
	bytes := size * 1024 * 1024
	nc := size * 10000 // this hasnt been thought through thoroughly
	return bytes, nc
}

func GetIgnoreCache() *IgnoreCache {
	if ignoreCache == nil {
		log.Fatal("ignore cache not initialized")
	}
	return ignoreCache
}

var vidThumbCache *VidThumbCache

func GetVidThumbCache() *VidThumbCache {
	if vidThumbCache == nil {
		log.Fatal("vidthumb cache not initialized")
	}
	return vidThumbCache
}

var resizeCache *ResizeCache

func GetResizeCache() *ResizeCache {
	if resizeCache == nil {
		log.Fatal("resize cache not initialized")
	}
	return resizeCache
}

// DimensionCache holds decoded image dimensions (image.Point{X: width,
// Y: height}) keyed by file path, so the gallery can emit width/height on grid
// <img>s without re-decoding every header on each listing/refilter render.
type DimensionCache struct {
	*ristretto.Cache[string, image.Point]
}

var dimensionCache *DimensionCache

func GetDimensionCache() *DimensionCache {
	if dimensionCache == nil {
		log.Fatal("dimension cache not initialized")
	}
	return dimensionCache
}

// NewDimensionCache uses a fixed, tiny footprint (each entry is one point, set
// with cost 1) rather than a configurable byte budget like the media caches.
func NewDimensionCache() error {
	cache, err := ristretto.NewCache(&ristretto.Config[string, image.Point]{
		NumCounters: 1_000_000, // ~100k tracked items
		MaxCost:     100_000,   // cost 1 per entry => up to ~100k cached dimensions
		BufferItems: 64,
	})
	if err != nil {
		return err
	}
	dimensionCache = &DimensionCache{cache}
	return nil
}

func NewIgnoreCache(size int64) error {
	size, nc := sizeToNCMB(size)
	cache, err := ristretto.NewCache(&ristretto.Config[string, []string]{
		NumCounters: nc,   // 1000*10 seems ok for heheignore files...
		MaxCost:     size, // 16MB
		BufferItems: 64,
	})
	if err != nil {
		return err
	}
	ignoreCache = &IgnoreCache{cache}
	return nil
}

type ResizeCacheItem struct {
	image.Image
	Transparent bool
}

type ResizeCache struct {
	*ristretto.Cache[string, ResizeCacheItem]
}

func NewResizeCache(size int64) error {
	size, nc := sizeToNCMB(size)
	cache, err := ristretto.NewCache(&ristretto.Config[string, ResizeCacheItem]{
		NumCounters: nc,
		MaxCost:     size,
		BufferItems: 64,
	})
	if err != nil {
		return err
	}
	resizeCache = &ResizeCache{cache}
	return nil
}

type VidThumbCache struct {
	*ristretto.Cache[string, image.Image]
}

func NewVidThumbCache(size int64) error {
	size, nc := sizeToNCMB(size)
	cache, err := ristretto.NewCache(&ristretto.Config[string, image.Image]{
		NumCounters: nc,
		MaxCost:     size,
		BufferItems: 64,
	})
	if err != nil {
		return err
	}
	vidThumbCache = &VidThumbCache{cache}
	return nil
}
