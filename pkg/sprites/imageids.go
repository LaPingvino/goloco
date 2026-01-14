package sprites

// ImageIDs from OpenLoco - sprite indices in G1.DAT

// UI Elements
const (
	WindowResizeHandle         = 2305
	ColourSwatchRecolourable   = 2306
	ColourSwatchRaised         = 2307
	ColourSwatchPressed        = 2308
	CompanyListDropdownIcon    = 2309
	IconParentFolder           = 2310
	IconFolder                 = 2311
	CurvedBorderLeftMedium     = 2315
	CurvedBorderRightMedium    = 2316
	CurvedBorderLeftMediumBold = 2317
	CurvedBorderRightMediumBold= 2318
	CurvedBorderLeftSmall      = 2319
	CurvedBorderRightSmall     = 2320
	CloseButton                = 2321
	FrameBackgroundImage       = 2322
	FrameBackgroundImageAlt    = 2323
	InlineGreenUpArrow         = 2324
	InlineRedDownArrow         = 2325
	ProgressbarStyle0Frame0    = 2326
	ProgressbarStyle0Frame1    = 2327
	ProgressbarStyle0Frame2    = 2328
	ProgressbarStyle0Frame3    = 2329
	ProgressbarTrack           = 2330
	Tab                        = 2387
	SelectedTab                = 2388
	TabDisplay                 = 2391
	TabControl                 = 2392
	TabSound                   = 2393
	TabMiscellaneous           = 2394
)

// Construction icons
const (
	ConstructionStraight            = 2335
	StepBack                        = 2336
	StepForward                     = 2337
	RedArrowUp                      = 2338
	RedArrowDown                    = 2339
	ConstructionLeftCurveVerySmall  = 2340
	ConstructionRightCurveVerySmall = 2341
	ConstructionLeftCurveSmall      = 2342
	ConstructionRightCurveSmall     = 2343
	ConstructionLeftCurve           = 2344
	ConstructionRightCurve          = 2345
	ConstructionLeftCurveLarge      = 2346
	ConstructionRightCurveLarge     = 2347
	ConstructionSteepSlopeDown      = 2348
	ConstructionSlopeDown           = 2349
	ConstructionLevel               = 2350
	ConstructionSlopeUp             = 2351
	ConstructionSteepSlopeUp        = 2352
	ConstructionRemove              = 2361
	ConstructionNewPosition         = 2362
	RubbishBin                      = 2363
	CentreViewport                  = 2364
	RotateObject                    = 2365
	PhotoCamera                     = 2366
	Paintbrush                      = 2367
)

// Toolbar icons - using available G1.DAT UI sprites
// Real Locomotion toolbar sprites come from InterfaceSkin objects in DAT files
// These are general UI icons that work for placeholder toolbar
const (
	ToolbarZoomIn      = 2379 // 22x22 zoom icon
	ToolbarZoomOut     = 2380 // 21x22 zoom icon
	ToolbarRotate      = 2365 // 20x19 rotate object
	ToolbarMap         = 2364 // 20x20 centre viewport
	ToolbarTracks      = 2335 // 14x20 construction straight
	ToolbarRoads       = 2344 // 20x20 construction curve
	ToolbarAirport     = 2368 // 20x20 icon
	ToolbarPorts       = 2369 // 18x19 icon
	ToolbarVehicles    = 2370 // 20x20 icon
	ToolbarStations    = 2371 // 20x20 icon
	ToolbarTowns       = 2372 // 20x20 icon
	ToolbarIndustries  = 2373 // 20x20 icon
	ToolbarLandscaping = 2367 // 20x20 paintbrush
	ToolbarTrees       = 2374 // 15x19 icon
	ToolbarWater       = 2375 // 15x16 icon
	ToolbarCompanies   = 2386 // 21x20 icon
	ToolbarFinances    = 2383 // 14x14 icon
	ToolbarNews        = 2384 // 14x13 icon
	ToolbarRemove      = 2361 // 27x19 rubbishbin
	ToolbarCamera      = 2366 // 17x12 camera
)

// Object tabs
const (
	TabObjectLand      = 3509
	TabObjectWater     = 3510
	TabObjectLandscape = 3511
)

// Terrain surfaces
const (
	SurfaceSmooth3Slope0 = 3746
	SurfaceSmooth3Slope1 = 3747
	SurfaceSmooth3Slope2 = 3748
	SurfaceSmooth3Slope3 = 3749
	SurfaceSmooth3Slope4 = 3750
	// ... slopes continue
	SurfaceSmooth1Slope0 = 3765
)

// Palette
const (
	DefaultPalette = 304
)

// Logo and title
const (
	ChrisRollercoasterGuy = 0  // First sprites are usually logo/intro
)
