include <BOSL2/std.scad>
include <BOSL2/shapes.scad>

// Parameters
radius = 50;
height = 100;
thickness = 2;
turns = 3;

// Create the helix shape
path = helix(l=height, r=radius, turns=turns);

// Generate the half-circle cross-section
cross_section = arc(r=thickness/2, angle=180);

// Create the helix using path_sweep
path_sweep(cross_section, path);
