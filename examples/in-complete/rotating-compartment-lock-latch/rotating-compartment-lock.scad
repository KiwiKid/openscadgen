include <BOSL2/std.scad>
include <BOSL2/constants.scad>
use <BOSL2/beziers.scad>

// Parameters
// Parameters
radius = 50;
height = 50;
thickness = 10;
turns = 1;  // Number of complete turns in the helix
baseDiameter = radius * 2;


path = [ [0, 0, 0], [33, 33, 33], [66, -33, -33], [100, 0, 0] ];
extrude_2d_shapes_along_bezier(path) difference(){
    circle(r=10);
    fwd(10/2) circle(r=8);
} 