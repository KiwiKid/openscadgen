package pkg

import "github.com/kiwikid/openscadgen/pkg/templates"

// OpenSCADToolRegistry returns the built-in list of OpenSCAD libraries and utilities
// that pair well with OpenSCADGen.
//
// Keep this hardcoded for now so the list can grow without introducing config
// plumbing before the shape of the catalog settles.
func OpenSCADToolRegistry() []templates.ToolLink {
	return []templates.ToolLink{
		{
			Title:                 "BOSL2",
			Slug:                  "bosl2",
			Description:           "Core helper library for many OpenSCAD models. OpenSCADGen can probe whether it is available locally.",
			URL:                   "/api/openscad/status",
			Badge:                 "Built in",
			Version:               "current",
			IsBOSL2InstallableLib: true,
			SupportsUpdateFlow:    true,
			CheckUpdatesURL:       "/api/openscad/libraries/bosl2/check",
			UpdateURL:             "/api/openscad/libraries/bosl2/update",
		},
		{
			Title:                 "JL_SCAD",
			Slug:                  "jl-scad",
			Description:           "Box enclosure library for OpenSCAD that builds on BOSL2 and is a strong fit for enclosure-style projects.",
			URL:                   "https://github.com/lijon/jl_scad",
			External:              true,
			Badge:                 "BOSL2-friendly",
			Version:               "untracked",
			IsBOSL2InstallableLib: true,
			SupportsUpdateFlow:    true,
			CheckUpdatesURL:       "/api/openscad/libraries/jl_scad/check",
			UpdateURL:             "/api/openscad/libraries/jl_scad/update",
		},
		{
			Title:                 "Attachable Text3D",
			Slug:                  "attachable-text3d",
			Description:           "Attachable 3D text modules that fit BOSL2-style workflows for labels, signs, and text-driven parts.",
			URL:                   "https://github.com/jon-gilbert/openscad_attachable_text3d",
			External:              true,
			Badge:                 "BOSL2-friendly",
			Version:               "untracked",
			IsBOSL2InstallableLib: true,
			SupportsUpdateFlow:    true,
			CheckUpdatesURL:       "/api/openscad/libraries/attachable_text3d/check",
			UpdateURL:             "/api/openscad/libraries/attachable_text3d/update",
		},
		{
			Title:                 "Pathbuilder",
			Slug:                  "pathbuilder",
			Description:           "SVG-like 2D path construction for OpenSCAD, useful when BOSL2 projects need more expressive path geometry.",
			URL:                   "https://github.com/dinther/pathbuilder",
			External:              true,
			Badge:                 "BOSL2-friendly",
			Version:               "untracked",
			IsBOSL2InstallableLib: true,
			SupportsUpdateFlow:    true,
			CheckUpdatesURL:       "/api/openscad/libraries/pathbuilder/check",
			UpdateURL:             "/api/openscad/libraries/pathbuilder/update",
		},
		{
			Title:       "Metaballs editor",
			Slug:        "metaballs-editor",
			Description: "A quick editor for shaping metaballs before you bring the geometry into OpenSCADGen.",
			URL:         "https://github.com/juliendorra/metaball-openscad-quick-editor",
			External:    true,
			Badge:       "External",
		},
	}
}
